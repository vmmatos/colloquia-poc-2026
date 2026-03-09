package usersclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserCreator is satisfied by the real gRPC client and by mocks in tests.
type UserCreator interface {
	CreateUser(ctx context.Context, id, email string) error
	Close() error
}

// UsersClient wraps the gRPC connection to the users service.
type UsersClient struct {
	conn   *grpc.ClientConn
	client UserServiceClient
}

func NewUsersClient(address string) (*UsersClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial users service at %s: %w", address, err)
	}
	return &UsersClient{conn: conn, client: NewUserServiceClient(conn)}, nil
}

func (c *UsersClient) CreateUser(ctx context.Context, id, email string) error {
	_, err := c.client.CreateUser(ctx, &CreateUserRequest{Id: id, Email: email})
	return err
}

func (c *UsersClient) Close() error {
	return c.conn.Close()
}
