package queue

import (
	"context"
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
