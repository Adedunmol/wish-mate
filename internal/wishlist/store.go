package wishlist

type Store interface {
	CreateWishlist(userID int, body Wishlist) (WishlistResponse, error)
	GetWishlistByID(wishlistID, userID int) (WishlistResponse, error)
	UpdateWishlistByID(id int, body Wishlist) (WishlistResponse, error)
	DeleteWishlistByID(id int) error
}
