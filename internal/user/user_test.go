package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"github.com/Adedunmol/wish-mate/internal/user"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type StubQueue struct {
	Tasks []queue.TaskPayload
}

func (q *StubQueue) Enqueue(taskPayload *queue.TaskPayload) error {
	q.Tasks = append(q.Tasks, *taskPayload)
	return nil
}

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

type StubFriendStore struct {
	friends []user.FriendshipResponse
}

func (s *StubFriendStore) CreateFriendship(userID, recipientID int) (user.FriendshipResponse, error) {

	data := user.FriendshipResponse{
		ID:       1,
		UserID:   userID,
		FriendID: recipientID,
		Status:   "pending",
	}

	s.friends = append(s.friends, data)

	return data, nil
}

func (s *StubFriendStore) UpdateFriendship(friendshipID int, status string) (user.FriendshipResponse, error) {

	for i, u := range s.friends {
		if u.ID == friendshipID {
			s.friends[i].Status = status
			return s.friends[i], nil
		}
	}
	return user.FriendshipResponse{}, helpers.ErrNotFound
}

func (s *StubFriendStore) GetAllFriendships(userID int, status string) ([]user.FriendshipResponse, error) {
	result := make([]user.FriendshipResponse, 0)

	log.Printf("status: %s", status)

	for _, u := range s.friends {
		if u.UserID == userID && (u.Status == status || u.Status == "all") {
			result = append(result, u)
		}
	}
	return result, nil
}

func (s *StubFriendStore) GetFriendship(requestID int) (user.FriendshipResponse, error) {

	for _, u := range s.friends {
		if u.ID == requestID {
			return u, nil
		}
	}

	return user.FriendshipResponse{}, helpers.ErrNotFound
}

type NotFoundFriendStore struct {
	friends []user.FriendshipResponse
}

func (s *NotFoundFriendStore) CreateFriendship(_, _ int) (user.FriendshipResponse, error) {

	return user.FriendshipResponse{}, helpers.ErrNotFound
}

func (s *NotFoundFriendStore) UpdateFriendship(_ int, _ string) (user.FriendshipResponse, error) {
	return user.FriendshipResponse{}, helpers.ErrNotFound
}

func (s *NotFoundFriendStore) GetAllFriendships(_ int, _ string) ([]user.FriendshipResponse, error) {
	return nil, nil
}

func (s *NotFoundFriendStore) GetFriendship(requestID int) (user.FriendshipResponse, error) {
	return user.FriendshipResponse{}, nil
}

type ConflictFriendStore struct {
	friends []user.FriendshipResponse
}

func (s *ConflictFriendStore) CreateFriendship(_, _ int) (user.FriendshipResponse, error) {

	return user.FriendshipResponse{}, helpers.ErrConflict
}

func (s *ConflictFriendStore) UpdateFriendship(_ int, _ string) (user.FriendshipResponse, error) {
	return user.FriendshipResponse{}, helpers.ErrConflict
}

func (s *ConflictFriendStore) GetAllFriendships(_ int, _ string) ([]user.FriendshipResponse, error) {
	return nil, nil
}

func (s *ConflictFriendStore) GetFriendship(requestID int) (user.FriendshipResponse, error) {
	return user.FriendshipResponse{}, nil
}

func TestSendRequest(t *testing.T) {
	authStore := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
		{ID: 2, FirstName: "Ade", LastName: "Oye", Password: "password", Email: "ade@gmail.com", Username: "Ade"},
	}}
	friendStore := StubFriendStore{friends: make([]user.FriendshipResponse, 0)}
	mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}

	server := &user.Handler{AuthStore: &authStore, FriendStore: &friendStore, Queue: &mockQueue}

	t.Run("send a request and return the entry", func(t *testing.T) {

		data := []byte(fmt.Sprintf(`{ "recipient_id": %d }`, 1))

		request := createSendRequest(2, data)
		response := httptest.NewRecorder()

		server.SendRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Friendship created successfully",
			"data": map[string]interface{}{
				"id":        float64(1),
				"user_id":   float64(2),
				"friend_id": float64(1),
				"status":    "pending",
			},
		}

		assertResponseCode(t, response.Code, http.StatusCreated)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for no user with the id", func(t *testing.T) {
		authStore := StubUserStore{users: []auth.User{
			{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
			{ID: 2, FirstName: "Ade", LastName: "Oye", Password: "password", Email: "ade@gmail.com", Username: "Ade"},
		}}
		friendStore := NotFoundFriendStore{friends: make([]user.FriendshipResponse, 0)}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}

		server := &user.Handler{AuthStore: &authStore, FriendStore: &friendStore, Queue: &mockQueue}
		data := []byte(fmt.Sprintf(`{ "recipient_id": %d }`, 3))

		request := createSendRequest(1, data)
		response := httptest.NewRecorder()

		server.SendRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource not found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("return 409 if friendship exists already", func(t *testing.T) {

		authStore := StubUserStore{users: []auth.User{
			{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
			{ID: 2, FirstName: "Ade", LastName: "Oye", Password: "password", Email: "ade@gmail.com", Username: "Ade"},
		}}
		friendStore := ConflictFriendStore{friends: make([]user.FriendshipResponse, 0)}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}

		server := &user.Handler{AuthStore: &authStore, FriendStore: &friendStore, Queue: &mockQueue}

		data := []byte(fmt.Sprintf(`{ "recipient_id": %d }`, 3))

		request := createSendRequest(1, data)
		response := httptest.NewRecorder()

		server.SendRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource already exists",
		}

		assertResponseCode(t, response.Code, http.StatusConflict)
		assertResponseBody(t, got, want)
	})

	t.Run("return bad request for empty user id", func(t *testing.T) {
		data := []byte(`{}`)

		request := createSendRequest(1, data)
		response := httptest.NewRecorder()

		server.SendRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "invalid request body",
			"problems": map[string][]string{
				"RecipientID": []string{"RecipientID required"},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})
}

