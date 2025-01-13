package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Handler struct {
	Store Store
}

func (h *Handler) CreateUserHandler(responseWriter http.ResponseWriter, request *http.Request) {

	body, _, err := helpers.DecodeAndValidate[*CreateUserBody](request)
	if err != nil && errors.Is(err, helpers.ErrValidate) {
		fmt.Errorf("err (create user): %s", err)
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	if err != nil && errors.Is(err, helpers.ErrDecode) {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	body.Password = string(hashedPassword)

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

	token, err := helpers.GenerateToken(data.ID, data.Email)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	response := Response{
		Status:  "Success",
		Message: "User logged in",
		Data:    map[string]interface{}{"token": token, "expiration": helpers.TokenExpiration},
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
