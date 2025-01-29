package wishlist

import (
	"github.com/Adedunmol/wish-mate/internal/helpers"
)

type Item struct {
	helpers.Validation
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Link        string `json:"link,omitempty"`
}

type Wishlist struct {
	helpers.Validation
	Name         string `json:"name" validate:"required"`
	Description  string `json:"description" validate:"required"`
	Items        []Item `json:"items,omitempty"`
	NotifyBefore int    `json:"notify_before" validate:"required"`
}

type ItemResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Taken       bool   `json:"taken"`
	Link        string `json:"link"`
}

type WishlistResponse struct {
	ID           int            `json:"id,omitempty"`
	UserID       int            `json:"user_id,omitempty"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	NotifyBefore int            `json:"notify_before,omitempty"`
	Items        []ItemResponse `json:"items,omitempty"`
}

type UpdateWishlist struct {
	helpers.Validation
	Name        string `json:"name"`
	Description string `json:"description"`
}
