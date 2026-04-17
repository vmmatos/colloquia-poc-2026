package service

import (
	"auth/internal/config"
	"auth/internal/repository"
	"auth/internal/usersclient"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ── Domain errors ─────────────────────────────────────────────────────────────

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account locked")
	ErrSessionNotFound    = errors.New("session not found")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("invalid token")
)

// ── Constants ─────────────────────────────────────────────────────────────────

const (
	maxFailedAttempts    = 5
	lockoutDuration      = 15 * time.Minute
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
	bcryptCost           = 12
)

// ── Result types (domain, not pb) ─────────────────────────────────────────────

type AuthResult struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type ValidateResult struct {
	UserID uuid.UUID
	Email  string
}

// ── Service ───────────────────────────────────────────────────────────────────

type AuthService struct {
	repo       repository.IAuthRepository
	cfg        *config.Config
	userCreator usersclient.UserCreator // nil-safe: skip if nil
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewAuthService(repo repository.IAuthRepository, cfg *config.Config, userCreator usersclient.UserCreator) (*AuthService, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(cfg.JwtPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(cfg.JwtPublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &AuthService{
		repo:        repo,
		cfg:         cfg,
		userCreator: userCreator,
		privateKey:  privateKey,
		publicKey:   publicKey,
	}, nil
}

// Register creates a new user and returns a session with tokens.
func (s *AuthService) Register(ctx context.Context, email, password string) (*AuthResult, error) {
	// Check if email is already taken.
	_, err := s.repo.FindUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	userID := uuid.New()
	if _, err = s.repo.CreateUser(ctx, userID, email, string(passwordHash)); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Best-effort: create profile in users service. If it fails, registration still succeeds.
	if s.userCreator != nil {
		if err := s.userCreator.CreateUser(ctx, userID.String(), email); err != nil {
			log.Printf("warn: create user profile: %v", err)
		}
	}

	return s.issueSession(ctx, userID, email)
}

// Login validates credentials and returns a session with tokens.
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		// Do not leak whether the email exists.
		return nil, ErrInvalidCredentials
	}

	// Account lockout check.
	if user.LockedUntil.Valid && user.LockedUntil.Time.After(time.Now()) {
		return nil, ErrAccountLocked
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if handleErr := s.handleFailedLogin(ctx, user.ID); handleErr != nil {
			return nil, handleErr
		}
		return nil, ErrInvalidCredentials
	}

	// Successful login: reset failure counter.
	if err = s.repo.ResetFailedLoginAttempts(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("reset failed attempts: %w", err)
	}

	return s.issueSession(ctx, user.ID, user.Email)
}

// Logout revokes the session associated with the given access token.
func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	hash := hashToken(accessToken)
	session, err := s.repo.FindSessionByAccessTokenHash(ctx, hash)
	if err != nil {
		return ErrSessionNotFound
	}
	return s.repo.RevokeSession(ctx, session.ID)
}

// RefreshToken rotates a refresh token: revokes the old session and issues a new one.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	hash := hashToken(refreshToken)
	session, err := s.repo.FindSessionByRefreshTokenHash(ctx, hash)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if session.ExpiresAt.Valid && session.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	user, err := s.repo.FindUserById(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	// Revoke old session before issuing a new one (rotation).
	if err = s.repo.RevokeSession(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("revoke session: %w", err)
	}

	return s.issueSession(ctx, user.ID, user.Email)
}

// ValidateToken verifies the JWT signature and confirms the session exists in DB.
func (s *AuthService) ValidateToken(ctx context.Context, accessToken string) (*ValidateResult, error) {
	claims, err := s.parseAccessToken(accessToken)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	// Double-check against DB to support revocation.
	hash := hashToken(accessToken)
	_, err = s.repo.FindSessionByAccessTokenHash(ctx, hash)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	user, err := s.repo.FindUserById(ctx, userID)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	return &ValidateResult{UserID: userID, Email: user.Email}, nil
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func (s *AuthService) issueSession(ctx context.Context, userID uuid.UUID, email string) (*AuthResult, error) {
	sessionID := uuid.New()
	expiresAt := time.Now().Add(refreshTokenDuration)

	accessToken, err := s.generateAccessToken(userID, sessionID, email)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	_, err = s.repo.CreateSession(ctx,
		sessionID,
		userID,
		hashToken(refreshToken),
		hashToken(accessToken),
		expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &AuthResult{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) handleFailedLogin(ctx context.Context, userID uuid.UUID) error {
	updated, err := s.repo.IncrementFailedLoginAttempts(ctx, userID)
	if err != nil {
		return fmt.Errorf("increment failed attempts: %w", err)
	}

	attempts := int(updated.FailedLoginAttempts.Int32)
	if attempts >= maxFailedAttempts {
		until := time.Now().Add(lockoutDuration)
		if err = s.repo.LockUser(ctx, userID, until); err != nil {
			return fmt.Errorf("lock user: %w", err)
		}
	}
	return nil
}

type accessTokenClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// generateAccessToken creates a signed RS256 JWT.
func (s *AuthService) generateAccessToken(userID uuid.UUID, sessionID uuid.UUID, email string) (string, error) {
	now := time.Now()
	claims := accessTokenClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ID:        sessionID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "colloquia-auth-key-1"
	return token.SignedString(s.privateKey)
}

func (s *AuthService) parseAccessToken(tokenStr string) (*accessTokenClaims, error) {
	claims := &accessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

// generateOpaqueToken returns a cryptographically random URL-safe token.
func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// hashToken returns the SHA-256 hex digest of the token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
