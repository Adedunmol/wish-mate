package main

import (
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/routes"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	defer handlePanics()

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal(err)
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
