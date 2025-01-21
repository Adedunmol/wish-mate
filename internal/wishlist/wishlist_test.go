package wishlist_test

import (
	"bytes"
	"encoding/json"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/user"
	"github.com/Adedunmol/wish-mate/internal/wishlist"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

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

type StubWishlistStore struct {
	wishlists []wishlist.Wishlist
}

func (s *StubWishlistStore) CreateWishlist(userID int, body wishlist.Wishlist) (wishlist.WishlistResponse, error) {
	return wishlist.WishlistResponse{}, nil
}

func (s *StubWishlistStore) GetWishlistByID(id int, verbose bool) (wishlist.WishlistResponse, error) {
	return wishlist.WishlistResponse{}, nil
}

func (s *StubWishlistStore) UpdateWishlistByID(id int, body wishlist.Wishlist) (wishlist.WishlistResponse, error) {
	return wishlist.WishlistResponse{}, nil
}

func (s *StubWishlistStore) DeleteWishlistByID(id int) error {
	return nil
}

func TestCreateWishlist(t *testing.T) {
	store := StubWishlistStore{wishlists: make([]wishlist.Wishlist, 0)}
	userStore := StubUserStore{users: []user.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	server := wishlist.Handler{Store: &store, UserStore: &userStore}

	t.Run("create and return a wishlist (with items)", func(t *testing.T) {

		data := map[string]interface{}{
			"name":        "Birthday list",
			"description": "some random description",
			"items": []map[string]interface{}{
				{"name": "phone", "description": "", "whole": true},
				{"name": "bag", "description": "", "whole": true},
			},
		}

		body, _ := json.Marshal(data)
		request := createWishlistRequest(body)
		response := httptest.NewRecorder()

		server.CreateWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Wishlist created successfully",
			"data": map[string]interface{}{
				"id":          float64(1),
				"user_id":     float64(1),
				"name":        "Birthday list",
				"description": "some random description",
				"items": []map[string]interface{}{
					{"id": float64(1), "name": "phone", "description": "", "whole": true, "taken": false},
					{"id": float64(2), "name": "bag", "description": "", "whole": true, "taken": false},
				},
			},
		}

		assertResponseCode(t, response.Code, http.StatusCreated)
		assertResponseBody(t, got, want)
	})

	t.Run("create and return a wishlist (without items)", func(t *testing.T) {

		data := map[string]interface{}{
			"name":        "Birthday list",
			"description": "some random description",
		}

		body, _ := json.Marshal(data)
		request := createWishlistRequest(body)
		response := httptest.NewRecorder()

		server.CreateWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Wishlist created successfully",
			"data": map[string]interface{}{
				"id":          float64(1),
				"user_id":     float64(1),
				"name":        "Birthday list",
				"description": "some random description",
			},
		}

		assertResponseCode(t, response.Code, http.StatusCreated)
		assertResponseBody(t, got, want)
	})

	t.Run("return error if no user is found with the email attached", func(t *testing.T) {

		data := map[string]interface{}{
			"name":        "Birthday list",
			"description": "some random description",
		}

		body, _ := json.Marshal(data)
		request := createWishlistRequest(body)
		response := httptest.NewRecorder()

		server.CreateWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "invalid credentials",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusUnauthorized)
		assertResponseBody(t, got, want)
	})

	t.Run("returns error for invalid request body", func(t *testing.T) {
		data := map[string]interface{}{
			"description": "some random description",
		}

		body, _ := json.Marshal(data)
		request := createWishlistRequest(body)
		response := httptest.NewRecorder()

		server.CreateWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "invalid request body",
			"problems": map[string][]string{
				"Name": []string{"Name required"},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})
}

func TestGetWishlist(t *testing.T) {

	t.Run("return a wishlist", func(t *testing.T) {})

	t.Run("return a 404", func(t *testing.T) {})

	t.Run("returns error for no id", func(t *testing.T) {})
}

func TestUpdateWishlist(t *testing.T) {

	t.Run("update and return a wishlist", func(t *testing.T) {})

	t.Run("return a 404", func(t *testing.T) {})

	t.Run("return a 403 if updating another user's resource", func(t *testing.T) {})

	t.Run("returns error for no id", func(t *testing.T) {})
}

func TestDeleteWishlist(t *testing.T) {

	t.Run("delete a wishlist", func(t *testing.T) {})

	t.Run("return a 404", func(t *testing.T) {})

	t.Run("return a 403 if updating another user's resource", func(t *testing.T) {})

	t.Run("returns error for no id", func(t *testing.T) {})
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

func createWishlistRequest(data []byte) *http.Request {

	request, _ := http.NewRequest(http.MethodPost, "/wishlist", bytes.NewReader(data))

	return request
}
