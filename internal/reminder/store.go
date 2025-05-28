package reminder

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

type CreateReminderBody struct {
	Template       string     `json:"template"`
	SourceType     string     `json:"source_type"`
	SourceId       int        `json:"source_id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	Username       string     `json:"username"`
	UserID         int        `json:"user_id"`
	Title          string     `json:"title"`
	Body           string     `json:"body"`
	FriendName     string     `json:"friend_name"`
	FriendUsername string     `json:"friend_username"`
	Type           string     `json:"type"` // wishlist_reminder // todo: event_reminder
	ExecuteAt      *time.Time `json:"execute_at"`
}

type UpdateReminder struct {
	Status string
}

type Store interface {
	CreateReminder(body CreateReminderBody) error
	GetReminders(currentTime *time.Time) ([]ReminderResponse, error)
	GetBirthdays(currentTime *time.Time) ([]ReminderResponse, error)
	UpdateReminder(ID int, data UpdateReminder) error
	DeleteReminder(sourceType string, sourceID int) error
}

type ReminderStore struct {
	DB *pgx.Conn
}

func NewReminderStore(db *pgx.Conn) ReminderStore {

	return ReminderStore{DB: db}
}

func (t *ReminderStore) DeleteReminder(sourceType string, sourceID int) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := t.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
	DELETE FROM reminders WHERE source_type = $1 and source_id = $2;
	`

	_, err = t.DB.Query(ctx, query, sourceType, sourceID)

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
		SELECT f.friend_id, u.email, u.username
		FROM friendships f
		JOIN users u ON u.id = f.friend_id
		WHERE f.user_id = $1 AND f.status = 'accepted';
`

	rows, err := t.DB.Query(ctx, query, body.UserID)

	if err != nil {
		return fmt.Errorf("error querying friendships (reminders): %w", err)
	}

	defer rows.Close()

	type friend struct {
		Email    string
		Username string
		ID       int
	}

	friends := make([]friend, 0)
	for rows.Next() {
		var friendData friend
		if err := rows.Scan(&friendData.ID, &friendData.Email, &friendData.Username); err != nil {
			return fmt.Errorf("error scanning friend: %w", err)
		}
		friends = append(friends, friendData)
	}

	// Create reminders for friends
	for _, friend := range friends {
		_, err = tx.Exec(ctx, `
			INSERT INTO reminders (user_id, title, body, type, execute_at, template, email, source_type, source_id, friend_name, friend_username, username)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
		`, friend.ID, body.Title, body.Body, body.Type, body.ExecuteAt, body.Template, friend.Email, body.SourceType, body.SourceId, body.FriendName, body.FriendUsername, friend.Username)
		if err != nil {
			return fmt.Errorf("create reminder for friend %d: %w", friend.ID, err)
		}
	}

	// Optionally create a reminder for the user themselves:
	//_, err = tx.Exec(ctx, `
	//	INSERT INTO reminders (user_id, title, body, type, execute_at, template, email, source_type, source_id)
	//	VALUES ($1, $2, $3, $4, $5, $6, $7);
	//`, body.UserID, body.Title, body.Body, body.Type, body.ExecuteAt, body.Template, body.Email, body.SourceType, body.SourceId)
	//if err != nil {
	//	return fmt.Errorf("create user reminder: %w", err)
	//}

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
		SELECT id, user_id, email, title, body, type, status, execute_at, template, friend_name, friend_username, username FROM reminders WHERE execute_at <= NOW();
`
	var reminders []ReminderResponse

	rows, err := tx.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("error querying reminders: %v", err)
	}

	for rows.Next() {
		var reminder ReminderResponse

		err = rows.Scan(&reminder.ID, &reminder.UserID, &reminder.Email, &reminder.Title, &reminder.Body, &reminder.Type, &reminder.Status, &reminder.ExecuteAt, &reminder.Template, &reminder.FriendName, &reminder.FriendUsername, &reminder.Username)
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
		SELECT 
			f.friend_id AS user_id, 
			u2.email AS email,
			u2.username AS username,
			u1.id AS birthday_user_id,
			u1.username AS birthday_user_name,
			u1.first_name || ' ' || u1.last_name AS birthday_name
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
		INSERT INTO reminders (user_id, email, title, body, type, status, template, execute_at, source_type, friend_name, friend_username, username)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'birthday', $10, $11, $12)
		RETURNING id;
	`

	type birthdayEntry struct {
		userID           int
		email            string
		birthdayUserID   int
		birthdayUserName string
		birthdayName     string
		username         string
	}

	var entries []birthdayEntry

	for rows.Next() {
		var entry birthdayEntry

		err := rows.Scan(&entry.userID, &entry.email, &entry.username, &entry.birthdayUserID, &entry.birthdayUserName, &entry.birthdayName, &entry.birthdayUserName, &entry.birthdayName)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		entries = append(entries, entry)
	}

	for _, entry := range entries {
		title := "🎉 It's " + entry.birthdayUserName + "'s Birthday!"
		body := "Wish " + entry.birthdayUserName + " a fantastic birthday today!"
		reminderType := "birthday"
		status := "pending"
		template := "birthday_reminder_mail.html"

		var id int
		err = tx.QueryRow(ctx, insertQuery,
			entry.userID, entry.email, title, body, reminderType, status, template, currentTime, "birthday", entry.birthdayName, entry.birthdayUserName, entry.username).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("error inserting reminder: %w", err)
		}

		reminders = append(reminders, ReminderResponse{
			ID:             id,
			UserID:         entry.userID,
			Email:          entry.email,
			Title:          title,
			Body:           body,
			Type:           reminderType,
			Status:         status,
			Template:       template,
			ExecuteAt:      currentTime,
			SourceType:     "birthday",
			FriendUsername: entry.birthdayUserName,
			FriendName:     entry.birthdayName,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return reminders, nil
}

func (t *ReminderStore) UpdateReminder(ID int, data UpdateReminder) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := t.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
			UPDATE reminders SET status = $1 WHERE id = $2;
			`

	_, err = tx.Exec(ctx, query, data.Status, ID)

	if err != nil {
		return fmt.Errorf("error updating reminder: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