func TestUpdateRequest(t *testing.T) {
	authStore := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
		{ID: 2, FirstName: "Ade", LastName: "Oye", Password: "password", Email: "ade@gmail.com", Username: "Ade"},
	}}
	friendStore := StubFriendStore{friends: []user.FriendshipResponse{
		{ID: 1, UserID: 1, FriendID: 2, Status: "pending"},
	}}
	mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}

	server := &user.Handler{AuthStore: &authStore, FriendStore: &friendStore, Queue: &mockQueue}

	t.Run("accept a request and return the entry", func(t *testing.T) {
		data := []byte(`{ "type": "accept" }`)

		request := createUpdateRequest(1, 1, data)
		response := httptest.NewRecorder()

		server.UpdateRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Friendship updated successfully",
			"data": map[string]interface{}{
				"id":        float64(1),
				"user_id":   float64(1),
				"friend_id": float64(2),
				"status":    "accepted",
			},
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("block a friendship and return the entry", func(t *testing.T) {
		authStore := StubUserStore{users: []auth.User{
			{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
			{ID: 2, FirstName: "Ade", LastName: "Oye", Password: "password", Email: "ade@gmail.com", Username: "Ade"},
		}}
		friendStore := StubFriendStore{friends: []user.FriendshipResponse{
			{ID: 1, UserID: 1, FriendID: 2, Status: "accepted"},
			{ID: 2, UserID: 2, FriendID: 1, Status: "accepted"},
		}}
		mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}

		server := &user.Handler{AuthStore: &authStore, FriendStore: &friendStore, Queue: &mockQueue}

		data := []byte(`{ "type": "block" }`)

		request := createUpdateRequest(1, 1, data)
		response := httptest.NewRecorder()

		server.UpdateRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Friendship updated successfully",
			"data": map[string]interface{}{
				"id":        float64(1),
				"user_id":   float64(1),
				"friend_id": float64(2),
				"status":    "blocked",
			},
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for no entry with the request id", func(t *testing.T) {
		data := []byte(`{ "type": "accept" }`)

		request := createUpdateRequest(1, 4, data)
		response := httptest.NewRecorder()

		server.UpdateRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource not found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for no entry with the request id", func(t *testing.T) {
		data := []byte(`{ "type": "accept" }`)

		request := createUpdateRequest(1, 4, data)
		response := httptest.NewRecorder()

		server.UpdateRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource not found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("return 400 for invalid request body", func(t *testing.T) {
		data := []byte(`{}`)

		request := createUpdateRequest(1, 4, data)
		response := httptest.NewRecorder()

		server.UpdateRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "invalid request body",
			"problems": map[string][]string{
				"Type": []string{"Type required"},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})

	t.Run("return 400 for empty request id", func(t *testing.T) {

		ctx := context.WithValue(context.Background(), "user_id", 1)
		request, _ := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/users/%d/friend_requests/", 1), bytes.NewReader([]byte(`{ "type": "accept" }`)))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("user_id", fmt.Sprint(1))

		request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
		response := httptest.NewRecorder()

		server.UpdateRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "request id is required",
		}

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})

	t.Run("return 400 for invalid type", func(t *testing.T) {
		data := []byte(`{ "type": "random" }`)

		request := createUpdateRequest(1, 1, data)
		response := httptest.NewRecorder()

		server.UpdateRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "invalid type",
		}

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})
}

func TestGetAllFriendships(t *testing.T) {
	authStore := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
		{ID: 2, FirstName: "Ade", LastName: "Oye", Password: "password", Email: "ade@gmail.com", Username: "Ade"},
		{ID: 3, FirstName: "Ayo", LastName: "Wale", Password: "password", Email: "ayo@gmail.com", Username: "Ayo"},
	}}

	friendStore := StubFriendStore{friends: []user.FriendshipResponse{
		{ID: 1, UserID: 1, FriendID: 2, Status: "accepted"},
		{ID: 2, UserID: 2, FriendID: 1, Status: "accepted"},
		{ID: 3, UserID: 1, FriendID: 3, Status: "pending"},
	}}
	mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}

	server := &user.Handler{AuthStore: &authStore, FriendStore: &friendStore, Queue: &mockQueue}

	t.Run("return all friendships", func(t *testing.T) {

		request := getAllRequests(1, 1, "")
		response := httptest.NewRecorder()

		server.GetAllRequestsHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Friendships retrieved successfully",
			"data": []map[string]interface{}{
				{"id": float64(1), "user_id": float64(1), "friend_id": float64(2), "status": "accepted"},
				{"id": float64(3), "user_id": float64(1), "friend_id": float64(3), "status": "pending"},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return all friendships (pending)", func(t *testing.T) {
		request := getAllRequests(1, 1, "pending")
		response := httptest.NewRecorder()

		server.GetAllRequestsHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Friendships retrieved successfully",
			"data": []map[string]interface{}{
				{"id": float64(3), "user_id": float64(1), "friend_id": float64(3), "status": "pending"},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return all friendships (accepted)", func(t *testing.T) {
		request := getAllRequests(1, 1, "accepted")
		response := httptest.NewRecorder()

		server.GetAllRequestsHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Friendships retrieved successfully",
			"data": []map[string]interface{}{
				{"id": float64(1), "user_id": float64(1), "friend_id": float64(2), "status": "accepted"},
			},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return all friendships (blocked)", func(t *testing.T) {
		request := getAllRequests(1, 1, "blocked")
		response := httptest.NewRecorder()

		server.GetAllRequestsHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Friendships retrieved successfully",
			"data":    []map[string]interface{}{},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return 400 for invalid status", func(t *testing.T) {
		request := getAllRequests(1, 1, "random")
		response := httptest.NewRecorder()

		server.GetAllRequestsHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "invalid status",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})

	t.Run("return 403 for accessing another user's requests", func(t *testing.T) {

		ctx := context.WithValue(context.Background(), "user_id", 1)
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/users/%d/friend_requests?status=%s", 2, "accepted"), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("user_id", fmt.Sprint(2))

		request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
		response := httptest.NewRecorder()

		server.GetAllRequestsHandler(response, request)

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
}

func TestGetFriendship(t *testing.T) {
	authStore := StubUserStore{users: []auth.User{
		{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale", Password: "password", Email: "adedunmola@gmail.com", Username: "Adedunmola"},
		{ID: 2, FirstName: "Ade", LastName: "Oye", Password: "password", Email: "ade@gmail.com", Username: "Ade"},
		{ID: 3, FirstName: "Ayo", LastName: "Wale", Password: "password", Email: "ayo@gmail.com", Username: "Ayo"},
	}}

	friendStore := StubFriendStore{friends: []user.FriendshipResponse{
		{ID: 1, UserID: 1, FriendID: 2, Status: "accepted"},
		{ID: 2, UserID: 2, FriendID: 1, Status: "accepted"},
		{ID: 3, UserID: 1, FriendID: 3, Status: "pending"},
	}}
	mockQueue := StubQueue{Tasks: make([]queue.TaskPayload, 0)}

	server := &user.Handler{AuthStore: &authStore, FriendStore: &friendStore, Queue: &mockQueue}

	t.Run("return a friendship", func(t *testing.T) {

		request := getARequest(1, 1)
		response := httptest.NewRecorder()

		server.GetRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"status":  "Success",
			"message": "Friendship retrieved successfully",
			"data":    map[string]interface{}{"id": float64(1), "user_id": float64(1), "friend_id": float64(2), "status": "accepted"},
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for friendship not found", func(t *testing.T) {

		request := getARequest(1, 10)
		response := httptest.NewRecorder()

		server.GetRequestHandler(response, request)

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

	t.Run("return 403 for accessing another user's friendship", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", 1)
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/users/%d/friend_requests/%d", 2, 1), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("user_id", fmt.Sprint(2))
		rctx.URLParams.Add("request_id", fmt.Sprint(1))

		request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
		response := httptest.NewRecorder()

		server.GetRequestHandler(response, request)

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

	t.Run("return 400 for no request id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", 1)
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/users/%d/friend_requests/%d", 2, 1), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("user_id", fmt.Sprint(2))

		request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
		response := httptest.NewRecorder()

		server.GetRequestHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		wantBody := map[string]interface{}{
			"message": "request id is required",
		}

		wantJSON, _ := json.Marshal(wantBody)

		var want map[string]interface{}
		_ = json.Unmarshal(wantJSON, &want)

		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})
}

func createSendRequest(userID int, data []byte) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/users/1/friend_requests", bytes.NewReader(data))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", fmt.Sprint(userID))

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}

func createUpdateRequest(userID, requestID int, data []byte) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/users/%d/friend_requests/%d", userID, requestID), bytes.NewReader(data))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", fmt.Sprint(userID))
	rctx.URLParams.Add("request_id", fmt.Sprint(requestID))

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}

func getAllRequests(userID, requestID int, status string) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/users/%d/friend_requests?status=%s", userID, status), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", fmt.Sprint(userID))
	rctx.URLParams.Add("request_id", fmt.Sprint(requestID))

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}

func getARequest(userID, requestID int) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/users/%d/friend_requests/%d", userID, requestID), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", fmt.Sprint(userID))
	rctx.URLParams.Add("request_id", fmt.Sprint(requestID))

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

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
