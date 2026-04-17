package service

import (
	"auth/internal/config"
	"auth/internal/db/sqlc"
	"auth/internal/repository"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// testKeys generates a test RSA key pair and returns PEM-encoded bytes.
func testKeys(t *testing.T) ([]byte, []byte) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return privPEM, pubPEM
}

// stubAuthRepo is a minimal IAuthRepository for testing AuthService.
type stubAuthRepo struct {
	repository.IAuthRepository
	users              map[uuid.UUID]*sqlc.User
	sessions           map[uuid.UUID]*sqlc.Session
	lastLockedUserID   uuid.UUID
	lastLockedUntilVal time.Time
}

func (s *stubAuthRepo) CreateUser(_ context.Context, id uuid.UUID, email, passwordHash string) (sqlc.User, error) {
	if _, exists := s.users[id]; exists {
		return sqlc.User{}, errors.New("unique violation")
	}
	for _, u := range s.users {
		if u.Email == email {
			return sqlc.User{}, errors.New("unique violation email")
		}
	}
	user := sqlc.User{ID: id, Email: email, PasswordHash: passwordHash}
	s.users[id] = &user
	return user, nil
}

func (s *stubAuthRepo) FindUserByEmail(_ context.Context, email string) (sqlc.User, error) {
	for _, u := range s.users {
		if u.Email == email {
			return *u, nil
		}
	}
	return sqlc.User{}, errors.New("not found")
}

func (s *stubAuthRepo) FindUserById(_ context.Context, id uuid.UUID) (sqlc.User, error) {
	if u, ok := s.users[id]; ok {
		return *u, nil
	}
	return sqlc.User{}, errors.New("not found")
}

func (s *stubAuthRepo) IncrementFailedLoginAttempts(_ context.Context, id uuid.UUID) (sqlc.User, error) {
	u, ok := s.users[id]
	if !ok {
		return sqlc.User{}, errors.New("not found")
	}
	if !u.FailedLoginAttempts.Valid {
		u.FailedLoginAttempts = pgtype.Int4{Int32: 1, Valid: true}
	} else {
		u.FailedLoginAttempts.Int32++
	}
	s.users[id] = u
	return *u, nil
}

func (s *stubAuthRepo) LockUser(_ context.Context, id uuid.UUID, until time.Time) error {
	s.lastLockedUserID = id
	s.lastLockedUntilVal = until
	u, ok := s.users[id]
	if !ok {
		return errors.New("not found")
	}
	u.LockedUntil = pgtype.Timestamp{Time: until, Valid: true}
	s.users[id] = u
	return nil
}

func (s *stubAuthRepo) ResetFailedLoginAttempts(_ context.Context, id uuid.UUID) error {
	u, ok := s.users[id]
	if !ok {
		return errors.New("not found")
	}
	u.FailedLoginAttempts = pgtype.Int4{Int32: 0, Valid: true}
	s.users[id] = u
	return nil
}

func (s *stubAuthRepo) CreateSession(_ context.Context, id, userID uuid.UUID, refreshTokenHash, accessTokenHash string, expiresAt time.Time) (sqlc.Session, error) {
	session := sqlc.Session{
		ID:                id,
		UserID:            userID,
		RefreshTokenHash:  refreshTokenHash,
		AccessTokenHash:   accessTokenHash,
		ExpiresAt:         pgtype.Timestamp{Time: expiresAt, Valid: true},
		Revoked:           pgtype.Bool{Bool: false, Valid: true},
	}
	s.sessions[id] = &session
	return session, nil
}

func (s *stubAuthRepo) FindSessionByAccessTokenHash(_ context.Context, hash string) (sqlc.Session, error) {
	for _, sess := range s.sessions {
		if sess.AccessTokenHash == hash && !sess.Revoked.Bool {
			return *sess, nil
		}
	}
	return sqlc.Session{}, errors.New("not found")
}

func (s *stubAuthRepo) FindSessionByRefreshTokenHash(_ context.Context, hash string) (sqlc.Session, error) {
	for _, sess := range s.sessions {
		if sess.RefreshTokenHash == hash && !sess.Revoked.Bool {
			return *sess, nil
		}
	}
	return sqlc.Session{}, errors.New("not found")
}

