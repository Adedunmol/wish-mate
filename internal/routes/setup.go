package routes

import (
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/user"
	"github.com/Adedunmol/wish-mate/internal/wishlist"
)

func SetupRoutes(config config.Config) {

	user.UserRoutes(config)
	wishlist.WishlistRoutes(config)
}
