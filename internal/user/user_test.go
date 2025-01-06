package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Adedunmol/wish-mate/internal/user"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	ErrCreate   = errors.New("error creating entry")
	ErrNoEntry  = errors.New("no entry found")
	ErrConflict = errors.New("conflict")
)

type StubUserStore struct {
	users []user.User
}

func (s *StubUserStore) CreateUser(body user.CreateUserBody) (user.CreateUserResponse, error) {

	for _, u := range s.users {
		if u.Email == body.Email {
			return user.CreateUserResponse{}, ErrCreate
		}
	}

	userData := user.User{ID: 1, FirstName: body.FirstName, LastName: body.LastName, Username: body.Username, Email: body.Email, Password: body.Password}

	s.users = append(s.users, userData)

	return user.CreateUserResponse{ID: userData.ID, FirstName: userData.FirstName, LastName: userData.LastName, Username: userData.Username}, nil
}

func (s *StubUserStore) FindUserByEmail(email string) (user.User, error) {

	return s.users[0], nil
}

func (s *StubUserStore) ComparePasswords(candidatePassword string) bool {
	return s.users[0].Password == candidatePassword
}

type FailingStubUserStore struct {
	users []user.User
}

func (s *FailingStubUserStore) CreateUser(body user.CreateUserBody) (user.CreateUserResponse, error) {

	return user.CreateUserResponse{}, ErrCreate
}

func (s *FailingStubUserStore) FindUserByEmail(email string) (user.User, error) {
	return user.User{}, ErrNoEntry
}

func (s *FailingStubUserStore) ComparePasswords(candidatePassword string) bool {
	return false
}

func TestPOSTUser(t *testing.T) {

	t.Run("create and send a user back", func(t *testing.T) {
		store := StubUserStore{users: make([]user.User, 0)}
		server := &user.Handler{Store: &store}

		data := []byte(`{ "first_name": "Adedunmola", "last_name": "Oyewale", "username": "Adedunmola", "password": "password", "email": "adedunmola@gmail.com" }`)

		request := createUserRequest(data)
		response := httptest.NewRecorder()

		server.CreateUserHandler(response, request)

		var got map[string]interface{}
		json.Unmarshal(response.Body.Bytes(), &got)

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
	})

	t.Run("fails in creating user", func(t *testing.T) {
		store := FailingStubUserStore{users: make([]user.User, 0)}
		server := &user.Handler{Store: &store}
		data := []byte(`{ "first_name": "Adedunmola", "last_name": "Oyewale", "username": "Adedunmola", "password": "password" }`)

		request := createUserRequest(data)
		response := httptest.NewRecorder()

		server.CreateUserHandler(response, request)

		var got map[string]interface{}
		json.Unmarshal(response.Body.Bytes(), &got)

		//want := map[string]interface{}{}

		assertResponseCode(t, response.Code, http.StatusInternalServerError)

		//assertResponseBody(t, got, want)
	})

	t.Run("email conflict", func(t *testing.T) {
		store := StubUserStore{users: []user.User{
			{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
		}}
		server := &user.Handler{Store: &store}

		data := []byte(`{ "first_name": "Adedunmola", "last_name": "Oyewale", "username": "Adedunmola", "password": "password", "email": "adedunmola@gmail.com" }`)

		request := createUserRequest(data)
		response := httptest.NewRecorder()

		server.CreateUserHandler(response, request)

		var got map[string]interface{}
		json.Unmarshal(response.Body.Bytes(), &got)

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

		assertResponseCode(t, response.Code, http.StatusConflict)
		assertResponseBody(t, got, want)
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
		json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "User logged in successfully",
			"data": map[string]interface{}{
				"token":     "somerandomaccesstoken",
				"expiresAt": float64(36000),
			},
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("does not find a user", func(t *testing.T) {
		data := []byte(`{ "email": "adedunmola1@gmail.com", "password": "password123" }`)

		request := loginUserRequest(data)
		response := httptest.NewRecorder()

		server.LoginUserHandler(response, request)

		var got map[string]interface{}
		json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Error",
			"message": "Invalid credentials",
			"data":    nil,
		}

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})

	t.Run("incorrect password", func(t *testing.T) {

		data := []byte(`{ "email": "adedunmola@gmail.com", "password": "password123" }`)

		request := loginUserRequest(data)
		response := httptest.NewRecorder()

		server.LoginUserHandler(response, request)

		var got map[string]interface{}
		json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Error",
			"message": "Invalid credentials",
			"data":    nil,
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
