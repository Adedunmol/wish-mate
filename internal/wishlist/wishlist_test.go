package wishlist_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/wishlist"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

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

func (s *StubWishlistStore) GetWishlistByID(wishlistID, userID int) (wishlist.WishlistResponse, error) {
	var response wishlist.WishlistResponse

	for _, w := range s.wishlists {
		if w.ID == wishlistID && w.UserID == userID {
			response = w
			return response, nil
		} else if w.ID == wishlistID && w.UserID != userID {
			response.UserID = w.UserID
			response.ID = w.ID
			response.Description = w.Description
			response.Name = w.Name

			var items []wishlist.ItemResponse
			for _, i := range w.Items {
				if !i.Taken {
					items = append(items, wishlist.ItemResponse{
						ID:          i.ID,
						Name:        i.Name,
						Description: i.Description,
						Whole:       i.Whole,
						Taken:       i.Taken,
					})
				}
			}

			response.Items = items

			return response, nil
		}

	}

	return wishlist.WishlistResponse{}, helpers.ErrNotFound
}

func (s *StubWishlistStore) UpdateWishlistByID(wishlistID, userID int, body wishlist.UpdateWishlist) (wishlist.WishlistResponse, error) {
	var response wishlist.WishlistResponse

	for _, w := range s.wishlists {

		if w.ID == wishlistID && w.UserID != userID {
			return response, helpers.ErrForbidden
		}

		if w.ID == wishlistID && w.UserID == userID {

			if body.Name != "" {
				response.Name = body.Name
			} else {
				response.Name = w.Name
			}

			if body.Description != "" {
				response.Description = body.Description
			} else {
				response.Description = w.Description
			}

			return response, nil
		}
	}

	return wishlist.WishlistResponse{}, helpers.ErrNotFound
}

func (s *StubWishlistStore) DeleteWishlistByID(wishlistID, userID int) error {

	for _, w := range s.wishlists {

		if w.ID == wishlistID && w.UserID != userID {
			return helpers.ErrForbidden
		}

		if w.ID == wishlistID && w.UserID == userID {
			return nil
		}
	}

	return helpers.ErrNotFound
}

