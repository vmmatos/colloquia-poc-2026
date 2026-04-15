package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"users/internal/broker"
	"users/internal/presence"
	"users/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const userIDKey = "userID"

type Handler struct {
	svc     *service.UsersService
	broker  *broker.Broker
	tracker *presence.Tracker
}

// ── JWT middleware ─────────────────────────────────────────────────────────────

// jwtMiddleware validates RS256 tokens locally using the public key.
// On success it stores the user UUID under userIDKey in the Gin context.
func jwtMiddleware(publicKeyPEM []byte) gin.HandlerFunc {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		panic("users: failed to parse JWT public key: " + err.Error())
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
	// Fallback for SSE via EventSource (browser cannot set custom headers)
	if t := c.Query("token"); t != "" {
		return t, true
	}
	return "", false
}

// ── Request / response types ──────────────────────────────────────────────────

type createUserRequest struct {
	ID    string `json:"id"    binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type updateProfileRequest struct {
	Name     *string `json:"name"`
	Avatar   *string `json:"avatar"`
	Bio      *string `json:"bio"`
	Timezone *string `json:"timezone"`
	Status   *string `json:"status"`
	Language *string `json:"language"`
}

type userResponse struct {
	ID        string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	Bio       string `json:"bio"`
	Timezone  string `json:"timezone"`
	Status    string `json:"status"`
	Language  string `json:"language"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func toResponse(r *service.UserResult) userResponse {
	return userResponse{
		ID:        r.ID.String(),
		Email:     r.Email,
		Name:      r.Name,
		Avatar:    r.Avatar,
		Bio:       r.Bio,
		Timezone:  r.Timezone,
		Status:    r.Status,
		Language:  r.Language,
		CreatedAt: r.CreatedAt.Unix(),
		UpdatedAt: r.UpdatedAt.Unix(),
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// POST /api/v1/users/
func (h *Handler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result, err := h.svc.CreateUser(c.Request.Context(), id, req.Email)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toResponse(result))
}

// GET /api/v1/users/me  (requires JWT)
func (h *Handler) Me(c *gin.Context) {
	userID := c.MustGet(userIDKey).(uuid.UUID)

	result, err := h.svc.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toResponse(result))
}

// GET /api/v1/users/:id
func (h *Handler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result, err := h.svc.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toResponse(result))
}

// PATCH /api/v1/users/me  (requires JWT)
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet(userIDKey).(uuid.UUID)

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.UpdateProfile(c.Request.Context(), userID, req.Name, req.Avatar, req.Bio, req.Timezone, req.Status, req.Language)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toResponse(result))
}

// GET /api/v1/users  (requires JWT)
func (h *Handler) ListUsers(c *gin.Context) {
	limit, offset := parsePagination(c)

	results, err := h.svc.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	resp := make([]userResponse, len(results))
	for i, r := range results {
		resp[i] = toResponse(r)
	}
	c.JSON(http.StatusOK, resp)
}

// GET /api/v1/users/search  (requires JWT)
func (h *Handler) SearchUsers(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing q parameter"})
		return
	}

	limit, offset := parsePagination(c)

	results, err := h.svc.SearchUsers(c.Request.Context(), q, limit, offset)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	resp := make([]userResponse, len(results))
	for i, r := range results {
		resp[i] = toResponse(r)
	}
	c.JSON(http.StatusOK, resp)
}

// ── Error mapping ─────────────────────────────────────────────────────────────

func serviceErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrUserAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, service.ErrInvalidLanguage):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func parsePagination(c *gin.Context) (limit, offset int32) {
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = int32(v)
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v >= 0 {
		offset = int32(v)
	}
	return
}

// POST /api/v1/users/heartbeat  (requires JWT)
func (h *Handler) Heartbeat(c *gin.Context) {
	userID := c.MustGet(userIDKey).(uuid.UUID)
	h.tracker.Heartbeat(c.Request.Context(), userID)
	c.Status(http.StatusNoContent)
}

// GET /api/v1/users/presence/stream  (requires JWT)
// Sends an initial snapshot of all currently-online users, then real-time presence changes.
func (h *Handler) StreamPresence(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	sub := h.broker.Subscribe("global")
	defer h.broker.Unsubscribe("global", sub)

	// Seed the client with the current snapshot so it doesn't wait for the next state change.
	now := time.Now().Unix()
	for uid, online := range h.tracker.OnlineUsers() {
		data, _ := json.Marshal(broker.PresenceEvent{UserID: uid, Online: online, LastSeen: now})
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	}
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-sub:
			if !ok {
				return
			}
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
