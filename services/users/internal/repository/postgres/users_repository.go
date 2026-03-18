package postgres

import (
	"context"
	"users/internal/db/sqlc"
	"users/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository struct {
	queries *sqlc.Queries
}

func NewUsersRepository(db *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{queries: sqlc.New(db)}
}

func (r *UsersRepository) CreateUserProfile(ctx context.Context, id uuid.UUID, email string) (sqlc.UserProfile, error) {
	return r.queries.CreateUserProfile(ctx, sqlc.CreateUserProfileParams{ID: id, Email: email})
}

func (r *UsersRepository) GetUserProfile(ctx context.Context, id uuid.UUID) (sqlc.UserProfile, error) {
	return r.queries.GetUserProfile(ctx, id)
}

func (r *UsersRepository) BatchGetUserProfiles(ctx context.Context, ids []uuid.UUID) ([]sqlc.UserProfile, error) {
	return r.queries.BatchGetUserProfiles(ctx, ids)
}

func (r *UsersRepository) UpdateUserProfile(ctx context.Context, id uuid.UUID, params repository.UpdateParams) (sqlc.UserProfile, error) {
	return r.queries.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		ID:       id,
		Name:     params.Name,
		Avatar:   params.Avatar,
		Bio:      params.Bio,
		Timezone: params.Timezone,
		Status:   params.Status,
	})
}

func (r *UsersRepository) ListUsers(ctx context.Context, limit, offset int32) ([]sqlc.UserProfile, error) {
	return r.queries.ListUsers(ctx, sqlc.ListUsersParams{Limit: limit, Offset: offset})
}

func (r *UsersRepository) SearchUsers(ctx context.Context, query string, limit, offset int32) ([]sqlc.UserProfile, error) {
	return r.queries.SearchUsers(ctx, sqlc.SearchUsersParams{
		Column1: pgtype.Text{String: query, Valid: true},
		Limit:   limit,
		Offset:  offset,
	})
}
