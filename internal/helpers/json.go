package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrDecode   = errors.New("error decoding json")
	ErrValidate = errors.New("error validating json body")
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
