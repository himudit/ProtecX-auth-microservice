package routes

import (
	"authService/config"
	"authService/internal/controllers"
	middlewares "authService/internal/middlewares"

	"authService/internal/repositories"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine, authController *controllers.AuthController, jwtKeyRepo repositories.ProjectJwtKeyRepository) {
	limited := router.Group("/iam")
	limited.Use(middlewares.ProjectContext(), middlewares.RateLimiter(config.RDB))
	{
		limited.POST("/login", authController.Login)
		limited.POST("/register", authController.Register)
		limited.POST("/refresh", authController.AccessRefreshToken)
	}

	auth := router.Group("/iam")
	auth.Use(middlewares.ProjectContext(), middlewares.AuthMiddleware(jwtKeyRepo))
	{
		auth.GET("/profile", authController.Profile)
		auth.POST("/logout", authController.Logout)
	}

	//  Non-rate-limited routes (cron / internal)
	ping := router.Group("/ping")
	ping.GET("", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
}
