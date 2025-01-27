package user

import (
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"time"
)

type FriendRequestBody struct {
	helpers.Validation
	RecipientID int `json:"recipient_id" validate:"required"`
}

type UpdateFriendRequestBody struct {
	helpers.Validation
	Type string `json:"type" validate:"required"`
}

type FriendshipResponse struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	FriendID    int        `json:"friend_id"`
	Status      string     `json:"status"`
	FriendSince *time.Time `json:"friend_since,omitempty"`
}
