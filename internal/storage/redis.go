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

func (r *RedisClient) SetToken(username, token string) error {
	err := r.client.Set(ctx, "token:"+username, token, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set token for user %s: %v", username, err)
	}
	return nil
}

func (r *RedisClient) GetToken(username string) (string, error) {
	token, err := r.client.Get(ctx, "token:"+username).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token for user %s does not exist", username)
	} else if err != nil {
		return "", fmt.Errorf("failed to get token for user %s: %v", username, err)
	}
	return token, nil
}

func (r *RedisClient) SetVote(username string) error {
	err := r.client.Set(ctx, "vote:"+username, "voted", 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set vote for user %s: %v", username, err)
	}
	return nil
}

func (r *RedisClient) HasVoted(username string) (bool, error) {
	vote, err := r.client.Get(ctx, "vote:"+username).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check vote for user %s: %v", username, err)
	}
	return vote == "voted", nil
}

func (r *RedisClient) IncrementVoteCount() (int64, error) {
	return r.client.Incr(ctx, "vote_count").Result()
}

func (r *RedisClient) GetVoteCount() (int64, error) {
	return r.client.Get(ctx, "vote_count").Int64()
}
