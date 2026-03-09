package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// context keys
type contextKey string

const (
	ContextProjectID  contextKey = "projectId"
	ContextProviderID contextKey = "providerId"
)

// ProjectContextInterceptor extracts project/provider IDs from metadata
func ProjectContextInterceptor() grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "metadata missing")
		}

		projectIDs := md.Get("x-project-id")
		if len(projectIDs) == 0 {
			return nil, status.Error(codes.InvalidArgument, "x-project-id missing")
		}

		providerIDs := md.Get("x-provider-id")
		if len(providerIDs) == 0 {
			return nil, status.Error(codes.InvalidArgument, "x-provider-id missing")
		}

		projectID := projectIDs[0]
		providerID := providerIDs[0]

		// basic sanity validation
		if len(projectID) < 10 || len(providerID) < 10 {
			return nil, status.Error(codes.InvalidArgument, "invalid projectId or providerId")
		}

		// store in context
		ctx = context.WithValue(ctx, ContextProjectID, projectID)
		ctx = context.WithValue(ctx, ContextProviderID, providerID)

		return handler(ctx, req)
	}
}
