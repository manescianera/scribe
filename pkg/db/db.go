package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type ConnParams struct {
	Driver  string
	ConnStr string
}

func New(ctx context.Context, params ConnParams) (*sql.DB, error) {
	for attempt := 1; attempt <= 5; attempt++ {
		db, err := sql.Open(params.Driver, params.ConnStr)
		if err != nil {
			log.Printf("Attempt %d: Failed to connect to PostgreSQL: %v\n", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			log.Printf("Attempt %d: Failed to ping PostgreSQL: %v\n", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		log.Println("Connected to PostgreSQL!")
		return db, nil
	}

	return nil, fmt.Errorf("failed to connect to PostgreSQL after multiple attempts")
}
