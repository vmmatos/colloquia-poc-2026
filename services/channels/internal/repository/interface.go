package repository

import (
	"context"

	"github.com/google/uuid"
)

// ChannelRow mirrors the channels table row plus a computed member_count.
type ChannelRow struct {
	ID          uuid.UUID
	Name        string
	Description string
	IsPrivate   bool
	CreatedBy   uuid.UUID
	Archived    bool
	Type        string
	DMKey       *string
	MemberCount int32
	CreatedAt   int64 // Unix timestamp
	UpdatedAt   int64 // Unix timestamp
}

// MemberRow mirrors the channel_members table row.
type MemberRow struct {
	ChannelID uuid.UUID
	UserID    uuid.UUID
	Role      string
	JoinedAt  int64 // Unix timestamp
}

type IChannelsRepository interface {
	// Channel operations
	CreateChannelWithOwner(ctx context.Context, name, description string, isPrivate bool, channelType string, dmKey *string, createdBy uuid.UUID, memberIDs []uuid.UUID) (*ChannelRow, error)
	GetChannel(ctx context.Context, channelID uuid.UUID) (*ChannelRow, error)
	GetChannelByDMKey(ctx context.Context, dmKey string) (*ChannelRow, error)
	ArchiveChannel(ctx context.Context, channelID uuid.UUID) error

	// Member operations
	AddMember(ctx context.Context, channelID, userID uuid.UUID, role string) (*MemberRow, error)
	RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error
	GetMember(ctx context.Context, channelID, userID uuid.UUID) (*MemberRow, error)

	// List operations
	ListUserChannels(ctx context.Context, userID uuid.UUID) ([]*ChannelRow, error)
	ListChannelMembers(ctx context.Context, channelID uuid.UUID) ([]*MemberRow, error)
	CountChannelMembers(ctx context.Context, channelID uuid.UUID) (int32, error)
}
