package grpc

import (
	"auth/internal/pb"
	"auth/internal/service"
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServicer defines the auth service interface for handlers.
type AuthServicer interface {
	Register(ctx context.Context, email, password string) (*service.AuthResult, error)
	Login(ctx context.Context, email, password string) (*service.AuthResult, error)
	Logout(ctx context.Context, accessToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*service.AuthResult, error)
	ValidateToken(ctx context.Context, accessToken string) (*service.ValidateResult, error)
}

// AuthHandler implements pb.AuthServiceServer.
type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService AuthServicer
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	result, err := h.authService.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toAuthPbResponse(result), nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	result, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toAuthPbResponse(result), nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	if err := h.authService.Logout(ctx, req.AccessToken); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.LogoutResponse{}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.AuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	result, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toAuthPbResponse(result), nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	result, err := h.authService.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: result.UserID.String(),
		Email:  result.Email,
	}, nil
}

func toAuthPbResponse(r *service.AuthResult) *pb.AuthResponse {
	return &pb.AuthResponse{
		UserId:       r.UserID.String(),
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		ExpiresAt:    r.ExpiresAt.Unix(),
	}
}

// toGRPCError maps domain errors to gRPC status codes.
func toGRPCError(err error) error {
	switch {
	case errors.Is(err, service.ErrEmailAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, service.ErrAccountLocked):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrSessionNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrTokenExpired):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, service.ErrTokenInvalid):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
