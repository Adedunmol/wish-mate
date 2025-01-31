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
	store := &StubStore{tasks: make([]scheduled_tasks.ScheduledTaskResponse, 0)}

	t.Run("create and return task", func(t *testing.T) {
		currentTime := time.Now()

		task, _ := scheduled_tasks.CreateTask(store, "birthday", []byte(``), &currentTime)

		if task.Status != "pending" {
			t.Error("task status should be pending")
		}

		if task.TaskName != "birthday" {
			t.Error("task name should be birthday")
		}

		if task.Payload == nil {
			t.Error("payload is empty")
		}

		if task.ExecuteAt != currentTime {
			t.Errorf("task executeAt should be %v", currentTime)
		}
	})

	t.Run("return error for invalid task body", func(t *testing.T) {
		_, err := scheduled_tasks.CreateTask(store, "", []byte(``), &time.Time{})

		if err == nil {
			t.Error("error should not be nil")
		}

		if err.Error() != "empty name" {
			t.Error("error should be 'empty name'")
		}
	})

}

func TestGetTasks(t *testing.T) {

	t.Run("return tasks that are before the current time (with pending status)", func(t *testing.T) {
		store := &StubStore{tasks: []scheduled_tasks.ScheduledTaskResponse{
			{ID: 1, TaskName: "birthday", Payload: []byte(`{"title": "some random title"}`), ExecuteAt: time.Now().Add(10 * time.Minute), Status: "pending"},
			{ID: 2, TaskName: "birthday", Payload: []byte(`{"title": "some random title"}`), ExecuteAt: time.Now().Add(-(1 * time.Minute)), Status: "pending"},
			{ID: 3, TaskName: "birthday", Payload: []byte(`{"title": "some random title"}`), ExecuteAt: time.Now().Add(-(1 * time.Minute)), Status: "pending"},
			{ID: 4, TaskName: "birthday", Payload: []byte(`{"title": "some random title"}`), ExecuteAt: time.Now().Add(-(1 * time.Minute)), Status: "scheduled"},
		}}

		currentTime := time.Now()
		tasks, _ := scheduled_tasks.GetTasks(store, &currentTime)

		if len(tasks) != 2 {
			t.Error("tasks should have two tasks")
		}
	})

	t.Run("return no tasks that are after the current time", func(t *testing.T) {
		store := &StubStore{tasks: []scheduled_tasks.ScheduledTaskResponse{
			{ID: 1, TaskName: "birthday", Payload: []byte(`{"title": "some random title"}`), ExecuteAt: time.Now().Add(10 * time.Minute), Status: "pending"},
			{ID: 2, TaskName: "birthday", Payload: []byte(`{"title": "some random title"}`), ExecuteAt: time.Now().Add(1 * time.Minute), Status: "pending"},
			{ID: 4, TaskName: "birthday", Payload: []byte(`{"title": "some random title"}`), ExecuteAt: time.Now().Add(-(1 * time.Minute)), Status: "scheduled"},
		}}

		currentTime := time.Now()
		tasks, _ := scheduled_tasks.GetTasks(store, &currentTime)

		if len(tasks) != 0 {
			t.Error("tasks should have no tasks")
		}
	})
}