func (s *stubAuthRepo) RevokeSession(_ context.Context, id uuid.UUID) error {
	sess, ok := s.sessions[id]
	if !ok {
		return errors.New("not found")
	}
	sess.Revoked = pgtype.Bool{Bool: true, Valid: true}
	s.sessions[id] = sess
	return nil
}

// TestRegisterHappyPath verifies successful registration.
func TestRegisterHappyPath(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	result, err := svc.Register(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Errorf("Register: got error %v; want nil", err)
	}
	if result == nil || result.AccessToken == "" || result.RefreshToken == "" {
		t.Errorf("Register: got empty tokens; want non-empty")
	}
}

// TestRegisterDuplicateEmail verifies duplicate email rejection.
func TestRegisterDuplicateEmail(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register first time
	_, _ = svc.Register(context.Background(), "test@example.com", "password123")

	// Register again with same email
	result, err := svc.Register(context.Background(), "test@example.com", "password456")
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("Register duplicate: got error %v; want ErrEmailAlreadyExists", err)
	}
	if result != nil {
		t.Errorf("Register duplicate: got result; want nil")
	}
}

// TestLoginHappyPath verifies successful login.
func TestLoginHappyPath(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register user first
	regResult, _ := svc.Register(context.Background(), "test@example.com", "password123")

	// Login with correct password
	result, err := svc.Login(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Errorf("Login: got error %v; want nil", err)
	}
	if result == nil || result.AccessToken == "" {
		t.Errorf("Login: got empty token; want non-empty")
	}
	if result.UserID != regResult.UserID {
		t.Errorf("Login: got UserID %v; want %v", result.UserID, regResult.UserID)
	}
}

// TestLoginWrongPassword verifies failed login on wrong password.
func TestLoginWrongPassword(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register user
	_, _ = svc.Register(context.Background(), "test@example.com", "password123")

	// Login with wrong password
	result, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login wrong password: got error %v; want ErrInvalidCredentials", err)
	}
	if result != nil {
		t.Errorf("Login wrong password: got result; want nil")
	}
}

// TestLoginNonexistentUser verifies no email leak on missing user.
func TestLoginNonexistentUser(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	result, err := svc.Login(context.Background(), "nonexistent@example.com", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login nonexistent: got error %v; want ErrInvalidCredentials", err)
	}
	if result != nil {
		t.Errorf("Login nonexistent: got result; want nil")
	}
}

// TestLoginLockedAccount verifies rejection of locked accounts.
func TestLoginLockedAccount(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register and manually lock
	regResult, _ := svc.Register(context.Background(), "test@example.com", "password123")

	// Lock the user until the future
	futureLock := time.Now().Add(1 * time.Hour)
	stub.LockUser(context.Background(), regResult.UserID, futureLock)

	// Attempt login on locked account
	result, err := svc.Login(context.Background(), "test@example.com", "password123")
	if !errors.Is(err, ErrAccountLocked) {
		t.Errorf("Login locked: got error %v; want ErrAccountLocked", err)
	}
	if result != nil {
		t.Errorf("Login locked: got result; want nil")
	}
}

// TestLoginFailedAttemptLockout verifies account lockout after 5 failed attempts.
func TestLoginFailedAttemptLockout(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register
	regResult, _ := svc.Register(context.Background(), "test@example.com", "password123")

	// Attempt 5 failed logins
	for i := 0; i < 5; i++ {
		_, _ = svc.Login(context.Background(), "test@example.com", "wrongpassword")
	}

	// Verify user was locked
	if stub.lastLockedUserID != regResult.UserID {
		t.Errorf("Lockout: user was not locked")
	}

	// Verify lockout prevents login
	result, err := svc.Login(context.Background(), "test@example.com", "password123")
	if !errors.Is(err, ErrAccountLocked) {
		t.Errorf("Login after lockout: got error %v; want ErrAccountLocked", err)
	}
	if result != nil {
		t.Errorf("Login after lockout: got result; want nil")
	}
}

