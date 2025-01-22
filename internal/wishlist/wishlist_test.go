package wishlist_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	wishlists []wishlist.WishlistResponse
}

func (s *StubWishlistStore) CreateWishlist(userID int, body wishlist.Wishlist) (wishlist.WishlistResponse, error) {

	var items []wishlist.ItemResponse
	id := 1
	for _, item := range body.Items {
		items = append(items, wishlist.ItemResponse{
			ID:          id,
			Name:        item.Name,
			Description: item.Description,
			Whole:       item.Whole,
			Taken:       false,
		})

		id += 1
	}

	wishlistData := wishlist.WishlistResponse{
		ID:           1,
		UserID:       userID,
		Name:         body.Name,
		Description:  body.Description,
		NotifyBefore: body.NotifyBefore,
		Items:        items,
	}

	s.wishlists = append(s.wishlists, wishlistData)

	return wishlistData, nil
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
	store := StubWishlistStore{wishlists: make([]wishlist.WishlistResponse, 0)}
	userStore := StubUserStore{users: []user.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	server := wishlist.Handler{Store: &store, UserStore: &userStore}

	t.Run("create and return a wishlist (with items)", func(t *testing.T) {

		data := map[string]interface{}{
			"name":          "Birthday list",
			"description":   "some random description",
			"notify_before": 7,
			"items": []map[string]interface{}{
				{"name": "phone", "description": "", "whole": true},
				{"name": "bag", "description": "", "whole": true},
			},
		}

		body, _ := json.Marshal(data)
		request := createWishlistRequest(body, "adedunmola@gmail.com")
		response := httptest.NewRecorder()

		server.CreateWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Wishlist created successfully",
			"data": map[string]interface{}{
				"id":            float64(1),
				"user_id":       float64(1),
				"name":          "Birthday list",
				"description":   "some random description",
				"notify_before": float64(7),
				"items": []map[string]interface{}{
					{"id": float64(1), "name": "phone", "description": "", "whole": true, "taken": false},
					{"id": float64(2), "name": "bag", "description": "", "whole": true, "taken": false},
				},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusCreated)
		assertResponseBody(t, got, want)
	})

	t.Run("create and return a wishlist (without items)", func(t *testing.T) {

		data := map[string]interface{}{
			"name":          "Birthday list",
			"description":   "some random description",
			"notify_before": 7,
		}

		body, _ := json.Marshal(data)
		request := createWishlistRequest(body, "adedunmola@gmail.com")
		response := httptest.NewRecorder()

		server.CreateWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Wishlist created successfully",
			"data": map[string]interface{}{
				"id":            float64(1),
				"user_id":       float64(1),
				"name":          "Birthday list",
				"description":   "some random description",
				"notify_before": float64(7),
			},
		}

		assertResponseCode(t, response.Code, http.StatusCreated)
		assertResponseBody(t, got, want)
	})

	t.Run("return error if no user is found with the email attached", func(t *testing.T) {

		data := map[string]interface{}{
			"name":          "Birthday list",
			"description":   "some random description",
			"notify_before": 7,
		}

		body, _ := json.Marshal(data)
		request := createWishlistRequest(body, "adedunmola123@gmail.com")
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
			"description":   "some random description",
			"notify_before": 7,
		}

		body, _ := json.Marshal(data)
		request := createWishlistRequest(body, "adedunmola@gmail.com")
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
	user1 := user.User{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"}
	user2 := user.User{ID: 2, FirstName: "Ade", LastName: "Oyewale", Password: "password", Email: "ade@gmail.com", Username: "Ade"}

	store := StubWishlistStore{wishlists: []wishlist.WishlistResponse{
		{ID: 1, UserID: user1.ID, Name: "Birthday list", Description: "some random description", NotifyBefore: 7, Items: []wishlist.ItemResponse{
			{ID: 2, Name: "bag", Description: "", Whole: true, Taken: false},
			{ID: 1, Name: "phone", Description: "", Whole: true, Taken: true},
		}},
	}}
	userStore := StubUserStore{users: []user.User{
		user1,
		user2,
	}}
	server := wishlist.Handler{Store: &store, UserStore: &userStore}

	t.Run("return a wishlist (owner)", func(t *testing.T) {

		request := getWishlistRequest(1, 1)
		response := httptest.NewRecorder()

		server.GetWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Wishlist retrieved successfully",
			"data": map[string]interface{}{
				"id":            float64(1),
				"user_id":       float64(1),
				"name":          "Birthday list",
				"description":   "some random description",
				"notify_before": float64(7),
				"items": []map[string]interface{}{
					{"id": float64(1), "name": "phone", "description": "", "whole": true, "taken": true},
					{"id": float64(2), "name": "bag", "description": "", "whole": true, "taken": false},
				},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return a wishlist (others)", func(t *testing.T) {
		request := getWishlistRequest(2, 1)
		response := httptest.NewRecorder()

		server.GetWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Wishlist retrieved successfully",
			"data": map[string]interface{}{
				"id":          float64(1),
				"user_id":     float64(1),
				"name":        "Birthday list",
				"description": "some random description",
				"items": []map[string]interface{}{
					{"id": float64(2), "name": "bag", "description": "", "whole": true, "taken": false},
				},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return a 404", func(t *testing.T) {
		request := getWishlistRequest(1, 2)
		response := httptest.NewRecorder()

		server.GetWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("returns error for no id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "id", 1)
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/wishlist", nil)
		response := httptest.NewRecorder()

		server.GetWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})
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

func createWishlistRequest(data []byte, email string) *http.Request {

	ctx := context.WithValue(context.Background(), "email", email)
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/wishlist", bytes.NewReader(data))

	return request
}

func getWishlistRequest(userID, wishlistID int) *http.Request {

	ctx := context.WithValue(context.Background(), "id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/wishlist/%s", fmt.Sprint(wishlistID)), nil)

	return request
}
