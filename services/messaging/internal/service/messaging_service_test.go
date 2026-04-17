package service

import (
	"context"
	"errors"
	"messaging/internal/broker"
	"messaging/internal/channelsclient"
	"messaging/internal/repository"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// stubMessagingRepo is a minimal IMessagingRepository for testing.
type stubMessagingRepo struct {
	repository.IMessagingRepository
	messages         map[uuid.UUID]*repository.MessageRow
	insertError      error
	lastInsertParams struct {
		channelID uuid.UUID
		userID    uuid.UUID
		content   string
	}
}

func (s *stubMessagingRepo) InsertMessage(ctx context.Context, channelID, userID uuid.UUID, content string) (*repository.MessageRow, error) {
	if s.insertError != nil {
		return nil, s.insertError
	}
	s.lastInsertParams.channelID = channelID
	s.lastInsertParams.userID = userID
	s.lastInsertParams.content = content

	id := uuid.New()
	msg := &repository.MessageRow{
		ID:        id,
		ChannelID: channelID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now().Unix(),
	}
	s.messages[id] = msg
	return msg, nil
}

func (s *stubMessagingRepo) ListMessages(ctx context.Context, channelID uuid.UUID, beforeID *uuid.UUID, limit int32) ([]*repository.MessageRow, error) {
	var result []*repository.MessageRow
	for _, msg := range s.messages {
		if msg.ChannelID == channelID {
			result = append(result, msg)
		}
	}
	if len(result) > int(limit) {
		result = result[:limit]
	}
	return result, nil
}

// stubMembershipValidator mocks MembershipValidator.
type stubMembershipValidator struct {
	isMember bool
	err      error
}

func (s *stubMembershipValidator) ValidateMembership(ctx context.Context, channelID, userID string) error {
	if s.err != nil {
		return s.err
	}
	if !s.isMember {
		return channelsclient.ErrNotMember
	}
	return nil
}

func (s *stubMembershipValidator) Close() error { return nil }

// TestSendMessageHappyPath verifies successful message sending.
func TestSendMessageHappyPath(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	valStub := &stubMembershipValidator{isMember: true}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	channelID := uuid.New()
	userID := uuid.New()
	content := "Hello, world!"

	result, err := svc.SendMessage(context.Background(), channelID, userID, content)

	if err != nil {
		t.Errorf("SendMessage: got error %v; want nil", err)
	}
	if result == nil {
		t.Errorf("SendMessage: got nil result; want message")
		return
	}
	if result.ChannelID != channelID {
		t.Errorf("SendMessage: channelID mismatch")
	}
	if result.UserID != userID {
		t.Errorf("SendMessage: userID mismatch")
	}
	if result.Content != content {
		t.Errorf("SendMessage: content mismatch")
	}
}

// TestSendMessageNotMember verifies rejection of non-members.
func TestSendMessageNotMember(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	valStub := &stubMembershipValidator{isMember: false}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	channelID := uuid.New()
	userID := uuid.New()

	result, err := svc.SendMessage(context.Background(), channelID, userID, "test")

	if !errors.Is(err, ErrNotMember) {
		t.Errorf("SendMessage not member: got error %v; want ErrNotMember", err)
	}
	if result != nil {
		t.Errorf("SendMessage not member: got result; want nil")
	}
}

// TestSendMessageValidatorUnavailable verifies error from membership check.
func TestSendMessageValidatorError(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	valStub := &stubMembershipValidator{err: errors.New("service unavailable")}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	channelID := uuid.New()
	userID := uuid.New()

	result, err := svc.SendMessage(context.Background(), channelID, userID, "test")

	if err == nil {
		t.Errorf("SendMessage validator error: got nil; want error")
	}
	if result != nil {
		t.Errorf("SendMessage validator error: got result; want nil")
	}
}

// TestSendMessageNilValidator verifies ErrChannelsUnavail when validator is nil.
func TestSendMessageNilValidator(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, nil, b)

	channelID := uuid.New()
	userID := uuid.New()

	result, err := svc.SendMessage(context.Background(), channelID, userID, "test")

	if !errors.Is(err, ErrChannelsUnavail) {
		t.Errorf("SendMessage nil validator: got error %v; want ErrChannelsUnavail", err)
	}
	if result != nil {
		t.Errorf("SendMessage nil validator: got result; want nil")
	}
}

// TestSendMessageInsertError verifies error propagation from repo.
func TestSendMessageInsertError(t *testing.T) {
	repoStub := &stubMessagingRepo{
		messages:    make(map[uuid.UUID]*repository.MessageRow),
		insertError: errors.New("database error"),
	}
	valStub := &stubMembershipValidator{isMember: true}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	channelID := uuid.New()
	userID := uuid.New()

	result, err := svc.SendMessage(context.Background(), channelID, userID, "test")

	if err == nil {
		t.Errorf("SendMessage insert error: got nil; want error")
	}
	if result != nil {
		t.Errorf("SendMessage insert error: got result; want nil")
	}
}