// TestLogoutSuccess verifies session revocation on logout.
func TestLogoutSuccess(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register and get token
	_, _ = svc.Register(context.Background(), "test@example.com", "password123")

	// Login to get access token
	loginResult, _ := svc.Login(context.Background(), "test@example.com", "password123")

	// Logout
	err = svc.Logout(context.Background(), loginResult.AccessToken)
	if err != nil {
		t.Errorf("Logout: got error %v; want nil", err)
	}

	// Verify token is revoked (no longer found)
	validateResult, err := svc.ValidateToken(context.Background(), loginResult.AccessToken)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("ValidateToken after logout: got error %v; want ErrSessionNotFound", err)
	}
	if validateResult != nil {
		t.Errorf("ValidateToken after logout: got result; want nil")
	}
}

// TestRefreshTokenSuccess verifies token rotation on refresh.
func TestRefreshTokenSuccess(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register and login
	_, _ = svc.Register(context.Background(), "test@example.com", "password123")
	loginResult, _ := svc.Login(context.Background(), "test@example.com", "password123")

	oldAccessToken := loginResult.AccessToken

	// Refresh token
	refreshResult, err := svc.RefreshToken(context.Background(), loginResult.RefreshToken)
	if err != nil {
		t.Errorf("RefreshToken: got error %v; want nil", err)
	}
	if refreshResult == nil || refreshResult.AccessToken == "" {
		t.Errorf("RefreshToken: got empty token; want non-empty")
	}

	// Old token should be revoked
	_, err = svc.ValidateToken(context.Background(), oldAccessToken)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("ValidateToken after refresh: old token still valid")
	}

	// New token should be valid
	_, err = svc.ValidateToken(context.Background(), refreshResult.AccessToken)
	if err != nil {
		t.Errorf("ValidateToken new: got error %v; want nil", err)
	}
}

// TestRefreshTokenExpired verifies rejection of expired refresh tokens.
func TestRefreshTokenExpired(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register and login
	_, _ = svc.Register(context.Background(), "test@example.com", "password123")
	loginResult, _ := svc.Login(context.Background(), "test@example.com", "password123")

	// Manually expire the session
	for _, sess := range stub.sessions {
		sess.ExpiresAt = pgtype.Timestamp{Time: time.Now().Add(-1 * time.Hour), Valid: true}
	}

	// Try to refresh
	result, err := svc.RefreshToken(context.Background(), loginResult.RefreshToken)
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("RefreshToken expired: got error %v; want ErrTokenExpired", err)
	}
	if result != nil {
		t.Errorf("RefreshToken expired: got result; want nil")
	}
}

// TestValidateTokenSuccess verifies valid token acceptance.
func TestValidateTokenSuccess(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register and login
	regResult, _ := svc.Register(context.Background(), "test@example.com", "password123")
	loginResult, _ := svc.Login(context.Background(), "test@example.com", "password123")

	// Validate
	validateResult, err := svc.ValidateToken(context.Background(), loginResult.AccessToken)
	if err != nil {
		t.Errorf("ValidateToken: got error %v; want nil", err)
	}
	if validateResult == nil || validateResult.UserID != regResult.UserID {
		t.Errorf("ValidateToken: got wrong UserID")
	}
	if validateResult.Email != "test@example.com" {
		t.Errorf("ValidateToken: got email %q; want 'test@example.com'", validateResult.Email)
	}
}

// TestValidateTokenRevokedSession verifies rejection of revoked tokens.
func TestValidateTokenRevokedSession(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register and login
	_, _ = svc.Register(context.Background(), "test@example.com", "password123")
	loginResult, _ := svc.Login(context.Background(), "test@example.com", "password123")

	// Logout to revoke
	_ = svc.Logout(context.Background(), loginResult.AccessToken)

	// Validate revoked token
	result, err := svc.ValidateToken(context.Background(), loginResult.AccessToken)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("ValidateToken revoked: got error %v; want ErrSessionNotFound", err)
	}
	if result != nil {
		t.Errorf("ValidateToken revoked: got result; want nil")
	}
}

