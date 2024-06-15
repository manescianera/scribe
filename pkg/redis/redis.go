package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisParams struct {
	Addr     string
	Password string
	DB       int
}

func New(ctx context.Context, params RedisParams) (*redis.Client, error) {
	for attempt := 1; attempt <= 5; attempt++ {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     params.Addr,
			Password: params.Password,
			DB:       params.DB,
		})

		pong, err := redisClient.Ping(ctx).Result()
		if err != nil {
			log.Printf("Attempt %d: Failed to connect to Redis: %v\n", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		log.Printf("Connected to Redis: %s\n", pong)
		return redisClient, nil
	}

	return nil, fmt.Errorf("failed to connect to Redis after multiple attempts")
}
