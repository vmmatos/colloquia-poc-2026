package postgres

import (
	"auth/internal/db/sqlc"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	queries *sqlc.Queries
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		queries: sqlc.New(db),
	}
}

func (ar *AuthRepository) CreateUser(
	ctx context.Context,
	id uuid.UUID,
	email string,
	passwordHash string,
) (sqlc.User, error) {
	return ar.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
	})
}

func (ar *AuthRepository) FindUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return ar.queries.FindUserByEmail(ctx, email)
}

func (ar *AuthRepository) FindUserById(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return ar.queries.FindUserById(ctx, id)
}

func (ar *AuthRepository) CreateSession(
	ctx context.Context, 
	id uuid.UUID, 
	userId uuid.UUID, 
	refreshTokenHash string, 
	accessTokenHash string, 
) (sqlc.Session, error) {
	return ar.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:               id,
		UserID:           userId,
		RefreshTokenHash: refreshTokenHash,
		AccessTokenHash:  accessTokenHash,
	})
}

func (ar *AuthRepository) FindSessionByAccessTokenHash(ctx context.Context, refreshTokenHash string) (sqlc.Session, error) {
	return ar.queries.FindSessionByAccessTokenHash(ctx, refreshTokenHash)
}

func (ar *AuthRepository) FindSessionById(ctx context.Context, id uuid.UUID) (sqlc.Session, error) {
	return ar.queries.FindSessionById(ctx, id)
}

func (ar *AuthRepository) RevokeSession(ctx context.Context, id uuid.UUID) error {
	return ar.queries.RevokeSession(ctx, id)
}