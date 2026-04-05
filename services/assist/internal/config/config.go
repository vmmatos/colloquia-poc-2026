package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPPort             string
	GRPCPort             string
	JwtPublicKey         []byte
	MessagingGRPCAddress string
	LLMProvider          string
	OllamaBaseURL        string
	OllamaModel          string
	OllamaTimeoutSeconds int
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
		HTTPPort:             required("ASSIST_HTTP_PORT"),
		GRPCPort:             required("ASSIST_GRPC_PORT"),
		JwtPublicKey:         decodeKey(required("JWT_PUBLIC_KEY")),
		MessagingGRPCAddress: required("ASSIST_MESSAGING_GRPC_ADDRESS"),
		LLMProvider:          required("ASSIST_LLM_PROVIDER"),
		OllamaBaseURL:        required("ASSIST_OLLAMA_BASE_URL"),
		OllamaModel:          required("ASSIST_OLLAMA_MODEL"),
	}

	timeoutStr := os.Getenv("ASSIST_OLLAMA_TIMEOUT_SECONDS")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		missing = append(missing, "ASSIST_OLLAMA_TIMEOUT_SECONDS (invalid int)")
	} else {
		cfg.OllamaTimeoutSeconds = timeout
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
