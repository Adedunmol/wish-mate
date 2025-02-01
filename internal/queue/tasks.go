package queue

import (
	"context"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/notification"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"sync"
)

type Task interface {
	NewTask() (*asynq.Task, error)
	//HandleTask(ctx context.Context, t *asynq.Task) error
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
	once   sync.Once
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
		break
	case TypeNotificationDelivery:
		notificationPayload := NotificationDeliveryPayload{
			ID:     taskPayload.Payload["id"].(int),
			UserID: taskPayload.Payload["user_id"].(int),
			Title:  taskPayload.Payload["title"].(string),
			Body:   taskPayload.Payload["body"].(string),
			Type:   taskPayload.Payload["type"].(string),
		}

		task, err := notificationPayload.NewTask()
		if err != nil {
			return fmt.Errorf("error creating new email task: %v", err)
		}

		_, err = qc.client.Enqueue(task)
		if err != nil {
			return fmt.Errorf("could not enqueue notification task for: notification_id = %d, user_id = %d: %v", notificationPayload.ID, notificationPayload.UserID, err)
		}
		break
	}

	return nil
}

func NewClient(ctx context.Context) (*Client, error) {
	var qc Client
	addr, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return &qc, fmt.Errorf("error parsing redis url: %v", err)
	}

	qc.once.Do(func() {
		log.Printf("setting up connection for asynq redis queue")

		qc.client = asynq.NewClient(asynq.RedisClientOpt{Addr: addr.Addr, Password: "", DB: 0})
	})

	return &qc, nil
}

func (qc *Client) GetClient() *asynq.Client {
	return qc.client
}

func (qc *Client) Close() error {
	log.Println("closing connection to asynq queue")
	return fmt.Errorf("error closing connection: %v", qc.client.Close())
}

func (qc *Client) Run(ctx context.Context, db *pgx.Conn) error {
	addr, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return fmt.Errorf("error parsing redis url: %v", err)
	}

	queueServer := asynq.NewServer(asynq.RedisClientOpt{Addr: addr.Addr}, asynq.Config{})

	mux := asynq.NewServeMux()

	mux.HandleFunc(TypeEmailDelivery, HandleEmailTask)
	mux.HandleFunc(TypeNotificationDelivery, WrapHandler(notification.NewNotificationStore(db)))

	if err := queueServer.Run(mux); err != nil {
		return fmt.Errorf("error running queue server: %v", err)
	}
	return nil
}
