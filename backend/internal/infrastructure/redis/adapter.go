package redis

import (
	"context"
	"time"
	"github.com/redis/go-redis/v9"
)

type Adapter struct {
	client *redis.Client
}

func NewAdapter(client *redis.Client) *Adapter {
	return &Adapter{client: client}
}

func (a *Adapter) Get(ctx context.Context, key string) (string, error) {
	return a.client.Get(ctx, key).Result()
}

func (a *Adapter) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return a.client.Set(ctx, key, value, expiration).Err()
}

func (a *Adapter) Delete(ctx context.Context, key string) error {
	return a.client.Del(ctx, key).Err()
}

func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	result, err := a.client.Exists(ctx, key).Result()
	return result > 0, err
}
