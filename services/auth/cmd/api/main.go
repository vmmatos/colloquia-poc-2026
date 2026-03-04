package main

import (
	"auth/internal/config"
	grpcserver "auth/internal/grpc"
	"auth/internal/repository/postgres"
	"auth/internal/service"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: connect: %v", err)
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		log.Fatalf("database: ping: %v", err)
	}
	log.Println("database: connection established")

	authRepo := postgres.NewAuthRepository(pool)
	authService := service.NewAuthService(authRepo, cfg)
	authHandler := grpcserver.NewAuthHandler(authService)
	server := grpcserver.NewServer(authHandler)

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down...")
		server.Stop()
	}()

	if err = server.Start(cfg.ServerPort); err != nil {
		log.Fatalf("server: %v", err)
	}
}
