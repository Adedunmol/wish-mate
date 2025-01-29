package main

import (
	"context"
	"errors"
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/Adedunmol/wish-mate/internal/routes"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	defer handlePanics()

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal(errors.Unwrap(err))
	}

	qc, err := queue.NewClient(context.Background())
	if err != nil {
		log.Fatal(errors.Unwrap(err))
	}

	r := chi.NewRouter()

	routes.SetupRoutes(config.Config{DB: db, Router: r, Queue: qc})

	go checkScheduledJobs(qc.GetClient(), db)

	log.Fatal(http.ListenAndServe(os.Getenv("PORT"), r))
}

func handlePanics() {
	if r := recover(); r != nil {
		log.Printf("panic occurred: %v", r)
	}
}

func checkScheduledJobs(client *asynq.Client, db *pgx.Conn) {

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for t := range ticker.C {
		// check db for jobs where scheduled = false AND scheduled_at <= now
		log.Printf("checking due scheduled jobs at: %v", t.UTC())
	}
}
