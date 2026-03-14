package service

import (
	"channels/internal/repository"
	"channels/internal/usersclient"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pgxerr "github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrChannelNotFound      = errors.New("channel not found")
	ErrChannelAlreadyExists = errors.New("channel already exists")
	ErrMemberNotFound       = errors.New("member not found")
	ErrMemberAlreadyExists  = errors.New("member already exists")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrChannelArchived      = errors.New("channel is archived")
	ErrUserNotFound         = errors.New("user not found")
)

type ChannelsService struct {
	repo    repository.IChannelsRepository
	userVal usersclient.UserValidator // nil-safe: validation skipped if nil
}

func NewChannelsService(repo repository.IChannelsRepository, userVal usersclient.UserValidator) *ChannelsService {
	return &ChannelsService{repo: repo, userVal: userVal}
}

func (s *ChannelsService) CreateChannel(ctx context.Context, name, description string, isPrivate bool, createdBy uuid.UUID) (*repository.ChannelRow, error) {
	ch, err := s.repo.CreateChannelWithOwner(ctx, name, description, isPrivate, createdBy)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrChannelAlreadyExists
		}
		return nil, fmt.Errorf("create channel: %w", err)
	}
	return ch, nil
}

func (s *ChannelsService) GetChannel(ctx context.Context, channelID, requestingUserID uuid.UUID) (*repository.ChannelRow, error) {
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChannelNotFound
		}
		return nil, fmt.Errorf("get channel: %w", err)
	}

	if ch.IsPrivate {
		_, err := s.repo.GetMember(ctx, channelID, requestingUserID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrPermissionDenied
			}
			return nil, fmt.Errorf("check membership: %w", err)
		}
	}

	return ch, nil
}

func (s *ChannelsService) DeleteChannel(ctx context.Context, channelID, requestingUserID uuid.UUID) error {
	if err := s.requireOwnerOrAdmin(ctx, channelID, requestingUserID); err != nil {
		return err
	}

	if err := s.repo.ArchiveChannel(ctx, channelID); err != nil {
		return fmt.Errorf("archive channel: %w", err)
	}
	return nil
}

func (s *ChannelsService) AddMember(ctx context.Context, channelID, userID uuid.UUID, role string, requestingUserID uuid.UUID) (*repository.MemberRow, error) {
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChannelNotFound
		}
		return nil, fmt.Errorf("get channel: %w", err)
	}

	if ch.Archived {
		return nil, ErrChannelArchived
	}

	if ch.IsPrivate {
		if err := s.requireOwnerOrAdmin(ctx, channelID, requestingUserID); err != nil {
			return nil, err
		}
	}

	if s.userVal != nil {
		if err := s.userVal.UsersExist(ctx, []string{userID.String()}); err != nil {
			if errors.Is(err, usersclient.ErrUserNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, fmt.Errorf("validate user: %w", err)
		}
	}

	member, err := s.repo.AddMember(ctx, channelID, userID, role)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrMemberAlreadyExists
		}
		return nil, fmt.Errorf("add member: %w", err)
	}
	return member, nil
}

func (s *ChannelsService) RemoveMember(ctx context.Context, channelID, userID, requestingUserID uuid.UUID) error {
	// User can always remove themselves; otherwise owner/admin required.
	if userID != requestingUserID {
		if err := s.requireOwnerOrAdmin(ctx, channelID, requestingUserID); err != nil {
			return err
		}
	}

	if err := s.repo.RemoveMember(ctx, channelID, userID); err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	return nil
}

func (s *ChannelsService) ListUserChannels(ctx context.Context, userID uuid.UUID) ([]*repository.ChannelRow, error) {
	channels, err := s.repo.ListUserChannels(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user channels: %w", err)
	}
	return channels, nil
}

func (s *ChannelsService) ListChannelMembers(ctx context.Context, channelID uuid.UUID) ([]*repository.MemberRow, error) {
	members, err := s.repo.ListChannelMembers(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("list channel members: %w", err)
	}
	return members, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *ChannelsService) requireOwnerOrAdmin(ctx context.Context, channelID, userID uuid.UUID) error {
	member, err := s.repo.GetMember(ctx, channelID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("get member: %w", err)
	}
	if member.Role != "owner" && member.Role != "admin" {
		return ErrPermissionDenied
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgxerr.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
