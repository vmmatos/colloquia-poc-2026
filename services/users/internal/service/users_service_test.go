package service

import (
	"context"
	"testing"
	"users/internal/db/sqlc"
	"users/internal/repository"

	"github.com/google/uuid"
)

// stubRepo is a minimal IUsersRepository that records the last search query.
type stubRepo struct {
	lastSearchQuery string
	repository.IUsersRepository // embed for unimplemented methods
}

func (s *stubRepo) SearchUsers(_ context.Context, query string, _, _ int32) ([]sqlc.UserProfile, error) {
	s.lastSearchQuery = query
	return nil, nil
}

func (s *stubRepo) GetUserProfile(_ context.Context, _ uuid.UUID) (sqlc.UserProfile, error) {
	return sqlc.UserProfile{}, nil
}

func (s *stubRepo) UpdateUserProfile(_ context.Context, _ uuid.UUID, _ repository.UpdateParams) (sqlc.UserProfile, error) {
	return sqlc.UserProfile{}, nil
}

func (s *stubRepo) TouchLastSeen(_ context.Context, _ uuid.UUID) error { return nil }

func TestSearchUsersEscapesLikeMetachars(t *testing.T) {
	stub := &stubRepo{}
	svc := &UsersService{repo: stub}

	tests := []struct {
		input string
		want  string
	}{
		{"%", `\%`},
		{"_", `\_`},
		{`\`, `\\`},
		{"%admin%", `\%admin\%`},
		{"foo_bar", `foo\_bar`},
		{"normal", "normal"},
	}

	for _, tc := range tests {
		_, _ = svc.SearchUsers(context.Background(), tc.input, 10, 0)
		if stub.lastSearchQuery != tc.want {
			t.Errorf("SearchUsers(%q): repo received %q; want %q", tc.input, stub.lastSearchQuery, tc.want)
		}
	}
}
