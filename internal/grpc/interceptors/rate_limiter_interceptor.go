package interceptors

import (
	"context"
	"strconv"
	"time"

	"authService/internal/utils"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type RateLimiterData struct {
	Tokens       float64
	LastRefillTs int64
}

func RateLimiterInterceptor(rdb *redis.Client) grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		skipMethods := map[string]bool{
			"/auth.v1.AuthService/Profile": true,
			"/auth.v1.AuthService/Logout":  true,
		}

		if skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "metadata missing")
		}

		projectIDs := md.Get("x-project-id")
		if len(projectIDs) == 0 {
			return nil, status.Error(codes.InvalidArgument, "x-project-id missing")
		}

		projectID := projectIDs[0]

		ip := utils.GetIPFromMetadata(md)
		if ip == "" {
			return nil, status.Error(codes.InvalidArgument, "unable to identify IP")
		}

		key := "rate_limit:" + projectID + ":" + ip

		val, err := rdb.HGetAll(ctx, key).Result()

		var data RateLimiterData

		if err != nil || len(val) == 0 {
			data = RateLimiterData{
				Tokens:       10,
				LastRefillTs: time.Now().Unix(),
			}
		} else {
			data.Tokens, _ = strconv.ParseFloat(val["tokens"], 64)
			data.LastRefillTs, _ = strconv.ParseInt(val["last_refill_ts"], 10, 64)
		}

		currentTime := time.Now().Unix()

		newTokens := float64(currentTime-data.LastRefillTs) / 6.0
		if newTokens > 0 {
			data.Tokens += newTokens
			if data.Tokens > 10 {
				data.Tokens = 10
			}
		}

		if data.Tokens < 1 {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		data.Tokens -= 1
		data.LastRefillTs = currentTime

		rdb.HSet(ctx, key, map[string]interface{}{
			"tokens":         data.Tokens,
			"last_refill_ts": data.LastRefillTs,
		})

		rdb.Expire(ctx, key, 2*time.Minute)

		return handler(ctx, req)
	}
}
