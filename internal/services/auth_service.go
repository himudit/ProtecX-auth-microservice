package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"authService/internal/domain"
	"authService/internal/models"
	"authService/internal/repositories"
	"authService/internal/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	projectUserRepo repositories.ProjectUserRepository
}

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

func NewAuthService(repo repositories.ProjectUserRepository) *AuthService {
	return &AuthService{projectUserRepo: repo}
}

func (s *AuthService) RegisterUser(
	ctx context.Context,
	req RegisterRequest,
	projectID string,
	providerID string,
) (*domain.ProjectUser, error) {

	// 1️⃣ Check if email already exists
	exists, err := s.projectUserRepo.ExistsByEmail(ctx, projectID, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists in this project")
	}

	// 2️⃣ Hash password using Argon2id from utils/password.go
	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 3️⃣ Create user object
	user := &domain.ProjectUser{
		ID:           uuid.NewString(), // unique ID for ProjectUser
		ProjectID:    projectID,        // tenant isolation
		ProviderID:   providerID,       // who created/provided this user
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPwd,
		Role:         domain.ProjectRole(req.Role),
		TokenVersion: 0,     // initial token version
		IsVerified:   false, // default
		CreatedAt:    time.Now(),
	}

	// 4️⃣ Persist to PostgreSQL
	if err := s.projectUserRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 5️⃣ Generate JWT tokens (access + refresh)
	// accessToken, err := utils.GenerateAccessToken(user.ID.Hex(), user.Email, user.Role, user.TokenVersion)
	// if err != nil {
	// 	return nil, nil, err
	// }

	// refreshToken, err := utils.GenerateRefreshToken(user.ID.Hex(), user.TokenVersion)
	// if err != nil {
	// 	return nil, nil, err
	// }

	// tokens := map[string]string{
	// 	"accessToken":  accessToken,
	// 	"refreshToken": refreshToken,
	// }

	return user, nil
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
