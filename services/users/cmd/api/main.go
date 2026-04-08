package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"users/internal/api"
	"users/internal/broker"
	"users/internal/config"
	grpcserver "users/internal/grpc"
	"users/internal/presence"
	"users/internal/repository/postgres"
	"users/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
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

	// Wire up layers.
	usersRepo := postgres.NewUsersRepository(pool)
	usersSvc := service.NewUsersService(usersRepo)

	presenceBroker := broker.NewBroker()
	presenceTracker := presence.NewTracker(presenceBroker, usersRepo)
	presenceTracker.StartReaper(ctx)

	grpcSrv := grpcserver.NewServer(grpcserver.NewUsersHandler(usersSvc))
	httpSrv := api.NewServer(usersSvc, presenceBroker, presenceTracker, cfg)

	// Run both servers concurrently; cancel context if either fails.
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
		if err := grpcSrv.Start(cfg.GRPCPort); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		log.Printf("HTTP server listening on :%s", cfg.HTTPPort)
		if err := httpSrv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	// Graceful shutdown on OS signal or on first server error.
	g.Go(func() error {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-quit:
			log.Printf("received signal: %s — shutting down", sig)
		case <-gCtx.Done():
		}

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		grpcSrv.Stop()
		if err := httpSrv.Stop(shutdownCtx); err != nil {
			log.Printf("HTTP shutdown error: %v", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("server: %v", err)
	}
	log.Println("shutdown complete")
}
