package database

import (
	"context"
	"log"
	"time"

	"github.com/chilljzz/gohub/pkg/config"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis(cfg config.RedisConfig) error {

	client := redis.NewClient(
		&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		},
	)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}

	RedisClient = client

	log.Printf(
		"redis connected: addr=%s db=%d",
		cfg.Addr,
		cfg.DB,
	)
	return nil
}
