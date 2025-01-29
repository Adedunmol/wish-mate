package config

import (
	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type Config struct {
	DB     *pgx.Conn
	Router *chi.Mux
	Queue  queue.Queue
}
