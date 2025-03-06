package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/golang-jwt/jwt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"
)

var (
	ErrCreate  = errors.New("error creating entry")
	ErrNoEntry = errors.New("no entry found")
)

type StubQueue struct {
	Tasks []queue.TaskPayload
}

func (q *StubQueue) Enqueue(taskPayload *queue.TaskPayload) error {
	q.Tasks = append(q.Tasks, *taskPayload)
	return nil
}

type StubOtpStore struct {
	otps []auth.OTP
}

func (s *StubOtpStore) CreateOTP(email, otp string, expiration int) error {
	currentTime := time.Now()
	futureTime := time.Now().Add(10 * time.Minute)

	data := auth.OTP{
		ID:        1,
		Email:     email,
		OTP:       otp,
		ExpiresAt: &futureTime,
		CreatedAt: &currentTime,
	}

	s.otps = append(s.otps, data)

	return nil
}

func (s *StubOtpStore) ValidateOTP(email string, otp string) (bool, error) {

	for _, otpData := range s.otps {
		if otpData.Email == email {
			if otpData.OTP != otp {
				return false, helpers.ErrBadRequest
			} else if otpData.ExpiresAt.Before(time.Now()) {
				return false, helpers.ErrBadRequest
			} else {
				return true, nil
			}
		}
	}

	return false, helpers.ErrNotFound
}

func (s *StubOtpStore) DeleteOTP(email string) error {
	return nil
}

type StubUserStore struct {
	users []auth.User
}

func (s *StubUserStore) CreateUser(body *auth.CreateUserBody) (auth.CreateUserResponse, error) {

	for _, u := range s.users {
		if u.Email == body.Email {
			return auth.CreateUserResponse{}, helpers.ErrConflict
		}
	}

	userData := auth.User{ID: 1, FirstName: body.FirstName, LastName: body.LastName, Username: body.Username, Email: body.Email, Password: body.Password}

	s.users = append(s.users, userData)

	return auth.CreateUserResponse{ID: userData.ID, FirstName: userData.FirstName, LastName: userData.LastName, Username: userData.Username}, nil
}

func (s *StubUserStore) FindUserByEmail(email string) (auth.User, error) {

	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return auth.User{}, helpers.ErrNotFound
}

func (s *StubUserStore) FindUserByID(id int) (auth.User, error) {

	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return auth.User{}, helpers.ErrNotFound
}

func (s *StubUserStore) UpdateUser(id int, data auth.UpdateUserBody) (auth.User, error) {
	for i, u := range s.users {
		if u.ID == id {
			s.users[i].Verified = data.Verified

			return s.users[i], nil
		}
	}

	return auth.User{}, helpers.ErrNotFound
}

func (s *StubUserStore) ComparePasswords(storedPassword, candidatePassword string) bool {
	return storedPassword == candidatePassword
}

func (s *StubUserStore) DeleteRefreshToken(refreshToken string) error {
	return nil
}

func (s *StubUserStore) UpdateRefreshToken(oldRefreshToken, refreshToken string) error {

	for i, u := range s.users {
		if oldRefreshToken == u.RefreshToken {
			s.users[i].RefreshToken = refreshToken
			return nil
		}
	}

	return helpers.ErrNotFound
}

type FailingStubUserStore struct {
	users []auth.User
}

func (s *FailingStubUserStore) CreateUser(_ *auth.CreateUserBody) (auth.CreateUserResponse, error) {

	return auth.CreateUserResponse{}, ErrCreate
}

func (s *FailingStubUserStore) FindUserByEmail(_ string) (auth.User, error) {
	return auth.User{}, ErrNoEntry
}

func (s *FailingStubUserStore) FindUserByID(id int) (auth.User, error) {

	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return auth.User{}, helpers.ErrNotFound
}

func (s *FailingStubUserStore) UpdateUser(id int, data auth.UpdateUserBody) (auth.User, error) {
	for i, u := range s.users {
		if u.ID == id {
			s.users[i].Verified = data.Verified

			return s.users[i], nil
		}
	}

	return auth.User{}, helpers.ErrNotFound
}

func (s *FailingStubUserStore) ComparePasswords(_, _ string) bool {
	return false
}

func (s *FailingStubUserStore) DeleteRefreshToken(refresToken string) error {
	return nil
}

func (s *FailingStubUserStore) UpdateRefreshToken(oldRefreshToken, refreshToken string) error {

	return nil
}