// TestGetMessagesLimitClamp verifies default limit clamping.
func TestGetMessagesLimitClamp(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	valStub := &stubMembershipValidator{isMember: true}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	channelID := uuid.New()

	// Test with limit 0 (should clamp to defaultLimit)
	messages, err := svc.GetMessages(context.Background(), channelID, nil, 0)

	if err != nil {
		t.Errorf("GetMessages limit 0: got error %v; want nil", err)
	}
	if len(messages) != 0 {
		t.Errorf("GetMessages limit 0: expected empty result, got %d messages", len(messages))
	}
}

// TestGetMessagesMaxLimitClamp verifies max limit enforcement.
func TestGetMessagesMaxLimitClamp(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	valStub := &stubMembershipValidator{isMember: true}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	channelID := uuid.New()

	// Test with limit > maxLimit
	messages, err := svc.GetMessages(context.Background(), channelID, nil, 500)

	if err != nil {
		t.Errorf("GetMessages max clamp: got error %v; want nil", err)
	}
	if len(messages) != 0 {
		t.Errorf("GetMessages max clamp: expected empty result, got %d messages", len(messages))
	}
}

// TestGetMessagesWithCursor verifies cursor handling.
func TestGetMessagesWithCursor(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	valStub := &stubMembershipValidator{isMember: true}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	channelID := uuid.New()
	cursorID := uuid.New()

	// Test with non-nil beforeID
	messages, err := svc.GetMessages(context.Background(), channelID, &cursorID, 50)

	if err != nil {
		t.Errorf("GetMessages with cursor: got error %v; want nil", err)
	}
	if len(messages) != 0 {
		t.Errorf("GetMessages with cursor: expected empty result, got %d messages", len(messages))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ERROR & EDGE CASE TESTS

// TestSendMessageEmptyContent verifies message with empty content is sent.
func TestSendMessageEmptyContent(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	valStub := &stubMembershipValidator{isMember: true}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	channelID := uuid.New()
	userID := uuid.New()

	// Empty content is allowed (e.g., reaction-only message)
	result, err := svc.SendMessage(context.Background(), channelID, userID, "")

	if err != nil {
		t.Errorf("SendMessage empty: got error %v; want nil", err)
	}
	if result == nil || result.Content != "" {
		t.Errorf("SendMessage empty: expected empty message content")
	}
}

// TestGetMessagesNegativeLimit verifies limit 0 defaults correctly.
func TestGetMessagesNegativeLimit(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	valStub := &stubMembershipValidator{isMember: true}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	// Negative limit should clamp to default
	messages, err := svc.GetMessages(context.Background(), uuid.New(), nil, -10)

	if err != nil {
		t.Errorf("GetMessages negative limit: got error %v; want nil", err)
	}
	// Result should be slice (even if empty), not nil
	if len(messages) != 0 {
		t.Errorf("GetMessages negative limit: expected 0 messages")
	}
}

// TestSendMessageLongContent verifies handling of long message content.
func TestSendMessageLongContent(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	valStub := &stubMembershipValidator{isMember: true}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	channelID := uuid.New()
	userID := uuid.New()
	// Long content
	longContent := "Lorem ipsum dolor sit amet. " + strings.Repeat("x", 10000)

	result, err := svc.SendMessage(context.Background(), channelID, userID, longContent)

	if err != nil {
		t.Errorf("SendMessage long: got error %v; want nil", err)
	}
	if result == nil || result.Content != longContent {
		t.Errorf("SendMessage long: content mismatch")
	}
}

// TestGetMessagesMultipleMessages verifies correct message retrieval.
func TestGetMessagesMultipleMessages(t *testing.T) {
	repoStub := &stubMessagingRepo{messages: make(map[uuid.UUID]*repository.MessageRow)}
	valStub := &stubMembershipValidator{isMember: true}
	b := broker.NewBroker()

	svc := NewMessagingService(repoStub, valStub, b)

	channelID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()

	// Send multiple messages
	msg1, _ := svc.SendMessage(context.Background(), channelID, userID1, "message 1")
	msg2, _ := svc.SendMessage(context.Background(), channelID, userID2, "message 2")

	// Retrieve messages
	messages, err := svc.GetMessages(context.Background(), channelID, nil, 50)

	if err != nil {
		t.Errorf("GetMessages multiple: got error %v; want nil", err)
	}
	if len(messages) != 2 {
		t.Errorf("GetMessages multiple: got %d messages; want 2", len(messages))
	}
	if messages != nil && len(messages) == 2 {
		if messages[0].Content != msg1.Content || messages[1].Content != msg2.Content {
			t.Errorf("GetMessages multiple: content mismatch")
		}
	}
}
