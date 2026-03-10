package config

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL  string
	JwtPublicKey []byte
	GRPCPort     string
	HTTPPort     string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DatabaseURL:  os.Getenv("USERS_DATABASE_URL"),
		JwtPublicKey: decodeKey(os.Getenv("JWT_PUBLIC_KEY")),
		GRPCPort:     os.Getenv("USERS_GRPC_PORT"),
		HTTPPort:     os.Getenv("USERS_HTTP_PORT"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("USERS_DATABASE_URL is required")
	}
	if len(cfg.JwtPublicKey) == 0 {
		return nil, errors.New("JWT_PUBLIC_KEY is required")
	}

	if cfg.GRPCPort == "" {
		cfg.GRPCPort = "50052"
	}
	if cfg.HTTPPort == "" {
		cfg.HTTPPort = "8082"
	}

	return cfg, nil
}

// decodeKey accepts either a base64-encoded PEM or a raw PEM string.
func decodeKey(s string) []byte {
	if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s)); err == nil {
		return decoded
	}
	return []byte(strings.ReplaceAll(s, `\n`, "\n"))
}
