package notification_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/notification"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type StubStore struct {
	notifications []notification.Notification
	users         []auth.User
}

func (s *StubStore) CreateNotification(body *notification.CreateNotificationBody) (notification.Notification, error) {
	var userData auth.User

	for _, u := range s.users {
		if u.ID == body.UserID {
			userData = u
		}
	}

	if userData.ID == 0 {
		return notification.Notification{}, errors.New("no user with the user id")
	}

	currentTime := time.Now()
	data := notification.Notification{
		ID:        1,
		UserID:    userData.ID,
		Title:     body.Title,
		Body:      body.Body,
		Type:      body.Type,
		Status:    "unread",
		Timestamp: &currentTime,
	}

	s.notifications = append(s.notifications, data)

	return data, nil
}

func (s *StubStore) GetNotification(id int) (notification.Notification, error) {

	for _, n := range s.notifications {
		if n.ID == id {
			return n, nil
		}
	}

	return notification.Notification{}, helpers.ErrNotFound
}
func (s *StubStore) UpdateNotification(ID int, status string) (notification.Notification, error) {
	return notification.Notification{}, nil
}

func (s *StubStore) DeleteNotification(id int) error {
	return nil
}

func (s *StubStore) GetUserNotifications(userID int) ([]notification.Notification, error) {
	return nil, nil
}

func TestCreateNotification(t *testing.T) {
	store := &StubStore{
		users: []auth.User{
			{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale"},
		},
	}

	server := &notification.Handler{Store: store}

	t.Run("create and return notification", func(t *testing.T) {
		body := notification.CreateNotificationBody{
			UserID: 1,
			Title:  "Birthday",
			Body:   "Wish someone",
			Type:   "alert",
		}

		notif, _ := server.CreateNotification(&body)
		currentTime := time.Now()

		want := notification.Notification{
			ID:        1,
			UserID:    body.UserID,
			Title:     body.Title,
			Body:      body.Body,
			Type:      body.Type,
			Status:    "unread",
			Timestamp: &currentTime,
		}

		if len(store.notifications) != 1 {
			t.Errorf("CreateNotification returned wrong number of notifications")
		}

		if !reflect.DeepEqual(notif, want) {
			t.Errorf("CreateNotification returned wrong notification")
		}
	})

	t.Run("return error for no user with the user id", func(t *testing.T) {
		body := notification.CreateNotificationBody{
			UserID: 10,
			Title:  "Birthday",
			Body:   "Wish someone",
			Type:   "alert",
		}

		_, err := server.CreateNotification(&body)

		if err == nil {
			t.Errorf("CreateNotification returned no error")
		}

		if err.Error() != "error creating notification: no user with the user id" {
			t.Errorf("wrong error returned")
		}
	})

	t.Run("return error for invalid body", func(t *testing.T) {
		body := notification.CreateNotificationBody{
			UserID: 1,
			Title:  "Birthday",
			Type:   "alert",
		}

		_, err := server.CreateNotification(&body)

		if err == nil {
			t.Errorf("CreateNotification returned no error")
		}
		if err.Error() != "body is required" {
			t.Errorf("wrong error returned")
		}

		body = notification.CreateNotificationBody{
			Title: "Birthday",
			Body:  "Wish someone",
			Type:  "alert",
		}

		_, err = server.CreateNotification(&body)

		if err == nil {
			t.Errorf("CreateNotification returned no error")
		}
		if err.Error() != "user id is required" {
			t.Errorf("wrong error returned")
		}
	})
}

func TestGetNotification(t *testing.T) {
	currentTime := time.Now()

	user := auth.User{
		ID: 1, FirstName: "Adedunmola", LastName: "Oyewale",
	}
	notif := notification.Notification{ID: 1, UserID: user.ID, Title: "", Body: "", Type: "", Status: "unread", Timestamp: &currentTime}
	store := &StubStore{
		users: []auth.User{
			user,
		},
		notifications: []notification.Notification{
			notif,
		},
	}

	server := &notification.Handler{Store: store}

	t.Run("get notification", func(t *testing.T) {
		request := getNotificationRequest(1, 1)
		response := httptest.NewRecorder()

		server.GetNotificationHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Notification retrieved successfully",
			"data": map[string]interface{}{
				"id":        float64(1),
				"user_id":   float64(user.ID),
				"title":     notif.Title,
				"body":      notif.Body,
				"type":      notif.Type,
				"status":    notif.Status,
				"timestamp": &currentTime,
			},
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for no notification with the id", func(t *testing.T) {
		request := getNotificationRequest(10, 1)
		response := httptest.NewRecorder()

		server.GetNotificationHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "no resource found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("return 400 for no notification id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", user.ID)
		request, _ := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/users/%d/notifications/", user.ID), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("user_id", fmt.Sprint(user.ID))

		request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

		response := httptest.NewRecorder()

		server.GetNotificationHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "no resource found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("return 403 for accessing another user's resource", func(t *testing.T) {

		request := getNotificationRequest(1, 2)
		response := httptest.NewRecorder()

		server.GetNotificationHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "forbidden from accessing the resource",
		}

		assertResponseCode(t, response.Code, http.StatusForbidden)
		assertResponseBody(t, got, want)
	})
}

func TestGetUserNotifications(t *testing.T) {

	t.Run("get user's notifications", func(t *testing.T) {})

	t.Run("return 400 for no user with the id", func(t *testing.T) {})

	t.Run("return 400 for no user id", func(t *testing.T) {})
}

func TestUpdateNotification(t *testing.T) {

	t.Run("update notification's status", func(t *testing.T) {})

	t.Run("return 404 for no notification with the id", func(t *testing.T) {})

	t.Run("return 400 for invalid status", func(t *testing.T) {})

	t.Run("return 403 for accessing another user's resource", func(t *testing.T) {})
}

func TestDeleteNotification(t *testing.T) {

	t.Run("delete notification's status", func(t *testing.T) {})

	t.Run("return 404 for no notification with the id", func(t *testing.T) {})

	t.Run("return 403 for accessing another user's resource", func(t *testing.T) {})
}

func getNotificationRequest(notificationID, userID int) *http.Request {

	ctx := context.WithValue(context.Background(), "user_id", userID)
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/users/%d/notifications/%d", userID, notificationID), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notification_id", fmt.Sprint(notificationID))
	rctx.URLParams.Add("user_id", fmt.Sprint(userID))

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
