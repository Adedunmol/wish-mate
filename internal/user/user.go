package user

import (
	"errors"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Handler struct {
	Store Store
	Queue queue.Queue
}

func (h *Handler) CreateUserHandler(responseWriter http.ResponseWriter, request *http.Request) {

	body, problems, err := helpers.DecodeAndValidate[*CreateUserBody](request)

	var clientError helpers.ClientError
	ok := errors.As(err, &clientError)

	if err != nil && problems == nil {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(err, http.StatusBadRequest, "invalid request body", nil))
		return
	}

	if err != nil && ok {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(err, http.StatusBadRequest, "invalid request body", problems))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	body.Password = string(hashedPassword)

	data, err := h.Store.CreateUser(body)
	if err != nil {
		var clientError helpers.ClientError
		ok := errors.As(err, &clientError)

		if ok {
			helpers.HandleError(responseWriter, helpers.ErrConflict)
			return
		}

		helpers.HandleError(responseWriter, helpers.NewHTTPError(err, http.StatusInternalServerError, "internal server error", nil))
		return
	}

	response := Response{
		Status:  "Success",
		Message: "User created successfully",
		Data:    data,
	}

	err = h.Queue.Enqueue(&queue.TaskPayload{
		Type: queue.TypeEmailDelivery,
		Payload: map[string]interface{}{
			"email":    body.Email,
			"template": "register",
			"subject":  "register",
			"data":     map[string]interface{}{},
		},
	})

	if err != nil {
		log.Printf("error enqueuing email task: %s", err)
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusCreated)
}

func (h *Handler) LoginUserHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, problems, err := helpers.DecodeAndValidate[*LoginUserBody](request)

	var clientError helpers.ClientError
	ok := errors.As(err, &clientError)

	if err != nil && problems == nil {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(err, http.StatusBadRequest, "invalid request body", nil))
		return
	}

	if err != nil && ok {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(err, http.StatusBadRequest, "invalid request body", problems))
		return
	}

	data, err := h.Store.FindUserByEmail(body.Email)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	match := h.Store.ComparePasswords(data.Password, body.Password)
	if !match {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	token, err := helpers.GenerateToken(data.ID, data.Email)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(err, http.StatusInternalServerError, "internal server error", nil))
		return
	}
	response := Response{
		Status:  "Success",
		Message: "User logged in",
		Data:    map[string]interface{}{"token": token, "expiration": helpers.TokenExpiration},
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}
