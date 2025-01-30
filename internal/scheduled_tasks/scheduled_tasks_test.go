package scheduled_tasks_test

import (
	"encoding/json"
	"errors"
	"github.com/Adedunmol/wish-mate/internal/scheduled_tasks"
	"testing"
	"time"
)

type StubStore struct {
	tasks []scheduled_tasks.ScheduledTaskResponse
}

func (s *StubStore) CreateTask(name string, payload json.RawMessage, executeAt *time.Time) (scheduled_tasks.ScheduledTaskResponse, error) {

	if name == "" {
		return scheduled_tasks.ScheduledTaskResponse{}, errors.New("empty name")
	}

	if executeAt == nil {
		return scheduled_tasks.ScheduledTaskResponse{}, errors.New("executeAt is empty")
	}

	if payload == nil {
		return scheduled_tasks.ScheduledTaskResponse{}, errors.New("payload is empty")
	}

	data := scheduled_tasks.ScheduledTaskResponse{
		ID:        1,
		TaskName:  name,
		Payload:   payload,
		ExecuteAt: *executeAt,
		Status:    "pending",
	}

	s.tasks = append(s.tasks, data)

	return data, nil
}

func (s *StubStore) GetTasks(currentTime *time.Time) ([]scheduled_tasks.ScheduledTaskResponse, error) {

	var result []scheduled_tasks.ScheduledTaskResponse

	for _, task := range s.tasks {
		if (task.ExecuteAt.Before(*currentTime) || task.ExecuteAt.Equal(*currentTime)) && task.Status == "pending" {
			result = append(result, task)
		}
	}

	return result, nil
}

func TestCreateTask(t *testing.T) {

	t.Run("create and return task", func(t *testing.T) {})

	t.Run("return error for invalid task body", func(t *testing.T) {})

}

func TestGetTasks(t *testing.T) {

	t.Run("return tasks that are before the current time (with pending status)", func(t *testing.T) {})

	t.Run("return no tasks that are after the current time", func(t *testing.T) {})

}
