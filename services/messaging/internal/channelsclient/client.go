package channelsclient

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ErrNotMember = errors.New("user is not a member of the channel")

// MembershipValidator is the interface the service layer depends on.
type MembershipValidator interface {
	ValidateMembership(ctx context.Context, channelID, userID string) error
	Close() error
}

type ChannelsClient struct {
	conn   *grpc.ClientConn
	client ChannelServiceClient
}

func NewChannelsClient(address string) (*ChannelsClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial channels service at %s: %w", address, err)
	}
	return &ChannelsClient{conn: conn, client: NewChannelServiceClient(conn)}, nil
}

// ValidateMembership returns ErrNotMember if the user is not a member.
// Any other error means the channels service is unavailable (treat as fail-closed).
func (c *ChannelsClient) ValidateMembership(ctx context.Context, channelID, userID string) error {
	resp, err := c.client.ValidateMembership(ctx, &ValidateMembershipRequest{
		ChannelId: channelID,
		UserId:    userID,
	})
	if err != nil {
		return fmt.Errorf("channels.ValidateMembership: %w", err)
	}
	if !resp.GetIsMember() {
		return ErrNotMember
	}
	return nil
}

func (c *ChannelsClient) Close() error { return c.conn.Close() }
