package utils

import "github.com/redis/go-redis/v9"

type ExponentialBackoffData struct {
	FailCount   int64 `json:"tokens"`
	NextAllowed int64 `json:"nextAllowed"`
}

func ExponentialBackoff(rdb *redis.Client) (string, error) {

	key := 
}
