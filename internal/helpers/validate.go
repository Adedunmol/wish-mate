package helpers

import (
	"github.com/go-playground/validator/v10"
)

type Validation struct{}

func (vs *Validation) Validate(data interface{}) map[string][]string {
	fieldErrors := make(map[string][]string)

	v := validator.New(validator.WithRequiredStructEnabled())

	if err := v.Struct(data); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fieldName := err.Field()
			fieldError := fieldName + " " + err.Tag()

			fieldErrors[fieldName] = append(fieldErrors[fieldName], fieldError)
		}

	}

	return fieldErrors
}
