package wishlist

import (
	"github.com/jackc/pgx/v5"
)

type Store interface {
	CreateWishlist(userID int, body Wishlist) (WishlistResponse, error)
	GetWishlistByID(wishlistID, userID int) (WishlistResponse, error)
	UpdateWishlistByID(wishlistID, userID int, body UpdateWishlist) (WishlistResponse, error)
	DeleteWishlistByID(wishlistID, userID int) error
	GetItem(wishlistID, itemID int) (ItemResponse, error)
	UpdateItem(wishlistID, itemID int, body Item) (ItemResponse, error)
	PickItem(wishlistID, itemID, userID int) (ItemResponse, error)
}

type WishlistStore struct {
	db *pgx.Conn
}

func NewWishlistStore(db *pgx.Conn) *WishlistStore {

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

func (w *WishlistStore) GetItem(wishlistID, itemID int) (ItemResponse, error) {
	return ItemResponse{}, nil
}

func (w *WishlistStore) UpdateItem(wishlistID, itemID int, body Item) (ItemResponse, error) {
	return ItemResponse{}, nil
}

func (w *WishlistStore) PickItem(wishlistID, itemID, userID int) (ItemResponse, error) {
	return ItemResponse{}, nil
}
