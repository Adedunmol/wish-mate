package notification

import "time"

type Notification struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"` // the receiver's id
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Type      string     `json:"type"`      // alert, update
	Status    string     `json:"status"`    // read or unread
	Timestamp *time.Time `json:"timestamp"` // time the notification was created
}

type CreateNotificationBody struct {
	UserID int    `json:"user_id" validate:"required"` // the receiver's id
	Title  string `json:"title" validate:"required"`
	Body   string `json:"body" validate:"required"`
	Type   string `json:"type" validate:"required"` // alert, update
}
