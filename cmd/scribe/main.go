package main

import (
	"context"
	"log"

	"github.com/manescianera/scribe/pkg/api"
	"github.com/manescianera/scribe/pkg/db"
	"github.com/manescianera/scribe/pkg/redis"
	"github.com/manescianera/scribe/pkg/transcribe"
	"github.com/manescianera/scribe/pkg/utils"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()

	cfg, err := utils.NewConfig()
	if err != nil {
		log.Fatal("Error loading configuration:", err)
	}

	redisParams := redis.RedisParams{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       0,
	}

	redisClient, err := redis.New(ctx, redisParams)
	if err != nil {
		log.Fatal("Error initializing Redis:", err)
	}

	dbParams := db.ConnParams{
		Driver:  cfg.PostgresDriver,
		ConnStr: cfg.PostgresDSN,
	}

	dbClient, err := db.New(ctx, dbParams)
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}

	go func() {
		apiHandler := api.NewHandler(dbClient)
		apiHandler.StartAPI()
	}()

	transcriber := transcribe.NewTranscriber(dbClient, redisClient, cfg.OpenAIKey)
	transcriber.Watch(ctx, cfg.AudioDir)
}
