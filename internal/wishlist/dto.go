package wishlist

import "github.com/Adedunmol/wish-mate/internal/helpers"

type Item struct {
	helpers.Validation
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Whole       bool   `json:"whole" validate:"required"`
}

type Wishlist struct {
	helpers.Validation
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Items       []Item `json:"items,omitempty"`
}

type WishlistResponse struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Items       []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Taken       bool   `json:"taken"`
		Whole       bool   `json:"whole"`
	} `json:"items,omitempty"`
}
