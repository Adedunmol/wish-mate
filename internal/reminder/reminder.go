package reminder

import (
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

type CreateReminderBody struct {
	Name      string     `json:"name"`
	UserID    int        `json:"user_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Type      string     `json:"type"`
	ExecuteAt *time.Time `json:"execute_at"`
}

type Store interface {
	CreateReminder(body CreateReminderBody) (ReminderResponse, error)
	GetReminders(currentTime *time.Time) ([]ReminderResponse, error)
	UpdateReminder(ID int) error
	DeleteReminder(ID int) error
}

type ReminderStore struct {
	DB *pgx.Conn
}

func (t *ReminderStore) DeleteReminder(ID int) error {
	//TODO implement me
	panic("implement me")
}

func (t *ReminderStore) CreateReminder(body CreateReminderBody) (ReminderResponse, error) {
	return ReminderResponse{}, nil
}

func (t *ReminderStore) GetReminders(currentTime *time.Time) ([]ReminderResponse, error) {

	return nil, nil
}

func (t *ReminderStore) UpdateReminder(ID int) error {
	return nil
}

type ReminderResponse struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	ExecuteAt *time.Time `json:"execute_at"`
}

func CreateReminder(store Store, body CreateReminderBody) (ReminderResponse, error) {

	if body.Name == "" {
		return ReminderResponse{}, errors.New("empty name")
	}

	if body.ExecuteAt == nil {
		return ReminderResponse{}, errors.New("executeAt is empty")
	}
	//
	//if body.Payload == nil {
	//	return ScheduledTaskResponse{}, errors.New("payload is empty")
	//}

	task, err := store.CreateReminder(body)
	if err != nil {
		return ReminderResponse{}, fmt.Errorf("error creating a task: %v", err)
	}

	return task, nil
}

func GetTasks(store Store, currentTime *time.Time) ([]ReminderResponse, error) {
	tasks, err := store.GetReminders(currentTime)
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

		err = store.UpdateReminder(task.ID)

		if err != nil {
			return fmt.Errorf("error updating task: %v", err)
		}
	}

	return nil
}

func DeleteTask(store Store, id int) error {
	err := store.DeleteReminder(id)

	if err != nil {
		return fmt.Errorf("error deleting task: %v", err)
	}

	return nil
}
