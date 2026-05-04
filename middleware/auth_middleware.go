package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"be-anis/helper"
	"be-anis/service"
)

const (
	ContextAccessToken = "access_token"
	ContextUser        = "auth_user"
)

func AuthRequired(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			helper.Err(c, http.StatusUnauthorized, "missing Authorization header", nil)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			helper.Err(c, http.StatusUnauthorized, "invalid Authorization header format", nil)
			return
		}

		token := parts[1]
		user, err := authService.GetCurrentUser(token)
		if err != nil {
			helper.Err(c, http.StatusUnauthorized, "invalid or expired token", err)
			return
		}

		c.Set(ContextAccessToken, token)
		c.Set(ContextUser, user)
		c.Next()
	}
}

func TokenFromContext(c *gin.Context) string {
	if v, ok := c.Get(ContextAccessToken); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
