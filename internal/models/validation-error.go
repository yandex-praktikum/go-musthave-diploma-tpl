package models

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"strings"
)

type ValidationError map[string]interface{}

func ExtractErrors(err error) ValidationError {
	res := ValidationError{}

	var invalidValidationError *validator.InvalidValidationError
	if errors.As(err, &invalidValidationError) {
		return res
	}

	for _, err := range err.(validator.ValidationErrors) {
		field := strings.ToLower(err.Field())
		actualTag := err.ActualTag()
		res[field] = map[string]bool{
			actualTag: true,
		}
	}

	return res
}
