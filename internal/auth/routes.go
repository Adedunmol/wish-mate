package auth

import (
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func AuthRoutes(config config.Config) {

	authRouter := chi.NewRouter()

	store := NewUserStore(config.DB)

	handler := Handler{Store: store, Queue: config.Queue}

	authRouter.Post("/register", http.HandlerFunc(handler.CreateUserHandler))
	authRouter.Post("/login", http.HandlerFunc(handler.LoginUserHandler))

	config.Router.Mount("/auth", authRouter)
}
