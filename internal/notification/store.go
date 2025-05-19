package notification

import (
	"context"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/jackc/pgx/v5"
	"time"
)

type Store interface {
	CreateNotification(body *CreateNotificationBody) (Notification, error)
	UpdateNotification(ID int, status string) (Notification, error)
	GetNotification(ID int) (Notification, error)
	GetUserNotifications(userID int) ([]Notification, error)
	DeleteNotification(ID int) error
}

type NotificationStore struct {
	db *pgx.Conn
}

func NewNotificationStore(db *pgx.Conn) *NotificationStore {

	return &NotificationStore{db: db}
}

func (s *NotificationStore) CreateNotification(body *CreateNotificationBody) (Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return Notification{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO notifications (user_id, title, body, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, title, body, type, read
	`

	var notification Notification
	err = tx.QueryRow(ctx, query, body.UserID, body.Title, body.Body, body.Type).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Title,
		&notification.Body,
		&notification.Type,
		&notification.Read,
	)
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return Notification{}, fmt.Errorf("insert notification: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return Notification{}, fmt.Errorf("commit tx: %w", err)
	}

	return notification, nil
}

func (s *NotificationStore) UpdateNotification(ID int, status string) (Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return Notification{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var readValue bool
	switch status {
	case "read":
		readValue = true
	case "unread":
		readValue = false
	default:
		return Notification{}, fmt.Errorf("invalid status: %s", status)
	}

	query := `
		UPDATE notifications
		SET read = $1
		WHERE id = $2
		RETURNING id, user_id, title, body, type, read
	`

	var notification Notification
	err = tx.QueryRow(ctx, query, readValue, ID).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Title,
		&notification.Body,
		&notification.Type,
		&notification.Read,
	)

	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return Notification{}, fmt.Errorf("update notification: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return Notification{}, fmt.Errorf("commit tx: %w", err)
	}

	return notification, nil
}

func (s *NotificationStore) GetUserNotifications(userID int) ([]Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return nil, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		SELECT id, user_id, title, body, type, read
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	notifications := make([]Notification, 0)
	for rows.Next() {
		var notification Notification
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Body,
			&notification.Type,
			&notification.Read,
		)
		if err != nil {
			err = errors.Join(helpers.ErrInternalServerError, err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		notifications = append(notifications, notification)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return notifications, nil
}

func (s *NotificationStore) DeleteNotification(ID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		DELETE FROM notifications
		WHERE id = $1
	`

	_, err = tx.Exec(ctx, query, ID)
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return fmt.Errorf("query error: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (s *NotificationStore) GetNotification(ID int) (Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, title, body, type, read
		FROM notifications
		WHERE id = $1
	`

	var notification Notification
	err := s.db.QueryRow(ctx, query, ID).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Title,
		&notification.Body,
		&notification.Type,
		&notification.Read,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Notification{}, helpers.ErrNotFound
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return Notification{}, fmt.Errorf("query error: %w", err)
	}

	return notification, nil
}
