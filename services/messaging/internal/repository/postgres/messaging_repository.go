package postgres

import (
	"context"
	"messaging/internal/db/sqlc"
	"messaging/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessagingRepository struct {
	queries *sqlc.Queries
}

func NewMessagingRepository(pool *pgxpool.Pool) *MessagingRepository {
	return &MessagingRepository{queries: sqlc.New(pool)}
}

func (r *MessagingRepository) InsertMessage(ctx context.Context, channelID, userID uuid.UUID, content string) (*repository.MessageRow, error) {
	msg, err := r.queries.InsertMessage(ctx, sqlc.InsertMessageParams{
		ChannelID: channelID,
		UserID:    userID,
		Content:   content,
	})
	if err != nil {
		return nil, err
	}
	return toMessageRow(msg), nil
}

func (r *MessagingRepository) ListMessages(ctx context.Context, channelID uuid.UUID, beforeID *uuid.UUID, limit int32) ([]*repository.MessageRow, error) {
	var rows []sqlc.Message
	var err error

	if beforeID == nil {
		rows, err = r.queries.ListMessagesFirst(ctx, sqlc.ListMessagesFirstParams{
			ChannelID: channelID,
			Limit:     limit,
		})
	} else {
		rows, err = r.queries.ListMessagesFromCursor(ctx, sqlc.ListMessagesFromCursorParams{
			ChannelID: channelID,
			ID:        *beforeID,
			Limit:     limit,
		})
	}
	if err != nil {
		return nil, err
	}

	result := make([]*repository.MessageRow, len(rows))
	for i, row := range rows {
		result[i] = toMessageRow(row)
	}
	return result, nil
}

func toMessageRow(m sqlc.Message) *repository.MessageRow {
	return &repository.MessageRow{
		ID:        m.ID,
		ChannelID: m.ChannelID,
		UserID:    m.UserID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt.Time.Unix(),
	}
}
