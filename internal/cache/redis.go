package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisCache struct {
	client *redis.Client
}

func NewRedis() *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	return &RedisCache{client: rdb}
}

func (r *RedisCache) Get(key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Set(key, value string) error {
	return r.client.Set(ctx, key, value, 0).Err()
}
