package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	defer handlePanics()

	r := chi.NewRouter()

	log.Fatal(http.ListenAndServe(os.Getenv("PORT"), r))
}

func handlePanics() {
	if r := recover(); r != nil {
		log.Printf("panic occurred: %v", r)
	}
}
