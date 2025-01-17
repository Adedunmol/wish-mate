package queue

import "context"

type Task interface {
	NewTask() (*asynq, error)
	HandleTask(ctx context.Context, t *asynq.Task) error
}