func TestPOSTUser(t *testing.T) {

	t.Run("create and send a auth back", func(t *testing.T) {
		store := StubUserStore{users: make([]auth.User, 0)}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}
		server := &auth.Handler{Store: &store, Queue: &mockQueue}

		data := []byte(`{ "first_name": "Adedunmola", "last_name": "Oyewale", "username": "Adedunmola", "password": "password", "email": "adedunmola@gmail.com" }`)

		request := createUserRequest(data)
		response := httptest.NewRecorder()

		server.CreateUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "User created successfully",
			"data": map[string]interface{}{
				"id":         float64(1),
				"first_name": "Adedunmola",
				"last_name":  "Oyewale",
				"username":   "Adedunmola",
			},
		}

		assertResponseCode(t, response.Code, http.StatusCreated)
		assertResponseBody(t, got, want)

		if len(store.users) != 1 {
			t.Errorf("got %d users, want 1", len(store.users))
		}

		if len(mockQueue.Tasks) != 1 {
			t.Errorf("got %d tasks, want 1", len(mockQueue.Tasks))
		}
	})

	t.Run("fails in creating auth", func(t *testing.T) {
		store := FailingStubUserStore{users: make([]auth.User, 0)}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}
		server := &auth.Handler{Store: &store, Queue: &mockQueue}
		data := []byte(`{ "first_name": "Adedunmola", "last_name": "Oyewale", "username": "Adedunmola", "password": "password" }`)

		request := createUserRequest(data)
		response := httptest.NewRecorder()

		server.CreateUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		//want := map[string]interface{}{}

		assertResponseCode(t, response.Code, http.StatusInternalServerError)

		//assertResponseBody(t, got, want)
	})

	t.Run("returns error for invalid request body", func(t *testing.T) {
		store := FailingStubUserStore{users: make([]auth.User, 0)}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}
		server := &auth.Handler{Store: &store, Queue: &mockQueue}
		data := []byte(`{ "first_name": "Adedunmola", "last_name": "Oyewale", "username": "Adedunmola" }`)

		request := createUserRequest(data)
		response := httptest.NewRecorder()

		server.CreateUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "invalid request body",
			"problems": map[string][]string{
				"Email":    []string{"Email required"},
				"Password": []string{"Password required"},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)

		if len(mockQueue.Tasks) == 1 {
			t.Errorf("got %d tasks, want 0", len(mockQueue.Tasks))
		}
	})

	t.Run("email conflict", func(t *testing.T) {
		store := StubUserStore{users: []auth.User{
			{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
		}}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}

		server := &auth.Handler{Store: &store, Queue: &mockQueue}

		data := []byte(`{ "first_name": "Adedunmola", "last_name": "Oyewale", "username": "Adedunmola", "password": "password", "email": "adedunmola@gmail.com" }`)

		request := createUserRequest(data)
		response := httptest.NewRecorder()

		server.CreateUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource already exists",
		}

		assertResponseCode(t, response.Code, http.StatusConflict)
		assertResponseBody(t, got, want)

		if len(mockQueue.Tasks) == 1 {
			t.Errorf("got %d tasks, want 0", len(mockQueue.Tasks))
		}
	})
}

func TestPOSTLogin(t *testing.T) {
	store := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	server := &auth.Handler{Store: &store}

	t.Run("find and log in a auth", func(t *testing.T) {

		data := []byte(`{ "email": "adedunmola@gmail.com", "password": "password" }`)

		request := loginUserRequest(data)
		response := httptest.NewRecorder()

		server.LoginUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		assertResponseCode(t, response.Code, http.StatusOK)
	})

	t.Run("returns error for invalid request body", func(t *testing.T) {
		store := FailingStubUserStore{users: make([]auth.User, 0)}
		server := &auth.Handler{Store: &store}
		data := []byte(`{ "password": "password" }`)

		request := createUserRequest(data)
		response := httptest.NewRecorder()

		server.LoginUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "invalid request body",
			"problems": map[string][]string{
				"Email": []string{"Email required"},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)

	})

	t.Run("does not find a auth", func(t *testing.T) {
		data := []byte(`{ "email": "adedunmola1@gmail.com", "password": "password123" }`)

		request := loginUserRequest(data)
		response := httptest.NewRecorder()

		server.LoginUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "invalid credentials",
		}

		assertResponseCode(t, response.Code, http.StatusUnauthorized)
		assertResponseBody(t, got, want)
	})

	t.Run("incorrect password", func(t *testing.T) {

		data := []byte(`{ "email": "adedunmola@gmail.com", "password": "password123" }`)

		request := loginUserRequest(data)
		response := httptest.NewRecorder()

		server.LoginUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "invalid credentials",
		}

		assertResponseCode(t, response.Code, http.StatusUnauthorized)
		assertResponseBody(t, got, want)
	})
}

