package form

import (
	"github.com/go-playground/validator/v10"
	"strings"
	"unicode"
)

func formatField(s string) string {
	var builder strings.Builder
	for _, c := range s {
		if unicode.IsUpper(c) {
			builder.WriteRune(' ')
			builder.WriteRune(unicode.ToLower(c))
		} else {
			builder.WriteRune(c)
		}
	}

	return strings.Trim(builder.String(), " ")
}

// Converts camel case to snake case
func camel2snake(s string) string {
	var builder strings.Builder
	for i, c := range s {
		if unicode.IsUpper(c) {
			if i > 0 && s[i-1] != '_' {
				builder.WriteRune('_')
			}
			builder.WriteRune(unicode.ToLower(c))
		} else {
			builder.WriteRune(c)
		}
	}

	return builder.String()
}

func formatError(err validator.FieldError) string {
	field := formatField(err.Field())

	switch err.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return "Invalid email address"
	case "min":
		return field + " is too short"
	case "max":
		return field + " is too long"
	case "eqfield":
		param := formatField(err.Param())
		return field + " must match " + param
	default:
		return field + " is invalid"
	}
}
