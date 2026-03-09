package handlers

import (
	"context"

	"authService/internal/services"
	authv1 "authService/proto/proto/iam/v1"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterUser(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	return h.authService.RegisterUser(ctx, req)
}

// func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
// 	return h.authService.Login(ctx, req)
// }

// func (h *AuthHandler) AccessRefreshToken(ctx context.Context, req *authv1.AccessRefreshTokenRequest) (*authv1.AccessRefreshTokenResponse, error) {
// 	return h.authService.AccessRefreshToken(ctx, req)
// }

// func (h *AuthHandler) Profile(ctx context.Context, req *authv1.ProfileRequest) (*authv1.ProfileResponse, error) {
// 	return h.authService.Profile(ctx, req)
// }

// func (h *AuthHandler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
// 	return h.authService.Logout(ctx, req)
// }
