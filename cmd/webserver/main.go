package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	log.Fatal(http.ListenAndServe(os.Getenv("PORT"), r))
}
