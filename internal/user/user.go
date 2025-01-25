package user

import "net/http"

type Handler struct{}

func (h *Handler) SendRequestHandler(responseWriter http.ResponseWriter, request *http.Request) {}

func (h *Handler) AcceptRequestHandler(responseWriter http.ResponseWriter, request *http.Request) {}
