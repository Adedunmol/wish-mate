package wishlist

import (
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/reminder"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func WishlistRoutes(config config.Config) {

	wishlistRouter := chi.NewRouter()

	store := NewWishlistStore(config.DB)
	userStore := auth.NewUserStore(config.DB)
	reminderStore := reminder.NewReminderStore(config.DB)

	handler := Handler{Store: store, UserStore: userStore, ReminderStore: reminderStore}

	wishlistRouter.Post("/", http.HandlerFunc(handler.CreateWishlist))
	wishlistRouter.Get("/{id}", http.HandlerFunc(handler.GetWishlist))
	wishlistRouter.Patch("/{id}", http.HandlerFunc(handler.UpdateWishlist))
	wishlistRouter.Delete("/{id}", http.HandlerFunc(handler.DeleteWishlist))

	config.Router.Mount("/wishlists", wishlistRouter)
}
