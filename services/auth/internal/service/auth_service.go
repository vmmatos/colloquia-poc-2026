package service

import (
	"auth/internal/config"
	"auth/internal/repository"
)

type AuthService struct {
	authRepo repository.IAuthRepository
	cfg *config.Config
}

func NewAuthService(authRepo repository.IAuthRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		cfg: cfg,
	}
}
