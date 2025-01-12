package user

import "database/sql"

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

	return CreateUserResponse{}, nil
}

func (s *UserStore) FindUserByEmail(email string) (User, error) {

	return User{}, nil
}

func (s *UserStore) ComparePasswords(storedPassword, candidatePassword string) bool {

	return false
}
