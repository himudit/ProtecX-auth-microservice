package interceptors

import (
	"context"
	"strings"

	"authService/internal/repositories"
	"authService/internal/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// context keys for user info
type userContextKey string

const (
	ContextUserID       userContextKey = "userId"
	ContextUserEmail    userContextKey = "email"
	ContextUserRole     userContextKey = "role"
	ContextTokenVersion userContextKey = "tokenVersion"
)

func AuthInterceptor(jwtKeyRepo repositories.ProjectJwtKeyRepository) grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		skipMethods := map[string]bool{
			"/auth.v1.AuthService/Login":              true,
			"/auth.v1.AuthService/RegisterUser":       true,
			"/auth.v1.AuthService/AccessRefreshToken": true,
		}

		if skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// projectID already set by ProjectContextInterceptor
		projectID, ok := ctx.Value(ContextProjectID).(string)
		if !ok || projectID == "" {
			return nil, status.Error(codes.Unauthenticated, "project identity missing")
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata missing")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization header missing")
		}

		authHeader := authHeaders[0]

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		tokenString := parts[1]

		// get project public key
		keyRow, err := jwtKeyRepo.GetActiveKeyByProjectID(ctx, projectID)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "failed to get project keys")
		}

		publicKey, err := utils.ParseRSAPublicKeyFromPEM(keyRow.PublicKey)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid public key")
		}

		claims, err := utils.VerifyAccessToken(tokenString, publicKey)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// attach user info to context
		ctx = context.WithValue(ctx, ContextUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextUserEmail, claims.Email)
		ctx = context.WithValue(ctx, ContextUserRole, claims.Role)
		ctx = context.WithValue(ctx, ContextTokenVersion, claims.TokenVersion)

		return handler(ctx, req)
	}
}
