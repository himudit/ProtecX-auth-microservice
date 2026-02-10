package utils

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ValidationErrors(err error) map[string]string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}

	out := make(map[string]string)
	for _, fe := range ve {
		field := strings.ToLower(fe.Field())

		switch fe.Tag() {
		case "required":
			out[field] = "is required"
		case "email":
			out[field] = "must be a valid email"
		case "min":
			out[field] = "must be at least " + fe.Param() + " characters"
		default:
			out[field] = "is invalid"
		}
	}

	return out
}
