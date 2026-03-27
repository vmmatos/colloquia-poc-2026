package service

import (
	"channels/internal/repository"
	"channels/internal/usersclient"
	"context"
	"crypto/sha256"
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
	ErrCannotModifyDM       = errors.New("cannot add or remove members from a DM channel")
)

type ChannelsService struct {
	repo    repository.IChannelsRepository
	userVal usersclient.UserValidator // nil-safe: validation skipped if nil
}

func NewChannelsService(repo repository.IChannelsRepository, userVal usersclient.UserValidator) *ChannelsService {
	return &ChannelsService{repo: repo, userVal: userVal}
}

func (s *ChannelsService) CreateChannel(ctx context.Context, name, description string, isPrivate bool, channelType string, createdBy uuid.UUID, memberIDs []uuid.UUID) (*repository.ChannelRow, error) {
	if channelType == "" {
		channelType = "channel"
	}

	switch channelType {
	case "dm":
		return nil, fmt.Errorf("use the CreateDM method to create DM channels")
	case "group":
		isPrivate = true
	case "channel":
		if name == "" {
			return nil, fmt.Errorf("channel name is required")
		}
	}

	if s.userVal != nil && len(memberIDs) > 0 {
		toValidate := make([]string, 0, len(memberIDs))
		for _, uid := range memberIDs {
			if uid != createdBy {
				toValidate = append(toValidate, uid.String())
			}
		}
		if len(toValidate) > 0 {
			if err := s.userVal.UsersExist(ctx, toValidate); err != nil {
				if errors.Is(err, usersclient.ErrUserNotFound) {
					return nil, ErrUserNotFound
				}
				return nil, fmt.Errorf("validate members: %w", err)
			}
		}
	}

	ch, err := s.repo.CreateChannelWithOwner(ctx, name, description, isPrivate, channelType, nil, createdBy, memberIDs)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrChannelAlreadyExists
		}
		return nil, fmt.Errorf("create channel: %w", err)
	}
	return ch, nil
}

// CreateDM creates or retrieves an existing DM channel between two users.
// Returns (channel, created, error) where created=false means an existing DM was returned.
func (s *ChannelsService) CreateDM(ctx context.Context, creatorID, otherUserID uuid.UUID) (*repository.ChannelRow, bool, error) {
	if s.userVal != nil {
		if err := s.userVal.UsersExist(ctx, []string{otherUserID.String()}); err != nil {
			if errors.Is(err, usersclient.ErrUserNotFound) {
				return nil, false, ErrUserNotFound
			}
			// best-effort: continue if users service is down
		}
	}

	dmKey := dmKeyFor(creatorID, otherUserID)
	ch, err := s.repo.CreateChannelWithOwner(ctx, "", "", true, "dm", &dmKey, creatorID, []uuid.UUID{otherUserID})
	if err != nil {
		if isUniqueViolation(err) {
			existing, err := s.repo.GetChannelByDMKey(ctx, dmKey)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil, false, ErrChannelNotFound
				}
				return nil, false, fmt.Errorf("get dm channel: %w", err)
			}
			return existing, false, nil
		}
		return nil, false, fmt.Errorf("create dm: %w", err)
	}
	return ch, true, nil
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

	if ch.Type == "dm" {
		return nil, ErrCannotModifyDM
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

func (s *ChannelsService) ValidateMembership(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	_, err := s.repo.GetMember(ctx, channelID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("validate membership: %w", err)
	}
	return true, nil
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

// dmKeyFor returns a canonical SHA-256 hex key for a DM between two users.
// The smaller UUID string always comes first to ensure idempotency.
func dmKeyFor(a, b uuid.UUID) string {
	sa, sb := a.String(), b.String()
	if sa > sb {
		sa, sb = sb, sa
	}
	h := sha256.Sum256([]byte(sa + ":" + sb))
	return fmt.Sprintf("%x", h)
}
