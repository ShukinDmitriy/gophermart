package models

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError map[string]interface{}

func ExtractErrors(err error) ValidationError {
	res := ValidationError{}

	var invalidValidationError *validator.InvalidValidationError
	if errors.As(err, &invalidValidationError) {
		return res
	}

	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	if !ok {
		return res
	}

	for _, err := range validationErrors {
		field := strings.ToLower(err.Field())
		actualTag := err.ActualTag()
		res[field] = map[string]bool{
			actualTag: true,
		}
	}

	return res
}
