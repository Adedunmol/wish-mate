package reminder

import (
	"context"
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
	GetBirthdays(currentTime *time.Time) ([]ReminderResponse, error)
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := t.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// add inner join to get the user_id friends (id, email), which the notifications and emails are going to be sent
	query := `
		SELECT id, user_id, title, body, execute_at FROM reminders WHERE execute_at <= NOW();
`
	var reminders []ReminderResponse

	rows, err := t.DB.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("error querying reminders: %v", err)
	}

	for rows.Next() {
		var reminder ReminderResponse

		err = rows.Scan(&reminder.ID, &reminder.UserID, &reminder.Email, &reminder.Title, &reminder.Body, &reminder.Type, &reminder.Status, &reminder.ExecuteAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

func (t *ReminderStore) GetBirthdays(currentTime *time.Time) ([]ReminderResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := t.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		SELECT id, user_id, 'Happy Birthday!' AS title, 
       		'Wishing you a wonderful day filled with joy!' AS body, 
       		'birthday' AS type, 'pending' AS status, email 
		FROM users 
		WHERE DATE_PART('month', birthdate) = DATE_PART('month', NOW()) 
		AND DATE_PART('day', birthdate) = DATE_PART('day', NOW());
`
	var birthdayReminders []ReminderResponse

	rows, err := t.DB.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("error querying users for birthdays: %v", err)
	}

	for rows.Next() {
		var reminder ReminderResponse

		err = rows.Scan(&reminder.ID, &reminder.UserID, &reminder.Title, &reminder.Body, &reminder.Type, &reminder.Status, &reminder.ExecuteAt, &reminder.Email)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		birthdayReminders = append(birthdayReminders, reminder)
	}

	return birthdayReminders, nil
}

func (t *ReminderStore) UpdateReminder(ID int) error {
	return nil
}

type ReminderResponse struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Email     int        `json:"email"`
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

		err = q.Enqueue(&queue.TaskPayload{
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
				"template": "birthday_mail",
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
