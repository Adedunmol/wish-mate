package scheduled_tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

type CreateTaskBody struct {
	Name      string     `json:"name"`
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Type      string     `json:"type"`
	ExecuteAt *time.Time `json:"execute_at"`
}

type Store interface {
	CreateTask(body CreateTaskBody) (ScheduledTaskResponse, error)
	GetTasks(currentTime *time.Time) ([]ScheduledTaskResponse, error)
	UpdateTask(ID int) error
	DeleteTask(ID int) error
}

type TaskStore struct {
	DB *pgx.Conn
}

func (t *TaskStore) DeleteTask(ID int) error {
	//TODO implement me
	panic("implement me")
}

func (t *TaskStore) CreateTask(body CreateTaskBody) (ScheduledTaskResponse, error) {
	return ScheduledTaskResponse{}, nil
}

func (t *TaskStore) GetTasks(currentTime *time.Time) ([]ScheduledTaskResponse, error) {

	return nil, nil
}

func (t *TaskStore) UpdateTask(ID int) error {
	return nil
}

type ScheduledTaskResponse struct {
	ID        int             `json:"id"`
	TaskName  string          `json:"task_name"`
	Payload   json.RawMessage `json:"payload"`
	Status    string          `json:"status"`
	ExecuteAt time.Time       `json:"execute_at"`
}

func CreateTask(store Store, body CreateTaskBody) (ScheduledTaskResponse, error) {

	if body.Name == "" {
		return ScheduledTaskResponse{}, errors.New("empty name")
	}

	if body.ExecuteAt == nil {
		return ScheduledTaskResponse{}, errors.New("executeAt is empty")
	}
	//
	//if body.Payload == nil {
	//	return ScheduledTaskResponse{}, errors.New("payload is empty")
	//}

	task, err := store.CreateTask(body)
	if err != nil {
		return ScheduledTaskResponse{}, fmt.Errorf("error creating a task: %v", err)
	}

	return task, nil
}

func GetTasks(store Store, currentTime *time.Time) ([]ScheduledTaskResponse, error) {
	tasks, err := store.GetTasks(currentTime)
	if err != nil {
		return nil, fmt.Errorf("error getting tasks: %v", err)
	}

	return tasks, nil
}

func GetTasksAndEnqueue(store Store, q queue.Queue, currentTime *time.Time) error {

	tasks, err := GetTasks(store, currentTime)
	if err != nil {
		return fmt.Errorf("error getting tasks: %v", err)
	}

	for _, task := range tasks {

		err = q.Enqueue(&queue.TaskPayload{
			Type: queue.TypeNotificationDelivery,
			Payload: map[string]interface{}{
				"id":      "",
				"user_id": "",
				"title":   "",
				"body":    "",
				"type":    "",
			},
		})
		if err != nil {
			log.Printf("error enqueuing scheduled task: %s : %v", err, task)
		}

		err = store.UpdateTask(task.ID)

		if err != nil {
			return fmt.Errorf("error updating task: %v", err)
		}
	}

	return nil
}

func DeleteTask(store Store, id int) error {
	err := store.DeleteTask(id)

	if err != nil {
		return fmt.Errorf("error deleting task: %v", err)
	}

	return nil
}
