package user

import (
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func UserRoutes(config config.Config) {

	userRouter := chi.NewRouter()

	handler := Handler{}

	userRouter.Post("/{user_id}/friend_requests", http.HandlerFunc(handler.SendRequestHandler))
	userRouter.Patch("/{user_id}/friend_requests", http.HandlerFunc(handler.AcceptRequestHandler))

	config.Router.Mount("/users", userRouter)
}
