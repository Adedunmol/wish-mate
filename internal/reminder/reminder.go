package reminder

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Adedunmol/wish-mate/internal/queue"
)

type ReminderResponse struct {
	ID             int        `json:"id"`
	UserID         int        `json:"user_id"`
	Email          string     `json:"email"`
	Username       string     `json:"username"`
	Title          string     `json:"title"`
	Body           string     `json:"body"`
	Type           string     `json:"type"`
	Status         string     `json:"status"`
	Template       string     `json:"template"`
	SourceType     string     `json:"source_type"`
	SourceId       int        `json:"source_id"`
	FriendName     string     `json:"friend_name"`
	FriendUsername string     `json:"friend_username"`
	ExecuteAt      *time.Time `json:"execute_at"`
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

func CreateBirthdayReminders(store Store, currentTime *time.Time) ([]ReminderResponse, error) {
	tasks, err := store.GetBirthdays(currentTime)
	if err != nil {
		return nil, fmt.Errorf("error creating birthday reminders: %v", err)
	}

	return tasks, nil
}

func EnqueueReminders(store Store, q queue.Queue, currentTime *time.Time) error {

	// this should send in reminders and the details of the users to send the reminders to
	reminders, err := GetReminders(store, currentTime)
	if err != nil {
		return fmt.Errorf("error getting reminders: %v", err)
	}

	for _, reminder := range reminders {

		err := q.Enqueue(&queue.TaskPayload{
			Type: queue.TypeNotificationDelivery,
			Payload: map[string]interface{}{
				"id":      reminder.ID,
				"user_id": reminder.UserID,
				"title":   reminder.Title,
				"body":    reminder.Body,
				"type":    reminder.Type,
			},
		})

		if err != nil {
			log.Printf("error enqueuing scheduled reminder (notification): %s : %v", err, reminder)
		}

		err = q.Enqueue(&queue.TaskPayload{
			Type: queue.TypeEmailDelivery,
			Payload: map[string]interface{}{
				"template": reminder.Template,
				"subject":  reminder.Title,
				"email":    reminder.Email,
				"data": struct {
					Username       string
					FriendName     string
					FriendUsername string
					Date           *time.Time
				}{
					Username:       reminder.Username,
					FriendName:     reminder.FriendName,
					FriendUsername: reminder.FriendUsername,
					Date:           reminder.ExecuteAt,
				},
			},
		})
		if err != nil {
			log.Printf("error enqueuing scheduled reminder (email notification): %s : %v", err, reminder)
		}

		data := UpdateReminder{Status: "scheduled"}

		err = store.UpdateReminder(reminder.ID, data)

		if err != nil {
			return fmt.Errorf("error updating reminder: %v", err)
		}
	}

	return nil
}

func DeleteReminder(store Store, sourceType string, sourceId int) error {
	err := store.DeleteReminder(sourceType, sourceId)

	if err != nil {
		return fmt.Errorf("error deleting reminder: %v", err)
	}

	return nil
}
