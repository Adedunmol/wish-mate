package auth

import (
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"time"
)

type User struct {
	ID           int
	FirstName    string
	LastName     string
	Username     string
	Email        string
	Password     string
	DateOfBirth  *time.Time
	Verified     bool
	RefreshToken string
}

type CreateUserBody struct {
	helpers.Validation
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Username    string `json:"username" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	DateOfBirth string `json:"date_of_birth" validate:"required"`
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

type OTP struct {
	ID        int        `json:"id"`
	Email     string     `json:"email"`
	OTP       string     `json:"otp"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt *time.Time `json:"created_at"`
}

type UpdateUserBody struct {
	helpers.Validation
	Verified     *bool  `json:"verified"` // take in a pointer cos of COALESCE in Postgres
	Password     string `json:"password"`
	RefreshToken string `json:"refresh_token"`
}

type VerifyOTPBody struct {
	helpers.Validation
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required"`
}

type RequestOTPBody struct {
	helpers.Validation
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordBody struct {
	helpers.Validation
	OldPassword        string `json:"old_password" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required"`
	NewPasswordConfirm string `json:"new_password_confirm" validate:"required"`
}

type ForgotPasswordBody struct {
	helpers.Validation
	Email              string `json:"email" validate:"required,email"`
	Code               string `json:"code" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required"`
	NewPasswordConfirm string `json:"new_password_confirm" validate:"required"`
}
