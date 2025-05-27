// delivery/http/middleware.go
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"otus-highload-arh-homework/internal/social/transport/service"
)

func AuthMiddleware(jwtService *service.JWTGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := jwtService.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
			return
		}

		// Сохраняем userID в контекст Gin
		c.Set("userID", userID)
		c.Next()
	}
}
