package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/user"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var CreateUserError = errors.New("error creating user")

type StubUserStore struct {
	users map[string]interface{}
}

func (s *StubUserStore) CreateUser(body user.CreateUserBody) (user.CreateUserResponse, error) {

	return user.CreateUserResponse{ID: 1, FirstName: body.FirstName, LastName: body.LastName, Username: body.Username}, nil
}

type FailingStubUserStore struct {
	users map[string]interface{}
}

func (s *FailingStubUserStore) CreateUser(body user.CreateUserBody) (user.CreateUserResponse, error) {

	return user.CreateUserResponse{}, CreateUserError
}

func TestPOSTUser(t *testing.T) {
	store := StubUserStore{make(map[string]interface{})}
	server := &user.Handler{Store: &store}
	t.Run("create and send a user back", func(t *testing.T) {
		data := []byte(`{ "first_name": "Adedunmola", "last_name": "Oyewale", "username": "Adedunmola", "password": "password" }`)

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
	})

	t.Run("fails in creating user", func(t *testing.T) {
		store := FailingStubUserStore{make(map[string]interface{})}
		server := &user.Handler{Store: &store}
		data := []byte(`{ "first_name": "Adedunmola", "last_name": "Oyewale", "username": "Adedunmola", "password": "password" }`)

		request := createUserRequest(data)
		response := httptest.NewRecorder()

		server.CreateUserHandler(response, request)

		var got map[string]interface{}
		json.Unmarshal(response.Body.Bytes(), &got)

		fmt.Println(got)
		//want := map[string]interface{}{}

		assertResponseCode(t, response.Code, http.StatusInternalServerError)

		//assertResponseBody(t, got, want)
	})
}

func createUserRequest(data []byte) *http.Request {

	request, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewReader(data))

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
