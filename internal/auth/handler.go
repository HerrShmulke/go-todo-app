package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req AuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	user, err := h.service.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ToRegisterResponse(user))
}

func (h *Handler) Login(c *gin.Context) {
	var req AuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	tokenPair, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.setRefreshCookie(c, tokenPair.RefreshToken)

	c.JSON(http.StatusOK, gin.H{"accessToken": tokenPair.AccessToken})
}

func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	tokens, err := h.service.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	h.setRefreshCookie(c, tokens.RefreshToken)

	c.JSON(http.StatusOK, gin.H{"accessToken": tokens.AccessToken})
}

func (h *Handler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil && refreshToken != "" {
		if err := h.service.Logout(c.Request.Context(), refreshToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to logout",
			})
			return
		}
	}

	c.SetSameSite(http.SameSiteLaxMode)

	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/api/auth",
		"",
		true,
		true,
	)

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidPassword):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrInvalidEmail):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})

	default:
		slog.Error("handler error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}

func (h *Handler) setRefreshCookie(c *gin.Context, t string) {
	c.SetSameSite(http.SameSiteLaxMode)

	c.SetCookie(
		"refresh_token",
		t,
		int(refreshTokenTTL.Seconds()),
		"/api/v1/auth",
		"",
		true,
		true,
	)
}
