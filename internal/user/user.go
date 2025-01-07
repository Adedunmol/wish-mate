package user

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type User struct {
	ID        int
	FirstName string
	LastName  string
	Username  string
	Email     string
	Password  string
}

type CreateUserBody struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}

type LoginUserBody struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type CreateUserResponse struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type Store interface {
	CreateUser(body CreateUserBody) (CreateUserResponse, error)
	FindUserByEmail(email string) (User, error)
	ComparePasswords(storedPassword, candidatePassword string) bool
}

type Handler struct {
	Store Store
}

func (h *Handler) CreateUserHandler(responseWriter http.ResponseWriter, request *http.Request) {

	var body CreateUserBody
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := h.Store.CreateUser(body)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "User created successfully",
		Data:    data,
	}

	WriteJSONResponse(responseWriter, response, http.StatusCreated)
}

func (h *Handler) LoginUserHandler(responseWriter http.ResponseWriter, request *http.Request) {
	var body LoginUserBody
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := h.Store.FindUserByEmail(body.Email)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	match := h.Store.ComparePasswords(data.Password, body.Password)
	if !match {
		http.Error(responseWriter, "Password does not match", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(data.ID, data.Email)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	response := Response{
		Status:  "Success",
		Message: "User logged in",
		Data:    map[string]interface{}{"token": token, "expiration": TokenExpiration},
	}

	WriteJSONResponse(responseWriter, response, http.StatusOK)
}

func WriteJSONResponse(responseWriter http.ResponseWriter, data Response, statusCode int) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)

	if err := json.NewEncoder(responseWriter).Encode(data); err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

const TokenExpiration = 30 * time.Minute

func generateToken(userID int, email string) (string, error) {
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
