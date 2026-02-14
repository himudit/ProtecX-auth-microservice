package routes

import (
	"authService/config"
	"authService/internal/controllers"
	middlewares "authService/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine, authController *controllers.AuthController) {
	limited := router.Group("/iam")
	limited.Use(middlewares.ProjectContext(), middlewares.RateLimiter(config.RDB))
	{
		limited.POST("/login", authController.Login)
		limited.POST("/register", authController.Register)
		limited.POST("/refresh", authController.AccessRefreshToken)
	}

	//  Non-rate-limited routes (cron / internal)
	open := router.Group("/iam")
	open.Use(middlewares.ProjectContext())
	{
		open.GET("/me", authController.Me)
		open.POST("/logout", authController.Logout)
	}
}
