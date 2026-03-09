package services

import (
	"context"
	"errors"
	"log"
	"time"

	"authService/internal/domain"
	"authService/internal/grpc/interceptors"
	"authService/internal/repositories"
	"authService/internal/utils"
	authv1 "authService/proto/proto/iam/v1"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthService struct {
	projectUserRepo repositories.ProjectUserRepository
	jwtKeyRepo      repositories.ProjectJwtKeyRepository
	rdb             *redis.Client
}

func NewAuthService(
	userRepo repositories.ProjectUserRepository,
	jwtKeyRepo repositories.ProjectJwtKeyRepository,
	rdb *redis.Client,
) *AuthService {
	return &AuthService{
		projectUserRepo: userRepo,
		jwtKeyRepo:      jwtKeyRepo,
		rdb:             rdb,
	}
}

func (s *AuthService) RegisterUser(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {

	projectID := ctx.Value(interceptors.ContextProjectID).(string)
	providerID := ctx.Value(interceptors.ContextProviderID).(string)

	exists, err := s.projectUserRepo.ExistsByEmail(ctx, projectID, req.Email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("email already exists in this project")
	}

	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	user := &domain.ProjectUser{
		ID:           uuid.NewString(),
		ProjectID:    projectID,
		ProviderID:   providerID,
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPwd,
		Role:         domain.RoleMember,
		TokenVersion: 0,
		IsVerified:   false,
		CreatedAt:    now,
		LastLoginAt:  &now,
	}

	if err := s.projectUserRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	keyRow, err := s.jwtKeyRepo.GetActiveKeyByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	privateKeyPEM, err := utils.DecryptAES256GCM(keyRow.PrivateKeyEncrypted)
	if err != nil {
		return nil, err
	}

	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Email,
		string(user.Role),
		user.TokenVersion,
		privateKeyPEM,
	)

	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		user.TokenVersion,
		privateKeyPEM,
	)

	if err != nil {
		return nil, err
	}

	var lastLogin *timestamppb.Timestamp
	if user.LastLoginAt != nil {
		lastLogin = timestamppb.New(*user.LastLoginAt)
	}

	go func() {
		ctx := context.Background()
		err := s.projectUserRepo.UpdateLastLoginAt(ctx, projectID, user.ID)
		if err != nil {
			log.Printf("failed to update last login: %v", err)
		}
	}()

	return &authv1.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &authv1.User{
			Id:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			Role:        mapDomainRoleToProto(user.Role),
			IsVerified:  user.IsVerified,
			LastLoginAt: lastLogin,
		},
	}, nil
}

func mapDomainRoleToProto(role domain.ProjectRole) authv1.ProjectRole {
	switch role {
	case domain.RoleOwner:
		return authv1.ProjectRole_OWNER
	case domain.RoleAdmin:
		return authv1.ProjectRole_ADMIN
	case domain.RoleMember:
		return authv1.ProjectRole_MEMBER
	default:
		return authv1.ProjectRole_MEMBER
	}
}

// func (s *AuthService) Login(
// 	ctx context.Context,
// 	req *authv1.LoginRequest,
// ) (*authv1.LoginResponse, error) {

// 	projectID := ctx.Value(interceptors.ContextProjectID).(string)

// 	status, remainingTime, err := utils.CheckBackoff(projectID, req.Email, s.rdb)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if status == "blocked" {
// 		return nil, fmt.Errorf("too many login attempts, try again in %s", remainingTime)
// 	}

// 	user, err := s.projectUserRepo.GetUserByEmail(ctx, projectID, req.Email)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if user == nil {
// 		utils.UpdateBackoff(projectID, req.Email, s.rdb)
// 		return nil, errors.New("invalid email or password")
// 	}

// 	valid, err := utils.VerifyPassword(user.PasswordHash, req.Password)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if !valid {
// 		utils.UpdateBackoff(projectID, req.Email, s.rdb)
// 		return nil, errors.New("invalid email or password")
// 	}

// 	utils.ResetBackoff(projectID, req.Email, s.rdb)