func TestCreateWishlist(t *testing.T) {
	store := StubWishlistStore{wishlists: make([]wishlist.WishlistResponse, 0)}
	userStore := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
	}}
	server := wishlist.Handler{Store: &store, UserStore: &userStore}

	t.Run("create and return a wishlist (with items)", func(t *testing.T) {

		data := map[string]interface{}{
			"name":          "Birthday list",
			"description":   "some random description",
			"notify_before": 7,
			"items": []map[string]interface{}{
				{"name": "phone", "description": ""},
				{"name": "bag", "description": ""},
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
					{"id": float64(1), "name": "phone", "description": "", "taken": false},
					{"id": float64(2), "name": "bag", "description": "", "taken": false},
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
	user1 := auth.User{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"}
	user2 := auth.User{ID: 2, FirstName: "Ade", LastName: "Oyewale", Password: "password", Email: "ade@gmail.com", Username: "Ade"}

	store := StubWishlistStore{wishlists: []wishlist.WishlistResponse{
		{ID: 1, UserID: user1.ID, Name: "Birthday list", Description: "some random description", NotifyBefore: 7, Items: []wishlist.ItemResponse{
			{ID: 1, Name: "phone", Description: "", Taken: true},
			{ID: 2, Name: "bag", Description: "", Taken: false},
		}},
	}}
	userStore := StubUserStore{users: []auth.User{
		user1,
		user2,
	}}
	server := wishlist.Handler{Store: &store, UserStore: &userStore}

	t.Run("return a wishlist (owner)", func(t *testing.T) {

		request := getWishlistRequest(user1.ID, 1)
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
					{"id": float64(1), "name": "phone", "description": "", "taken": true},
					{"id": float64(2), "name": "bag", "description": "", "taken": false},
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
		request := getWishlistRequest(user2.ID, 1)
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
					{"id": float64(2), "name": "bag", "description": "", "taken": false},
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

		wantBody := map[string]interface{}{
			"message": "resource not found",
		}

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

		wantBody := map[string]interface{}{
			"message": "id is required",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})
}

func TestUpdateWishlist(t *testing.T) {
	user1 := auth.User{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"}
	user2 := auth.User{ID: 2, FirstName: "Ade", LastName: "Oyewale", Password: "password", Email: "ade@gmail.com", Username: "Ade"}

	store := StubWishlistStore{wishlists: []wishlist.WishlistResponse{
		{ID: 1, UserID: user1.ID, Name: "Birthday list", Description: "some random description", NotifyBefore: 7, Items: []wishlist.ItemResponse{
			{ID: 1, Name: "phone", Description: "", Taken: true},
			{ID: 2, Name: "bag", Description: "", Taken: false},
		}},
	}}
	userStore := StubUserStore{users: []auth.User{
		user1,
		user2,
	}}
	server := wishlist.Handler{Store: &store, UserStore: &userStore}

	t.Run("update and return a wishlist", func(t *testing.T) {
		data := []byte(`{ "name": "Birthday list 2" }`)

		request := updateWishlistRequest(user1.ID, 1, data)
		response := httptest.NewRecorder()

		server.UpdateWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Wishlist updated successfully",
			"data": map[string]interface{}{
				"name":        "Birthday list 2",
				"description": "some random description",
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return a 404", func(t *testing.T) {
		data := []byte(`{}`)

		request := updateWishlistRequest(user1.ID, 2, data)
		response := httptest.NewRecorder()

		server.UpdateWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "resource not found",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("return a 403 if updating another auth's resource", func(t *testing.T) {
		data := []byte(`{}`)
		request := updateWishlistRequest(user2.ID, 1, data)
		response := httptest.NewRecorder()

		server.UpdateWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "forbidden from accessing the resource",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusForbidden)
		assertResponseBody(t, got, want)
	})

	t.Run("returns error for no id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "id", 1)
		request, _ := http.NewRequestWithContext(ctx, http.MethodPatch, "/wishlist/", nil)
		response := httptest.NewRecorder()

		server.UpdateWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "id is required",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})
}

func TestDeleteWishlist(t *testing.T) {
	user1 := auth.User{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"}
	user2 := auth.User{ID: 2, FirstName: "Ade", LastName: "Oyewale", Password: "password", Email: "ade@gmail.com", Username: "Ade"}

	store := StubWishlistStore{wishlists: []wishlist.WishlistResponse{
		{ID: 1, UserID: user1.ID, Name: "Birthday list", Description: "some random description", NotifyBefore: 7, Items: []wishlist.ItemResponse{
			{ID: 1, Name: "phone", Description: "", Taken: true},
			{ID: 2, Name: "bag", Description: "", Taken: false},
		}},
	}}
	userStore := StubUserStore{users: []auth.User{
		user1,
		user2,
	}}
	server := wishlist.Handler{Store: &store, UserStore: &userStore}

	t.Run("delete a wishlist", func(t *testing.T) {

		request := deleteWishlistRequest(user1.ID, 1)
		response := httptest.NewRecorder()

		server.DeleteWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Wishlist deleted successfully",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return a 404", func(t *testing.T) {
		request := deleteWishlistRequest(user1.ID, 2)
		response := httptest.NewRecorder()

		server.DeleteWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "resource not found",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("return a 403 if updating another auth's resource", func(t *testing.T) {
		request := deleteWishlistRequest(user2.ID, 1)
		response := httptest.NewRecorder()

		server.DeleteWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "forbidden from accessing the resource",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusForbidden)
		assertResponseBody(t, got, want)
	})

	t.Run("returns error for no id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "id", 1)
		request, _ := http.NewRequestWithContext(ctx, http.MethodDelete, "/wishlist/", nil)
		response := httptest.NewRecorder()

		server.DeleteWishlist(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "id is required",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})
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

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/wishlist/%s", fmt.Sprint(wishlistID)), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprint(wishlistID))

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}

func updateWishlistRequest(userID, wishlistID int, body []byte) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/wishlist/%s", fmt.Sprint(wishlistID)), bytes.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprint(wishlistID))

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}

func deleteWishlistRequest(userID, wishlistID int) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/wishlist/%s", fmt.Sprint(wishlistID)), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprint(wishlistID))

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}
