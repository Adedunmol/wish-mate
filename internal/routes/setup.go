package routes

import (
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/friendship"
	"github.com/Adedunmol/wish-mate/internal/wishlist"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/swaggest/swgui/v5emb"
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

	config.Router.Route("/docs", func(r chi.Router) {
		// Serve the embedded Swagger UI assets
		r.Mount("/", v5emb.New(
			"My API Docs",
			"/docs/swagger.json", // URL Swagger UI fetches
			"/docs",              // BasePath must match route
		))

		// Serve swagger.json file from ./docs folder
		r.Get("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
			http.ServeFile(w, req, "./docs/swagger.json")
		})
	})

	auth.AuthRoutes(config)
	friendship.FriendshipRoutes(config)
	wishlist.WishlistRoutes(config)
}
