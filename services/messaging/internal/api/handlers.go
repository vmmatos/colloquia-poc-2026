package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"messaging/internal/broker"
	"messaging/internal/repository"
	"messaging/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const userIDKey = "userID"

type Handler struct {
	svc    *service.MessagingService
	broker *broker.Broker
}

// ── JWT middleware ─────────────────────────────────────────────────────────────

func jwtMiddleware(publicKeyPEM []byte) gin.HandlerFunc {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		panic("messaging: failed to parse JWT public key: " + err.Error())
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
	// Fallback para SSE via EventSource (não suporta headers customizados)
	if t := c.Query("token"); t != "" {
		return t, true
	}
	return "", false
}

// ── Request / response types ──────────────────────────────────────────────────

type sendMessageRequest struct {
	ChannelID string `json:"channel_id" binding:"required"`
	Content   string `json:"content"    binding:"required,max=4000"`
}

type messageResponse struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

func toMessageResponse(m *repository.MessageRow) messageResponse {
	return messageResponse{
		ID:        m.ID.String(),
		ChannelID: m.ChannelID.String(),
		UserID:    m.UserID.String(),
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// POST /api/v1/messages
func (h *Handler) SendMessage(c *gin.Context) {
	userID := c.MustGet(userIDKey).(uuid.UUID)

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channelID, err := uuid.Parse(req.ChannelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	msg, err := h.svc.SendMessage(c.Request.Context(), channelID, userID, req.Content)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toMessageResponse(msg))
}

// GET /api/v1/messages?channel_id=&before_id=&limit=
func (h *Handler) GetMessages(c *gin.Context) {
	channelID, err := uuid.Parse(c.Query("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	var beforeID *uuid.UUID
	if raw := c.Query("before_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid before_id"})
			return
		}
		beforeID = &id
	}

	var limit int32 = 50
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err == nil && n > 0 {
			limit = int32(n)
		}
	}

	msgs, err := h.svc.GetMessages(c.Request.Context(), channelID, beforeID, limit)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	resp := make([]messageResponse, len(msgs))
	for i, m := range msgs {
		resp[i] = toMessageResponse(m)
	}
	c.JSON(http.StatusOK, resp)
}

// GET /api/v1/messages/stream?channel_id=id1&channel_id=id2...
// Accepts one or more channel_id query parameters in a single SSE connection.
func (h *Handler) StreamMessages(c *gin.Context) {
	channelIDs := c.QueryArray("channel_id")
	if len(channelIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id is required"})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Subscribe to all channels and fan-in events into a single merged channel.
	merged := make(chan broker.SSEEvent, 32)
	subs := make([]chan broker.SSEEvent, len(channelIDs))
	for i, id := range channelIDs {
		sub := h.broker.Subscribe(id)
		subs[i] = sub
		go func(s chan broker.SSEEvent) {
			for evt := range s {
				select {
				case merged <- evt:
				default:
				}
			}
		}(sub)
	}
	defer func() {
		for i, id := range channelIDs {
			h.broker.Unsubscribe(id, subs[i])
		}
	}()

	// Initial flush to establish the connection.
	fmt.Fprintf(c.Writer, ": connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event := <-merged:
			data, _ := json.Marshal(event)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(c.Writer, ": heartbeat\n\n")
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

// ── Error mapping ─────────────────────────────────────────────────────────────

func serviceErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrNotMember):
		return http.StatusForbidden
	case errors.Is(err, service.ErrChannelsUnavail):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
