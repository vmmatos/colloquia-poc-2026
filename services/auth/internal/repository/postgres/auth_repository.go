package postgres

import (
	"auth/internal/db/sqlc"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

// ── Users ─────────────────────────────────────────────────────────────────────

func (r *AuthRepository) CreateUser(ctx context.Context, id uuid.UUID, email string, passwordHash string) (sqlc.User, error) {
	return r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
	})
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.queries.FindUserByEmail(ctx, email)
}

func (r *AuthRepository) FindUserById(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return r.queries.FindUserById(ctx, id)
}

func (r *AuthRepository) IncrementFailedLoginAttempts(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return r.queries.IncrementFailedLoginAttempts(ctx, id)
}

func (r *AuthRepository) LockUser(ctx context.Context, id uuid.UUID, until time.Time) error {
	return r.queries.LockUser(ctx, sqlc.LockUserParams{
		ID:          id,
		LockedUntil: pgtype.Timestamp{Time: until, Valid: true},
	})
}

func (r *AuthRepository) ResetFailedLoginAttempts(ctx context.Context, id uuid.UUID) error {
	return r.queries.ResetFailedLoginAttempts(ctx, id)
}

// ── Sessions ──────────────────────────────────────────────────────────────────

func (r *AuthRepository) CreateSession(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
	refreshTokenHash string,
	accessTokenHash string,
	expiresAt time.Time,
) (sqlc.Session, error) {
	return r.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:               id,
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		AccessTokenHash:  accessTokenHash,
		ExpiresAt:        pgtype.Timestamp{Time: expiresAt, Valid: true},
		Revoked:          pgtype.Bool{Bool: false, Valid: true},
	})
}

func (r *AuthRepository) FindSessionByAccessTokenHash(ctx context.Context, accessTokenHash string) (sqlc.Session, error) {
	return r.queries.FindSessionByAccessTokenHash(ctx, accessTokenHash)
}

func (r *AuthRepository) FindSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (sqlc.Session, error) {
	return r.queries.FindSessionByRefreshTokenHash(ctx, refreshTokenHash)
}

func (r *AuthRepository) FindSessionById(ctx context.Context, id uuid.UUID) (sqlc.Session, error) {
	return r.queries.FindSessionById(ctx, id)
}

func (r *AuthRepository) RevokeSession(ctx context.Context, id uuid.UUID) error {
	return r.queries.RevokeSession(ctx, id)
}

func (r *AuthRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return r.queries.RevokeAllUserSessions(ctx, userID)
}
