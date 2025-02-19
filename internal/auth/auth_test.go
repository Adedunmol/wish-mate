package auth_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"net/http"
	"net/http/httptest"
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
			}

			if otpData.ExpiresAt.Before(time.Now()) {
				return false, helpers.ErrBadRequest
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

	t.Run("validate the otp and update user's verified status", func(t *testing.T) {
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

	t.Run("send otp to user", func(t *testing.T) {
		data := []byte(`{ "email": "adedunmola@gmail.com" }`)

		request := verifyOTPRequest(data)
		response := httptest.NewRecorder()

		server.RequestCodeHandler(response, request)

		assertResponseCode(t, response.Code, http.StatusOK)
	})

	t.Run("no user found with email", func(t *testing.T) {
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

func createUserRequest(data []byte) *http.Request {

	request, _ := http.NewRequest("POST", "/api/v1/users/register", bytes.NewReader(data))

	return request
}

func loginUserRequest(data []byte) *http.Request {

	request, _ := http.NewRequest("POST", "/api/v1/users/login", bytes.NewReader(data))

	return request
}

func verifyOTPRequest(data []byte) *http.Request {
	request, _ := http.NewRequest("POST", "/auth/verify", bytes.NewReader(data))

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
