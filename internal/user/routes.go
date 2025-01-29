package user

import (
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func UserRoutes(config config.Config) {

	userRouter := chi.NewRouter()

	authStore := auth.NewUserStore(config.DB)
	friendshipStore := NewFriendshipStore(config.DB)

	handler := Handler{AuthStore: authStore, FriendStore: friendshipStore, Queue: config.Queue}

	userRouter.Post("/{user_id}/friend_requests", http.HandlerFunc(handler.SendRequestHandler))
	userRouter.Patch("/{user_id}/friend_requests/{request_id}", http.HandlerFunc(handler.UpdateRequestHandler))

	config.Router.Mount("/users", userRouter)
}
