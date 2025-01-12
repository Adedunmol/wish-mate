package config

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
)

type Config struct {
	DB     *sql.DB
	Router *chi.Mux
}
