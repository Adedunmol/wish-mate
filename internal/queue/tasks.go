package queue

import (
	"context"
	"fmt"
	"github.com/hibiken/asynq"
)

type Task interface {
	NewTask() (*asynq.Task, error)
	HandleTask(ctx context.Context, t *asynq.Task) error
}

type TaskPayload struct {
	Type    string
	Payload map[string]interface{}
}

type Queue interface {
	Enqueue(taskPayload *TaskPayload) error
}

type Client struct {
	client *asynq.Client
}

func (qc *Client) Enqueue(taskPayload *TaskPayload) error {

	switch taskPayload.Type {
	case TypeEmailDelivery:
		emailPayload := EmailDeliveryPayload{
			Email:    taskPayload.Payload["email"].(string),
			Template: taskPayload.Payload["template"].(string),
			Subject:  taskPayload.Payload["subject"].(string),
			Data:     map[string]interface{}{},
		}

		task, err := emailPayload.NewTask()
		if err != nil {
			return fmt.Errorf("error creating new email task: %v", err)
		}

		_, err = qc.client.Enqueue(task)
		if err != nil {
			return fmt.Errorf("could not enqueue mail task for: %s: %v", emailPayload.Email, err)
		}
	}

	return nil
}
