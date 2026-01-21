package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Context keys (avoid collisions)
const (
	ContextProjectID  = "projectId"
	ContextProviderID = "providerId"
)

// ProjectContext extracts and stores projectId in gin.Context
func ProjectContext() gin.HandlerFunc {
	return func(c *gin.Context) {

		projectID := c.GetHeader("x-project-id")
		if projectID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "x-project-id header missing",
			})
			return
		}

		// Extract providerId
		providerID := c.GetHeader("x-provider-id")
		if providerID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "x-provider-id header missing",
			})
			return
		}

		// Basic sanity check (UUID format)
		if len(projectID) < 10 || len(providerID) < 10 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid projectId or providerId",
			})
			return
		}

		// Store in context for downstream usage
		c.Set(ContextProjectID, projectID)
		c.Set(ContextProviderID, providerID)

		c.Next()
	}
}
