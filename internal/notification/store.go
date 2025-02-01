package notification

type Store interface {
	CreateNotification(body *CreateNotificationBody) (Notification, error)
	UpdateNotification(ID int, status string) (Notification, error)
	GetNotification(ID int) (Notification, error)
	GetUserNotifications(userID int) ([]Notification, error)
	DeleteNotification(ID int) error
}
