package database

import (
	"card-payment-service/internal/config"
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.GetRedisAddr(),
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if e := client.Ping(ctx).Err(); e != nil {
		return nil, e
	}

	log.Println("connected to redis")
	return client, nil
}
