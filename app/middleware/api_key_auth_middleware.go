package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func APIKeyAuthMiddleware() gin.HandlerFunc {
	expectedApiKey := os.Getenv("API_KEY")
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		apiKey := c.GetHeader("X-API-KEY")
		if apiKey == "" || apiKey != expectedApiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   true,
				"message": "Unauthorized: Invalid API Key",
			})
			return
		}
		c.Next()
	}
}
