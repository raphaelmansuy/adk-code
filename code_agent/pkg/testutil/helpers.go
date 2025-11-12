// Package testutil provides common testing utilities and helper functions.
package testutil

// IntPtr returns a pointer to the given int value.
// Useful for tests that need to create pointers to primitive values.
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the given bool value.
// Useful for tests that need to create pointers to primitive values.
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr returns a pointer to the given string value.
// Useful for tests that need to create pointers to primitive values.
func StringPtr(s string) *string {
	return &s
}
