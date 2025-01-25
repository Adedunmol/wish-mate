package user_test

import "testing"

func TestSendRequest(t *testing.T) {

	t.Run("send a request and return the entry", func(t *testing.T) {})

	t.Run("return 404 for no auth with the id", func(t *testing.T) {})

	t.Run("return 409 if friendship exists already", func(t *testing.T) {})

	t.Run("return bad request for empty auth id", func(t *testing.T) {})
}

func TestAcceptRequest(t *testing.T) {

	t.Run("accept a request and return the entry", func(t *testing.T) {})

	t.Run("return 404 for no entry with the request id", func(t *testing.T) {})

	t.Run("return bad request for empty request id", func(t *testing.T) {})
}

func TestGetAllFriendships(t *testing.T) {

	t.Run("return all friendships (pending)", func(t *testing.T) {})

	t.Run("return all friendships (accepted)", func(t *testing.T) {})

	t.Run("return all friendships (blocked)", func(t *testing.T) {})
}

func TestGetFriendship(t *testing.T) {

	t.Run("return a friendship", func(t *testing.T) {})
}

func UpdateFriendship(t *testing.T) {

	t.Run("update and return a friendship", func(t *testing.T) {})

	t.Run("return forbidden", func(t *testing.T) {})

	t.Run("invalid request body", func(t *testing.T) {})

	t.Run("return not found for friendship id not found", func(t *testing.T) {})
}
