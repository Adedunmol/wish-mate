package config

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
)

func ConnectDB() (*sql.DB, error) {
	connectionStr, exists := os.LookupEnv("DATABASE_URL")

	if !exists {
		return nil, errors.New("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", connectionStr)

	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return db, nil
}