func TestVerifyOTP(t *testing.T) {
	currentTime := time.Now()
	futureTime := time.Now().Add(10 * time.Minute)

	store := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	otpStore := StubOtpStore{
		otps: []auth.OTP{
			{ID: 1, Email: "adedunmola@gmail.com", OTP: "123456", ExpiresAt: &futureTime, CreatedAt: &currentTime},
		},
	}
	server := &auth.Handler{Store: &store, OTPStore: &otpStore}

	t.Run("validate the otp and update friendship's verified status", func(t *testing.T) {
		data := []byte(`{ "email": "adedunmola@gmail.com", "code": "123456" }`)
		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.VerifyUserHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusOK)
	})

	t.Run("invalid otp", func(t *testing.T) {
		data := []byte(`{ "email": "adedunmola@gmail.com", "code": "123478" }`)
		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.VerifyUserHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})

	t.Run("expired otp", func(t *testing.T) {
		pastTime := time.Now().Add(-10 * time.Minute)
		store := StubUserStore{users: []auth.User{
			{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
		}}
		otpStore := StubOtpStore{
			otps: []auth.OTP{
				{ID: 1, Email: "adedunmola@gmail.com", OTP: "123456", ExpiresAt: &pastTime, CreatedAt: &currentTime},
			},
		}
		server := &auth.Handler{Store: &store, OTPStore: &otpStore}

		data := []byte(`{ "email": "adedunmola@gmail.com", "code": "123456" }`)
		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.VerifyUserHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})

	t.Run("no otp found with email", func(t *testing.T) {
		data := []byte(`{ "email": "ade@gmail.com", "code": "123456" }`)
		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.VerifyUserHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})

	t.Run("invalid body", func(t *testing.T) {
		data := []byte(`{ "email": "ade@gmail.com" }`)
		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.VerifyUserHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})
}

func TestRequestOTP(t *testing.T) {
	currentTime := time.Now()
	futureTime := time.Now().Add(10 * time.Minute)

	store := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}
	otpStore := StubOtpStore{
		otps: []auth.OTP{
			{ID: 1, Email: "adedunmola@gmail.com", OTP: "123456", ExpiresAt: &futureTime, CreatedAt: &currentTime},
		},
	}
	server := &auth.Handler{Store: &store, OTPStore: &otpStore, Queue: &mockQueue}

	t.Run("send otp to friendship", func(t *testing.T) {
		data := []byte(`{ "email": "adedunmola@gmail.com" }`)

		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.RequestCodeHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusOK)
	})

	t.Run("no friendship found with email", func(t *testing.T) {
		data := []byte(`{ "email": "ade@gmail.com" }`)

		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.RequestCodeHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})

	t.Run("invalid body", func(t *testing.T) {
		data := []byte(`{}`)
		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.RequestCodeHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})
}

func TestLogout(t *testing.T) {

	store := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	server := &auth.Handler{Store: &store}

	t.Run("log out user (with cookie)", func(t *testing.T) {
		request := logoutUserRequest()
		response := httptest.NewRecorder()

		server.LogoutUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "User logged out successfully",
		}

		assertResponseBody(t, got, want)
		assertResponseCode(t, response.Code, http.StatusOK)
	})

	t.Run("log out user (without cookie)", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/api/v1/users/logout", nil)
		response := httptest.NewRecorder()

		server.LogoutUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "User logged out successfully",
		}

		assertResponseBody(t, got, want)
		assertResponseCode(t, response.Code, http.StatusOK)
	})

}

