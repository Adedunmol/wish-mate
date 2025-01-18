package wishlist

type Store interface {
	CreateWishlist(body interface{}) (interface{}, error)
	GetWishlistByID(id int, verbose bool) (interface{}, error)
	UpdateWishlistByID(id int, body interface{}) (interface{}, error)
	DeleteWishlistByID(id int) error
}
