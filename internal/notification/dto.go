package notification

import "time"

type Notification struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"` // the receiver's id
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Type      string     `json:"type"` // birthday, wishlist_reminder, item_picked, birthday_greeting, friend_request, friend_request_accepted // todo: wishlist_shared, wishlist_expiring
	Read      bool       `json:"read"`
	Timestamp *time.Time `json:"timestamp"` // time the notification was created
}

type CreateNotificationBody struct {
	UserID int    `json:"user_id" validate:"required"` // the receiver's id
	Title  string `json:"title" validate:"required"`
	Body   string `json:"body" validate:"required"`
	Type   string `json:"type" validate:"required"` // alert, update
}
