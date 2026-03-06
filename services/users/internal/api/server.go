package api

import (
	"context"
	"net/http"
	"time"
	"users/internal/config"
	"users/internal/service"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(svc *service.UsersService, cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	h := &Handler{svc: svc}
	jwtMw := jwtMiddleware(cfg.JwtPublicKey)

	router.GET("/__health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1/users")
	{
		v1.POST("/", h.CreateUser)
		v1.GET("/me", jwtMw, h.Me)
		v1.GET("/:id", h.GetUser)
		v1.PATCH("/me", jwtMw, h.UpdateProfile)
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
