package api

import (
	"channels/internal/repository"
	"channels/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const userIDKey = "userID"

type Handler struct {
	svc *service.ChannelsService
}

// ── JWT middleware ─────────────────────────────────────────────────────────────

func jwtMiddleware(publicKeyPEM []byte) gin.HandlerFunc {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		panic("channels: failed to parse JWT public key: " + err.Error())
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

type createChannelRequest struct {
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

type addMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role"`
}

type channelResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
	CreatedBy   string `json:"created_by"`
	Archived    bool   `json:"archived"`
	MemberCount int32  `json:"member_count"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type memberResponse struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	JoinedAt  int64  `json:"joined_at"`
}

func toChannelResponse(ch *repository.ChannelRow) channelResponse {
	return channelResponse{
		ID:          ch.ID.String(),
		Name:        ch.Name,
		Description: ch.Description,
		IsPrivate:   ch.IsPrivate,
		CreatedBy:   ch.CreatedBy.String(),
		Archived:    ch.Archived,
		MemberCount: ch.MemberCount,
		CreatedAt:   ch.CreatedAt,
		UpdatedAt:   ch.UpdatedAt,
	}
}

func toMemberResponse(m *repository.MemberRow) memberResponse {
	return memberResponse{
		ChannelID: m.ChannelID.String(),
		UserID:    m.UserID.String(),
		Role:      m.Role,
		JoinedAt:  m.JoinedAt,
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// POST /api/v1/channels
func (h *Handler) CreateChannel(c *gin.Context) {
	userID := c.MustGet(userIDKey).(uuid.UUID)

	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.svc.CreateChannel(c.Request.Context(), req.Name, req.Description, req.IsPrivate, userID)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toChannelResponse(ch))
}

// GET /api/v1/channels/me  — must be registered BEFORE /:id
func (h *Handler) ListMyChannels(c *gin.Context) {
	userID := c.MustGet(userIDKey).(uuid.UUID)

	channels, err := h.svc.ListUserChannels(c.Request.Context(), userID)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	resp := make([]channelResponse, len(channels))
	for i, ch := range channels {
		resp[i] = toChannelResponse(ch)
	}
	c.JSON(http.StatusOK, resp)
}

// GET /api/v1/channels/:id
func (h *Handler) GetChannel(c *gin.Context) {
	userID := c.MustGet(userIDKey).(uuid.UUID)

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	ch, err := h.svc.GetChannel(c.Request.Context(), channelID, userID)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toChannelResponse(ch))
}

// DELETE /api/v1/channels/:id
func (h *Handler) DeleteChannel(c *gin.Context) {
	userID := c.MustGet(userIDKey).(uuid.UUID)

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	if err := h.svc.DeleteChannel(c.Request.Context(), channelID, userID); err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// POST /api/v1/channels/:id/members
func (h *Handler) AddMember(c *gin.Context) {
	userID := c.MustGet(userIDKey).(uuid.UUID)

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	var req addMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	member, err := h.svc.AddMember(c.Request.Context(), channelID, targetUserID, role, userID)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toMemberResponse(member))
}

// DELETE /api/v1/channels/:id/members/:userId
func (h *Handler) RemoveMember(c *gin.Context) {
	requestingUserID := c.MustGet(userIDKey).(uuid.UUID)

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.svc.RemoveMember(c.Request.Context(), channelID, targetUserID, requestingUserID); err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GET /api/v1/channels/:id/members
func (h *Handler) ListMembers(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	members, err := h.svc.ListChannelMembers(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(serviceErrorStatus(err), gin.H{"error": err.Error()})
		return
	}

	resp := make([]memberResponse, len(members))
	for i, m := range members {
		resp[i] = toMemberResponse(m)
	}
	c.JSON(http.StatusOK, resp)
}

// ── Error mapping ─────────────────────────────────────────────────────────────

func serviceErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrChannelNotFound), errors.Is(err, service.ErrMemberNotFound), errors.Is(err, service.ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrChannelAlreadyExists), errors.Is(err, service.ErrMemberAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, service.ErrPermissionDenied):
		return http.StatusForbidden
	case errors.Is(err, service.ErrChannelArchived):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
