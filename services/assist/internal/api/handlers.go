package api

import (
	"errors"
	"log"
	"net/http"

	"assist/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const userIDKey = "userID"

type Handler struct {
	svc *service.AssistService
}

// ── JWT middleware ─────────────────────────────────────────────────────────────

func jwtMiddleware(publicKeyPEM []byte) gin.HandlerFunc {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		panic("assist: failed to parse JWT public key: " + err.Error())
	}

	return func(c *gin.Context) {
		tokenStr, ok := bearerToken(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return publicKey, nil
		}, jwt.WithExpirationRequired())

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		sub, _ := claims.GetSubject()
		userID, err := uuid.Parse(sub)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid subject claim"})
			return
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}

func bearerToken(c *gin.Context) (string, bool) {
	h := c.GetHeader("Authorization")
	if len(h) > 7 && h[:7] == "Bearer " {
		return h[7:], true
	}
	return "", false
}

// ── Request / response types ──────────────────────────────────────────────────

type suggestionsRequest struct {
	ChannelID    string `json:"channel_id"    binding:"required"`
	CurrentInput string `json:"current_input" binding:"required"`
	MessageLimit int32  `json:"message_limit"`
}

type suggestionsResponse struct {
	Suggestions []string `json:"suggestions"`
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// POST /api/v1/assist/suggestions
func (h *Handler) GetSuggestions(c *gin.Context) {
	var req suggestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := uuid.Parse(req.ChannelID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	suggestions, err := h.svc.GetSuggestions(c.Request.Context(), req.ChannelID, req.CurrentInput, req.MessageLimit)
	if err != nil {
		log.Printf("assist: GetSuggestions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate suggestions"})
		return
	}

	c.JSON(http.StatusOK, suggestionsResponse{Suggestions: suggestions})
}
