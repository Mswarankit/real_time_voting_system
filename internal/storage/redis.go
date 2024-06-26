package storage

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient() *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return &RedisClient{client: rdb}
}

func (r *RedisClient) SetUser(username, password string) error {
	err := r.client.HSet(ctx, "user:"+username, "password", password).Err()
	if err != nil {
		return fmt.Errorf("failed to set user %s: %v", username, err)
	}
	return nil
}

func (r *RedisClient) GetUser(username string) (string, error) {
	password, err := r.client.HGet(ctx, "user:"+username, "password").Result()
	if err == redis.Nil {
		return "", fmt.Errorf("user %s does not exist", username)
	} else if err != nil {
		return "", fmt.Errorf("failed to get user %s: %v", username, err)
	}
	return password, nil
}
