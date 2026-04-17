package main

import (
	"auth/internal/api"
	"auth/internal/config"
	grpcserver "auth/internal/grpc"
	"auth/internal/repository/postgres"
	"auth/internal/service"
	"auth/internal/usersclient"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
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

	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: parse config: %v", err)
	}
	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatalf("database: connect: %v", err)
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		log.Fatalf("database: ping: %v", err)
	}
	log.Println("database: connection established")

	// Connect to users service (best-effort — nil if unavailable).
	var usersClient usersclient.UserCreator
	uc, err := usersclient.NewUsersClient(cfg.UsersGRPCAddress)
	if err != nil {
		log.Printf("warn: users service unavailable at %s: %v — user profiles will not be created", cfg.UsersGRPCAddress, err)
	} else {
		usersClient = uc
		defer uc.Close()
	}

	// Wire up layers.
	authRepo := postgres.NewAuthRepository(pool)
	authService, err := service.NewAuthService(authRepo, cfg, usersClient)
	if err != nil {
		log.Fatalf("auth service: %v", err)
	}

	grpcSrv := grpcserver.NewServer(grpcserver.NewAuthHandler(authService))
	httpSrv := api.NewServer(authService, cfg)

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

		// Give servers up to 10 s to drain in-flight requests.
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
