package middlewares

import (
	"context"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {

		authHeader := request.Header.Get("Authorization")
		if authHeader == "" {
			helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
			return
		}

		tokenString := strings.Split(authHeader, " ")[1]
		if len(tokenString) != 2 {
			helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
			return
		}

		data, err := helpers.DecodeToken(tokenString)
		if err != nil {
			helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
			return
		}

		if !data["verified"].(bool) {
			helpers.HandleError(responseWriter, helpers.ErrForbidden)
			return
		}

		ctx := context.WithValue(request.Context(), "email", data["email"])
		ctx = context.WithValue(ctx, "user_id", data["user_id"])
		ctx = context.WithValue(ctx, "verified", data["verified"])

		newRequest := request.WithContext(ctx)
		next.ServeHTTP(responseWriter, newRequest)
	})
}
