package scheduled_tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"log"
	"time"
)

type Store interface {
	CreateTask(name string, payload json.RawMessage, executeAt *time.Time) (ScheduledTaskResponse, error)
	GetTasks(currentTime *time.Time) ([]ScheduledTaskResponse, error)
	UpdateTask(ID int) error
}

type ScheduledTaskResponse struct {
	ID        int             `json:"id"`
	TaskName  string          `json:"task_name"`
	Payload   json.RawMessage `json:"payload"`
	Status    string          `json:"status"`
	ExecuteAt time.Time       `json:"execute_at"`
}

func CreateTask(store Store, name string, payload json.RawMessage, executeAt *time.Time) (ScheduledTaskResponse, error) {

	if name == "" {
		return ScheduledTaskResponse{}, errors.New("empty name")
	}

	if executeAt == nil {
		return ScheduledTaskResponse{}, errors.New("executeAt is empty")
	}

	if payload == nil {
		return ScheduledTaskResponse{}, errors.New("payload is empty")
	}

	task, err := store.CreateTask(name, payload, executeAt)
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
			Type:    queue.TypeEmailDelivery,
			Payload: map[string]interface{}{},
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
