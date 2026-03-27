package service

import (
	"context"
	"errors"
	"fmt"
	"messaging/internal/broker"
	"messaging/internal/channelsclient"
	"messaging/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrNotMember       = errors.New("user is not a member of the channel")
	ErrChannelsUnavail = errors.New("channels service unavailable")
)

const (
	defaultLimit = 50
	maxLimit     = 100
)

type MessagingService struct {
	repo    repository.IMessagingRepository
	chanVal channelsclient.MembershipValidator // fail-closed: must not be nil
	broker  *broker.Broker
}

func NewMessagingService(
	repo repository.IMessagingRepository,
	chanVal channelsclient.MembershipValidator,
	b *broker.Broker,
) *MessagingService {
	return &MessagingService{repo: repo, chanVal: chanVal, broker: b}
}

func (s *MessagingService) SendMessage(ctx context.Context, channelID, userID uuid.UUID, content string) (*repository.MessageRow, error) {
	if s.chanVal == nil {
		return nil, ErrChannelsUnavail
	}

	if err := s.chanVal.ValidateMembership(ctx, channelID.String(), userID.String()); err != nil {
		if errors.Is(err, channelsclient.ErrNotMember) {
			return nil, ErrNotMember
		}
		return nil, fmt.Errorf("membership check: %w", err)
	}

	msg, err := s.repo.InsertMessage(ctx, channelID, userID, content)
	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}

	s.broker.Publish(channelID.String(), broker.SSEEvent{
		MessageID: msg.ID.String(),
		ChannelID: msg.ChannelID.String(),
		UserID:    msg.UserID.String(),
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
	})

	return msg, nil
}

func (s *MessagingService) GetMessages(ctx context.Context, channelID uuid.UUID, beforeID *uuid.UUID, limit int32) ([]*repository.MessageRow, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	msgs, err := s.repo.ListMessages(ctx, channelID, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return msgs, nil
}
