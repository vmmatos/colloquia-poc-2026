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

	"assist/internal/api"
	"assist/internal/config"
	grpcserver "assist/internal/grpc"
	"assist/internal/messagingclient"
	"assist/internal/provider"
	"assist/internal/provider/ollama"
	"assist/internal/service"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Wire up LLM provider.
	var llmProvider provider.LLMProvider
	switch cfg.LLMProvider {
	case "ollama":
		llmProvider = ollama.New(cfg)
		log.Printf("LLM provider: ollama (%s) at %s", cfg.OllamaModel, cfg.OllamaBaseURL)
	default:
		log.Fatalf("unknown LLM_PROVIDER: %s", cfg.LLMProvider)
	}

	// Wire up messaging client (best-effort).
	var msgClient messagingclient.MessageFetcher
	mc, err := messagingclient.NewMessagingClient(cfg.MessagingGRPCAddress)
	if err != nil {
		log.Printf("warning: messaging client unavailable at %s: %v — context will be empty", cfg.MessagingGRPCAddress, err)
	} else {
		defer mc.Close()
		msgClient = mc
		log.Printf("messaging client connected to %s", cfg.MessagingGRPCAddress)
	}

	assistSvc := service.NewAssistService(msgClient, llmProvider)

	// Warm up the model so the first real request doesn't pay the cold-load cost.
	// This also ensures keep_alive:-1 is applied via the real inference path.
	go func() {
		wCtx, wCancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer wCancel()
		if _, err := llmProvider.Complete(wCtx, provider.CompletionRequest{
			SystemPrompt: "respond with []",
			UserPrompt:   "hi",
		}); err != nil {
			log.Printf("model warm-up: %v (non-fatal)", err)
		} else {
			log.Printf("model warm-up complete")
		}
	}()

	grpcSrv := grpcserver.NewServer(grpcserver.NewAssistHandler(assistSvc))
	httpSrv := api.NewServer(assistSvc, cfg)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
		return grpcSrv.Start(cfg.GRPCPort)
	})

	g.Go(func() error {
		log.Printf("HTTP server listening on :%s", cfg.HTTPPort)
		if err := httpSrv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

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
