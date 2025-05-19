package friendship

import (
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/notification"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Handler struct {
	AuthStore         auth.Store
	FriendStore       FriendStore
	Queue             queue.Queue
	NotificationStore *notification.NotificationStore
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
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("friendship id is required"), http.StatusBadRequest, "id is required", nil))
		return
	}

	newUserID, err := strconv.Atoi(userID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	email := request.Context().Value("email")

	if email == nil || email == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	userData, err := h.AuthStore.FindUserByEmail(email.(string))
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	if newUserID != userData.ID {
		helpers.HandleError(responseWriter, helpers.ErrForbidden)
		return
	}

	data, err := h.FriendStore.CreateFriendship(newUserID, body.RecipientID)
	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	_, err = h.NotificationStore.CreateNotification(&notification.CreateNotificationBody{
		UserID: body.RecipientID,
		Title:  "Friend request",
		Body:   fmt.Sprintf("You have a request from %s", userData.Username),
		Type:   "friend_request",
	})

	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
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

	friendID := chi.URLParam(request, "friend_id")

	if friendID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(nil, http.StatusBadRequest, "request id is required", nil))
		return
	}

	newFriendID, err := strconv.Atoi(friendID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	currentUserID := request.Context().Value("user_id").(int)

	friendshipData, err := h.FriendStore.GetFriendship(currentUserID, newFriendID)

	if err != nil {
		if errors.Is(err, ErrNoFriendship) {
			helpers.HandleError(responseWriter, helpers.ErrNotFound)
			return
		}
		helpers.HandleError(responseWriter, err)
		return
	}

	if currentUserID != friendshipData.UserID {
		helpers.HandleError(responseWriter, helpers.ErrForbidden)
		return
	}

	var status string

	switch strings.ToLower(body.Type) {
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

	data, err := h.FriendStore.UpdateFriendship(currentUserID, newFriendID, status)

	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	email := request.Context().Value("email")

	if email == nil || email == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	userData, err := h.AuthStore.FindUserByEmail(email.(string))
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	_, err = h.NotificationStore.CreateNotification(&notification.CreateNotificationBody{
		UserID: data.FriendID,
		Title:  "Friend request update",
		Body:   fmt.Sprintf("You have an update for the request from %s", userData.Username),
		Type:   "friend_request",
	})

	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
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
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("friendship id is required"), http.StatusBadRequest, "id is required", nil))
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
		Message: "Friendships retrieved successfully",
		Data:    data,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}

func (h *Handler) GetRequestHandler(responseWriter http.ResponseWriter, request *http.Request) {
	userID := chi.URLParam(request, "user_id")

	if userID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("friendship id is required"), http.StatusBadRequest, "id is required", nil))
		return
	}

	newUserID, err := strconv.Atoi(userID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	friendID := chi.URLParam(request, "friend_id")

	if friendID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(nil, http.StatusBadRequest, "request id is required", nil))
		return
	}

	newFriendID, err := strconv.Atoi(friendID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	currentUserID := request.Context().Value("user_id").(int)

	if currentUserID != newUserID {
		helpers.HandleError(responseWriter, helpers.ErrForbidden)
		return
	}

	data, err := h.FriendStore.GetFriendship(newUserID, newFriendID)

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

//func (h *Handler) GetAllFriendsHandler(responseWriter http.ResponseWriter, request *http.Request) {}
//
//func (h *Handler) GetUser(responseWriter http.ResponseWriter, request *http.Request) {}
