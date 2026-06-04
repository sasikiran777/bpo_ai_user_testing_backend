package helpers

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationErrors(err error) map[string]string {

	errors := make(map[string]string)

	validationErrors, ok := err.(validator.ValidationErrors)

	if !ok {

		errors["error"] = err.Error()

		return errors
	}

	for _, fieldError := range validationErrors {

		field := strings.ToLower(fieldError.Field())

		switch fieldError.Tag() {

		case "required":
			errors[field] = field + " is required"

		case "email":
			errors[field] = "invalid email address"

		case "min":
			errors[field] = field + " is too short"

		default:
			errors[field] = "invalid value"
		}
	}

	return errors
}
