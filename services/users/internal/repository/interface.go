package repository

import (
	"context"
	"users/internal/db/sqlc"

	"github.com/google/uuid"
)

// UpdateParams holds all fields for a profile update (caller merges current values).
type UpdateParams struct {
	Name     string
	Avatar   string
	Bio      string
	Timezone string
	Status   string
}

type IUsersRepository interface {
	CreateUserProfile(ctx context.Context, id uuid.UUID, email string) (sqlc.UserProfile, error)
	GetUserProfile(ctx context.Context, id uuid.UUID) (sqlc.UserProfile, error)
	BatchGetUserProfiles(ctx context.Context, ids []uuid.UUID) ([]sqlc.UserProfile, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, params UpdateParams) (sqlc.UserProfile, error)
}
