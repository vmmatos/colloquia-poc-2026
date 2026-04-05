package api

import (
	"context"
	"assist/internal/config"
	"assist/internal/service"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(svc *service.AssistService, cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost"},
		AllowMethods:     []string{"POST", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
	}))

	h := &Handler{svc: svc}
	jwtMw := jwtMiddleware(cfg.JwtPublicKey)

	router.GET("/__health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1/assist")
	{
		v1.POST("/suggestions", jwtMw, h.GetSuggestions)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.HTTPPort,
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 0, // disabled: LLM responses can take 60-90s on cold start
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
