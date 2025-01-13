package user

import (
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func UserRoutes(config config.Config) {

	userRouter := chi.NewRouter()

	store := NewUserStore(config.DB)

	handler := Handler{Store: store}

	userRouter.Patch("/register", http.HandlerFunc(handler.CreateUserHandler))
	userRouter.Post("/login", http.HandlerFunc(handler.LoginUserHandler))

	config.Router.Mount("/users", userRouter)
}
