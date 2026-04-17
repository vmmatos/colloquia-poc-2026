package grpc

import (
	"channels/internal/pb"
	"channels/internal/repository"
	"channels/internal/service"
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChannelsHandler struct {
	pb.UnimplementedChannelServiceServer
	svc *service.ChannelsService
}

func NewChannelsHandler(svc *service.ChannelsService) *ChannelsHandler {
	return &ChannelsHandler{svc: svc}
}

func (h *ChannelsHandler) CreateChannel(ctx context.Context, req *pb.CreateChannelRequest) (*pb.ChannelResponse, error) {
	createdBy, err := uuid.Parse(req.GetCreatedBy())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid created_by: %v", err)
	}

	memberIDs := make([]uuid.UUID, 0, len(req.GetMemberIds()))
	for _, s := range req.GetMemberIds() {
		uid, err := uuid.Parse(s)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid member_id %q: %v", s, err)
		}
		memberIDs = append(memberIDs, uid)
	}

	channelType := req.GetType()
	if channelType == "" {
		channelType = "channel"
	}

	ch, err := h.svc.CreateChannel(ctx, req.GetName(), req.GetDescription(), req.GetIsPrivate(), channelType, createdBy, memberIDs)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.ChannelResponse{Channel: toProtoChannel(ch)}, nil
}

func (h *ChannelsHandler) CreateDM(ctx context.Context, req *pb.CreateDMRequest) (*pb.CreateDMResponse, error) {
	requestingUserID, err := uuid.Parse(req.GetRequestingUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid requesting_user_id: %v", err)
	}

	otherUserID, err := uuid.Parse(req.GetOtherUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid other_user_id: %v", err)
	}

	ch, created, err := h.svc.CreateDM(ctx, requestingUserID, otherUserID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateDMResponse{Channel: toProtoChannel(ch), Created: created}, nil
}

func (h *ChannelsHandler) GetChannel(ctx context.Context, req *pb.GetChannelRequest) (*pb.ChannelResponse, error) {
	channelID, err := uuid.Parse(req.GetChannelId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid channel_id: %v", err)
	}

	requestingUserID, err := uuid.Parse(req.GetRequestingUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid requesting_user_id: %v", err)
	}

	ch, err := h.svc.GetChannel(ctx, channelID, requestingUserID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.ChannelResponse{Channel: toProtoChannel(ch)}, nil
}

func (h *ChannelsHandler) DeleteChannel(ctx context.Context, req *pb.DeleteChannelRequest) (*pb.DeleteChannelResponse, error) {
	channelID, err := uuid.Parse(req.GetChannelId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid channel_id: %v", err)
	}

	requestingUserID, err := uuid.Parse(req.GetRequestingUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid requesting_user_id: %v", err)
	}

	if err := h.svc.DeleteChannel(ctx, channelID, requestingUserID); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.DeleteChannelResponse{Success: true}, nil
}

func (h *ChannelsHandler) AddMember(ctx context.Context, req *pb.AddMemberRequest) (*pb.MemberResponse, error) {
	channelID, err := uuid.Parse(req.GetChannelId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid channel_id: %v", err)
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	requestingUserID, err := uuid.Parse(req.GetRequestingUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid requesting_user_id: %v", err)
	}

	role := req.GetRole()
	if role == "" {
		role = "member"
	}

	member, err := h.svc.AddMember(ctx, channelID, userID, role, requestingUserID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.MemberResponse{Member: toProtoMember(member)}, nil
}

func (h *ChannelsHandler) RemoveMember(ctx context.Context, req *pb.RemoveMemberRequest) (*pb.RemoveMemberResponse, error) {
	channelID, err := uuid.Parse(req.GetChannelId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid channel_id: %v", err)
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	requestingUserID, err := uuid.Parse(req.GetRequestingUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid requesting_user_id: %v", err)
	}

	if err := h.svc.RemoveMember(ctx, channelID, userID, requestingUserID); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.RemoveMemberResponse{Success: true}, nil
}

func (h *ChannelsHandler) ListUserChannels(ctx context.Context, req *pb.ListUserChannelsRequest) (*pb.ListUserChannelsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	channels, err := h.svc.ListUserChannels(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	result := make([]*pb.Channel, len(channels))
	for i, ch := range channels {
		result[i] = toProtoChannel(ch)
	}

	return &pb.ListUserChannelsResponse{Channels: result}, nil
}

func (h *ChannelsHandler) ListChannelMembers(ctx context.Context, req *pb.ListChannelMembersRequest) (*pb.ListChannelMembersResponse, error) {
	channelID, err := uuid.Parse(req.GetChannelId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid channel_id: %v", err)
	}

	members, err := h.svc.ListChannelMembers(ctx, channelID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	result := make([]*pb.Member, len(members))
	for i, m := range members {
		result[i] = toProtoMember(m)
	}

	return &pb.ListChannelMembersResponse{Members: result}, nil
}

func (h *ChannelsHandler) ValidateMembership(ctx context.Context, req *pb.ValidateMembershipRequest) (*pb.ValidateMembershipResponse, error) {
	channelID, err := uuid.Parse(req.GetChannelId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid channel_id: %v", err)
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	isMember, err := h.svc.ValidateMembership(ctx, channelID, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.ValidateMembershipResponse{IsMember: isMember}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toProtoChannel(ch *repository.ChannelRow) *pb.Channel {
	return &pb.Channel{
		Id:          ch.ID.String(),
		Name:        ch.Name,
		Description: ch.Description,
		IsPrivate:   ch.IsPrivate,
		CreatedBy:   ch.CreatedBy.String(),
		Archived:    ch.Archived,
		MemberCount: ch.MemberCount,
		CreatedAt:   ch.CreatedAt,
		UpdatedAt:   ch.UpdatedAt,
		Type:        ch.Type,
	}
}

func toProtoMember(m *repository.MemberRow) *pb.Member {
	return &pb.Member{
		ChannelId: m.ChannelID.String(),
		UserId:    m.UserID.String(),
		Role:      m.Role,
		JoinedAt:  m.JoinedAt,
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, service.ErrChannelNotFound):
		return status.Errorf(codes.NotFound, "%v", err)
	case errors.Is(err, service.ErrChannelAlreadyExists):
		return status.Errorf(codes.AlreadyExists, "%v", err)
	case errors.Is(err, service.ErrMemberNotFound):
		return status.Errorf(codes.NotFound, "%v", err)
	case errors.Is(err, service.ErrMemberAlreadyExists):
		return status.Errorf(codes.AlreadyExists, "%v", err)
	case errors.Is(err, service.ErrPermissionDenied):
		return status.Errorf(codes.PermissionDenied, "%v", err)
	case errors.Is(err, service.ErrChannelArchived):
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	case errors.Is(err, service.ErrUserNotFound):
		return status.Errorf(codes.NotFound, "%v", err)
	case errors.Is(err, service.ErrCannotModifyDM):
		return status.Errorf(codes.PermissionDenied, "%v", err)
	case errors.Is(err, service.ErrInvalidRole):
		return status.Errorf(codes.InvalidArgument, "%v", err)
	default:
		return status.Errorf(codes.Internal, "internal error")
	}
}
