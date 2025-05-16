package routes

import (
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/friendship"
	"github.com/Adedunmol/wish-mate/internal/wishlist"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func SetupRoutes(config config.Config) {

	config.Router.Use(middleware.Logger)
	config.Router.Use(middleware.CleanPath)
	config.Router.Use(middleware.Recoverer)

	config.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("Hello from wishmate API\n")); err != nil {
			log.Println(fmt.Errorf("error writing response: %w", err))

			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	auth.AuthRoutes(config)
	friendship.FriendshipRoutes(config)
	wishlist.WishlistRoutes(config)
}
