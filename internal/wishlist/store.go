package wishlist

import "database/sql"

type Store interface {
	CreateWishlist(userID int, body Wishlist) (WishlistResponse, error)
	GetWishlistByID(wishlistID, userID int) (WishlistResponse, error)
	UpdateWishlistByID(wishlistID, userID int, body UpdateWishlist) (WishlistResponse, error)
	DeleteWishlistByID(wishlistID, userID int) error
}

type WishlistStore struct {
	db *sql.DB
}

func NewWishlistStore(db *sql.DB) *WishlistStore {

	return &WishlistStore{db: db}
}

func (w *WishlistStore) CreateWishlist(userID int, body Wishlist) (WishlistResponse, error) {

	return WishlistResponse{}, nil
}

func (w *WishlistStore) GetWishlistByID(wishlistID, userID int) (WishlistResponse, error) {
	return WishlistResponse{}, nil
}

func (w *WishlistStore) UpdateWishlistByID(wishlistID, userID int, body UpdateWishlist) (WishlistResponse, error) {

	return WishlistResponse{}, nil
}

func (w *WishlistStore) DeleteWishlistByID(wishlistID, userID int) error {
	return nil
}
