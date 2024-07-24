package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisRepository struct {
	client *redis.Client
}

type IRedisRepository interface {
	SetToken(string, string, time.Duration) error
	GetToken(string) (string, error)
	DeleteToken(string)
	Close()
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func (c *RedisRepository) SetToken(token, email string, expiration time.Duration) error {
	ctx := context.Background()
	return c.client.Set(ctx, token, email, expiration).Err()
}

func (c *RedisRepository) GetToken(token string) (string, error) {
	ctx := context.Background()
	return c.client.Get(ctx, token).Result()
}

func (c *RedisRepository) DeleteToken(token string) {
	ctx := context.Background()
	c.client.Del(ctx, token)
}

func (c *RedisRepository) Close() {
	c.client.Close()
}
