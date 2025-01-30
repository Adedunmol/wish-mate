package auth

import "github.com/Adedunmol/wish-mate/internal/helpers"

type User struct {
	ID          int
	FirstName   string
	LastName    string
	Username    string
	Email       string
	Password    string
	DateOfBirth string
}

type CreateUserBody struct {
	helpers.Validation
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Password  string `json:"password" validate:"required"`
	Username  string `json:"username" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
}

type LoginUserBody struct {
	helpers.Validation
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required"`
}

type CreateUserResponse struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}
