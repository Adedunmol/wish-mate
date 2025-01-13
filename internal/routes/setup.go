package routes

import (
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/Adedunmol/wish-mate/internal/user"
)

func SetupRoutes(config config.Config) {

	user.UserRoutes(config)
}
