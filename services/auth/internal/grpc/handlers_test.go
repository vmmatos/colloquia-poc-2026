package grpc

import (
	"auth/internal/pb"
	"auth/internal/service"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stubAuthService is a minimal auth service for handler testing.
type stubAuthService struct {
	registerResult    *service.AuthResult
	registerError     error
	loginResult       *service.AuthResult
	loginError        error
	logoutError       error
	refreshResult     *service.AuthResult
	refreshError      error
	validateResult    *service.ValidateResult
	validateError     error
}

func (s *stubAuthService) Register(context.Context, string, string) (*service.AuthResult, error) {
	return s.registerResult, s.registerError
}

func (s *stubAuthService) Login(context.Context, string, string) (*service.AuthResult, error) {
	return s.loginResult, s.loginError
}

func (s *stubAuthService) Logout(context.Context, string) error {
	return s.logoutError
}

func (s *stubAuthService) RefreshToken(context.Context, string) (*service.AuthResult, error) {
	return s.refreshResult, s.refreshError
}

func (s *stubAuthService) ValidateToken(context.Context, string) (*service.ValidateResult, error) {
	return s.validateResult, s.validateError
}

// TestToGRPCError verifies error mapping.
func TestToGRPCError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{"ErrEmailAlreadyExists", service.ErrEmailAlreadyExists, codes.AlreadyExists},
		{"ErrInvalidCredentials", service.ErrInvalidCredentials, codes.Unauthenticated},
		{"ErrAccountLocked", service.ErrAccountLocked, codes.PermissionDenied},
		{"ErrSessionNotFound", service.ErrSessionNotFound, codes.NotFound},
		{"ErrTokenExpired", service.ErrTokenExpired, codes.Unauthenticated},
		{"ErrTokenInvalid", service.ErrTokenInvalid, codes.Unauthenticated},
		{"unknown error", errors.New("unknown"), codes.Internal},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			grpcErr := toGRPCError(tc.err)
			s, _ := status.FromError(grpcErr)
			if s.Code() != tc.wantCode {
				t.Errorf("toGRPCError: got code %v; want %v", s.Code(), tc.wantCode)
			}
		})
	}
}

// TestRegisterHandlerEmptyEmail verifies empty email rejection.
func TestRegisterHandlerEmptyEmail(t *testing.T) {
	stub := &stubAuthService{}
	handler := &AuthHandler{authService: stub}

	req := &pb.RegisterRequest{Email: "", Password: "password123"}
	resp, err := handler.Register(context.Background(), req)

	s, ok := status.FromError(err)
	if !ok || s.Code() != codes.InvalidArgument {
		t.Errorf("Register empty email: expected InvalidArgument, got %v", err)
	}
	if resp != nil {
		t.Errorf("Register empty email: expected nil response")
	}
}

// TestRegisterHandlerEmptyPassword verifies empty password rejection.
func TestRegisterHandlerEmptyPassword(t *testing.T) {
	stub := &stubAuthService{}
	handler := NewAuthHandler(&service.AuthService{})
	handler.authService = stub

	req := &pb.RegisterRequest{Email: "test@example.com", Password: ""}
	resp, err := handler.Register(context.Background(), req)

	s, ok := status.FromError(err)
	if !ok || s.Code() != codes.InvalidArgument {
		t.Errorf("Register empty password: expected InvalidArgument, got %v", err)
	}
	if resp != nil {
		t.Errorf("Register empty password: expected nil response")
	}
}

// TestLoginHandlerEmptyEmail verifies empty email rejection.
func TestLoginHandlerEmptyEmail(t *testing.T) {
	handler := NewAuthHandler(&service.AuthService{})

	req := &pb.LoginRequest{Email: "", Password: "password123"}
	resp, err := handler.Login(context.Background(), req)

	s, ok := status.FromError(err)
	if !ok || s.Code() != codes.InvalidArgument {
		t.Errorf("Login empty email: expected InvalidArgument, got %v", err)
	}
	if resp != nil {
		t.Errorf("Login empty email: expected nil response")
	}
}

// TestLoginHandlerEmptyPassword verifies empty password rejection.
func TestLoginHandlerEmptyPassword(t *testing.T) {
	handler := NewAuthHandler(&service.AuthService{})

	req := &pb.LoginRequest{Email: "test@example.com", Password: ""}
	resp, err := handler.Login(context.Background(), req)

	s, ok := status.FromError(err)
	if !ok || s.Code() != codes.InvalidArgument {
		t.Errorf("Login empty password: expected InvalidArgument, got %v", err)
	}
	if resp != nil {
		t.Errorf("Login empty password: expected nil response")
	}
}