func TestPasswordReset(t *testing.T) {

	store := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	server := &auth.Handler{Store: &store}

	t.Run("update user password", func(t *testing.T) {
		data := []byte(`{ "old_password": "password", "new_password": "newpassword", "new_password_confirm": "newpassword" }`)
		request := resetPasswordRequest("adedunmola@gmail.com", data)
		response := httptest.NewRecorder()

		server.ResetPasswordHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Password has been reset successfully",
		}

		assertResponseBody(t, got, want)
		assertResponseCode(t, response.Code, http.StatusOK)
	})

	t.Run("return unauthorized if user not logged in", func(t *testing.T) {
		data := []byte(`{ "old_password": "password", "new_password": "newpassword", "new_password_confirm": "newpassword" }`)
		request, _ := http.NewRequest("POST", "/auth/verify", bytes.NewReader(data))
		response := httptest.NewRecorder()

		server.ResetPasswordHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "user not logged in",
		}

		assertResponseCode(t, response.Code, http.StatusUnauthorized)
		assertResponseBody(t, got, want)
	})

	t.Run("return unauthorized if old password is incorrect", func(t *testing.T) {
		data := []byte(`{ "old_password": "oldpassword", "new_password": "newpassword", "new_password_confirm": "newpassword" }`)
		request := resetPasswordRequest("adedunmola@gmail.com", data)
		response := httptest.NewRecorder()

		server.ResetPasswordHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "invalid credentials",
		}

		assertResponseCode(t, response.Code, http.StatusUnauthorized)
		assertResponseBody(t, got, want)
	})

	t.Run("return bad request if required fields are not sent", func(t *testing.T) {
		data := []byte(`{ "old_password": "oldpassword", "new_password": "newpassword" }`)
		request := resetPasswordRequest("adedunmola@gmail.com", data)
		response := httptest.NewRecorder()

		server.ResetPasswordHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "invalid request body",
			"problems": map[string][]string{
				"NewPasswordConfirm": []string{"NewPasswordConfirm required"},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})
}

func TestForgotPasswordRequest(t *testing.T) {
	currentTime := time.Now()
	futureTime := time.Now().Add(10 * time.Minute)

	store := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}
	otpStore := StubOtpStore{
		otps: []auth.OTP{
			{ID: 1, Email: "adedunmola@gmail.com", OTP: "123456", ExpiresAt: &futureTime, CreatedAt: &currentTime},
		},
	}
	server := &auth.Handler{Store: &store, OTPStore: &otpStore, Queue: &mockQueue}

	t.Run("send otp to user", func(t *testing.T) {
		data := []byte(`{ "email": "adedunmola@gmail.com" }`)

		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.ForgotPasswordRequestHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusOK)
	})

	t.Run("no user found with email", func(t *testing.T) {
		data := []byte(`{ "email": "ade@gmail.com" }`)

		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.ForgotPasswordRequestHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})

	t.Run("invalid body", func(t *testing.T) {
		data := []byte(`{}`)
		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.ForgotPasswordRequestHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})
}

