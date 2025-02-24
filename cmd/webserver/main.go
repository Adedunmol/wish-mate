package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/Adedunmol/wish-mate/internal/reminder"
	"github.com/Adedunmol/wish-mate/internal/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %s", err)
	}

	defer handlePanics()

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal(errors.Unwrap(err))
	}

	defer db.Close(ctx)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	qc, err := queue.NewClient(ctxWithTimeout)
	if err != nil {
		log.Fatal(errors.Unwrap(err))
	}

	// create a new scheduler
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal(fmt.Errorf("error starting gocron scheduler: %v", errors.Unwrap(err)))
	}

	// schedule a task to run every minute
	_, err = scheduler.NewJob(
		gocron.DurationJob(time.Minute),
		gocron.NewTask(func() {
			enqueueReminders(qc, db)
		}),
	)
	if err != nil {
		log.Fatalf("failed to schedule job: %v", errors.Unwrap(err))
	}

	go scheduler.Start()

	r := chi.NewRouter()

	routes.SetupRoutes(config.Config{DB: db, Router: r, Queue: qc})

	go qc.Run(ctx, db)

	// handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	server := &http.Server{Addr: fmt.Sprintf(":%d", os.Getenv("PORT")), Handler: r}
	go func() {
		log.Printf("starting web server on port %d", os.Getenv("PORT"))
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("error starting web server on port %d: %v", os.Getenv("PORT"), err)
		}
	}()

	<-stop

	err = scheduler.Shutdown()
	if err != nil {
		log.Fatal(fmt.Errorf("error shutting down scheduler: %v", errors.Unwrap(err)))
	}

	// gracefully shutdown the server
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shut down: %v", err)
	}

	log.Println("server exited properly")
}

func handlePanics() {
	if r := recover(); r != nil {
		log.Printf("panic occurred: %v", r)
	}
}

func enqueueReminders(client *queue.Client, db *pgx.Conn) {
	currentTime := time.Now()
	taskStore := &reminder.ReminderStore{DB: db}

	// check db for reminders where scheduled = pending AND scheduled_at <= now
	log.Printf("checking due scheduled reminders at: %v", currentTime.UTC())

	if err := reminder.EnqueueReminders(taskStore, client, &currentTime); err != nil {
		log.Printf(errors.Unwrap(err).Error())
	}

	// get today's birthdays and send notifications and mails to their friends
	log.Printf("checking birthdays due today: %v", currentTime.UTC())

	if err := reminder.EnqueueBirthdays(taskStore, client, &currentTime); err != nil {
		log.Printf(errors.Unwrap(err).Error())
	}
}
