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

func (h *AuthHandler) LoginUser(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return h.authService.LoginUser(ctx, req)
}

func (h *AuthHandler) RefreshSession(ctx context.Context, req *authv1.RefreshSessionRequest) (*authv1.RefreshSessionResponse, error) {
	return h.authService.RefreshSession(ctx, req)
}

func (h *AuthHandler) GetProfile(ctx context.Context, req *authv1.GetProfileRequest) (*authv1.GetProfileResponse, error) {
	return h.authService.GetProfile(ctx, req)
}

func (h *AuthHandler) LogoutUser(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	return h.authService.Logout(ctx, req)
}
