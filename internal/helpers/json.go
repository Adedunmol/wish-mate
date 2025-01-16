package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrDecode = errors.New("error decoding json")
)

type Validator interface {
	Validate(data context.Context) map[string][]string
}

func DecodeAndValidate[V Validator](r *http.Request) (V, map[string][]string, error) {
	var body V
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return body, nil, fmt.Errorf("%s: %s", ErrDecode, err)
	}

	if validationErrors := body.Validate(r.Context()); len(validationErrors) != 0 {
		return body, nil, ErrValidate
	}

	return body, nil, nil
}

func WriteJSONResponse(responseWriter http.ResponseWriter, data interface{}, statusCode int) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)

	if err := json.NewEncoder(responseWriter).Encode(data); err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}
