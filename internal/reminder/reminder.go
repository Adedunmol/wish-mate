package reminder

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Adedunmol/wish-mate/internal/queue"
)

type ReminderResponse struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Email     string     `json:"email"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	Template  string     `json:"template"`
	ExecuteAt *time.Time `json:"execute_at"`
}

func CreateReminder(store Store, body CreateReminderBody) error {

	if body.Name == "" {
		return errors.New("empty name")
	}

	if body.ExecuteAt == nil {
		return errors.New("executeAt is empty")
	}
	//
	//if body.Payload == nil {
	//	return ScheduledTaskResponse{}, errors.New("payload is empty")
	//}

	err := store.CreateReminder(body)
	if err != nil {
		return fmt.Errorf("error creating a task: %v", err)
	}

	return nil
}

func GetReminders(store Store, currentTime *time.Time) ([]ReminderResponse, error) {
	tasks, err := store.GetReminders(currentTime)
	if err != nil {
		return nil, fmt.Errorf("error getting tasks: %v", err)
	}

	return tasks, nil
}

func GetBirthdays(store Store, currentTime *time.Time) ([]ReminderResponse, error) {
	tasks, err := store.GetBirthdays(currentTime)
	if err != nil {
		return nil, fmt.Errorf("error getting tasks: %v", err)
	}

	return tasks, nil
}

func EnqueueReminders(store Store, q queue.Queue, currentTime *time.Time) error {

	// this should send in reminders and the details of the users to send the reminders to
	tasks, err := GetReminders(store, currentTime)
	if err != nil {
		return fmt.Errorf("error getting tasks: %v", err)
	}

	for _, task := range tasks {

		err := q.Enqueue(&queue.TaskPayload{
			Type: queue.TypeNotificationDelivery,
			Payload: map[string]interface{}{
				"id":      task.ID,
				"user_id": task.UserID,
				"title":   task.Title,
				"body":    task.Title,
				"type":    task.Type,
			},
		})

		if err != nil {
			log.Printf("error enqueuing scheduled task: %s : %v", err, task)
		}

		err = q.Enqueue(&queue.TaskPayload{
			Type: queue.TypeEmailDelivery,
			Payload: map[string]interface{}{
				"template": "reminder_mail",
				"subject":  "Wishlist Reminder",
				"email":    task.Email,
				"data":     "",
				// embed the data below into a map and then pass into data
				//"id":       task.ID,
				//"user_id":  task.UserID,
				//"title":    task.Title,
				//"body":     task.Title,
				//"type":     task.Type,
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

func EnqueueBirthdays(store Store, q queue.Queue, currentTime *time.Time) error {

	tasks, err := GetBirthdays(store, currentTime)
	if err != nil {
		return fmt.Errorf("error getting birthdays: %v", err)
	}

	for _, task := range tasks {

		err = q.Enqueue(&queue.TaskPayload{
			Type: queue.TypeNotificationDelivery,
			Payload: map[string]interface{}{
				"id":      task.ID,
				"user_id": task.UserID,
				"title":   task.Title,
				"body":    task.Body,
				"type":    task.Type,
			},
		})

		if err != nil {
			log.Printf("error enqueuing scheduled task: %s : %v", err, task)
		}

		err = q.Enqueue(&queue.TaskPayload{
			Type: queue.TypeEmailDelivery,
			Payload: map[string]interface{}{
				"template": task.Template,
				"subject":  "Birthday",
				"email":    task.Email,
				"data":     "",
				// embed the data below into a map and then pass into data
				//"id":       task.ID,
				//"user_id":  task.UserID,
				//"title":    task.Title,
				//"body":     task.Title,
				//"type":     task.Type,
			},
		})
		if err != nil {
			log.Printf("error enqueuing scheduled task: %s : %v", err, task)
		}
	}

	return nil
}

func DeleteReminder(store Store, id int) error {
	err := store.DeleteReminder(id)

	if err != nil {
		return fmt.Errorf("error deleting task: %v", err)
	}

	return nil
}
