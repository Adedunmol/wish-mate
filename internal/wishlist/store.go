package wishlist

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	CreateItem(userID, wishlistID int, body *Item) (ItemResponse, error)
}

type WishlistStore struct {
	db *pgx.Conn
}

const UniqueViolation = "23505"

func NewWishlistStore(db *pgx.Conn) *WishlistStore {

	return &WishlistStore{db: db}
}

func (w *WishlistStore) CreateWishlist(userID int, body Wishlist) (WishlistResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	//date, _ := time.Parse("2006-01-02", body.Date)

	var wishlist WishlistResponse

	query := `INSERT INTO wishlists (created_by, name, description, notify_before, date) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_by, name, description, notify_before, date;`

	err = w.db.QueryRow(ctx, query, userID, body.Name, body.Description, body.NotifyBefore, body.Date).
		Scan(&wishlist.ID, &wishlist.UserID, &wishlist.Name, &wishlist.Description, &wishlist.NotifyBefore, &wishlist.Date)
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("error inserting wishlist: %w", err)
	}

	wishlist.Items = make([]ItemResponse, 0)

	insertItemQuery := `INSERT INTO items (wishlist_id, name, description, link, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, description, link;`
	for _, item := range body.Items {
		var newItem ItemResponse
		err = w.db.QueryRow(ctx, insertItemQuery, wishlist.ID, item.Name, item.Description, item.Link, userID).
			Scan(&newItem.ID, &newItem.Name, &newItem.Description, &newItem.Link)

		if err != nil {
			err = errors.Join(helpers.ErrInternalServerError, err)
			return WishlistResponse{}, fmt.Errorf("error inserting item: %w", err)
		}
		wishlist.Items = append(wishlist.Items, newItem)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return wishlist, nil
}

var ErrNoWishlistFound = errors.New("no wishlist found")

func (w *WishlistStore) GetWishlistByID(wishlistID, userID int) (WishlistResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var wishlist WishlistResponse

	query := `SELECT id, created_by, name, description, notify_before, date 
		FROM wishlists WHERE id = $1;`
	err = tx.QueryRow(ctx, query, wishlistID).
		Scan(&wishlist.ID, &wishlist.UserID, &wishlist.Name, &wishlist.Description, &wishlist.NotifyBefore, &wishlist.Date)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WishlistResponse{}, ErrNoWishlistFound
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("error fetching wishlist: %w", err)
	}

	isOwner := userID == wishlist.UserID
	isPast := wishlist.Date.Before(time.Now())

	items, err := fetchItemsForWishlist(ctx, tx, wishlist.ID, isOwner, isPast)
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("fetch items: %w", err)
	}
	wishlist.Items = items

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return wishlist, nil
}

func (w *WishlistStore) GetUserWishlists(userID int, isOwner bool) ([]WishlistResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return nil, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Step 1: Fetch all wishlists
	query := `SELECT id, created_by, name, description, notify_before, date FROM wishlists WHERE created_by = $1;`
	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return nil, fmt.Errorf("error fetching wishlists: %w", err)
	}
	defer rows.Close()

	type wishlistData struct {
		Wishlist WishlistResponse
	}

	var all []wishlistData

	for rows.Next() {
		var wl WishlistResponse
		err := rows.Scan(&wl.ID, &wl.UserID, &wl.Name, &wl.Description, &wl.NotifyBefore, &wl.Date)
		if err != nil {
			err = errors.Join(helpers.ErrInternalServerError, err)
			return nil, fmt.Errorf("error scanning wishlist: %w", err)
		}
		wl.Items = []ItemResponse{}
		all = append(all, wishlistData{Wishlist: wl})
	}

	// Step 2: Loop and fetch items for each wishlist
	for i := range all {
		wishlist := &all[i].Wishlist
		isPast := wishlist.Date.Before(time.Now())

		items, err := fetchItemsForWishlist(ctx, tx, wishlist.ID, isOwner, isPast)
		if err != nil {
			err = errors.Join(helpers.ErrInternalServerError, err)
			return nil, fmt.Errorf("fetch items: %w", err)
		}
		wishlist.Items = items
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	var result []WishlistResponse
	for _, w := range all {
		result = append(result, w.Wishlist)
	}
	return result, nil
}

