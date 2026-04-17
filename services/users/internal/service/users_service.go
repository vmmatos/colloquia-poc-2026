package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"users/internal/db/sqlc"
	"users/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pgxerr "github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidLanguage   = errors.New("invalid language")
)

var allowedLanguages = map[string]bool{"en": true, "pt": true}

type UserResult struct {
	ID        uuid.UUID
	Email     string
	Name      string
	Avatar    string
	Bio       string
	Timezone  string
	Status    string
	Language  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UsersService struct {
	repo repository.IUsersRepository
}

func NewUsersService(repo repository.IUsersRepository) *UsersService {
	return &UsersService{repo: repo}
}

func (s *UsersService) CreateUser(ctx context.Context, id uuid.UUID, email string) (*UserResult, error) {
	profile, err := s.repo.CreateUserProfile(ctx, id, email)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create user profile: %w", err)
	}
	return toResult(profile), nil
}

func (s *UsersService) GetUser(ctx context.Context, id uuid.UUID) (*UserResult, error) {
	profile, err := s.repo.GetUserProfile(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user profile: %w", err)
	}
	return toResult(profile), nil
}

func (s *UsersService) BatchGetUsers(ctx context.Context, ids []uuid.UUID) ([]*UserResult, error) {
	profiles, err := s.repo.BatchGetUserProfiles(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("batch get user profiles: %w", err)
	}
	return mapSlice(profiles, func(p sqlc.UserProfile) *UserResult { return toResult(p) }), nil
}

// UpdateProfile applies partial updates: only non-nil fields override the current value.
func (s *UsersService) UpdateProfile(ctx context.Context, id uuid.UUID, name, avatar, bio, timezone, status, language *string) (*UserResult, error) {
	if language != nil && !allowedLanguages[*language] {
		return nil, ErrInvalidLanguage
	}

	current, err := s.repo.GetUserProfile(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	params := repository.UpdateParams{
		Name:     current.Name,
		Avatar:   current.Avatar,
		Bio:      current.Bio,
		Timezone: current.Timezone,
		Status:   current.Status,
		Language: current.Language,
	}
	if name != nil {
		params.Name = *name
	}
	if avatar != nil {
		params.Avatar = *avatar
	}
	if bio != nil {
		params.Bio = *bio
	}
	if timezone != nil {
		params.Timezone = *timezone
	}
	if status != nil {
		params.Status = *status
	}
	if language != nil {
		params.Language = *language
	}

	profile, err := s.repo.UpdateUserProfile(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("update user profile: %w", err)
	}
	return toResult(profile), nil
}

func (s *UsersService) ListUsers(ctx context.Context, limit, offset int32) ([]*UserResult, error) {
	profiles, err := s.repo.ListUsers(ctx, clampLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return mapSlice(profiles, func(p sqlc.UserProfile) *UserResult { return toResult(p) }), nil
}

func (s *UsersService) SearchUsers(ctx context.Context, query string, limit, offset int32) ([]*UserResult, error) {
	profiles, err := s.repo.SearchUsers(ctx, EscapeLikePattern(query), clampLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	return mapSlice(profiles, func(p sqlc.UserProfile) *UserResult { return toResult(p) }), nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func mapSlice[A, B any](in []A, fn func(A) B) []B {
	out := make([]B, len(in))
	for i, a := range in {
		out[i] = fn(a)
	}
	return out
}

func clampLimit(l int32) int32 {
	if l <= 0 {
		return 20
	}
	if l > 100 {
		return 100
	}
	return l
}

func toResult(p sqlc.UserProfile) *UserResult {
	return &UserResult{
		ID:        p.ID,
		Email:     p.Email,
		Name:      p.Name,
		Avatar:    p.Avatar,
		Bio:       p.Bio,
		Timezone:  p.Timezone,
		Status:    p.Status,
		Language:  p.Language,
		CreatedAt: p.CreatedAt.Time,
		UpdatedAt: p.UpdatedAt.Time,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgxerr.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
