package factories

import (
	"fmt"
)

// ValidateRequiredFields is a helper function that validates a set of required fields
// It returns an error if any required field is missing
func ValidateRequiredFields(fields map[string]string) error {
	for name, value := range fields {
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}

// ValidateRequiredField validates a single required field
func ValidateRequiredField(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// ValidateAllRequiredFields validates multiple fields in the order specified
// Returns on the first missing field
func ValidateAllRequiredFields(checks ...fieldCheck) error {
	for _, check := range checks {
		if err := ValidateRequiredField(check.name, check.value); err != nil {
			return err
		}
	}
	return nil
}

// fieldCheck represents a single field validation check
type fieldCheck struct {
	name  string
	value string
}

// NewFieldCheck creates a field check tuple
func NewFieldCheck(name, value string) fieldCheck {
	return fieldCheck{name: name, value: value}
}
