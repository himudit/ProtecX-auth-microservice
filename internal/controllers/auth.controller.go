package controllers

import (
	"net/http"

	"authService/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type AuthController struct {
	redisClient *redis.Client
}

func NewAuthController(rdb *redis.Client) *AuthController {
	return &AuthController{
		redisClient: rdb,
	}
}

// Register request payload
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
}

func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest

	// Bind incoming JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service layer
	user, tokens, err := services.RegisterUser(services.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return user info and JWT tokens
	c.JSON(http.StatusOK, gin.H{
		"message":      "Signed up successfully",
		"user":         user,
		"accessToken":  tokens["accessToken"],
		"refreshToken": tokens["refreshToken"],
	})
}

func (ac *AuthController) Login(c *gin.Context) {
	var req RegisterRequest

	// Bind incoming JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service layer
	user, tokens, err := services.LoginUser(services.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}, ac.redisClient)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return user info and JWT tokens
	c.JSON(http.StatusOK, gin.H{
		"message":      "Loged in successfully",
		"user":         user,
		"accessToken":  tokens["accessToken"],
		"refreshToken": tokens["refreshToken"],
	})
}
