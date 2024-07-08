package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
	ctx         context.Context
)

const (
	BlacklistedTokensSet = "blacklisted_tokens"
)

func Init() {
	ctx = context.Background()
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis %v", err)
	}
}

func GetRedisClient() (*redis.Client, context.Context) {
	return redisClient, ctx
}

func BlackListToken(token string, expiry time.Time) error {
	pipe := redisClient.Pipeline()
	pipe.SAdd(ctx, BlacklistedTokensSet, token)
	pipe.ExpireAt(ctx, BlacklistedTokensSet, expiry)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

func IsTokenBlacklisted(token string) (bool, error) {
	exists, err := redisClient.SIsMember(ctx, BlacklistedTokensSet, token).Result()
	if err != nil {
		return true, err
	}
	if exists {
		return true, nil
	}

	return false, nil
}
