package wishlist

import (
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"time"
)

type Item struct {
	helpers.Validation
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Link        string `json:"link,omitempty"`
}

type Wishlist struct {
	helpers.Validation
	Name         string     `json:"name" validate:"required"`
	Description  string     `json:"description" validate:"required"`
	Items        []Item     `json:"items,omitempty"`
	NotifyBefore int        `json:"notify_before" validate:"required"`
	Date         *time.Time `json:"date,omitempty"`
}

type ItemResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Taken       bool      `json:"taken"`
	Link        string    `json:"link"`
	PickedBy    auth.User `json:"picked_by,omitempty"`
	WishlistID  int       `json:"wishlist_id,omitempty"`
}

type WishlistResponse struct {
	ID           int            `json:"id,omitempty"`
	UserID       int            `json:"user_id,omitempty"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	NotifyBefore int            `json:"notify_before,omitempty"`
	Date         *time.Time     `json:"date,omitempty"`
	Items        []ItemResponse `json:"items,omitempty"`
}

type UpdateWishlist struct {
	helpers.Validation
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateItem struct {
	helpers.Validation
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Link        string `json:"link,omitempty"`
}
