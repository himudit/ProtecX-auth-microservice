package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"authService/internal/domain"
	"authService/internal/repositories"
	"authService/internal/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	projectUserRepo repositories.ProjectUserRepository
	jwtKeyRepo      repositories.ProjectJwtKeyRepository
}

type RegisterRequest struct {
	Name     string             `json:"name"`
	Email    string             `json:"email"`
	Password string             `json:"password"`
	Role     domain.ProjectRole `json:"role"` // optional
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthService(
	userRepo repositories.ProjectUserRepository,
	jwtKeyRepo repositories.ProjectJwtKeyRepository,
) *AuthService {
	return &AuthService{
		projectUserRepo: userRepo,
		jwtKeyRepo:      jwtKeyRepo,
	}
}

func (s *AuthService) RegisterUser(
	ctx context.Context,
	req RegisterRequest,
	projectID string,
	providerID string,
) (*domain.ProjectUser, map[string]string, error) {

	// 1️⃣ Check if email already exists
	exists, err := s.projectUserRepo.ExistsByEmail(ctx, projectID, req.Email)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, errors.New("email already exists in this project")
	}

	// 2️⃣ Hash password using Argon2id from utils/password.go
	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	// Default role logic
	if req.Role == "" {
		req.Role = domain.RoleMember
	}

	// 3️⃣ Create user object
	user := &domain.ProjectUser{
		ID:           uuid.NewString(), // unique ID for ProjectUser
		ProjectID:    projectID,        // tenant isolation
		ProviderID:   providerID,       // who created/provided this user
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPwd,
		Role:         req.Role,
		TokenVersion: 0,     // initial token version
		IsVerified:   false, // default
		CreatedAt:    time.Now(),
	}

	// 4️⃣ Persist to PostgreSQL
	if err := s.projectUserRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	keyRow, err := s.jwtKeyRepo.GetActiveKeyByProjectID(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}
	privateKeyPEM, err := utils.DecryptAES256GCM(keyRow.PrivateKeyEncrypted)
	if err != nil {
		return nil, nil, err
	}
	// 5️⃣ Generate JWT tokens (access + refresh)
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, string(user.Role), user.TokenVersion, privateKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.TokenVersion, privateKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	tokens := map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}

	return user, tokens, nil
}

func (s *AuthService) LoginUser(ctx context.Context, req LoginRequest,
	projectID string,
	providerID string, rdb *redis.Client) (*domain.ProjectUser, map[string]string, error) {

	status, remainingTime, err := utils.CheckBackoff(projectID, req.Email, rdb)

	if err != nil {
		return nil, nil, err
	}

	if status == "blocked" {
		return nil, nil, fmt.Errorf("too many login attempts, try again in %s", remainingTime)
	}

	user, err := s.projectUserRepo.GetUserByEmail(ctx, projectID, req.Email)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		_ = utils.UpdateBackoff(projectID, req.Email, rdb)
		return nil, nil, fmt.Errorf("invalid email or password")
	}

	valid, err := utils.VerifyPassword(user.PasswordHash, req.Password)
	if err != nil {
		return nil, nil, err
	}

	if !valid {
		utils.UpdateBackoff(projectID, req.Email, rdb)
		return nil, nil, fmt.Errorf("invalid email or password")
	}

	utils.ResetBackoff(projectID, req.Email, rdb)

	keyRow, err := s.jwtKeyRepo.GetActiveKeyByProjectID(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}
	privateKeyPEM, err := utils.DecryptAES256GCM(keyRow.PrivateKeyEncrypted)
	if err != nil {
		return nil, nil, err
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, string(user.Role), user.TokenVersion, privateKeyPEM)

	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.TokenVersion, privateKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	tokens := map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}

	return user, tokens, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, projectID string, refreshToken string) (map[string]string, error) {
	// 1. Get active key for project
	keyRow, err := s.jwtKeyRepo.GetActiveKeyByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 2. Parse public key
	publicKey, err := utils.ParseRSAPublicKeyFromPEM(keyRow.PublicKey)
	if err != nil {
		return nil, err
	}

	// 3. Verify refresh token
	claims, err := utils.VerifyRefreshToken(refreshToken, publicKey)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 4. Get user
	user, err := s.projectUserRepo.GetUserByID(ctx, projectID, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 5. Check token version
	if claims.TokenVersion != user.TokenVersion {
		return nil, errors.New("refresh token expired or revoked")
	}

	// 6. Increment token version in DB
	err = s.projectUserRepo.IncrementTokenVersion(ctx, projectID, user.ID)
	if err != nil {
		return nil, err
	}
	newTokenVersion := user.TokenVersion + 1

	// 7. Decrypt private key for signing new tokens
	privateKeyPEM, err := utils.DecryptAES256GCM(keyRow.PrivateKeyEncrypted)
	if err != nil {
		return nil, err
	}

	// 8. Generate new tokens
	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Email, string(user.Role), newTokenVersion, privateKeyPEM)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID, newTokenVersion, privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"accessToken":  newAccessToken,
		"refreshToken": newRefreshToken,
	}, nil
}

func (s *AuthService) LogoutUser(ctx context.Context, projectID string, accessToken string) error {
	// 1. Get active key for project
	keyRow, err := s.jwtKeyRepo.GetActiveKeyByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	// 2. Parse public key
	publicKey, err := utils.ParseRSAPublicKeyFromPEM(keyRow.PublicKey)
	if err != nil {
		return err
	}

	// 3. Verify access token
	claims, err := utils.VerifyAccessToken(accessToken, publicKey)
	if err != nil {
		return errors.New("invalid access token")
	}

	// 4. Increment token version to invalidate all current tokens
	return s.projectUserRepo.IncrementTokenVersion(ctx, projectID, claims.UserID)
}
