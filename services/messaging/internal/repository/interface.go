package repository

import (
	"context"

	"github.com/google/uuid"
)

// MessageRow is the domain type for a persisted message.
type MessageRow struct {
	ID        uuid.UUID
	ChannelID uuid.UUID
	UserID    uuid.UUID
	Content   string
	CreatedAt int64 // Unix timestamp
}

type IMessagingRepository interface {
	InsertMessage(ctx context.Context, channelID, userID uuid.UUID, content string) (*MessageRow, error)
	// ListMessages returns up to limit messages in a channel, ordered newest-first.
	// If beforeID is non-nil, only messages older than that cursor are returned.
	ListMessages(ctx context.Context, channelID uuid.UUID, beforeID *uuid.UUID, limit int32) ([]*MessageRow, error)
}
