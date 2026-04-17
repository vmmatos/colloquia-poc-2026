package api

import (
	"channels/internal/config"
	"channels/internal/service"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(svc *service.ChannelsService, cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1 MB
		c.Next()
	})

	h := &Handler{svc: svc}
	jwtMw := jwtMiddleware(cfg.JwtPublicKey)

	router.GET("/__health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1/channels")
	{
		v1.POST("", jwtMw, h.CreateChannel)
		v1.POST("/dm", jwtMw, h.CreateDM)
		v1.GET("/me", jwtMw, h.ListMyChannels)
		v1.GET("/:id", jwtMw, h.GetChannel)
		v1.DELETE("/:id", jwtMw, h.DeleteChannel)
		v1.POST("/:id/members", jwtMw, h.AddMember)
		v1.DELETE("/:id/members/:userId", jwtMw, h.RemoveMember)
		v1.GET("/:id/members", jwtMw, h.ListMembers)
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

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
