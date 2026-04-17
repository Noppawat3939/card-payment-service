package redis

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Locker interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (string, error)
	Release(ctx context.Context, key, value string) error
}

type redisLock struct {
	client *redis.Client
}

func NewRedisLocker(client *redis.Client) Locker {
	return &redisLock{client}
}

func (r *redisLock) Acquire(ctx context.Context, key string, ttl time.Duration) (string, error) {
	value := uuid.NewString()
	res, err := r.client.SetArgs(ctx, key, value, redis.SetArgs{
		Mode: "NX",
		TTL:  ttl,
	}).Result()

	if err != nil {
		return "", err
	}

	if res != "OK" {
		// cannot lock
		return "", nil
	}

	return value, nil
}

func (r *redisLock) Release(ctx context.Context, key, value string) error {
	script := redis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`)

	_, err := script.Run(ctx, r.client, []string{key}, value).Result()
	if err != nil {
		return err
	}

	return nil
}
