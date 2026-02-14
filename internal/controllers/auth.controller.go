package controllers

import (
	"net/http"

	"authService/internal/domain"
	"authService/internal/middlewares"
	"authService/internal/services"
	"authService/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type AuthController struct {
	redisClient *redis.Client
	authService *services.AuthService
}

func NewAuthController(
	rdb *redis.Client,
	authService *services.AuthService,
) *AuthController {
	return &AuthController{
		redisClient: rdb,
		authService: authService,
	}
}

// Register request payload
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type LogoutRequest struct {
	AccessToken string `json:"accessToken"`
}

func (ac *AuthController) Register(c *gin.Context) {
	ctx := c.Request.Context()
	projectID := c.GetString(middlewares.ContextProjectID)
	providerID := c.GetString(middlewares.ContextProviderID)

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		if errs := utils.ValidationErrors(err); errs != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errs,
			})
			return
		}

		// non-validation error (bad JSON, wrong types, etc.)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	user, tokens, err := ac.authService.RegisterUser(ctx, services.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     domain.ProjectRole(req.Role),
	}, projectID, providerID)

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
	ctx := c.Request.Context()
	projectID := c.GetString(middlewares.ContextProjectID)
	providerID := c.GetString(middlewares.ContextProviderID)

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		if errs := utils.ValidationErrors(err); errs != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errs,
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	user, tokens, err := ac.authService.LoginUser(ctx, services.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}, projectID, providerID, ac.redisClient)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Loged in successfully",
		"user":         user,
		"accessToken":  tokens["accessToken"],
		"refreshToken": tokens["refreshToken"],
	})
}

func (ac *AuthController) AccessRefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	projectID := c.GetString(middlewares.ContextProjectID)

	var payload RefreshRequest
	if err := c.ShouldBindJSON(&payload); err != nil || payload.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refreshToken required in body"})
		return
	}

	tokens, err := ac.authService.RefreshToken(ctx, projectID, payload.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  tokens["accessToken"],
		"refreshToken": tokens["refreshToken"],
	})
}

func (ac *AuthController) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	projectID := c.GetString(middlewares.ContextProjectID)

	var payload LogoutRequest
	if err := c.ShouldBindJSON(&payload); err != nil || payload.AccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accessToken required in body"})
		return
	}

	err := ac.authService.LogoutUser(ctx, projectID, payload.AccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

func (ac *AuthController) Me(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "hello from 8080",
	})
}
