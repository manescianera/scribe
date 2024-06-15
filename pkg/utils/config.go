package utils

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AudioDir string

	PostgresDSN    string
	PostgresDriver string
	RedisAddr      string
	RedisPass      string

	OpenAIKey      string
	OpenAIEndpoint string
}

func NewConfig() (*Config, error) {
	e := &Config{
		AudioDir:       "audio",
		OpenAIEndpoint: "https://api.openai.com/v1/audio/transcriptions",
		PostgresDriver: "pgx",
	}

	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	e.PostgresDSN = os.Getenv("POSTGRES_DSN")
	e.RedisAddr = os.Getenv("REDIS_ADDR")
	e.RedisPass = os.Getenv("REDIS_PASS")
	e.OpenAIKey = os.Getenv("OPENAI_KEY")

	return e, nil
}
