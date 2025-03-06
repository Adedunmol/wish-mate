package auth

import (
	"github.com/Adedunmol/wish-mate/internal/config"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func AuthRoutes(config config.Config) {

	authRouter := chi.NewRouter()

	store := NewUserStore(config.DB)

	handler := Handler{Store: store, Queue: config.Queue}

	authRouter.Post("/register", http.HandlerFunc(handler.CreateUserHandler))
	authRouter.Post("/login", http.HandlerFunc(handler.LoginUserHandler))
	authRouter.Post("/logout", http.HandlerFunc(handler.LogoutUserHandler))
	authRouter.Post("/verify", http.HandlerFunc(handler.VerifyUserHandler))
	authRouter.Get("/refresh-token", http.HandlerFunc(handler.RefreshTokenHandler))
	authRouter.Post("/request-code", http.HandlerFunc(handler.RequestCodeHandler))
	authRouter.Post("/forgot-password", http.HandlerFunc(handler.ForgotPasswordHandler))
	authRouter.Post("/forgot-password-request", http.HandlerFunc(handler.ForgotPasswordRequestHandler))
	authRouter.Post("/reset-password", http.HandlerFunc(handler.ResetPasswordHandler))

	config.Router.Mount("/auth", authRouter)
}
