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
		return notification.Notification{}, errors.New("no friendship with the friendship id")
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

	notif, _ := s.GetNotification(ID)

	notif.Status = status

	return notif, nil
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

	t.Run("return error for no friendship with the friendship id", func(t *testing.T) {
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

		if err.Error() != "error creating notification: no friendship with the friendship id" {
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
		if err.Error() != "friendship id is required" {
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
		request := getNotificationRequest(1, 1, false)
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
		request := getNotificationRequest(10, 1, false)
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

	t.Run("return 403 for accessing another friendship's resource", func(t *testing.T) {

		request := getNotificationRequest(1, 2, false)
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
	currentTime := time.Now()

	user1 := auth.User{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale"}
	user2 := auth.User{ID: 2, FirstName: "Ade", LastName: "Oye"}

	notif1 := notification.Notification{ID: 1, UserID: user1.ID, Title: "", Body: "", Type: "", Status: "unread", Timestamp: &currentTime}
	notif2 := notification.Notification{ID: 2, UserID: user1.ID, Title: "", Body: "", Type: "", Status: "unread", Timestamp: &currentTime}
	notif3 := notification.Notification{ID: 3, UserID: user2.ID, Title: "", Body: "", Type: "", Status: "unread", Timestamp: &currentTime}

	store := &StubStore{
		users: []auth.User{
			user1,
		},
		notifications: []notification.Notification{
			notif1,
			notif2,
			notif3,
		},
	}

	server := &notification.Handler{Store: store}

	t.Run("get friendship's notifications", func(t *testing.T) {
		request := getNotificationRequest(1, 1, true)
		response := httptest.NewRecorder()

		server.GetUserNotificationsHandler(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Notification retrieved successfully",
			"data": []map[string]interface{}{
				{"id": float64(notif1.ID), "user_id": float64(user1.ID), "title": notif1.Title, "body": notif1.Body, "type": notif1.Type, "status": notif1.Status, "timestamp": &currentTime},
				{"id": float64(notif2.ID), "user_id": float64(user1.ID), "title": notif2.Title, "body": notif2.Body, "type": notif2.Type, "status": notif2.Status, "timestamp": &currentTime},
			},
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for no friendship with the id", func(t *testing.T) {
		request := getNotificationRequest(10, 1, true)
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
}

func TestUpdateNotification(t *testing.T) {
	currentTime := time.Now()

	user1 := auth.User{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale"}
	user2 := auth.User{ID: 2, FirstName: "Ade", LastName: "Oye"}

	notif1 := notification.Notification{ID: 1, UserID: user1.ID, Title: "", Body: "", Type: "", Status: "unread", Timestamp: &currentTime}
	notif2 := notification.Notification{ID: 2, UserID: user1.ID, Title: "", Body: "", Type: "", Status: "unread", Timestamp: &currentTime}
	notif3 := notification.Notification{ID: 3, UserID: user2.ID, Title: "", Body: "", Type: "", Status: "unread", Timestamp: &currentTime}

	store := &StubStore{
		users: []auth.User{
			user1,
		},
		notifications: []notification.Notification{
			notif1,
			notif2,
			notif3,
		},
	}

	server := &notification.Handler{Store: store}

	t.Run("update notification's status", func(t *testing.T) {
		request := updateNotificationRequest(notif1.ID, user1.ID)
		response := httptest.NewRecorder()

		server.UpdateNotification(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Notification retrieved successfully",
			"data": map[string]interface{}{
				"id":        float64(1),
				"user_id":   float64(user1.ID),
				"title":     notif1.Title,
				"body":      notif1.Body,
				"type":      notif1.Type,
				"status":    "read",
				"timestamp": &currentTime,
			},
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for no notification with the id", func(t *testing.T) {
		request := updateNotificationRequest(10, user1.ID)
		response := httptest.NewRecorder()

		server.UpdateNotification(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource not found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("return 400 for invalid status", func(t *testing.T) {
		request := updateNotificationRequest(notif1.ID, user1.ID)
		response := httptest.NewRecorder()

		server.UpdateNotification(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "invalid status",
		}
		assertResponseCode(t, response.Code, http.StatusBadRequest)
		assertResponseBody(t, got, want)
	})

	t.Run("return 403 for accessing another friendship's resource", func(t *testing.T) {
		request := updateNotificationRequest(notif1.ID, user2.ID)
		response := httptest.NewRecorder()

		server.UpdateNotification(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "forbidden from accessing another friendship's resource",
		}

		assertResponseCode(t, response.Code, http.StatusForbidden)
		assertResponseBody(t, got, want)
	})
}

func TestDeleteNotification(t *testing.T) {
	currentTime := time.Now()

	user1 := auth.User{ID: 1, FirstName: "Adedunmola", LastName: "Oyewale"}
	user2 := auth.User{ID: 2, FirstName: "Ade", LastName: "Oye"}

	notif1 := notification.Notification{ID: 1, UserID: user1.ID, Title: "", Body: "", Type: "", Status: "unread", Timestamp: &currentTime}
	notif2 := notification.Notification{ID: 2, UserID: user1.ID, Title: "", Body: "", Type: "", Status: "unread", Timestamp: &currentTime}
	notif3 := notification.Notification{ID: 3, UserID: user2.ID, Title: "", Body: "", Type: "", Status: "unread", Timestamp: &currentTime}

	store := &StubStore{
		users: []auth.User{
			user1,
		},
		notifications: []notification.Notification{
			notif1,
			notif2,
			notif3,
		},
	}

	server := &notification.Handler{Store: store}

	t.Run("delete notification", func(t *testing.T) {
		request := deleteNotificationRequest(notif1.ID, user1.ID)
		response := httptest.NewRecorder()

		server.DeleteNotification(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"status":  "Success",
			"message": "Notification deleted successfully",
		}

		assertResponseCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, got, want)
	})

	t.Run("return 404 for no notification with the id", func(t *testing.T) {
		request := deleteNotificationRequest(10, user1.ID)
		response := httptest.NewRecorder()

		server.DeleteNotification(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "resource not found",
		}

		assertResponseCode(t, response.Code, http.StatusNotFound)
		assertResponseBody(t, got, want)
	})

	t.Run("return 403 for accessing another friendship's resource", func(t *testing.T) {
		request := deleteNotificationRequest(notif1.ID, user2.ID)
		response := httptest.NewRecorder()

		server.DeleteNotification(response, request)

		var got map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &got)

		want := map[string]interface{}{
			"message": "forbidden from accessing another friendship's resource",
		}

		assertResponseCode(t, response.Code, http.StatusForbidden)
		assertResponseBody(t, got, want)
	})
}

func getNotificationRequest(notificationID, userID int, all bool) *http.Request {

	var request *http.Request
	var rctx *chi.Context
	ctx := context.WithValue(context.Background(), "user_id", userID)

	if !all {
		request, _ = http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/users/%d/notifications/%d", userID, notificationID), nil)

		rctx = chi.NewRouteContext()
		rctx.URLParams.Add("notification_id", fmt.Sprint(notificationID))
		rctx.URLParams.Add("user_id", fmt.Sprint(userID))
	} else {
		request, _ = http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/users/%d/notifications/", userID), nil)

		rctx = chi.NewRouteContext()
		rctx.URLParams.Add("user_id", fmt.Sprint(userID))
	}

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}

func updateNotificationRequest(notificationID, userID int) *http.Request {
	ctx := context.WithValue(context.Background(), "user_id", userID)
	ctx = context.WithValue(ctx, "notification_id", notificationID)

	request, _ := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/users/%d/notifications/%d", userID, notificationID), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notification_id", fmt.Sprint(notificationID))
	rctx.URLParams.Add("user_id", fmt.Sprint(userID))

	return request
}

func deleteNotificationRequest(notificationID, userID int) *http.Request {
	ctx := context.WithValue(context.Background(), "user_id", userID)

	request, _ := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/users/%d/notifications/%d", userID, notificationID), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notification_id", fmt.Sprint(notificationID))
	rctx.URLParams.Add("user_id", fmt.Sprint(userID))

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
