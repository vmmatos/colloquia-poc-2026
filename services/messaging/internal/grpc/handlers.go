package grpc

import (
	"context"
	"errors"
	"messaging/internal/pb"
	"messaging/internal/repository"
	"messaging/internal/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MessagingHandler struct {
	pb.UnimplementedMessagingServiceServer
	svc *service.MessagingService
}

func NewMessagingHandler(svc *service.MessagingService) *MessagingHandler {
	return &MessagingHandler{svc: svc}
}

func (h *MessagingHandler) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.MessageResponse, error) {
	channelID, err := uuid.Parse(req.GetChannelId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid channel_id: %v", err)
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	if req.GetContent() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "content is required")
	}

	msg, err := h.svc.SendMessage(ctx, channelID, userID, req.GetContent())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.MessageResponse{Message: toProtoMessage(msg)}, nil
}

func (h *MessagingHandler) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	channelID, err := uuid.Parse(req.GetChannelId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid channel_id: %v", err)
	}

	var beforeID *uuid.UUID
	if raw := req.GetBeforeId(); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid before_id: %v", err)
		}
		beforeID = &id
	}

	msgs, err := h.svc.GetMessages(ctx, channelID, beforeID, req.GetLimit())
	if err != nil {
		return nil, toGRPCError(err)
	}

	result := make([]*pb.Message, len(msgs))
	for i, m := range msgs {
		result[i] = toProtoMessage(m)
	}

	return &pb.GetMessagesResponse{Messages: result}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toProtoMessage(m *repository.MessageRow) *pb.Message {
	return &pb.Message{
		Id:        m.ID.String(),
		ChannelId: m.ChannelID.String(),
		UserId:    m.UserID.String(),
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, service.ErrNotMember):
		return status.Errorf(codes.PermissionDenied, "%v", err)
	case errors.Is(err, service.ErrChannelsUnavail):
		return status.Errorf(codes.Unavailable, "%v", err)
	default:
		return status.Errorf(codes.Internal, "internal error")
	}
}
