package notification

import (
	"errors"
	"fmt"
)

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
