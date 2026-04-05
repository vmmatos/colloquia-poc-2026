package grpc

import (
	"context"
	"assist/internal/pb"
	"assist/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AssistHandler struct {
	pb.UnimplementedAssistServiceServer
	svc *service.AssistService
}

func NewAssistHandler(svc *service.AssistService) *AssistHandler {
	return &AssistHandler{svc: svc}
}

func (h *AssistHandler) GetSuggestions(ctx context.Context, req *pb.GetSuggestionsRequest) (*pb.GetSuggestionsResponse, error) {
	if req.GetChannelId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "channel_id is required")
	}

	suggestions, err := h.svc.GetSuggestions(ctx, req.GetChannelId(), req.GetCurrentInput(), req.GetMessageLimit())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &pb.GetSuggestionsResponse{Suggestions: suggestions}, nil
}
