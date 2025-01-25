package user

import (
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/queue"
	"net/http"
)

type Handler struct {
	AuthStore   auth.Store
	FriendStore FriendStore
	Queue       queue.Queue
}

func (h *Handler) SendRequestHandler(responseWriter http.ResponseWriter, request *http.Request) {}

func (h *Handler) AcceptRequestHandler(responseWriter http.ResponseWriter, request *http.Request) {}
