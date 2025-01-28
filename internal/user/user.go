package user

import (
	"errors"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Handler struct {
	AuthStore   auth.Store
	FriendStore FriendStore
	Queue       queue.Queue
}

func (h *Handler) SendRequestHandler(responseWriter http.ResponseWriter, request *http.Request) {

	body, problems, err := helpers.DecodeAndValidate[*FriendRequestBody](request)

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

	userID := chi.URLParam(request, "user_id")

	if userID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("user id is required"), http.StatusBadRequest, "id is required", nil))
		return
	}

	newUserID, err := strconv.Atoi(userID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	data, err := h.FriendStore.CreateFriendship(newUserID, body.RecipientID)
	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Friendship created successfully",
		Data:    data,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusCreated)
}

func (h *Handler) UpdateRequestHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, problems, err := helpers.DecodeAndValidate[*UpdateFriendRequestBody](request)

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

	requestID := chi.URLParam(request, "request_id")

	if requestID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(nil, http.StatusBadRequest, "request id is required", nil))
		return
	}

	newRequestID, err := strconv.Atoi(requestID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	var status string

	switch body.Type {
	case "accept":
		status = "accepted"
		break
	case "block":
		status = "blocked"
		break
	default:
		helpers.HandleError(responseWriter, helpers.NewHTTPError(nil, http.StatusBadRequest, "invalid type", nil))
		return
	}

	data, err := h.FriendStore.UpdateFriendship(newRequestID, status)

	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Friendship updated successfully",
		Data:    data,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}

func (h *Handler) GetAllRequestsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	userID := chi.URLParam(request, "user_id")

	if userID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("user id is required"), http.StatusBadRequest, "id is required", nil))
		return
	}

	newUserID, err := strconv.Atoi(userID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	currentUserID := request.Context().Value("user_id").(int)

	if currentUserID != newUserID {
		helpers.HandleError(responseWriter, helpers.ErrForbidden)
		return
	}

	status := request.URL.Query().Get("status")

	var data []FriendshipResponse

	switch status {
	case "accepted", "blocked", "pending":
		data, err = h.FriendStore.GetAllFriendships(newUserID, status)
		break
	case "":
		status = "all"
		data, err = h.FriendStore.GetAllFriendships(newUserID, status)
		break
	default:
		helpers.HandleError(responseWriter, helpers.NewHTTPError(nil, http.StatusBadRequest, "invalid status", nil))
		return
	}

	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Friendship retrieved successfully",
		Data:    data,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}
