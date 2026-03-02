package repository

import (
	"auth/internal/db/sqlc"
	"context"

	"github.com/google/uuid"
)

type IAuthRepository interface {
    CreateUser(ctx context.Context, id uuid.UUID, email string, passwordHash string) (sqlc.User, error)
    FindUserByEmail(ctx context.Context, email string) (sqlc.User, error)
    FindUserById(ctx context.Context, id uuid.UUID) (sqlc.User, error)
    CreateSession(ctx context.Context, id uuid.UUID, userId uuid.UUID, refreshTokenHash string, accessTokenHash string) (sqlc.Session, error)
    FindSessionByAccessTokenHash(ctx context.Context, accessTokenHash string) (sqlc.Session, error)
    FindSessionById(ctx context.Context, id uuid.UUID) (sqlc.Session, error)
    RevokeSession(ctx context.Context, id uuid.UUID) error
}

