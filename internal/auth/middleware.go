package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AccessTokenVerifier interface {
	Verify(rawToken string) (string, error)
}

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(verifier AccessTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			unauthorized(c)
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(c)
			return
		}

		userID, err := verifier.Verify(parts[1])
		if err != nil || userID == "" {
			unauthorized(c)
			return
		}

		ctx := context.WithValue(
			c.Request.Context(),
			userIDKey,
			userID,
		)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", "Bearer")
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": "unauthorized",
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok && userID != ""
}
