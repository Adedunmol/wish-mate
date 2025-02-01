package notification

type Store interface {
	CreateNotification(body *CreateNotificationBody) (Notification, error)
	UpdateNotification(ID int, status string) (Notification, error)
	DeleteNotification(ID int) error
}
