package grpc

import (
	"context"
	"errors"
	"users/internal/pb"
	"users/internal/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UsersHandler struct {
	pb.UnimplementedUserServiceServer
	svc *service.UsersService
}

func NewUsersHandler(svc *service.UsersService) *UsersHandler {
	return &UsersHandler{svc: svc}
}

func (h *UsersHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	result, err := h.svc.CreateUser(ctx, id, req.GetEmail())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UserResponse{User: toProto(result)}, nil
}

func (h *UsersHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	result, err := h.svc.GetUser(ctx, id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UserResponse{User: toProto(result)}, nil
}

func (h *UsersHandler) BatchGetUsers(ctx context.Context, req *pb.BatchGetUsersRequest) (*pb.BatchGetUsersResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.GetIds()))

	for _, s := range req.GetIds() {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid id %q: %v", s, err)
		}
		ids = append(ids, id)
	}

	results, err := h.svc.BatchGetUsers(ctx, ids)
	if err != nil {
		return nil, toGRPCError(err)
	}

	users := make([]*pb.UserProfile, len(results))
	for i, r := range results {
		users[i] = toProto(r)
	}

	return &pb.BatchGetUsersResponse{Users: users}, nil
}

func (h *UsersHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UserResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	result, err := h.svc.UpdateProfile(ctx, id, req.Name, req.Avatar, req.Bio, req.Timezone, req.Status)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UserResponse{User: toProto(result)}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toProto(r *service.UserResult) *pb.UserProfile {
	return &pb.UserProfile{
		Id:        r.ID.String(),
		Email:     r.Email,
		Name:      r.Name,
		Avatar:    r.Avatar,
		Bio:       r.Bio,
		Timezone:  r.Timezone,
		Status:    r.Status,
		CreatedAt: r.CreatedAt.Unix(),
		UpdatedAt: r.UpdatedAt.Unix(),
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		return status.Errorf(codes.NotFound, "%v", err)
	case errors.Is(err, service.ErrUserAlreadyExists):
		return status.Errorf(codes.AlreadyExists, "%v", err)
	default:
		return status.Errorf(codes.Internal, "internal error")
	}
}
