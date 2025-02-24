package reminder_test

import (
	"errors"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/reminder"
	"testing"
	"time"
)

type StubStore struct {
	reminders []reminder.ReminderResponse
}

func (s *StubStore) CreateReminder(body reminder.CreateReminderBody) (reminder.ReminderResponse, error) {

	if body.Name == "" {
		return reminder.ReminderResponse{}, errors.New("empty name")
	}

	if body.ExecuteAt == nil {
		return reminder.ReminderResponse{}, errors.New("executeAt is empty")
	}

	//if body.Payload == nil {
	//	return reminder.ReminderResponse{}, errors.New("payload is empty")
	//}

	data := reminder.ReminderResponse{
		ID:        1,
		ExecuteAt: body.ExecuteAt,
		UserID:    body.UserID,
		Title:     body.Title,
		Body:      body.Body,
		Type:      body.Type,
		Status:    "pending",
	}

	s.reminders = append(s.reminders, data)

	return data, nil
}

func (s *StubStore) GetReminders(currentTime *time.Time) ([]reminder.ReminderResponse, error) {

	var result []reminder.ReminderResponse

	for _, r := range s.reminders {
		if (r.ExecuteAt.Before(*currentTime) || r.ExecuteAt.Equal(*currentTime)) && r.Status == "pending" {
			result = append(result, r)
		}
	}

	return result, nil
}

func (s *StubStore) GetBirthdays(currentTime *time.Time) ([]reminder.ReminderResponse, error) {
	return nil, nil
}

func (s *StubStore) UpdateReminder(id int) error {
	return nil
}

func (s *StubStore) DeleteReminder(id int) error {

	for index, r := range s.reminders {
		if r.ID == id {
			s.reminders = append(s.reminders[:index], s.reminders[index+1:]...)
			return nil
		}
	}

	return helpers.ErrNotFound
}

func TestCreateReminder(t *testing.T) {
	store := &StubStore{reminders: make([]reminder.ReminderResponse, 0)}

	t.Run("create and return task", func(t *testing.T) {
		currentTime := time.Now()
		executeAt := time.Now().Add(10 * time.Minute)

		body := reminder.CreateReminderBody{
			UserID:    1,
			ExecuteAt: &executeAt,
			Title:     "birthday",
			Body:      "some random text",
			Type:      "birthday",
		}

		task, _ := reminder.CreateReminder(store, body)

		if task.Status != "pending" {
			t.Error("task status should be pending")
		}

		if task.Title != "birthday" {
			t.Error("task name should be birthday")
		}

		if task.ExecuteAt != &executeAt {
			t.Errorf("task executeAt should be %v", currentTime)
		}
	})

	t.Run("return error for invalid task body", func(t *testing.T) {
		executeAt := time.Now().Add(10 * time.Minute)
		body := reminder.CreateReminderBody{
			UserID:    1,
			ExecuteAt: &executeAt,
			Title:     "birthday",
			Body:      "some random text",
			Type:      "birthday",
		}

		_, err := reminder.CreateReminder(store, body)

		if err == nil {
			t.Error("error should not be nil")
		}

		if err.Error() != "empty name" {
			t.Error("error should be 'empty name'")
		}
	})

}

func TestGetReminders(t *testing.T) {

	t.Run("return tasks that are before the current time (with pending status)", func(t *testing.T) {
		futureTime := time.Now().Add(10 * time.Minute)
		pastTime := time.Now().Add(-(1 * time.Minute))

		store := &StubStore{reminders: []reminder.ReminderResponse{
			{ID: 1, Title: "birthday", Body: "some random text", ExecuteAt: &futureTime, Status: "pending"},
			{ID: 2, Title: "birthday", Body: "some random text", ExecuteAt: &pastTime, Status: "pending"},
			{ID: 3, Title: "birthday", Body: "some random text", ExecuteAt: &pastTime, Status: "pending"},
			{ID: 4, Title: "birthday", Body: "some random text", ExecuteAt: &pastTime, Status: "scheduled"},
		}}

		currentTime := time.Now()
		tasks, _ := reminder.GetReminders(store, &currentTime)

		if len(tasks) != 2 {
			t.Error("tasks should have two tasks")
		}
	})

	t.Run("return no tasks that are after the current time", func(t *testing.T) {
		futureTime := time.Now().Add(10 * time.Minute)
		pastTime := time.Now().Add(-(1 * time.Minute))

		store := &StubStore{reminders: []reminder.ReminderResponse{
			{ID: 1, Title: "birthday", Body: "some random text", ExecuteAt: &futureTime, Status: "pending"},
			{ID: 2, Title: "birthday", Body: "some random text", ExecuteAt: &futureTime, Status: "pending"},
			{ID: 4, Title: "birthday", Body: "some random text", ExecuteAt: &pastTime, Status: "scheduled"},
		}}

		currentTime := time.Now()
		tasks, _ := reminder.GetReminders(store, &currentTime)

		if len(tasks) != 0 {
			t.Error("tasks should have no tasks")
		}
	})
}

func DeleteReminder(t *testing.T) {
	futureTime := time.Now().Add(10 * time.Minute)
	pastTime := time.Now().Add(-(1 * time.Minute))

	store := &StubStore{reminders: []reminder.ReminderResponse{
		{ID: 1, Title: "birthday", Body: "some random text", ExecuteAt: &futureTime, Status: "pending"},
		{ID: 2, Title: "birthday", Body: "some random text", ExecuteAt: &futureTime, Status: "pending"},
		{ID: 4, Title: "birthday", Body: "some random text", ExecuteAt: &pastTime, Status: "scheduled"},
	}}

	t.Run("delete a task", func(t *testing.T) {

		err := reminder.DeleteReminder(store, 1)
		if err != nil {
			t.Error("error should be nil")
		}

		if len(store.reminders) != 2 {
			t.Error("tasks should have two tasks")
		}
	})

	t.Run("return error for no task found with id", func(t *testing.T) {

		err := reminder.DeleteReminder(store, 10)

		if err == nil {
			t.Error("error should not be nil for no task found with id")
		}
	})
}
