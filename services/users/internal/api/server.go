package api

import (
	"context"
	"net/http"
	"time"
	"users/internal/broker"
	"users/internal/config"
	"users/internal/presence"
	"users/internal/service"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(svc *service.UsersService, b *broker.Broker, tracker *presence.Tracker, cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1 MB
		c.Next()
	})

	h := &Handler{svc: svc, broker: b, tracker: tracker}
	jwtMw := jwtMiddleware(cfg.JwtPublicKey)

	router.GET("/__health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1/users")
	{
		v1.POST("/", h.CreateUser)
		v1.GET("", jwtMw, h.ListUsers)
		v1.GET("/me", jwtMw, h.Me)
		v1.GET("/search", jwtMw, h.SearchUsers)
		v1.GET("/:id", h.GetUser)
		v1.PATCH("/me", jwtMw, h.UpdateProfile)
		v1.POST("/heartbeat", jwtMw, h.Heartbeat)
		v1.GET("/presence/stream", jwtMw, h.StreamPresence)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.HTTPPort,
			Handler: router,
			// WriteTimeout must be 0 to support long-lived SSE connections.
			ReadTimeout: 10 * time.Second,
			WriteTimeout: 0,
			IdleTimeout: 120 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
