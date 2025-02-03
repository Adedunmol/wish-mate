package notification

import (
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/helpers"
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
	Store Store
}

func (h *Handler) CreateNotification(body *CreateNotificationBody) (Notification, error) {
	if body.UserID == 0 {
		return Notification{}, errors.New("user id is required")
	}

	if body.Body == "" {
		return Notification{}, errors.New("body is required")
	}

	if body.Type == "" {
		return Notification{}, errors.New("type is required")
	}

	if body.Title == "" {
		return Notification{}, errors.New("title is required")
	}

	notification, err := h.Store.CreateNotification(body)

	if err != nil {
		return Notification{}, fmt.Errorf("error creating notification: %v", err)
	}

	return notification, nil
}

func (h *Handler) GetNotificationHandler(responseWriter http.ResponseWriter, request *http.Request) {
	notificationID := chi.URLParam(request, "notification_id")

	if notificationID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("id is required"), http.StatusBadRequest, "id is required", nil))
		return
	}

	userID := request.Context().Value("user_id")

	if userID == nil || userID == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	newNotificationID, err := strconv.Atoi(notificationID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newUserID := userID.(int)

	notification, err := h.Store.GetNotification(newNotificationID)
	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	if notification.UserID != newUserID {
		helpers.HandleError(responseWriter, helpers.ErrForbidden)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Notification retrieved successfully",
		Data:    notification,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}
