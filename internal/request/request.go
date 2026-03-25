package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// DecodeJSON parses the request body as JSON into the provided struct pointer and runs validation.
func DecodeJSON(r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	if err := validate.Struct(dst); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			var errMsgs []string
			for _, vErr := range validationErrors {
				switch vErr.Tag() {
				case "required":
					errMsgs = append(errMsgs, fmt.Sprintf("%s is required", strings.ToLower(vErr.Field())))
				case "min":
					errMsgs = append(errMsgs, fmt.Sprintf("%s must be at least %s characters", strings.ToLower(vErr.Field()), vErr.Param()))
				case "max":
					errMsgs = append(errMsgs, fmt.Sprintf("%s must be at most %s characters", strings.ToLower(vErr.Field()), vErr.Param()))
				case "oneof":
					errMsgs = append(errMsgs, fmt.Sprintf("%s must be one of: %s", strings.ToLower(vErr.Field()), vErr.Param()))
				default:
					errMsgs = append(errMsgs, fmt.Sprintf("%s failed validation: %s", strings.ToLower(vErr.Field()), vErr.Tag()))
				}
			}
			return fmt.Errorf("validation failed: %s", strings.Join(errMsgs, ", "))
		}
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
