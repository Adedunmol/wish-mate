package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type OTPStore interface {
	CreateOTP(email string, code string, expiration int) error
	ValidateOTP(email string, otp string) (bool, error)
	DeleteOTP(email string) error
}

type Store interface {
	CreateUser(body *CreateUserBody) (CreateUserResponse, error)
	FindUserByEmail(email string) (User, error)
	FindUserByID(id int) (User, error)
	UpdateUser(id int, data UpdateUserBody) (User, error)
	ComparePasswords(storedPassword, candidatePassword string) bool
	UpdateRefreshToken(oldRefreshToken, refreshToken string) error
	DeleteRefreshToken(refreshToken string) error
}

type UserStore struct {
	db *pgx.Conn
}

func NewUserStore(db *pgx.Conn) *UserStore {

	return &UserStore{db: db}
}

const UniqueViolation = "23505"

func (s *UserStore) CreateUser(body *CreateUserBody) (CreateUserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return CreateUserResponse{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var user CreateUserResponse

	dob, _ := time.Parse("2006-01-02", body.DateOfBirth)

	formattedDOB := dob.Format("2006-01-02")

	row := tx.QueryRow(
		ctx,
		"INSERT INTO users (email, username, first_name, last_name, password, date_of_birth) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, username, first_name, last_name;",
		body.Email, body.Username, body.FirstName, body.LastName, body.Password, formattedDOB)

	err = row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.LastName)

	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == UniqueViolation {
			return CreateUserResponse{}, helpers.ErrConflict
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return CreateUserResponse{}, fmt.Errorf("error scanning row (create user): %w", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return CreateUserResponse{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return user, nil
}

func (s *UserStore) FindUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return User{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var user User

	row := tx.QueryRow(ctx, "SELECT id, username, email, first_name, last_name, password, date_of_birth FROM users WHERE email = $1;", email)

	err = row.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.DateOfBirth)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, helpers.ErrNotFound
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return User{}, fmt.Errorf("error scanning row (find auth by email): %w", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return User{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return user, nil
}

func (s *UserStore) FindUserByID(id int) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return User{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var user User

	row := tx.QueryRow(
		ctx,
		"SELECT id, username, email, first_name, last_name, password, date_of_birth FROM users WHERE id = $1;",
		id,
	)

	err = row.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.DateOfBirth)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, helpers.ErrNotFound
		}
		err = errors.Join(helpers.ErrInternalServerError, err)
		return User{}, fmt.Errorf("error scanning row (find auth by email): %w", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return User{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return user, nil
}

func (s *UserStore) UpdateUser(id int, data UpdateUserBody) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return User{}, fmt.Errorf("error creating transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := tx.Commit(ctx); err != nil {
		err = errors.Join(helpers.ErrInternalServerError, err)
		return User{}, fmt.Errorf("commit tx: %w", err)
	}

	return User{}, err
}

func (s *UserStore) ComparePasswords(storedPassword, candidatePassword string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(candidatePassword))

	if err != nil {
		return false
	}
	return true
}

func (s *UserStore) DeleteRefreshToken(refreshToken string) error {
	return nil
}

func (s *UserStore) UpdateRefreshToken(oldRefreshToken, refreshToken string) error {
	return nil
}
