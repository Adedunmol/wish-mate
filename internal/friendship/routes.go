package friendship

import (
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/middlewares"
	"github.com/Adedunmol/wish-mate/internal/notification"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func FriendshipRoutes(config config.Config) {

	friendshipRouter := chi.NewRouter()

	friendshipRouter.Use(middlewares.AuthMiddleware)

	authStore := auth.NewUserStore(config.DB)
	friendshipStore := NewFriendshipStore(config.DB)
	notificationStore := notification.NewNotificationStore(config.DB)

	handler := Handler{AuthStore: authStore, FriendStore: friendshipStore, Queue: config.Queue, NotificationStore: notificationStore}

	friendshipRouter.Post("/{user_id}/friend_requests", http.HandlerFunc(handler.SendRequestHandler))
	friendshipRouter.Patch("/{user_id}/friend_requests/{friend_id}", http.HandlerFunc(handler.UpdateRequestHandler))
	friendshipRouter.Get("/{user_id}/friend_requests/{friend_id}", http.HandlerFunc(handler.GetRequestHandler))
	friendshipRouter.Get("/{user_id}/friend_requests", http.HandlerFunc(handler.GetAllRequestsHandler))

	config.Router.Mount("/users", friendshipRouter)
}
