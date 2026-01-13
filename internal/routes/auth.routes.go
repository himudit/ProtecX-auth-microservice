package routes

import (
	"authService/config"
	"authService/internal/controllers"
	ratelimiter "authService/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine, authController *controllers.AuthController) {
	limited := router.Group("/auth")
	limited.Use(ratelimiter.RateLimiter(config.RDB))
	{
		limited.POST("/login", authController.Login)
		limited.POST("/register", authController.Register)
		limited.POST("/refresh", authController.AccessRefreshToken)
	}

	//  Non-rate-limited routes (cron / internal)
	open := router.Group("/auth")
	{
		open.GET("/me", authController.Me)
		open.POST("/logout", authController.Logout)
	}
}
