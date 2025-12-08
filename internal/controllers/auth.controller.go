package controllers

import "github.com/gin-gonic/gin"

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (ac *AuthController) Register(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Signed up successfully",
	})
}
