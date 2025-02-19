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

func (s *StubUserStore) FindUserByID(id int) (auth.User, error) {

	for _, u := range s.users {
		if u.ID == id {
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
			Taken:       false,
			Link:        item.Link,
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
		Date:         body.Date,
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

func (s *StubWishlistStore) GetUserWishlists(userID int, isOwner bool) ([]wishlist.WishlistResponse, error) {
	response := make([]wishlist.WishlistResponse, 0)

	for _, w := range s.wishlists {
		var data wishlist.WishlistResponse
		if w.UserID == userID {

			data.UserID = w.UserID
			data.ID = w.ID
			data.Description = w.Description
			data.Name = w.Name
			data.NotifyBefore = w.NotifyBefore

			var items []wishlist.ItemResponse

			if !isOwner {
				for _, i := range w.Items {
					if !i.Taken {
						items = append(items, wishlist.ItemResponse{
							ID:          i.ID,
							Name:        i.Name,
							Description: i.Description,
							Taken:       i.Taken,
						})
					}
				}
				data.Items = items
			} else {
				data.Items = w.Items
			}

			response = append(response, data)
		}
	}

	return response, nil
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

func (s *StubWishlistStore) GetItem(wishlistID, itemID int) (wishlist.ItemResponse, error) {

	for _, w := range s.wishlists {
		if w.ID == wishlistID {
			for _, i := range w.Items {
				if i.ID == itemID {
					return i, nil
				}
			}
		}
	}

	return wishlist.ItemResponse{}, helpers.ErrNotFound
}

func (s *StubWishlistStore) UpdateItem(wishlistID, itemID int, body *wishlist.UpdateItem) (wishlist.ItemResponse, error) {

	var wish wishlist.WishlistResponse

	for i, w := range s.wishlists {

		if w.ID == wishlistID {

			wish = s.wishlists[i]
		}
	}

	for idx, i := range wish.Items {

		if i.ID == itemID {

			if body.Name != "" {
				wish.Items[idx].Name = body.Name
			}

			if body.Description != "" {
				wish.Items[idx].Description = body.Description
			}

			if body.Link != "" {
				wish.Items[idx].Link = body.Link
			}

			return wish.Items[idx], nil
		}
	}

	return wishlist.ItemResponse{}, helpers.ErrNotFound
}

func (s *StubWishlistStore) DeleteItem(wishlistID, itemID int) error {

	for _, w := range s.wishlists {
		if w.ID == wishlistID {
			for _, i := range w.Items {
				if i.ID == itemID {
					return nil
				}
			}
		}
	}

	return helpers.ErrNotFound
}

func (s *StubWishlistStore) PickItem(wishlistID, itemID, userID int) (wishlist.ItemResponse, error) {

	for _, w := range s.wishlists {
		if w.ID == wishlistID {
			for _, i := range w.Items {
				if i.ID == itemID && !i.Taken {
					i.Taken = true

					return i, nil
				} else if i.ID == itemID && i.Taken {
					return wishlist.ItemResponse{}, helpers.NewHTTPError(nil, http.StatusConflict, "Item picked already", nil)
				}
			}
		}
	}

	return wishlist.ItemResponse{}, helpers.ErrNotFound
}

func TestCreateWishlist(t *testing.T) {
	store := StubWishlistStore{wishlists: make([]wishlist.WishlistResponse, 0)}
	userStore := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola", DateOfBirth: "2020-01-01"},
	}}
	server := wishlist.Handler{Store: &store, UserStore: &userStore}

	t.Run("create and return a wishlist (with items)", func(t *testing.T) {

		data := map[string]interface{}{
			"name":          "Birthday list",
			"description":   "some random description",
			"notify_before": 7,
			"items": []map[string]interface{}{
				{"name": "phone", "description": "", "link": ""},
				{"name": "bag", "description": "", "link": ""},
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
				"date":          "2020-01-01",
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

	t.Run("create and return a wishlist (with date)", func(t *testing.T) {

		data := map[string]interface{}{
			"name":          "Birthday list",
			"description":   "some random description",
			"notify_before": 7,
			"date":          "2025-02-10",
			"items": []map[string]interface{}{
				{"name": "phone", "description": "", "link": ""},
				{"name": "bag", "description": "", "link": ""},
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
				"date":          "2025-02-10",
				"items": []map[string]interface{}{
					{"id": float64(1), "name": "phone", "description": "", "taken": false, "link": ""},
					{"id": float64(2), "name": "bag", "description": "", "taken": false, "link": ""},
				},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusCreated)
		assertResponseBody(t, got, want)
	})

	t.Run("create and return a wishlist (without date uses the friendship's birthday)", func(t *testing.T) {

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
				"date":          "2020-01-01",
				"items": []map[string]interface{}{
					{"id": float64(1), "name": "phone", "description": "", "taken": false, "link": ""},
					{"id": float64(2), "name": "bag", "description": "", "taken": false, "link": ""},
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
				"date":          "2020-01-01",
			},
		}

		assertResponseCode(t, response.Code, http.StatusCreated)
		assertResponseBody(t, got, want)
	})

	t.Run("return error if no friendship is found with the email attached", func(t *testing.T) {

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

func TestGetUserWishlists(t *testing.T) {
	user1 := auth.User{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"}
	user2 := auth.User{ID: 2, FirstName: "Ade", LastName: "Oyewale", Password: "password", Email: "ade@gmail.com", Username: "Ade"}

	store := StubWishlistStore{wishlists: []wishlist.WishlistResponse{
		{ID: 1, UserID: user1.ID, Name: "Birthday list", Description: "some random description", NotifyBefore: 7, Items: []wishlist.ItemResponse{
			{ID: 1, Name: "phone", Description: "", Taken: true},
			{ID: 2, Name: "bag", Description: "", Taken: false},
		}},
		{ID: 2, UserID: user1.ID, Name: "Grad list", Description: "some random description", NotifyBefore: 7, Items: []wishlist.ItemResponse{
			{ID: 1, Name: "clothes", Description: "", Taken: true},
			{ID: 2, Name: "shoes", Description: "", Taken: false},
		}},
	}}
	userStore := StubUserStore{users: []auth.User{
		user1,
		user2,
	}}
	server := wishlist.Handler{Store: &store, UserStore: &userStore}

	t.Run("get friendship's wishlists (owner)", func(t *testing.T) {

		request := getUserWishlistsRequest(user1.ID)
		response := httptest.NewRecorder()

		server.GetAllWishlists(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Wishlists retrieved successfully",
			"data": []map[string]interface{}{
				{
					"id":            float64(1),
					"user_id":       float64(1),
					"name":          "Birthday list",
					"description":   "some random description",
					"notify_before": float64(7),
					"items": []map[string]interface{}{
						{"id": float64(1), "name": "phone", "description": "", "link": "", "taken": true},
						{"id": float64(2), "name": "bag", "description": "", "link": "", "taken": false},
					},
				},
				{
					"id":            float64(2),
					"user_id":       float64(1),
					"name":          "Grad list",
					"description":   "some random description",
					"notify_before": float64(7),
					"items": []map[string]interface{}{
						{"id": float64(1), "name": "clothes", "description": "", "link": "", "taken": true},
						{"id": float64(2), "name": "shoes", "description": "", "link": "", "taken": false},
					},
				},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("get friendship's wishlists (others)", func(t *testing.T) {

		ctx := context.WithValue(context.Background(), "user_id", user2.ID)
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("users/%s/wishlists", fmt.Sprint(user1.ID)), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("user_id", fmt.Sprint(user1.ID))

		request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

		response := httptest.NewRecorder()

		server.GetAllWishlists(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Wishlists retrieved successfully",
			"data": []map[string]interface{}{
				{
					"id":            float64(1),
					"user_id":       float64(1),
					"name":          "Birthday list",
					"description":   "some random description",
					"notify_before": float64(7),
					"items": []map[string]interface{}{
						{"id": float64(2), "name": "bag", "description": "", "link": "", "taken": false},
					},
				},
				{
					"id":            float64(2),
					"user_id":       float64(1),
					"name":          "Grad list",
					"description":   "some random description",
					"notify_before": float64(7),
					"items": []map[string]interface{}{
						{"id": float64(2), "name": "shoes", "description": "", "link": "", "taken": false},
					},
				},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for no friendship with the id", func(t *testing.T) {

		request := getUserWishlistsRequest(3)
		response := httptest.NewRecorder()

		server.GetAllWishlists(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "no friendship found with the id",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusNotFound)
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

func TestUpdateItem(t *testing.T) {

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

	t.Run("update and return item", func(t *testing.T) {
		data := []byte(`{ "link": "https://random.com/item" }`)

		request := updateItemRequest(user1.ID, 1, 1, data)
		response := httptest.NewRecorder()

		server.UpdateWishlistItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Item updated successfully",
			"data": map[string]interface{}{
				"id":          float64(1),
				"name":        "phone",
				"description": "",
				"link":        "https://random.com/item",
				"taken":       true,
			},
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for no item found with id", func(t *testing.T) {
		data := []byte(`{ "link": "https://random.com/item" }`)

		request := updateItemRequest(user1.ID, 1, 10, data)
		response := httptest.NewRecorder()

		server.UpdateWishlistItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource not found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("return 403 for accessing another friendship's resource", func(t *testing.T) {
		data := []byte(`{ "link": "https://random.com/item" }`)

		request := updateItemRequest(user2.ID, 1, 1, data)
		response := httptest.NewRecorder()

		server.UpdateWishlistItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "forbidden from accessing the resource",
		}

		assertResponseCode(t, response.Code, http.StatusForbidden)
		assertResponseBody(t, got, want)
	})
}

func TestPickItem(t *testing.T) {

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

	t.Run("pick and return item", func(t *testing.T) {

		request := pickItemRequest(user1.ID, 1, 2)
		response := httptest.NewRecorder()

		server.PickWishlistItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Item picked successfully",
			"data": map[string]interface{}{
				"id":          float64(2),
				"name":        "bag",
				"description": "",
				"link":        "",
				"taken":       true,
			},
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return 409 for trying to pick a picked item", func(t *testing.T) {

		request := pickItemRequest(user1.ID, 1, 1)
		response := httptest.NewRecorder()

		server.PickWishlistItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "Item picked already",
		}

		assertResponseCode(t, response.Code, http.StatusConflict)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for no item found with id", func(t *testing.T) {

		request := pickItemRequest(user1.ID, 10, 10)
		response := httptest.NewRecorder()

		server.PickWishlistItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource not found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})
}

func TestGetItem(t *testing.T) {
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

	t.Run("return item", func(t *testing.T) {
		request := getItemRequest(user1.ID, 1, 1)
		response := httptest.NewRecorder()

		server.GetItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Item retrieved successfully",
			"data": map[string]interface{}{
				"id":          float64(1),
				"name":        "phone",
				"description": "",
				"link":        "",
				"taken":       true,
			},
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return a 400 for no item/wishlist id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", user1.ID)
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/wishlists/%s/items/%s", fmt.Sprint(1), ""), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("wishlist_id", fmt.Sprint(1))

		request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
		response := httptest.NewRecorder()

		server.GetItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "item id is required",
		}

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})

	t.Run("return a 404 for no item with id", func(t *testing.T) {
		request := getItemRequest(user1.ID, 1, 10)
		response := httptest.NewRecorder()

		server.GetItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource not found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})
}

func TestDeleteItem(t *testing.T) {
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

	t.Run("delete item", func(t *testing.T) {
		request := deleteItemRequest(user1.ID, 1, 1)
		response := httptest.NewRecorder()

		server.DeleteItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Item deleted successfully",
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return a 403 for accessing another friendship's item", func(t *testing.T) {
		request := deleteItemRequest(user2.ID, 1, 1)
		response := httptest.NewRecorder()

		server.DeleteItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "forbidden from accessing the resource",
		}

		assertResponseCode(t, response.Code, http.StatusForbidden)
		assertResponseBody(t, got, want)
	})

	t.Run("return a 400 for no item/wishlist id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", user1.ID)
		request, _ := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/wishlists/%s/items/%s", fmt.Sprint(1), ""), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("wishlist_id", fmt.Sprint(1))

		request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
		response := httptest.NewRecorder()

		server.DeleteItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "item id is required",
		}

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})

	t.Run("return a 404 for no item with id", func(t *testing.T) {
		request := deleteItemRequest(user1.ID, 1, 10)
		response := httptest.NewRecorder()

		server.DeleteItemHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource not found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
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

func getUserWishlistsRequest(userID int) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("users/%s/wishlists", fmt.Sprint(userID)), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", fmt.Sprint(userID))

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}

func getItemRequest(userID, wishlistID, itemID int) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/wishlists/%s/items/%s", fmt.Sprint(wishlistID), fmt.Sprint(itemID)), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("wishlist_id", fmt.Sprint(wishlistID))
	rctx.URLParams.Add("item_id", fmt.Sprint(itemID))

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}

func deleteItemRequest(userID, wishlistID, itemID int) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/wishlists/%s/items/%s", fmt.Sprint(wishlistID), fmt.Sprint(itemID)), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("wishlist_id", fmt.Sprint(wishlistID))
	rctx.URLParams.Add("item_id", fmt.Sprint(itemID))

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

func updateItemRequest(userID, wishlistID, itemID int, body []byte) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/wishlists/%s/items/%s", fmt.Sprint(wishlistID), fmt.Sprint(itemID)), bytes.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("wishlist_id", fmt.Sprint(wishlistID))
	rctx.URLParams.Add("item_id", fmt.Sprint(itemID))

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}

func pickItemRequest(userID, wishlistID, itemID int) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/wishlists/%s/items/%s", fmt.Sprint(wishlistID), fmt.Sprint(itemID)), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("wishlist_id", fmt.Sprint(wishlistID))
	rctx.URLParams.Add("item_id", fmt.Sprint(itemID))

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
