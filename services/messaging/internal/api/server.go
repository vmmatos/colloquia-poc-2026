package api

import (
	"context"
	"messaging/internal/broker"
	"messaging/internal/config"
	"messaging/internal/service"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(svc *service.MessagingService, b *broker.Broker, cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1 MB
		c.Next()
	})
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
	}))

	h := &Handler{svc: svc, broker: b}
	jwtMw := jwtMiddleware(cfg.JwtPublicKey)

	router.GET("/__health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1/messages")
	{
		v1.POST("", jwtMw, h.SendMessage)
		v1.GET("", jwtMw, h.GetMessages)
		v1.GET("/stream", jwtMw, h.StreamMessages)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:        ":" + cfg.HTTPPort,
			Handler:     router,
			ReadTimeout: 10 * time.Second,
			// WriteTimeout disabled (0) to allow long-lived SSE connections.
			WriteTimeout: 0,
			IdleTimeout:  120 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
