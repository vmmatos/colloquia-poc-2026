package api

import (
	"auth/internal/config"
	"auth/internal/service"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(authService *service.AuthService, cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1 MB
		c.Next()
	})

	// Parse RSA public key — required for the JWKS endpoint consumed by the
	// API gateway. Fail fast so misconfiguration is caught at startup.
	rsaPub := mustParseRSAPublicKey(cfg.JwtPublicKey)

	h := &Handler{authService: authService}

	router.GET("/__health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Standard OIDC discovery path — consumed by KrakenD to verify JWTs.
	router.GET("/.well-known/jwks.json", jwksHandler(rsaPub))

	v1 := router.Group("/api/v1/auth")
	{
		v1.POST("/register", h.Register)
		v1.POST("/login", h.Login)
		v1.POST("/logout", h.Logout)
		v1.POST("/refresh", h.RefreshToken)
		v1.GET("/validate", h.ValidateToken)
		v1.GET("/me", h.Me)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.HTTPPort,
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func mustParseRSAPublicKey(pemBytes []byte) *rsa.PublicKey {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		log.Fatal("api: failed to decode PEM block for RSA public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatalf("api: failed to parse RSA public key: %v", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		log.Fatal("api: JWT_PUBLIC_KEY is not an RSA public key")
	}
	return rsaPub
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
