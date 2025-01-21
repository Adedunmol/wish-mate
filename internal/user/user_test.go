package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/Adedunmol/wish-mate/internal/user"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
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

type StubUserStore struct {
	users []user.User
}

func (s *StubUserStore) CreateUser(body *user.CreateUserBody) (user.CreateUserResponse, error) {

	for _, u := range s.users {
		if u.Email == body.Email {
			return user.CreateUserResponse{}, helpers.ErrConflict
		}
	}

	userData := user.User{ID: 1, FirstName: body.FirstName, LastName: body.LastName, Username: body.Username, Email: body.Email, Password: body.Password}

	s.users = append(s.users, userData)

	return user.CreateUserResponse{ID: userData.ID, FirstName: userData.FirstName, LastName: userData.LastName, Username: userData.Username}, nil
}

func (s *StubUserStore) FindUserByEmail(email string) (user.User, error) {

	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return user.User{}, helpers.ErrNotFound
}

func (s *StubUserStore) ComparePasswords(storedPassword, candidatePassword string) bool {
	return storedPassword == candidatePassword
}

type FailingStubUserStore struct {
	users []user.User
}

func (s *FailingStubUserStore) CreateUser(_ *user.CreateUserBody) (user.CreateUserResponse, error) {

	return user.CreateUserResponse{}, ErrCreate
}

func (s *FailingStubUserStore) FindUserByEmail(_ string) (user.User, error) {
	return user.User{}, ErrNoEntry
}

func (s *FailingStubUserStore) ComparePasswords(_, _ string) bool {
	return false
}

func TestPOSTUser(t *testing.T) {

	t.Run("create and send a user back", func(t *testing.T) {
		store := StubUserStore{users: make([]user.User, 0)}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}
		server := &user.Handler{Store: &store, Queue: &mockQueue}

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

	t.Run("fails in creating user", func(t *testing.T) {
		store := FailingStubUserStore{users: make([]user.User, 0)}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}
		server := &user.Handler{Store: &store, Queue: &mockQueue}
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
		store := FailingStubUserStore{users: make([]user.User, 0)}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}
		server := &user.Handler{Store: &store, Queue: &mockQueue}
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
		store := StubUserStore{users: []user.User{
			{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
		}}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}

		server := &user.Handler{Store: &store, Queue: &mockQueue}

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
	store := StubUserStore{users: []user.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	server := &user.Handler{Store: &store}

	t.Run("find and log in a user", func(t *testing.T) {

		data := []byte(`{ "email": "adedunmola@gmail.com", "password": "password" }`)

		request := loginUserRequest(data)
		response := httptest.NewRecorder()

		server.LoginUserHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		assertResponseCode(t, response.Code, http.StatusOK)
	})

	t.Run("returns error for invalid request body", func(t *testing.T) {
		store := FailingStubUserStore{users: make([]user.User, 0)}
		server := &user.Handler{Store: &store}
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

	t.Run("does not find a user", func(t *testing.T) {
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

func createUserRequest(data []byte) *http.Request {

	request, _ := http.NewRequest("POST", "/api/v1/users/register", bytes.NewReader(data))

	return request
}

func loginUserRequest(data []byte) *http.Request {

	request, _ := http.NewRequest("POST", "/api/v1/users/login", bytes.NewReader(data))

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
