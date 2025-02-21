package wishlist

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/jackc/pgx/v5"
	"time"
)

type Store interface {
	CreateWishlist(userID int, body Wishlist) (WishlistResponse, error)
	GetWishlistByID(wishlistID, userID int) (WishlistResponse, error)
	GetUserWishlists(userID int, isOwner bool) ([]WishlistResponse, error)
	UpdateWishlistByID(wishlistID, userID int, body UpdateWishlist) (WishlistResponse, error)
	DeleteWishlistByID(wishlistID, userID int) error
	GetItem(wishlistID, itemID int) (ItemResponse, error)
	UpdateItem(wishlistID, itemID int, body *UpdateItem) (ItemResponse, error)
	PickItem(wishlistID, itemID, userID int) (ItemResponse, error)
	DeleteItem(wishlistID, itemID int) error
}

type WishlistStore struct {
	db *pgx.Conn
}

func NewWishlistStore(db *pgx.Conn) *WishlistStore {

	return &WishlistStore{db: db}
}

func (w *WishlistStore) CreateWishlist(userID int, body Wishlist) (WishlistResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WishlistResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var wishlist WishlistResponse

	query := `INSERT INTO wishlists (user_id, name, description, notify_before, date) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id, user_id, name, description, notify_before, date;`

	err = w.db.QueryRow(ctx, query, userID, body.Name, body.Description, body.NotifyBefore, body.Date).
		Scan(&wishlist.ID, &wishlist.UserID, &wishlist.Name, &wishlist.Description, &wishlist.NotifyBefore, &wishlist.Date)
	if err != nil {
		return WishlistResponse{}, fmt.Errorf("error inserting wishlist: %w", err)
	}

	wishlist.Items = make([]ItemResponse, 0)

	insertItemQuery := `INSERT INTO items (wishlist_id, name, description, link) VALUES ($1, $2, $3, $4) RETURNING id, name, description, price;`
	for _, item := range body.Items {
		var newItem ItemResponse
		err = w.db.QueryRow(ctx, insertItemQuery, wishlist.ID, item.Name, item.Description, item.Link).
			Scan(&newItem.ID, &newItem.Name, &newItem.Description, &newItem.Link)

		if err != nil {
			return WishlistResponse{}, fmt.Errorf("error inserting item: %w", err)
		}
		wishlist.Items = append(wishlist.Items, newItem)
	}

	return wishlist, nil
}

func (w *WishlistStore) GetWishlistByID(wishlistID, userID int) (WishlistResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WishlistResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var wishlist WishlistResponse

	query := `SELECT id, user_id, name, description, notify_before, date 
		FROM wishlists WHERE id = $1;`
	err = w.db.QueryRow(ctx, query, wishlistID).
		Scan(&wishlist.ID, &wishlist.UserID, &wishlist.Name, &wishlist.Description, &wishlist.NotifyBefore, &wishlist.Date)
	if err != nil {
		return WishlistResponse{}, fmt.Errorf("error fetching wishlist: %w", err)
	}

	wishlist.Items = make([]ItemResponse, 0)

	var itemsQuery string
	if wishlist.Date <= fmt.Sprintf("%s", sql.NullString{String: "CURRENT_DATE", Valid: true}) {
		itemsQuery = `SELECT i.id, i.name, i.description, i.price, u.id, u.username, u.first_name, u.last_name
			FROM items i 
			LEFT JOIN item_picks ip ON i.id = ip.item_id
			LEFT JOIN users u ON ip.user_id = u.id
			WHERE i.wishlist_id = $1;`
	} else {
		itemsQuery = `SELECT id, name, description, price, NULL, NULL, NULL FROM items WHERE wishlist_id = $1;`
	}

	rows, err := w.db.Query(ctx, itemsQuery, wishlistID)
	if err != nil {
		return WishlistResponse{}, fmt.Errorf("error fetching items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item ItemResponse
		var user auth.User
		var userID sql.NullInt64
		var username sql.NullString
		var firstName sql.NullString
		var lastName sql.NullString

		err = rows.Scan(&item.ID, &item.Name, &item.Description, &item.Link, &userID, &username, &firstName, &lastName)
		if err != nil {
			return WishlistResponse{}, fmt.Errorf("error scanning item: %w", err)
		}

		if userID.Valid {
			user = auth.User{
				ID:        int(userID.Int64),
				Username:  username.String,
				FirstName: firstName.String,
				LastName:  lastName.String,
			}
			item.PickedBy = user
		}

		wishlist.Items = append(wishlist.Items, item)
	}
	return WishlistResponse{}, nil
}

func (w *WishlistStore) GetUserWishlists(userID int, isOwner bool) ([]WishlistResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var wishlists []WishlistResponse

	query := `SELECT id, user_id, name, description, notify_before, date 
		FROM wishlists WHERE user_id = $1;`

	rows, err := w.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching wishlists: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var wishlist WishlistResponse
		err := rows.Scan(&wishlist.ID, &wishlist.UserID, &wishlist.Name, &wishlist.Description, &wishlist.NotifyBefore, &wishlist.Date)
		if err != nil {
			return nil, fmt.Errorf("error scanning wishlist: %w", err)
		}

		wishlist.Items = []ItemResponse{}
		var itemsQuery string

		if isOwner {
			// Owner: Show all items, picked and unpicked
			if wishlist.Date <= fmt.Sprintf("%s", sql.NullString{String: "CURRENT_DATE", Valid: true}) {
				itemsQuery = `SELECT i.id, i.name, i.description, i.price, u.id, u.username, u.first_name, u.last_name
					FROM items i 
					LEFT JOIN item_picks ip ON i.id = ip.item_id
					LEFT JOIN users u ON ip.user_id = u.id
					WHERE i.wishlist_id = $1;`
			} else {
				itemsQuery = `SELECT id, name, description, link FROM items WHERE wishlist_id = $1;`
			}
		} else {
			// Non-owner: Show only unpicked items
			itemsQuery = `SELECT id, name, description, link FROM items 
				WHERE wishlist_id = $1 AND id NOT IN (SELECT item_id FROM item_picks);`
		}

		itemRows, err := w.db.Query(ctx, itemsQuery, wishlist.ID)
		if err != nil {
			return nil, fmt.Errorf("error fetching items: %w", err)
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item ItemResponse
			var user auth.User
			var userID sql.NullInt64
			var username sql.NullString
			var firstName sql.NullString
			var lastName sql.NullString

			err := itemRows.Scan(&item.ID, &item.Name, &item.Description, &item.Link, &userID, &username, &firstName, &lastName)
			if err != nil {
				return nil, fmt.Errorf("error scanning item: %w", err)
			}

			if userID.Valid {
				user = auth.User{
					ID:        int(userID.Int64),
					Username:  username.String,
					FirstName: firstName.String,
					LastName:  lastName.String,
				}
				item.PickedBy = user
			}

			wishlist.Items = append(wishlist.Items, item)
		}

		wishlists = append(wishlists, wishlist)
	}

	return wishlists, nil
}

