package friendship

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

type FriendStore interface {
	CreateFriendship(userID, recipientID int) (FriendshipResponse, error)
	UpdateFriendship(friendshipID int, status string) (FriendshipResponse, error)
	GetAllFriendships(userID int, status string) ([]FriendshipResponse, error)
	GetFriendship(requestID int) (FriendshipResponse, error)
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
		return FriendshipResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO friendships (user_id, friend_id, status, friend_since)
	VALUES ($1, $2, 'pending', NULL)
	ON CONFLICT (user_id, friend_id) DO NOTHING
	RETURNING id, user_id, friend_id, status, friend_since;
	`

	var friendship FriendshipResponse

	err = f.db.QueryRow(ctx, query, userID, recipientID).Scan(&friendship.ID, &friendship.UserID, &friendship.FriendID, &friendship.Status, &friendship.FriendSince)

	if err != nil {
		return FriendshipResponse{}, fmt.Errorf("error inserting friendship: %w", err)
	}

	return friendship, nil
}

func (f *FriendshipStore) UpdateFriendship(friendshipID int, status string) (FriendshipResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	tx, err := f.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return FriendshipResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `SELECT user_id, friend_id, status FROM friendships WHERE id = $1;`
	var friendship FriendshipResponse

	err = f.db.QueryRow(ctx, query, friendshipID).Scan(&friendship.UserID, &friendship.FriendID, &friendship.Status)
	if err != nil {
		return FriendshipResponse{}, fmt.Errorf("error getting friendship: %w", err)
	}

	if friendship.Status == "pending" {
		updateQuery := `
		UPDATE friendships SET status = $1, friend_since = $2 WHERE id = $3 RETURNING id, user_id, friend_id;
	`

		err = f.db.QueryRow(ctx, updateQuery, "accepted", time.Now(), friendshipID).Scan(&friendship.ID, &friendship.UserID, &friendship.FriendID)
		if err != nil {
			return FriendshipResponse{}, fmt.Errorf("error updating friendship: %w", err)
		}

		insertQuery := `
	INSERT INTO friendships (user_id, friend_id, status, friend_since)
	VALUES ($1, $2, 'accepted', $3)
	ON CONFLICT (user_id, friend_id) DO NOTHING
	`

		_, err = f.db.Exec(ctx, insertQuery, &friendship.FriendID, &friendship.UserID, time.Now())

		if err != nil {
			return FriendshipResponse{}, fmt.Errorf("error inserting friendship: %w", err)
		}
	} else {
		updateQuery := `
		UPDATE friendships SET status = $1 WHERE id = $2 RETURNING id, user_id, friend_id, status;
	`

		err = f.db.QueryRow(ctx, updateQuery, status, friendshipID).Scan(&friendship.ID, &friendship.UserID, &friendship.FriendID, &friendship.Status)
		if err != nil {
			return FriendshipResponse{}, fmt.Errorf("error updating friendship: %w", err)
		}

	}

	return friendship, nil
}

func (f *FriendshipStore) GetAllFriendships(userID int, status string) ([]FriendshipResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	tx, err := f.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `SELECT id, user_id, friend_id, friend_since FROM frienships WHERE user_id = $1 AND status = $2;`
	var friendships []FriendshipResponse

	rows, err := f.db.Query(ctx, query, userID, status)

	if err != nil {
		return nil, fmt.Errorf("error querying friendhips: %v", err)
	}

	for rows.Next() {
		var friendship FriendshipResponse

		err = rows.Scan(&friendship.ID, &friendship.UserID, &friendship.FriendID, &friendship.FriendSince)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		friendships = append(friendships, friendship)
	}

	return friendships, nil
}

func (f *FriendshipStore) GetFriendship(requestID int) (FriendshipResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	tx, err := f.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return FriendshipResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `SELECT id, user_id, friend_id, friend_since FROM frienships WHERE id = $1;`

	var friendship FriendshipResponse

	err = f.db.QueryRow(ctx, query, requestID).Scan(&friendship.ID, &friendship.UserID, &friendship.FriendID, &friendship.FriendSince)
	if err != nil {
		return FriendshipResponse{}, fmt.Errorf("error getting friendship: %w", err)
	}

	return friendship, nil
}
