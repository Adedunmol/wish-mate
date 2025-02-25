package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"log"
)

const TypeEmailDelivery = "mail:deliver"

type EmailDeliveryPayload struct {
	Template string
	Subject  string
	Email    string
	Data     interface{}
}

func (e *EmailDeliveryPayload) NewTask() (*asynq.Task, error) {
	payload, err := json.Marshal(e)

	if err != nil {
		return nil, fmt.Errorf("marshal email delivery payload: %w", err)
	}

	return asynq.NewTask(TypeEmailDelivery, payload), nil
}

func HandleEmailTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailDeliveryPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("error decoding email delivery payload: %w", err)
	}
	log.Printf("sending mail to user: %s", payload.Email)

	// send mail to user

	return nil
}
