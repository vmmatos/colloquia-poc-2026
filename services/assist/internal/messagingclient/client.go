package messagingclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MessageFetcher is the interface the service layer depends on.
type MessageFetcher interface {
	GetMessages(ctx context.Context, channelID string, limit int32) ([]*Message, error)
}

type MessagingClient struct {
	conn   *grpc.ClientConn
	client MessagingServiceClient
}

func NewMessagingClient(address string) (*MessagingClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial messaging service at %s: %w", address, err)
	}
	return &MessagingClient{conn: conn, client: NewMessagingServiceClient(conn)}, nil
}

// GetMessages fetches up to limit messages from the given channel (no cursor — latest messages).
func (c *MessagingClient) GetMessages(ctx context.Context, channelID string, limit int32) ([]*Message, error) {
	resp, err := c.client.GetMessages(ctx, &GetMessagesRequest{
		ChannelId: channelID,
		Limit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("messaging.GetMessages: %w", err)
	}
	return resp.GetMessages(), nil
}

func (c *MessagingClient) Close() error { return c.conn.Close() }