func (w *WishlistStore) UpdateWishlistByID(wishlistID, userID int, body UpdateWishlist) (WishlistResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WishlistResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var wishlist WishlistResponse

	// Check if the user is the owner of the wishlist
	var ownerID int
	err = w.db.QueryRow(ctx, "SELECT user_id FROM wishlists WHERE id = $1", wishlistID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WishlistResponse{}, errors.New("wishlist not found")
		}
		return WishlistResponse{}, fmt.Errorf("error checking wishlist ownership: %w", err)
	}

	if ownerID != userID {
		return WishlistResponse{}, errors.New("forbidden: you do not own this wishlist")
	}

	// Update the wishlist with non-empty fields
	query := `UPDATE wishlists SET 
		name = COALESCE(NULLIF($1, ''), name),
		description = COALESCE(NULLIF($2, ''), description)
		WHERE id = $3 RETURNING id, user_id, name, description, notify_before, date;`

	err = w.db.QueryRow(ctx, query, body.Name, body.Description, wishlistID).Scan(&wishlist.ID, &wishlist.UserID, &wishlist.Name, &wishlist.Description, &wishlist.NotifyBefore, &wishlist.Date)
	if err != nil {
		return WishlistResponse{}, fmt.Errorf("error updating wishlist: %w", err)
	}

	return wishlist, nil
}

func (w *WishlistStore) DeleteWishlistByID(wishlistID, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if the user is the owner of the wishlist
	var ownerID int
	err = w.db.QueryRow(ctx, "SELECT user_id FROM wishlists WHERE id = $1", wishlistID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("wishlist not found")
		}
		return fmt.Errorf("error checking wishlist ownership: %w", err)
	}

	if ownerID != userID {
		return errors.New("forbidden: you do not own this wishlist")
	}

	// Delete the wishlist
	_, err = w.db.Exec(ctx, "DELETE FROM wishlists WHERE id = $1", wishlistID)
	if err != nil {
		return fmt.Errorf("error deleting wishlist: %w", err)
	}

	return nil
}

func (w *WishlistStore) GetItem(wishlistID, itemID int) (ItemResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ItemResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var item ItemResponse
	var pickedBy auth.User

	query := `
	SELECT i.id, i.name, i.description, i.link
	FROM items i
	WHERE i.id = $1 AND i.wishlist_id = $2;`

	err = w.db.QueryRow(ctx, query, itemID, wishlistID).Scan(&item.ID, &item.Name, &item.Description, &item.Link)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ItemResponse{}, errors.New("item not found in wishlist")
		}
		return ItemResponse{}, fmt.Errorf("error retrieving item: %w", err)
	}

	// If the wishlist date has passed or is today, fetch the user who picked the item
	var wishlistDate string
	err = w.db.QueryRow(ctx, "SELECT date FROM wishlists WHERE id = $1", wishlistID).Scan(&wishlistDate)
	if err == nil && wishlistDate <= fmt.Sprintf("%v", sql.NullString{String: "CURRENT_DATE", Valid: true}) {
		pickQuery := `
		SELECT u.id, u.username, u.first_name, u.last_name
		FROM users u
		JOIN item_picks ip ON u.id = ip.user_id
		WHERE ip.item_id = $1 LIMIT 1;`

		err = w.db.QueryRow(ctx, pickQuery, itemID).Scan(&pickedBy.ID, &pickedBy.Username, &pickedBy.FirstName, &pickedBy.LastName)
		if err == nil {
			item.PickedBy = pickedBy
		}
	}

	return item, nil
}

func (w *WishlistStore) UpdateItem(wishlistID, itemID int, body *UpdateItem) (ItemResponse, error) {
	return ItemResponse{}, nil
}

func (w *WishlistStore) PickItem(wishlistID, itemID, userID int) (ItemResponse, error) {
	return ItemResponse{}, nil
}

func (w *WishlistStore) DeleteItem(wishlistID, itemID int) error {
	return nil
}
