package wishlist_test

import (
	"bytes"
	"encoding/json"
	"github.com/Adedunmol/wish-mate/internal/wishlist"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type Store struct {
	wishlists []wishlist.Wishlist
}

func (s *Store) CreateWishlist(userID int, body wishlist.Wishlist) (wishlist.WishlistResponse, error) {
	return wishlist.WishlistResponse{}, nil
}

func (s *Store) GetWishlistByID(id int, verbose bool) (wishlist.WishlistResponse, error) {
	return wishlist.WishlistResponse{}, nil
}

func (s *Store) UpdateWishlistByID(id int, body wishlist.Wishlist) (wishlist.WishlistResponse, error) {
	return wishlist.WishlistResponse{}, nil
}

func (s *Store) DeleteWishlistByID(id int) error {
	return nil
}

func TestCreateWishlist(t *testing.T) {

	t.Run("create and return a wishlist (with items)", func(t *testing.T) {
		store := Store{wishlists: make([]wishlist.Wishlist, 0)}
		server := wishlist.Handler{Store: &store}

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

	t.Run("returns error for invalid request body", func(t *testing.T) {})
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
