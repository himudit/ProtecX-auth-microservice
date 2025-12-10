package services

import (
	"errors"
	"fmt"
	"time"

	"authService/internal/models"
	"authService/internal/utils"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // optional
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterUser(req RegisterRequest) (*models.User, map[string]string, error) {

	// 1️⃣ Check if email already exists
	exists, err := models.IsEmailExists(req.Email)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, errors.New("email already exists")
	}

	// 2️⃣ Hash password using Argon2id from utils/password.go
	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	// 3️⃣ Create user object
	user := &models.User{
		ID:           primitive.NewObjectID(),
		Name:         req.Name,
		Email:        req.Email,
		Password:     hashedPwd,
		CreatedAt:    time.Now().Unix(),
		TokenVersion: 1, // default for new users
	}

	// 4️⃣ Insert into MongoDB
	err = models.InsertUser(user)
	if err != nil {
		return nil, nil, err
	}

	// 5️⃣ Generate JWT tokens (access + refresh)
	accessToken, err := utils.GenerateAccessToken(user.ID.Hex(), user.Email, user.Role, user.TokenVersion)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID.Hex(), user.TokenVersion)
	if err != nil {
		return nil, nil, err
	}

	tokens := map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}

	return user, tokens, nil
}

func LoginUser(req LoginRequest, rdb *redis.Client) (*models.User, map[string]string, error) {

	status, remainingTime, err := utils.CheckBackoff(req.Email, rdb)

	if err != nil {
		return nil, nil, err
	}

	if status == "blocked" {
		return nil, nil, fmt.Errorf("too many login attempts, try again in %s", remainingTime)
	}

	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid email or password")
	}

	valid, err := utils.VerifyPassword(user.Password, req.Password)
	if err != nil {
		return nil, nil, err
	}

	if !valid {
		utils.UpdateBackoff(req.Email, rdb)
		return nil, nil, fmt.Errorf("invalid email or password")
	}

	utils.ResetBackoff(req.Email, rdb)

	accessToken, err := utils.GenerateAccessToken(user.ID.Hex(), user.Email, user.Role, user.TokenVersion)

	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID.Hex(), user.TokenVersion)
	if err != nil {
		return nil, nil, err
	}

	tokens := map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}

	return user, tokens, nil
}
