package config

import "os"

type Config struct {
	DatabaseURL string
	JwtPrivateKey []byte
	JwtPublicKey []byte
	TokenExpirationTime string
	ServerPort string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JwtPrivateKey: []byte(os.Getenv("JWT_PRIVATE_KEY")),
		JwtPublicKey: []byte(os.Getenv("JWT_PUBLIC_KEY")),
		TokenExpirationTime: os.Getenv("TOKEN_EXPIRATION_TIME"),
		ServerPort: os.Getenv("SERVER_PORT"),
	}
}
