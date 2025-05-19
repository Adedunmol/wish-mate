package friendship

import (
	"context"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/jackc/pgx/v5"
	"time"
)

type FriendStore interface {
	CreateFriendship(userID, recipientID int) (FriendshipResponse, error)
	UpdateFriendship(userID, friendID int, status string) (FriendshipResponse, error)
	GetAllFriendships(userID int, status string) ([]FriendshipResponse, error)
	GetFriendship(userID, friendID int) (FriendshipResponse, error)
}

type FriendshipStore struct {
	db *pgx.Conn
}

func NewFriendshipStore(db *pgx.Conn) *FriendshipStore {

	return &FriendshipStore{db: db}
}

func (f *FriendshipStore) CreateFriendship(userID, recipientID int) (FriendshipResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := f.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return FriendshipResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO friendships (user_id, friend_id, status, friend_since)
	VALUES ($1, $2, 'pending', NULL)
	ON CONFLICT (user_id, friend_id) DO NOTHING
	RETURNING user_id, friend_id, status, friend_since;
	`

	var friendship FriendshipResponse

	err = f.db.QueryRow(ctx, query, userID, recipientID).Scan(&friendship.UserID, &friendship.FriendID, &friendship.Status, &friendship.FriendSince)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {

			query = `
					SELECT user_id, friend_id, status, friend_since FROM friendships WHERE user_id = $1 AND friend_id = $2;
					`
			err = f.db.QueryRow(ctx, query, userID, recipientID).Scan(&friendship.UserID, &friendship.FriendID, &friendship.Status, &friendship.FriendSince)

			return friendship, nil // or fetch it from DB explicitly
		}

		err = errors.Join(helpers.ErrInternalServerError, err)
		return FriendshipResponse{}, fmt.Errorf("error inserting friendship: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return FriendshipResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return friendship, nil
}

func (f *FriendshipStore) UpdateFriendship(userID, friendID int, status string) (FriendshipResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	tx, err := f.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return FriendshipResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `SELECT user_id, friend_id, status FROM friendships WHERE user_id = $1 AND friend_id = $2;`
	var friendship FriendshipResponse

	err = f.db.QueryRow(ctx, query, userID, friendID).Scan(&friendship.UserID, &friendship.FriendID, &friendship.Status)
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return FriendshipResponse{}, fmt.Errorf("error getting friendship: %w", err)
	}

	if friendship.Status == "pending" {
		updateQuery := `
		UPDATE friendships SET status = $1, friend_since = $2 WHERE user_id = $3 AND friend_id = $4 RETURNING id, user_id, friend_id;
	`

		err = f.db.QueryRow(ctx, updateQuery, "accepted", time.Now(), userID, friendID).Scan(&friendship.UserID, &friendship.FriendID)
		if err != nil {
			err = errors.Join(helpers.ErrInternalServerError, err)
			return FriendshipResponse{}, fmt.Errorf("error updating friendship: %w", err)
		}

		insertQuery := `
	INSERT INTO friendships (user_id, friend_id, status, friend_since)
	VALUES ($1, $2, 'accepted', $3)
	ON CONFLICT (user_id, friend_id) DO NOTHING
	`

		_, err = f.db.Exec(ctx, insertQuery, &friendship.FriendID, &friendship.UserID, time.Now())

		if err != nil {
			err = errors.Join(helpers.ErrInternalServerError, err)
			return FriendshipResponse{}, fmt.Errorf("error inserting friendship: %w", err)
		}
	} else {
		updateQuery := `
		UPDATE friendships SET status = $1 WHERE user_id = $2 AND friend_id = $3 RETURNING id, user_id, friend_id, status;
	`

		err = f.db.QueryRow(ctx, updateQuery, status, userID, friendID).Scan(&friendship.UserID, &friendship.FriendID, &friendship.Status)
		if err != nil {
			err = errors.Join(helpers.ErrInternalServerError, err)
			return FriendshipResponse{}, fmt.Errorf("error updating friendship: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return FriendshipResponse{}, fmt.Errorf("commit tx: %w", err)
	}

	return friendship, nil
}

func (f *FriendshipStore) GetAllFriendships(userID int, status string) ([]FriendshipResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	tx, err := f.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return nil, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var query string
	var rows pgx.Rows
	if status == "all" {
		query = `SELECT user_id, friend_id, friend_since, status FROM friendships WHERE user_id = $1;`
		rows, err = f.db.Query(ctx, query, userID)
	} else {
		query = `SELECT user_id, friend_id, friend_since, status FROM friendships WHERE user_id = $1 AND status = $2;`
		rows, err = f.db.Query(ctx, query, userID, status)
	}

	var friendships []FriendshipResponse

	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return nil, fmt.Errorf("error querying friendships: %v", err)
	}

	for rows.Next() {
		var friendship FriendshipResponse

		err = rows.Scan(&friendship.UserID, &friendship.FriendID, &friendship.FriendSince, &friendship.Status)
		if err != nil {
			err = errors.Join(helpers.ErrInternalServerError, err)
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		friendships = append(friendships, friendship)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return friendships, nil
}

func (f *FriendshipStore) GetFriendship(userID, friendID int) (FriendshipResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	tx, err := f.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return FriendshipResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `SELECT user_id, friend_id, friend_since, status FROM friendships WHERE user_id = $1 AND friend_id = $2;`

	var friendship FriendshipResponse

	err = f.db.QueryRow(ctx, query, userID, friendID).Scan(&friendship.UserID, &friendship.FriendID, &friendship.FriendSince, &friendship.Status)
	if err != nil {
		return FriendshipResponse{}, fmt.Errorf("error getting friendship: %w", err)
	}

	return friendship, nil
}
