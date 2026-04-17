package service

import (
	"channels/internal/repository"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// stubChannelsRepo is a minimal IChannelsRepository for testing AddMember role validation.
type stubChannelsRepo struct {
	repository.IChannelsRepository // embed for unimplemented methods
}

func (s *stubChannelsRepo) GetChannel(_ context.Context, _ uuid.UUID) (*repository.ChannelRow, error) {
	return &repository.ChannelRow{
		Type:      "channel",
		IsPrivate: false,
		Archived:  false,
	}, nil
}

func (s *stubChannelsRepo) GetMember(_ context.Context, _, _ uuid.UUID) (*repository.MemberRow, error) {
	return nil, nil // no pre-existing member
}

func (s *stubChannelsRepo) AddMember(_ context.Context, _, _ uuid.UUID, role string) (*repository.MemberRow, error) {
	return &repository.MemberRow{Role: role}, nil
}

func TestAddMemberRoleValidation(t *testing.T) {
	stub := &stubChannelsRepo{}
	svc := NewChannelsService(stub, nil)

	channelID := uuid.New()
	userID := uuid.New()
	requestingID := uuid.New()

	validRoles := []string{"member", "admin", "owner"}
	for _, role := range validRoles {
		_, err := svc.AddMember(context.Background(), channelID, userID, role, requestingID)
		if err != nil {
			t.Errorf("AddMember with valid role %q: got error %v; want nil", role, err)
		}
	}

	invalidRoles := []string{"superadmin", "root", "hacker", "MEMBER", "Admin", "owner "}
	for _, role := range invalidRoles {
		_, err := svc.AddMember(context.Background(), channelID, userID, role, requestingID)
		if !errors.Is(err, ErrInvalidRole) {
			t.Errorf("AddMember with invalid role %q: got %v; want ErrInvalidRole", role, err)
		}
	}
}

func TestAddMemberEmptyRoleDefaultsToMember(t *testing.T) {
	stub := &stubChannelsRepo{}
	svc := NewChannelsService(stub, nil)

	channelID := uuid.New()
	userID := uuid.New()
	requestingID := uuid.New()

	member, err := svc.AddMember(context.Background(), channelID, userID, "", requestingID)
	if err != nil {
		t.Fatalf("AddMember with empty role: got error %v; want nil", err)
	}
	if member.Role != "member" {
		t.Errorf("AddMember with empty role: got role %q; want %q", member.Role, "member")
	}
}
