package auth

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Store interface {
	CreateUser(body *CreateUserBody) (CreateUserResponse, error)
	FindUserByEmail(email string) (User, error)
	ComparePasswords(storedPassword, candidatePassword string) bool
}

type UserStore struct {
	db *pgx.Conn
}

func NewUserStore(db *pgx.Conn) *UserStore {

	return &UserStore{db: db}
}

func (s *UserStore) CreateUser(body *CreateUserBody) (CreateUserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var user CreateUserResponse

	row := tx.QueryRow(
		ctx,
		"INSERT INTO users (username, first_name, last_name, password) VALUES ($1, $2, $3, $4) RETURNING username, first_name, last_name;",
		body.Username, body.FirstName, body.LastName, body.Password)

	err = row.Scan(&user.Username, &user.FirstName, &user.LastName, &user.LastName)

	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("error scanning row (insert user): %w", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return user, nil
}

func (s *UserStore) FindUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return User{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var user User

	row := tx.QueryRow(ctx, "SELECT id, username, email, first_name, last_name, password, date_of_birth FROM users WHERE email = $1;", email)

	err = row.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.DateOfBirth)

	if err != nil {
		return User{}, fmt.Errorf("error scanning row (find auth by email): %w", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return User{}, fmt.Errorf("error committing transaction: %w", err)
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
