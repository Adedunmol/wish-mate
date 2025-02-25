package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/notification"
	"github.com/hibiken/asynq"
	"log"
)

const TypeNotificationDelivery = "notification:deliver"

type NotificationDeliveryPayload struct {
	ID     int
	UserID int
	Title  string
	Body   string
	Type   string
}

func (e *NotificationDeliveryPayload) NewTask() (*asynq.Task, error) {
	payload, err := json.Marshal(e)

	if err != nil {
		return nil, fmt.Errorf("marshal notification delivery payload: %w", err)
	}

	return asynq.NewTask(TypeNotificationDelivery, payload), nil
}

func WrapHandler(store notification.Store) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {

		var payload NotificationDeliveryPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("error decoding notification delivery payload: %w", err)
		}
		log.Printf("creating notification %d for: %d", payload.ID, payload.UserID)

		// send in a collection of the user's friends from sql query

		_, err := store.CreateNotification(&notification.CreateNotificationBody{
			UserID: payload.UserID,
			Title:  payload.Title,
			Body:   payload.Body,
			Type:   payload.Type,
		})

		if err != nil {
			return fmt.Errorf("error creating notification: %w", err)
		}

		return nil
	}
}