// 	keyRow, err := s.jwtKeyRepo.GetActiveKeyByProjectID(ctx, projectID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	privateKeyPEM, err := utils.DecryptAES256GCM(keyRow.PrivateKeyEncrypted)
// 	if err != nil {
// 		return nil, err
// 	}

// 	accessToken, err := utils.GenerateAccessToken(
// 		user.ID,
// 		user.Email,
// 		string(user.Role),
// 		user.TokenVersion,
// 		privateKeyPEM,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	refreshToken, err := utils.GenerateRefreshToken(
// 		user.ID,
// 		user.TokenVersion,
// 		privateKeyPEM,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	go func() {
// 		ctx := context.Background()
// 		err := s.projectUserRepo.UpdateLastLoginAt(ctx, projectID, user.ID)
// 		if err != nil {
// 			log.Printf("failed to update last login: %v", err)
// 		}
// 	}()

// 	return &authv1.LoginResponse{
// 		AccessToken:  accessToken,
// 		RefreshToken: refreshToken,
// 	}, nil
// }

// func (s *AuthService) AccessRefreshToken(
// 	ctx context.Context,
// 	req *authv1.AccessRefreshTokenRequest,
// ) (*authv1.AccessRefreshTokenResponse, error) {

// 	projectID := ctx.Value(interceptors.ContextProjectID).(string)

// 	keyRow, err := s.jwtKeyRepo.GetActiveKeyByProjectID(ctx, projectID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	publicKey, err := utils.ParseRSAPublicKeyFromPEM(keyRow.PublicKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	claims, err := utils.VerifyRefreshToken(req.RefreshToken, publicKey)
// 	if err != nil {
// 		return nil, errors.New("invalid refresh token")
// 	}

// 	user, err := s.projectUserRepo.GetUserByID(ctx, projectID, claims.UserID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if user == nil {
// 		return nil, errors.New("user not found")
// 	}

// 	if claims.TokenVersion != user.TokenVersion {
// 		return nil, errors.New("refresh token expired or revoked")
// 	}

// 	err = s.projectUserRepo.IncrementTokenVersion(ctx, projectID, user.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	newTokenVersion := user.TokenVersion + 1

// 	privateKeyPEM, err := utils.DecryptAES256GCM(keyRow.PrivateKeyEncrypted)
// 	if err != nil {
// 		return nil, err
// 	}

// 	newAccessToken, err := utils.GenerateAccessToken(
// 		user.ID,
// 		user.Email,
// 		string(user.Role),
// 		newTokenVersion,
// 		privateKeyPEM,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	newRefreshToken, err := utils.GenerateRefreshToken(
// 		user.ID,
// 		newTokenVersion,
// 		privateKeyPEM,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return &authv1.AccessRefreshTokenResponse{
// 		AccessToken:  newAccessToken,
// 		RefreshToken: newRefreshToken,
// 	}, nil
// }

// func (s *AuthService) Profile(
// 	ctx context.Context,
// 	req *authv1.ProfileRequest,
// ) (*authv1.ProfileResponse, error) {

// 	projectID := ctx.Value(interceptors.ContextProjectID).(string)
// 	userID := ctx.Value(interceptors.ContextUserID).(string)
// 	tokenVersion := ctx.Value(interceptors.ContextTokenVersion).(int)

// 	user, err := s.projectUserRepo.GetUserByID(ctx, projectID, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if user == nil {
// 		return nil, errors.New("user not found")
// 	}

// 	if user.TokenVersion != tokenVersion {
// 		return nil, errors.New("token expired or revoked")
// 	}

// 	return &authv1.ProfileResponse{
// 		Id:    user.ID,
// 		Name:  user.Name,
// 		Email: user.Email,
// 		Role:  string(user.Role),
// 	}, nil
// }

// func (s *AuthService) Logout(
// 	ctx context.Context,
// 	req *authv1.LogoutRequest,
// ) (*authv1.LogoutResponse, error) {

// 	projectID := ctx.Value(interceptors.ContextProjectID).(string)
// 	userID := ctx.Value(interceptors.ContextUserID).(string)

// 	err := s.projectUserRepo.IncrementTokenVersion(ctx, projectID, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &authv1.LogoutResponse{
// 		Message: "logged out successfully",
// 	}, nil
// }
