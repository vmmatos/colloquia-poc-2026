package grpc

import (
	"auth/internal/pb"
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

type Server struct {
	grpcServer *grpc.Server
}

func NewServer(authHandler *AuthHandler) *Server {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryInterceptor,
			loggingInterceptor,
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              2 * time.Hour,
			Timeout:           20 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	pb.RegisterAuthServiceServer(s, authHandler)

	return &Server{grpcServer: s}
}

func (s *Server) Start(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	log.Printf("gRPC server listening on :%s", port)
	return s.grpcServer.Serve(listener)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}

// ── Interceptors ──────────────────────────────────────────────────────────────

func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	code := codes.OK
	if err != nil {
		if s, ok := status.FromError(err); ok {
			code = s.Code()
		}
	}
	log.Printf("gRPC %s | %s | %s", info.FullMethod, code, time.Since(start))
	return resp, err
}

func recoveryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic recovered in %s: %v", info.FullMethod, r)
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()
	return handler(ctx, req)
}
