package notification

import (
	"errors"
	"github.com/jackc/pgx/v5"
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
	return Notification{}, errors.New("not implemented")
}

func (s *NotificationStore) UpdateNotification(ID int, status string) (Notification, error) {
	return Notification{}, errors.New("not implemented")
}

func (s *NotificationStore) GetUserNotifications(userID int) ([]Notification, error) {
	return make([]Notification, 0), errors.New("not implemented")
}

func (s *NotificationStore) DeleteNotification(ID int) error {
	return errors.New("not implemented")
}

func (s *NotificationStore) GetNotification(ID int) (Notification, error) {
	return Notification{}, errors.New("not implemented")
}
