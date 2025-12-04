package routes

import (
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine, authController *controllers.authController) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		// auth.POST("/login", authController.Login)
		// auth.POST("/refresh", authController.RefreshToken)
		// auth.POST("/logout", authController.Logout)
		// auth.GET("/me", authController.Me)
	}
}
