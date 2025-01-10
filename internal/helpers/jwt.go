package helpers

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"os"
	"time"
)

const TokenExpiration = 30 * time.Minute

func GenerateToken(userID int, email string) (string, error) {
	var signingKey = []byte(os.Getenv("SECRET_KEY"))
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["email"] = email
	claims["id"] = userID
	claims["exp"] = time.Now().Add(TokenExpiration).Unix()

	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		fmt.Printf("error generating token: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}
