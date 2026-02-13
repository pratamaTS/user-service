package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"harjonan.id/user-service/app/helpers"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "missing Authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "invalid Authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		secret := helpers.ProvideJWTSecret()
		if secret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": "server misconfigured: missing JWT_SECRET",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "invalid token claims"})
			c.Abort()
			return
		}

		log.Print("claims: ", claims)

		sessionID, _ := claims["sessionId"].(string)
		subject, _ := claims["subject"].(string)
		tokenType, _ := claims["type"].(string)

		if sessionID == "" || subject == "" || tokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "invalid access token"})
			c.Abort()
			return
		}

		c.Set("session_id", sessionID)
		c.Set("subject", subject)
		c.Set("access_token", strings.TrimSpace(tokenString))

		c.Next()
	}
}
