package repository

import (
	"auth/internal/db/sqlc"
	"context"
	"time"

	"github.com/google/uuid"
)

type IAuthRepository interface {
	// Users
	CreateUser(ctx context.Context, id uuid.UUID, email string, passwordHash string) (sqlc.User, error)
	FindUserByEmail(ctx context.Context, email string) (sqlc.User, error)
	FindUserById(ctx context.Context, id uuid.UUID) (sqlc.User, error)
	IncrementFailedLoginAttempts(ctx context.Context, id uuid.UUID) (sqlc.User, error)
	LockUser(ctx context.Context, id uuid.UUID, until time.Time) error
	ResetFailedLoginAttempts(ctx context.Context, id uuid.UUID) error

	// Sessions
	CreateSession(ctx context.Context, id uuid.UUID, userID uuid.UUID, refreshTokenHash string, accessTokenHash string, expiresAt time.Time) (sqlc.Session, error)
	FindSessionByAccessTokenHash(ctx context.Context, accessTokenHash string) (sqlc.Session, error)
	FindSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (sqlc.Session, error)
	FindSessionById(ctx context.Context, id uuid.UUID) (sqlc.Session, error)
	RevokeSession(ctx context.Context, id uuid.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
}