// TestValidateTokenHandlerValidToken verifies valid token response.
func TestValidateTokenHandlerValidToken(t *testing.T) {
	userID := uuid.New()
	stub := &stubAuthService{
		validateResult: &service.ValidateResult{
			UserID: userID,
			Email:  "test@example.com",
		},
	}
	handler := NewAuthHandler(&service.AuthService{})
	handler.authService = stub

	req := &pb.ValidateTokenRequest{AccessToken: "valid-token"}
	resp, err := handler.ValidateToken(context.Background(), req)

	if err != nil {
		t.Errorf("ValidateToken valid: got error %v; want nil", err)
	}
	if !resp.Valid {
		t.Errorf("ValidateToken valid: expected Valid=true")
	}
	if resp.UserId != userID.String() {
		t.Errorf("ValidateToken valid: got UserID %s; want %s", resp.UserId, userID.String())
	}
	if resp.Email != "test@example.com" {
		t.Errorf("ValidateToken valid: got email %q; want 'test@example.com'", resp.Email)
	}
}

// TestValidateTokenHandlerInvalidToken verifies error returns ValidFalse.
func TestValidateTokenHandlerInvalidToken(t *testing.T) {
	stub := &stubAuthService{
		validateError: service.ErrTokenInvalid,
	}
	handler := NewAuthHandler(&service.AuthService{})
	handler.authService = stub

	req := &pb.ValidateTokenRequest{AccessToken: "invalid-token"}
	resp, err := handler.ValidateToken(context.Background(), req)

	if err != nil {
		t.Errorf("ValidateToken invalid: got error %v; want nil", err)
	}
	if resp == nil {
		t.Errorf("ValidateToken invalid: got nil response")
		return
	}
	if resp.Valid {
		t.Errorf("ValidateToken invalid: expected Valid=false, got true")
	}
}

// TestRegisterHandlerSuccess verifies successful registration response.
func TestRegisterHandlerSuccess(t *testing.T) {
	userID := uuid.New()
	stub := &stubAuthService{
		registerResult: &service.AuthResult{
			UserID:       userID,
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresAt:    time.Now().Add(15 * time.Minute),
		},
	}
	handler := NewAuthHandler(&service.AuthService{})
	handler.authService = stub

	req := &pb.RegisterRequest{Email: "test@example.com", Password: "password123"}
	resp, err := handler.Register(context.Background(), req)

	if err != nil {
		t.Errorf("Register success: got error %v; want nil", err)
	}
	if resp == nil {
		t.Errorf("Register success: got nil response")
		return
	}
	if resp.UserId != userID.String() {
		t.Errorf("Register success: got UserID %s; want %s", resp.UserId, userID.String())
	}
	if resp.AccessToken != "access-token" {
		t.Errorf("Register success: access token mismatch")
	}
	if resp.RefreshToken != "refresh-token" {
		t.Errorf("Register success: refresh token mismatch")
	}
}

// TestLoginHandlerSuccess verifies successful login response.
func TestLoginHandlerSuccess(t *testing.T) {
	userID := uuid.New()
	stub := &stubAuthService{
		loginResult: &service.AuthResult{
			UserID:       userID,
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresAt:    time.Now().Add(15 * time.Minute),
		},
	}
	handler := NewAuthHandler(&service.AuthService{})
	handler.authService = stub

	req := &pb.LoginRequest{Email: "test@example.com", Password: "password123"}
	resp, err := handler.Login(context.Background(), req)

	if err != nil {
		t.Errorf("Login success: got error %v; want nil", err)
	}
	if resp == nil {
		t.Errorf("Login success: got nil response")
		return
	}
	if resp.UserId != userID.String() {
		t.Errorf("Login success: got UserID %s; want %s", resp.UserId, userID.String())
	}
}

// TestLogoutHandlerSuccess verifies successful logout.
func TestLogoutHandlerSuccess(t *testing.T) {
	stub := &stubAuthService{logoutError: nil}
	handler := NewAuthHandler(&service.AuthService{})
	handler.authService = stub

	req := &pb.LogoutRequest{AccessToken: "valid-token"}
	resp, err := handler.Logout(context.Background(), req)

	if err != nil {
		t.Errorf("Logout success: got error %v; want nil", err)
	}
	if resp == nil {
		t.Errorf("Logout success: got nil response")
	}
}

// TestRefreshTokenHandlerSuccess verifies successful token refresh response.
func TestRefreshTokenHandlerSuccess(t *testing.T) {
	userID := uuid.New()
	stub := &stubAuthService{
		refreshResult: &service.AuthResult{
			UserID:       userID,
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		},
	}
	handler := NewAuthHandler(&service.AuthService{})
	handler.authService = stub

	req := &pb.RefreshTokenRequest{RefreshToken: "old-refresh-token"}
	resp, err := handler.RefreshToken(context.Background(), req)

	if err != nil {
		t.Errorf("RefreshToken success: got error %v; want nil", err)
	}
	if resp == nil {
		t.Errorf("RefreshToken success: got nil response")
		return
	}
	if resp.AccessToken != "new-access-token" {
		t.Errorf("RefreshToken success: access token mismatch")
	}
}
