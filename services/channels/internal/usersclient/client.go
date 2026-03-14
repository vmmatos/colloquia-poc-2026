package usersclient

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ErrUserNotFound is returned by UsersExist when one or more IDs are not found.
var ErrUserNotFound = errors.New("user not found")

// UserValidator is satisfied by the real gRPC client and by mocks in tests.
type UserValidator interface {
	UsersExist(ctx context.Context, ids []string) error
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

// UsersExist calls BatchGetUsers and returns ErrUserNotFound if any id is missing.
func (c *UsersClient) UsersExist(ctx context.Context, ids []string) error {
	resp, err := c.client.BatchGetUsers(ctx, &BatchGetUsersRequest{Ids: ids})
	if err != nil {
		return fmt.Errorf("batch get users: %w", err)
	}
	if len(resp.GetUsers()) != len(ids) {
		return ErrUserNotFound
	}
	return nil
}

func (c *UsersClient) Close() error {
	return c.conn.Close()
}
