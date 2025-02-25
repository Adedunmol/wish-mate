package helpers

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"os"
	"time"
)

const TokenExpiration = 30 * time.Minute

func GenerateToken(userID int, email string, verified bool) (string, error) {
	var signingKey = []byte(os.Getenv("SECRET_KEY"))
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["email"] = email
	claims["user_id"] = userID
	claims["verified"] = verified
	claims["exp"] = time.Now().Add(TokenExpiration).Unix()

	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		fmt.Printf("error generating token: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}

func DecodeToken(tokenString string) (map[string]interface{}, error) {
	var err error

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var data map[string]interface{}

		data["email"] = claims["email"].(string)
		data["user_id"] = claims["user_id"].(string)
		data["verified"] = claims["verified"].(bool)
		return data, nil
	}
	return nil, fmt.Errorf("invalid token")
}
