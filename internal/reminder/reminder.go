package reminder

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/jackc/pgx/v5"
)

type CreateReminderBody struct {
	Template  string     `json:"template"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	UserID    int        `json:"user_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Type      string     `json:"type"` // wishlist_reminder // todo: event_reminder
	ExecuteAt *time.Time `json:"execute_at"`
}

type Store interface {
	CreateReminder(body CreateReminderBody) error
	GetReminders(currentTime *time.Time) ([]ReminderResponse, error)
	GetBirthdays(currentTime *time.Time) ([]ReminderResponse, error)
	UpdateReminder(ID int) error
	DeleteReminder(ID int) error
}

type ReminderStore struct {
	DB *pgx.Conn
}

func NewReminderStore(db *pgx.Conn) ReminderStore {

	return ReminderStore{DB: db}
}

func (t *ReminderStore) DeleteReminder(ID int) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := t.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
	DELETE FROM reminders WHERE id = $1;
	`

	_, err = t.DB.Query(ctx, query, ID)

	if err != nil {
		return fmt.Errorf("error deleting reminder: %w", err)
	}

	return nil
}

func (t *ReminderStore) CreateReminder(body CreateReminderBody) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := t.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
	SELECT f.friend_id, u.email FROM friendships f
	WHERE f.user_id = $1 AND f.status = 'accepted'
	JOIN users u ON u.id = f.friend_id;
	`

	rows, err := t.DB.Query(ctx, query, body.UserID)

	if err != nil {
		return fmt.Errorf("error querying friendships (reminders): %w", err)
	}

	defer rows.Close()

	type friend struct {
		Email string
		ID    int
	}

	friends := make([]friend, 0)
	for rows.Next() {
		var friendData friend
		if err := rows.Scan(&friendData.ID, &friendData.Email); err != nil {
			return fmt.Errorf("scan friend: %w", err)
		}
		friends = append(friends, friendData)
	}

	// Create reminders for friends
	for _, friend := range friends {
		_, err = tx.Exec(ctx, `
			INSERT INTO reminders (name, user_id, title, body, type, execute_at, template, email)
			VALUES ($1, $2, $3, $4, $5, $6, $7);
		`, body.Name, friend.ID, body.Title, body.Body, body.Type, body.ExecuteAt, body.Template, friend.Email)
		if err != nil {
			return fmt.Errorf("create reminder for friend %d: %w", friend.ID, err)
		}
	}

	// Optionally create a reminder for the user themselves:
	_, err = tx.Exec(ctx, `
		INSERT INTO reminders (name, user_id, title, body, type, execute_at, template, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`, body.Name, body.UserID, body.Title, body.Body, body.Type, body.ExecuteAt, body.Template, body.Email)
	if err != nil {
		return fmt.Errorf("create user reminder: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (t *ReminderStore) GetReminders(currentTime *time.Time) ([]ReminderResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := t.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		SELECT id, user_id, email, title, body, type, status, execute_at, template FROM reminders WHERE execute_at <= NOW();
`
	var reminders []ReminderResponse

	rows, err := t.DB.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("error querying reminders: %v", err)
	}

	for rows.Next() {
		var reminder ReminderResponse

		err = rows.Scan(&reminder.ID, &reminder.UserID, &reminder.Email, &reminder.Title, &reminder.Body, &reminder.Type, &reminder.Status, &reminder.ExecuteAt, &reminder.Template)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		_, err = tx.Exec(ctx, `
        UPDATE reminders SET status = 'scheduled' WHERE id = $1`,
			reminder.ID)

		if err != nil {
			return nil, err
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
		SELECT 
			f.friend_id AS user_id, 
			u2.email AS email,
			u1.id AS birthday_user_id,
			u1.first_name || ' ' || u1.last_name AS birthday_user_name
		FROM friendships f
		JOIN users u1 ON u1.id = f.user_id
		JOIN users u2 ON u2.id = f.friend_id
		WHERE f.status = 'accepted'
		AND DATE_PART('month', u1.date_of_birth) = DATE_PART('month', NOW())
		AND DATE_PART('day', u1.date_of_birth) = DATE_PART('day', NOW());
	`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying users for birthdays: %v", err)
	}
	defer rows.Close()

	var reminders []ReminderResponse
	// now := time.Now()

	insertQuery := `
		INSERT INTO reminders (user_id, email, title, body, type, status, template, execute_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id;
	`

	for rows.Next() {
		var userID int
		var email string
		var birthdayUserID int
		var birthdayUserName string

		err := rows.Scan(&userID, &email, &birthdayUserID, &birthdayUserName)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		title := "🎉 It's " + birthdayUserName + "'s Birthday!"
		body := "Wish " + birthdayUserName + " a fantastic birthday today!"
		reminderType := "birthday"
		status := "pending"
		template := "birthday_reminder.html" // your HTML file name

		var id int
		err = tx.QueryRow(ctx, insertQuery,
			userID, email, title, body, reminderType, status, template, currentTime).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("error inserting reminder: %w", err)
		}

		reminders = append(reminders, ReminderResponse{
			ID:        id,
			UserID:    userID,
			Email:     email,
			Title:     title,
			Body:      body,
			Type:      reminderType,
			Status:    status,
			Template:  template,
			ExecuteAt: currentTime,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return reminders, nil
}

func (t *ReminderStore) UpdateReminder(ID int) error {
	return nil
}

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
