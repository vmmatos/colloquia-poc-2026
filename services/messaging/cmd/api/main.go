package main

import (
	"context"
	"errors"
	"log"
	"messaging/internal/api"
	"messaging/internal/broker"
	"messaging/internal/channelsclient"
	"messaging/internal/config"
	grpcserver "messaging/internal/grpc"
	"messaging/internal/repository/postgres"
	"messaging/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	messagingRepo := postgres.NewMessagingRepository(pool)

	chanClient, err := channelsclient.NewChannelsClient(cfg.ChannelsGRPCAddress)
	if err != nil {
		log.Fatalf("channels service unavailable at %s: %v", cfg.ChannelsGRPCAddress, err)
	}
	defer chanClient.Close()
	log.Printf("channels client connected to %s", cfg.ChannelsGRPCAddress)

	sseBroker := broker.NewBroker()

	messagingSvc := service.NewMessagingService(messagingRepo, chanClient, sseBroker)

	grpcSrv := grpcserver.NewServer(grpcserver.NewMessagingHandler(messagingSvc))
	httpSrv := api.NewServer(messagingSvc, sseBroker, cfg)

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
