package config

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL   string
	JwtPrivateKey []byte
	JwtPublicKey  []byte
	ServerPort    string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JwtPrivateKey: decodeKey(os.Getenv("JWT_PRIVATE_KEY")),
		JwtPublicKey:  decodeKey(os.Getenv("JWT_PUBLIC_KEY")),
		ServerPort:    os.Getenv("SERVER_PORT"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if len(cfg.JwtPrivateKey) == 0 {
		return nil, errors.New("JWT_PRIVATE_KEY is required")
	}
	if len(cfg.JwtPublicKey) == 0 {
		return nil, errors.New("JWT_PUBLIC_KEY is required")
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "50051"
	}

	return cfg, nil
}

// decodeKey accepts either a base64-encoded PEM (convenient for env vars / Docker)
// or a raw PEM string (with real or literal-\n newlines).
func decodeKey(s string) []byte {
	if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s)); err == nil {
		return decoded
	}
	// Fall back: replace literal \n with real newlines (common in shell exports).
	return []byte(strings.ReplaceAll(s, `\n`, "\n"))
}
