package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"os"
)

func ConnectDB() (*pgx.Conn, error) {
	connectionStr, exists := os.LookupEnv("DATABASE_URL")

	if !exists {
		return nil, errors.New("DATABASE_URL environment variable not set")
	}

	conn, err := pgx.Connect(context.Background(), connectionStr)

	defer conn.Close(context.Background())

	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return conn, nil
}
