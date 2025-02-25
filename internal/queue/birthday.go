package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"log"
)

const TypeBirthdayMailDelivery = "birthday:mail"

type BirthdayMailPayload struct {
	Template string
	Subject  string
	Email    string
	Data     interface{}
}

func (e *BirthdayMailPayload) NewTask() (*asynq.Task, error) {
	payload, err := json.Marshal(e)

	if err != nil {
		return nil, fmt.Errorf("marshal birthday email delivery payload: %w", err)
	}

	return asynq.NewTask(TypeBirthdayMailDelivery, payload), nil
}

func HandleBirthdayMailTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailDeliveryPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("error decoding birthday email delivery payload: %w", err)
	}
	log.Printf("sending mails to user's friends: %s", payload.Email)

	// get user's friends and send mails to them

	return nil
}
