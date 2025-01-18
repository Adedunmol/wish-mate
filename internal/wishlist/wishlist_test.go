package wishlist_test

import (
	"reflect"
	"testing"
)

func TestCreateWishlist(t *testing.T) {

	t.Run("create and return a wishlist", func(t *testing.T) {})

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
