package middlewares

import (
	"net/http"
	"strings"

	"authService/internal/repositories"
	"authService/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtKeyRepo repositories.ProjectJwtKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.GetString(ContextProjectID)
		if projectID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "project identity missing"})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]

		// Get active key for project to verify token
		keyRow, err := jwtKeyRepo.GetActiveKeyByProjectID(c.Request.Context(), projectID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "failed to get project keys"})
			return
		}

		publicKey, err := utils.ParseRSAPublicKeyFromPEM(keyRow.PublicKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid public key"})
			return
		}

		claims, err := utils.VerifyAccessToken(tokenString, publicKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Store user info in context
		c.Set("userId", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("tokenVersion", claims.TokenVersion)

		c.Next()
	}
}
