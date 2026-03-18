package postgres

import (
	"channels/internal/db/sqlc"
	"channels/internal/repository"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChannelsRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewChannelsRepository(pool *pgxpool.Pool) *ChannelsRepository {
	return &ChannelsRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// CreateChannelWithOwner creates a channel, inserts the creator as 'owner', and adds any extra memberIDs in one transaction.
func (r *ChannelsRepository) CreateChannelWithOwner(ctx context.Context, name, description string, isPrivate bool, createdBy uuid.UUID, memberIDs []uuid.UUID) (*repository.ChannelRow, error) {
	var channel sqlc.Channel

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		var err error
		channel, err = q.CreateChannel(ctx, sqlc.CreateChannelParams{
			Name:        name,
			Description: description,
			IsPrivate:   isPrivate,
			CreatedBy:   createdBy,
		})
		if err != nil {
			return err
		}

		_, err = q.AddChannelMember(ctx, sqlc.AddChannelMemberParams{
			ChannelID: channel.ID,
			UserID:    createdBy,
			Role:      "owner",
		})
		if err != nil {
			return err
		}

		for _, uid := range memberIDs {
			if uid == createdBy {
				continue // owner already inserted
			}
			_, err = q.AddChannelMember(ctx, sqlc.AddChannelMemberParams{
				ChannelID: channel.ID,
				UserID:    uid,
				Role:      "member",
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	count, err := r.queries.CountChannelMembers(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	return toChannelRow(channel, count), nil
}

func (r *ChannelsRepository) GetChannel(ctx context.Context, channelID uuid.UUID) (*repository.ChannelRow, error) {
	channel, err := r.queries.GetChannelByID(ctx, channelID)
	if err != nil {
		return nil, err
	}

	count, err := r.queries.CountChannelMembers(ctx, channelID)
	if err != nil {
		return nil, err
	}

	return toChannelRow(channel, count), nil
}

func (r *ChannelsRepository) ArchiveChannel(ctx context.Context, channelID uuid.UUID) error {
	return r.queries.ArchiveChannel(ctx, channelID)
}

func (r *ChannelsRepository) AddMember(ctx context.Context, channelID, userID uuid.UUID, role string) (*repository.MemberRow, error) {
	member, err := r.queries.AddChannelMember(ctx, sqlc.AddChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
		Role:      role,
	})
	if err != nil {
		return nil, err
	}
	return toMemberRow(member), nil
}

func (r *ChannelsRepository) RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error {
	return r.queries.RemoveChannelMember(ctx, sqlc.RemoveChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
}

func (r *ChannelsRepository) GetMember(ctx context.Context, channelID, userID uuid.UUID) (*repository.MemberRow, error) {
	member, err := r.queries.GetChannelMember(ctx, sqlc.GetChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		return nil, err
	}
	return toMemberRow(member), nil
}

func (r *ChannelsRepository) ListUserChannels(ctx context.Context, userID uuid.UUID) ([]*repository.ChannelRow, error) {
	rows, err := r.queries.ListUserChannels(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*repository.ChannelRow, len(rows))
	for i, row := range rows {
		result[i] = &repository.ChannelRow{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			IsPrivate:   row.IsPrivate,
			CreatedBy:   row.CreatedBy,
			Archived:    row.Archived,
			MemberCount: row.MemberCount,
			CreatedAt:   row.CreatedAt.Time.Unix(),
			UpdatedAt:   row.UpdatedAt.Time.Unix(),
		}
	}
	return result, nil
}

func (r *ChannelsRepository) ListChannelMembers(ctx context.Context, channelID uuid.UUID) ([]*repository.MemberRow, error) {
	members, err := r.queries.ListChannelMembers(ctx, channelID)
	if err != nil {
		return nil, err
	}

	result := make([]*repository.MemberRow, len(members))
	for i, m := range members {
		result[i] = toMemberRow(m)
	}
	return result, nil
}

func (r *ChannelsRepository) CountChannelMembers(ctx context.Context, channelID uuid.UUID) (int32, error) {
	return r.queries.CountChannelMembers(ctx, channelID)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toChannelRow(c sqlc.Channel, count int32) *repository.ChannelRow {
	return &repository.ChannelRow{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		IsPrivate:   c.IsPrivate,
		CreatedBy:   c.CreatedBy,
		Archived:    c.Archived,
		MemberCount: count,
		CreatedAt:   c.CreatedAt.Time.Unix(),
		UpdatedAt:   c.UpdatedAt.Time.Unix(),
	}
}

func toMemberRow(m sqlc.ChannelMember) *repository.MemberRow {
	return &repository.MemberRow{
		ChannelID: m.ChannelID,
		UserID:    m.UserID,
		Role:      m.Role,
		JoinedAt:  m.JoinedAt.Time.Unix(),
	}
}
