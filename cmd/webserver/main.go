package main

import (
	"database/sql"
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/routes"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	defer handlePanics()

	connectionStr, exists := os.LookupEnv("DATABASE_URL")

	if !exists {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", connectionStr)

	if err != nil {
		log.Fatalf("error connecting to db: %s", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("error pinging db: %s", err)
	}

	r := chi.NewRouter()

	routes.SetupRoutes(config.Config{DB: db, Router: r})

	log.Fatal(http.ListenAndServe(os.Getenv("PORT"), r))
}

func handlePanics() {
	if r := recover(); r != nil {
		log.Printf("panic occurred: %v", r)
	}
}