func (w *WishlistStore) UpdateWishlistByID(wishlistID, userID int, body UpdateWishlist) (WishlistResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var wishlist WishlistResponse

	// Check if the user is the owner of the wishlist
	var ownerID int
	err = w.db.QueryRow(ctx, "SELECT created_by FROM wishlists WHERE id = $1", wishlistID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WishlistResponse{}, helpers.ErrNotFound
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("error checking wishlist ownership: %w", err)
	}

	if ownerID != userID {
		return WishlistResponse{}, helpers.ErrForbidden
	}

	// Update the wishlist with non-empty fields
	query := `UPDATE wishlists SET 
		name = COALESCE(NULLIF($1, ''), name),
		description = COALESCE(NULLIF($2, ''), description)
		WHERE id = $3 RETURNING id, created_by, name, description, notify_before, date;`

	err = w.db.QueryRow(ctx, query, body.Name, body.Description, wishlistID).Scan(&wishlist.ID, &wishlist.UserID, &wishlist.Name, &wishlist.Description, &wishlist.NotifyBefore, &wishlist.Date)
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("error updating wishlist: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return WishlistResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return wishlist, nil
}

func (w *WishlistStore) DeleteWishlistByID(wishlistID, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if the user is the owner of the wishlist
	var ownerID int
	err = w.db.QueryRow(ctx, "SELECT created_by FROM wishlists WHERE id = $1", wishlistID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helpers.ErrNotFound
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return fmt.Errorf("error checking wishlist ownership: %w", err)
	}

	if ownerID != userID {
		return helpers.ErrForbidden
	}

	// Delete the wishlist
	_, err = w.db.Exec(ctx, "DELETE FROM wishlists WHERE id = $1", wishlistID)
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return fmt.Errorf("error deleting wishlist: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (w *WishlistStore) GetItem(wishlistID, itemID int) (ItemResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
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
		if errors.Is(err, pgx.ErrNoRows) {
			return ItemResponse{}, helpers.ErrNotFound
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("error retrieving item: %w", err)
	}

	itemResponse := ItemResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Link:        item.Link,
		Taken:       item.Taken,
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
			item.PickedBy = &pickedBy
		}
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return itemResponse, nil
}

func (w *WishlistStore) UpdateItem(wishlistID, itemID int, body *UpdateItem) (ItemResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var item ItemResponse

	query := `
	UPDATE items SET 
		name = COALESCE(NULLIF($1, ''), name),
		description = COALESCE(NULLIF($2, ''), description),
		link = COALESCE(NULLIF($3, ''), link)
	WHERE id = $4 AND wishlist_id = $5
	RETURNING id, name, description, link;`

	err = w.db.QueryRow(ctx, query, body.Name, body.Description, body.Link, itemID, wishlistID).
		Scan(&item.ID, &item.Name, &item.Description, &item.Link)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ItemResponse{}, helpers.ErrNotFound
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("error updating item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return item, nil
}

func (w *WishlistStore) PickItem(wishlistID, itemID, userID int) (ItemResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var item ItemResponse

	// Ensure item is not already picked
	var existingItemID int

	err = tx.QueryRow(ctx, "SELECT id FROM items WHERE id = $1 AND wishlist_id = $2", itemID, wishlistID).Scan(&existingItemID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ItemResponse{}, helpers.ErrNotFound
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("error checking item status: %w", err)
	}

	pickedQuery := `SELECT item_id FROM item_picks WHERE item_id = $1 LIMIT 1;`

	err = tx.QueryRow(ctx, pickedQuery, itemID).Scan(&existingItemID)
	if err == nil {
		return ItemResponse{}, helpers.ErrConflict
	}

	query := `INSERT INTO item_picks (item_id, user_id) VALUES ($1, $2);`

	_, err = tx.Exec(ctx, query, itemID, userID)
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == UniqueViolation {
			return ItemResponse{}, helpers.ErrConflict
		}
		return ItemResponse{}, fmt.Errorf("error picking item: %w", err)
	}

	query = `UPDATE items SET taken = true WHERE id = $1 RETURNING id, name, description, link, taken;`

	err = tx.QueryRow(ctx, query, itemID).Scan(&item.ID, &item.Name, &item.Description, &item.Link, &item.Taken)
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("error updating item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return item, nil
}

func (w *WishlistStore) DeleteItem(wishlistID, itemID int) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := "DELETE FROM items WHERE id = $1 AND wishlist_id = $2"
	result, err := w.db.Exec(ctx, query, itemID, wishlistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helpers.ErrNotFound
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return fmt.Errorf("error deleting item: %w", err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return helpers.ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (w *WishlistStore) CreateItem(userID, wishlistID int, body *Item) (ItemResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var item ItemResponse

	query := `INSERT INTO items (name, description, link, created_by, wishlist_id) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id, name, description, link, wishlist_id;`

	err = tx.QueryRow(ctx, query, body.Name, body.Description, body.Link, userID, wishlistID).Scan(&item.ID, &item.Name, &item.Description, &item.Link, &item.WishlistID)
	if err != nil {
		return ItemResponse{}, fmt.Errorf("error inserting wishlist: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return ItemResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return item, nil
}

func fetchItemsForWishlist(
	ctx context.Context,
	tx pgx.Tx,
	wishlistID int,
	isOwner bool,
	isPast bool,
) ([]ItemResponse, error) {
	var items []ItemResponse
	var query string
	includePickedInfo := false

	if isOwner && isPast {
		query = `
			SELECT i.id, i.name, i.description, i.link, u.id, u.username, u.first_name, u.last_name
			FROM items i
			LEFT JOIN item_picks ip ON i.id = ip.item_id
			LEFT JOIN users u ON ip.user_id = u.id
			WHERE i.wishlist_id = $1;`
		includePickedInfo = true
	} else if isOwner {
		query = `SELECT id, name, description, link FROM items WHERE wishlist_id = $1;`
	} else {
		query = `
			SELECT id, name, description, link FROM items
			WHERE wishlist_id = $1 AND id NOT IN (SELECT item_id FROM item_picks);`
	}

	rows, err := tx.Query(ctx, query, wishlistID)
	if err != nil {
		return nil, fmt.Errorf("error fetching items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item ItemResponse
		if includePickedInfo {
			var pickedByID sql.NullInt64
			var username, firstName, lastName sql.NullString

			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Link,
				&pickedByID, &username, &firstName, &lastName); err != nil {
				return nil, fmt.Errorf("error scanning item with picked info: %w", err)
			}

			if pickedByID.Valid {
				item.PickedBy = &auth.User{
					ID:        int(pickedByID.Int64),
					Username:  username.String,
					FirstName: firstName.String,
					LastName:  lastName.String,
				}
			}
		} else {
			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Link); err != nil {
				return nil, fmt.Errorf("error scanning item: %w", err)
			}
		}

		items = append(items, item)
	}

	return items, nil
}
