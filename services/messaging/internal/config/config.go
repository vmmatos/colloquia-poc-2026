package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL          string
	JwtPublicKey         []byte
	GRPCPort             string
	HTTPPort             string
	ChannelsGRPCAddress  string
}

func LoadConfig() (*Config, error) {
	var missing []string
	
	required := func (key string) string {
		v, ok := os.LookupEnv(key)
		if !ok || strings.TrimSpace(v) == "" {
			missing = append(missing, key)
		}
		return v
	}

	cfg := &Config{
		DatabaseURL:         required("MESSAGING_DATABASE_URL"),
		JwtPublicKey:        decodeKey(required("JWT_PUBLIC_KEY")),
		GRPCPort:            required("MESSAGING_GRPC_PORT"),
		HTTPPort:            required("MESSAGING_HTTP_PORT"),
		ChannelsGRPCAddress: required("CHANNELS_GRPC_ADDRESS"),
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing or invalid env vars: %s", strings.Join(missing, ", "))
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
