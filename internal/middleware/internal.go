package middleware

import (
	"store/internal/config"
	"net/http"
	"github.com/gin-gonic/gin"
)

func InternalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" || apiKey != config.ApplicationConfig.InternalAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized internal access"})
			c.Abort()
			return
		}
		c.Next()
	}
}