// TestHashToken verifies deterministic SHA-256 hashing.
func TestHashToken(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"token1"},
		{"token2"},
		{""},
		{"very long token string with special chars !@#$%"},
	}

	for _, tc := range tests {
		hash1 := hashToken(tc.input)
		hash2 := hashToken(tc.input)
		if hash1 != hash2 {
			t.Errorf("hashToken(%q): not deterministic", tc.input)
		}
	}

	hash1 := hashToken("token1")
	hash2 := hashToken("token2")
	if hash1 == hash2 {
		t.Errorf("hashToken: different inputs produced same hash")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ERROR & EDGE CASE TESTS

// TestLogoutInvalidToken verifies session not found on invalid token.
func TestLogoutInvalidToken(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Logout with non-existent token
	err = svc.Logout(context.Background(), "invalid-token-hash")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("Logout invalid token: got error %v; want ErrSessionNotFound", err)
	}
}

// TestRefreshTokenNotFound verifies error when session doesn't exist.
func TestRefreshTokenNotFound(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	result, err := svc.RefreshToken(context.Background(), "nonexistent-refresh-token")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("RefreshToken not found: got error %v; want ErrSessionNotFound", err)
	}
	if result != nil {
		t.Errorf("RefreshToken not found: got result; want nil")
	}
}

// TestValidateTokenNotInDB verifies rejection when session not in DB.
func TestValidateTokenNotInDB(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Create a valid token but don't store session in DB
	userID := uuid.New()
	sessionID := uuid.New()
	// Create user for validation to work
	stub.CreateUser(context.Background(), userID, "test@example.com", "hash")

	validToken, _ := svc.generateAccessToken(userID, sessionID, "test@example.com")

	// Validate token that's not in DB (token is valid cryptographically but session revoked)
	result, err := svc.ValidateToken(context.Background(), validToken)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("ValidateToken not in DB: got error %v; want ErrSessionNotFound", err)
	}
	if result != nil {
		t.Errorf("ValidateToken not in DB: got result; want nil")
	}
}

// TestLoginAfter4FailedAttempts verifies counter increments without lockout.
func TestLoginAfter4FailedAttempts(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, err := NewAuthService(stub, cfg, nil)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Register
	_, _ = svc.Register(context.Background(), "test@example.com", "password123")

	// Attempt 4 failed logins
	for i := 0; i < 4; i++ {
		_, _ = svc.Login(context.Background(), "test@example.com", "wrongpassword")
	}

	// 5th attempt with correct password should still work (lockout only after 5 failures)
	result, err := svc.Login(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Errorf("Login after 4 failures: got error %v; want nil", err)
	}
	if result == nil {
		t.Errorf("Login after 4 failures: got nil; want token")
	}
}

// TestValidateTokenMalformedJWT verifies rejection of malformed tokens.
func TestValidateTokenMalformedJWT(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, _ := NewAuthService(stub, cfg, nil)

	// Validate malformed token
	result, err := svc.ValidateToken(context.Background(), "not-a-jwt-token")
	if !errors.Is(err, ErrTokenInvalid) {
		t.Errorf("ValidateToken malformed: got error %v; want ErrTokenInvalid", err)
	}
	if result != nil {
		t.Errorf("ValidateToken malformed: got result; want nil")
	}
}

// TestValidateTokenEmptyString verifies rejection of empty token.
func TestValidateTokenEmptyString(t *testing.T) {
	privPEM, pubPEM := testKeys(t)
	cfg := &config.Config{JwtPrivateKey: privPEM, JwtPublicKey: pubPEM}

	stub := &stubAuthRepo{users: make(map[uuid.UUID]*sqlc.User), sessions: make(map[uuid.UUID]*sqlc.Session)}
	svc, _ := NewAuthService(stub, cfg, nil)

	result, err := svc.ValidateToken(context.Background(), "")
	if !errors.Is(err, ErrTokenInvalid) {
		t.Errorf("ValidateToken empty: got error %v; want ErrTokenInvalid", err)
	}
	if result != nil {
		t.Errorf("ValidateToken empty: got result; want nil")
	}
}
