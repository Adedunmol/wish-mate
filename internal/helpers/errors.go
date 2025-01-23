package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrUnprocessable       = NewHTTPError(nil, http.StatusUnprocessableEntity, "unable to process request", nil)
	ErrDecode              = NewHTTPError(nil, http.StatusInternalServerError, "error decoding json body", nil)
	ErrBadRequest          = NewHTTPError(nil, http.StatusBadRequest, "invalid request body", nil)
	ErrUnauthorized        = NewHTTPError(nil, http.StatusUnauthorized, "invalid credentials", nil)
	ErrConflict            = NewHTTPError(nil, http.StatusConflict, "resource already exists", nil)
	ErrValidate            = NewHTTPError(nil, http.StatusBadRequest, "error validating request body", nil)
	ErrNotFound            = NewHTTPError(nil, http.StatusNotFound, "resource not found", nil)
	ErrInternalServerError = NewHTTPError(nil, http.StatusInternalServerError, "internal server error", nil)
)

type ClientError interface {
	Error() string
	ResponseBody() ([]byte, error)
	ResponseHeaders() (int, map[string]string)
}

type HTTPError struct {
	Cause    error               `json:"-"`
	Message  string              `json:"message"`
	Status   int                 `json:"-"`
	Problems map[string][]string `json:"problems,omitempty"`
}

func (e *HTTPError) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return e.Message + ": " + e.Cause.Error()
}

func (e *HTTPError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("error while parsing response body: %v", err)
	}
	return body, nil
}

func (e *HTTPError) ResponseHeaders() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
}

func NewHTTPError(err error, status int, message string, problems map[string][]string) error {
	return &HTTPError{
		Cause:    err,
		Message:  message,
		Status:   status,
		Problems: problems,
	}
}

func HandleError(responseWriter http.ResponseWriter, err error) {
	var clientError ClientError
	ok := errors.As(err, &clientError)
	if !ok {
		body := struct {
			Message string `json:"message"`
		}{
			Message: "An internal server error has occurred.",
		}
		WriteJSONResponse(responseWriter, body, http.StatusInternalServerError)
	}
	status, headers := clientError.ResponseHeaders()

	for k, v := range headers {
		responseWriter.Header().Add(k, v)
	}

	//body, err := clientError.ResponseBody()
	//if err != nil {
	//	log.Printf("an error occurred: %s", err)
	//	WriteJSONResponse(responseWriter, nil, http.StatusInternalServerError)
	//}

	WriteJSONResponse(responseWriter, clientError, status)
}
