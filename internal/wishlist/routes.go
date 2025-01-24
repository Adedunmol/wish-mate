package wishlist

import (
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func WishlistRoutes(config config.Config) {

	wishlistRouter := chi.NewRouter()

	store := NewWishlistStore(config.DB)

	handler := Handler{Store: store}

	wishlistRouter.Post("/", http.HandlerFunc(handler.CreateWishlist))
	wishlistRouter.Get("/{id}", http.HandlerFunc(handler.GetWishlist))
	wishlistRouter.Patch("/{id}", http.HandlerFunc(handler.UpdateWishlist))
	wishlistRouter.Delete("/{id}", http.HandlerFunc(handler.DeleteWishlist))

	config.Router.Mount("/wishlists", wishlistRouter)
}
