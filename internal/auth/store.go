package auth

import (
	"context"
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Store interface {
	CreateUser(body *CreateUserBody) (CreateUserResponse, error)
	FindUserByEmail(email string) (User, error)
	ComparePasswords(storedPassword, candidatePassword string) bool
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {

	return &UserStore{db: db}
}

func (s *UserStore) CreateUser(body *CreateUserBody) (CreateUserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback()

	var user CreateUserResponse

	row := tx.QueryRowContext(
		ctx,
		"INSERT INTO users (username, first_name, last_name, password) VALUES ($1, $2, $3, $4) RETURNING username, first_name, last_name;",
		body.Username, body.FirstName, body.LastName, body.Password)

	err = row.Scan(&user.Username, &user.FirstName, &user.LastName, &user.LastName)

	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("error scanning row (insert auth): %w", err)
	}

	return user, nil
}

func (s *UserStore) FindUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback()

	var user User

	row := tx.QueryRowContext(ctx, "SELECT id, username, email, first_name, last_name, password FROM users WHERE email = $1;", email)

	err = row.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Password)

	if err != nil {
		return User{}, fmt.Errorf("error scanning row (find auth by email): %w", err)
	}

	return user, nil
}

func (s *UserStore) ComparePasswords(storedPassword, candidatePassword string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(candidatePassword))

	if err != nil {
		return false
	}
	return true
}
