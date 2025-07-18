package utils


import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Allowed statuses
var allowedStatuses = []string{"active", "on_break", "busy"}

// ValidateWorkerStatus validates a status string
func ValidateWorkerStatus(status string) error {
	return validation.Validate(status,
		validation.Required,
		validation.In(allowedStatuses),
	)
}



