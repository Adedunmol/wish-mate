package routes

import (
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/friendship"
	"github.com/Adedunmol/wish-mate/internal/wishlist"
)

func SetupRoutes(config config.Config) {

	auth.AuthRoutes(config)
	friendship.UserRoutes(config)
	wishlist.WishlistRoutes(config)
}
