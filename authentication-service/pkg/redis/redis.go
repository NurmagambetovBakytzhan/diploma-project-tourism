package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(context.Background(), key, value, expiration).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(context.Background(), key).Result()
}

func (r *RedisClient) Delete(key string) error {
	return r.client.Del(context.Background(), key).Err()
}