func TestForgotPassword(t *testing.T) {
	currentTime := time.Now()
	futureTime := time.Now().Add(10 * time.Minute)

	store := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	otpStore := StubOtpStore{
		otps: []auth.OTP{
			{ID: 1, Email: "adedunmola@gmail.com", OTP: "123456", ExpiresAt: &futureTime, CreatedAt: &currentTime},
		},
	}
	server := &auth.Handler{Store: &store, OTPStore: &otpStore}

	t.Run("update user password", func(t *testing.T) {
		data := []byte(`{ "email": "adedunmola@gmail.com", "code": "123456", "old_password": "password", "new_password": "newpassword", "new_password_confirm": "newpassword" }`)
		request := forgotPasswordRequest(data)
		response := httptest.NewRecorder()

		server.ForgotPasswordHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Password has been reset successfully",
		}

		assertResponseBody(t, got, want)
		assertResponseCode(t, response.Code, http.StatusOK)
	})

	t.Run("return bad request if required fields are not sent", func(t *testing.T) {
		data := []byte(`{ "email": "adedunmola@gmail.com", "old_password": "oldpassword", "new_password": "newpassword", "new_password_confirm": "newpassword" }`)
		request := forgotPasswordRequest(data)
		response := httptest.NewRecorder()

		server.ForgotPasswordHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "invalid request body",
			"problems": map[string][]string{
				"Code": []string{"Code required"},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})

	t.Run("invalid otp", func(t *testing.T) {
		data := []byte(`{ "email": "adedunmola@gmail.com", "code": "123478", "old_password": "password", "new_password": "newpassword", "new_password_confirm": "newpassword" }`)
		request := forgotPasswordRequest(data)
		response := httptest.NewRecorder()

		server.ForgotPasswordHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})

	t.Run("expired otp", func(t *testing.T) {
		pastTime := time.Now().Add(-10 * time.Minute)
		store := StubUserStore{users: []auth.User{
			{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
		}}
		otpStore := StubOtpStore{
			otps: []auth.OTP{
				{ID: 1, Email: "adedunmola@gmail.com", OTP: "123456", ExpiresAt: &pastTime, CreatedAt: &currentTime},
			},
		}
		server := &auth.Handler{Store: &store, OTPStore: &otpStore}

		data := []byte(`{ "email": "adedunmola@gmail.com", "code": "123456", "old_password": "password", "new_password": "newpassword", "new_password_confirm": "newpassword" }`)
		request := forgotPasswordRequest(data)
		response := httptest.NewRecorder()

		server.ForgotPasswordHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})

	t.Run("no otp found with email", func(t *testing.T) {
		data := []byte(`{ "email": "ade@gmail.com", "code": "123456", "old_password": "password", "new_password": "newpassword", "new_password_confirm": "newpassword" }`)
		request := forgotPasswordRequest(data)
		response := httptest.NewRecorder()

		server.ForgotPasswordHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
	})
}

func TestRefreshToken(t *testing.T) {
	store := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola", RefreshToken: "somerandomtoken"},
	}}
	server := &auth.Handler{Store: &store}

	t.Run("return new access token and update old refresh token", func(t *testing.T) {
		request, token := refreshTokenRequest("adedunmola@gmail.com", 1, true)
		response := httptest.NewRecorder()

		store.users[0].RefreshToken = token

		server.RefreshTokenHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		assertResponseCode(t, response.Code, http.StatusOK)
	})

	t.Run("return unauthorized if cookie not found", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/api/v1/users/refresh-token", nil)
		response := httptest.NewRecorder()

		server.RefreshTokenHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		assertResponseCode(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("return unauthorized if refresh token is not found in an entry", func(t *testing.T) {

		store := StubUserStore{users: []auth.User{
			{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola", RefreshToken: "somerandomtokenunique"},
		}}
		server := &auth.Handler{Store: &store}

		request, _ := refreshTokenRequest("adedunmola@gmail.com", 1, true)
		response := httptest.NewRecorder()

		server.RefreshTokenHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		assertResponseCode(t, response.Code, http.StatusUnauthorized)
	})
}

func refreshTokenRequest(email string, userID int, verified bool) (*http.Request, string) {
	request, _ := http.NewRequest("POST", "/api/v1/users/refresh-token", nil)

	var signingKey = []byte(os.Getenv("SECRET_KEY"))
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["email"] = email
	claims["user_id"] = userID
	claims["verified"] = verified
	claims["exp"] = time.Now().Add(30 * time.Minute).Unix()

	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		fmt.Printf("error generating token: %s", err.Error())
	}

	cookie := http.Cookie{Name: "refresh_token", Value: tokenString, MaxAge: 10, HttpOnly: true}

	request.AddCookie(&cookie)

	return request, tokenString
}

func createUserRequest(data []byte) *http.Request {

	request, _ := http.NewRequest("POST", "/api/v1/users/register", bytes.NewReader(data))

	return request
}

func loginUserRequest(data []byte) *http.Request {

	request, _ := http.NewRequest("POST", "/api/v1/users/login", bytes.NewReader(data))

	return request
}

func logoutUserRequest() *http.Request {

	request, _ := http.NewRequest("POST", "/api/v1/users/logout", nil)

	cookie := http.Cookie{Name: "refresh_token", Value: "somereandomtoken", MaxAge: 10, HttpOnly: true}

	request.AddCookie(&cookie)

	return request
}

func verifyOTPRequest(data []byte) *http.Request {
	request, _ := http.NewRequest("POST", "/auth/verify", bytes.NewReader(data))

	return request
}

func resetPasswordRequest(email string, data []byte) *http.Request {
	ctx := context.WithValue(context.Background(), "email", email)
	request, _ := http.NewRequestWithContext(ctx, "POST", "/auth/reset", bytes.NewReader(data))

	return request
}

func forgotPasswordRequest(data []byte) *http.Request {
	request, _ := http.NewRequest("POST", "/auth/forgot-password", bytes.NewReader(data))

	return request
}

func assertResponseCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("response code = %d, want %d", got, want)
	}
}

func assertResponseBody(t *testing.T, got, want map[string]interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("response body = %v, want %v", got, want)
	}
}
