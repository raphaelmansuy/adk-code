// Package common provides shared utilities and error types for tools.
package common

import (
	"fmt"
)

// ErrorCode represents a structured error type
type ErrorCode string

const (
	ErrorCodeFileNotFound     ErrorCode = "FILE_NOT_FOUND"
	ErrorCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	ErrorCodePathTraversal    ErrorCode = "PATH_TRAVERSAL"
	ErrorCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrorCodeOperationFailed  ErrorCode = "OPERATION_FAILED"
	ErrorCodePatchFailed      ErrorCode = "PATCH_FAILED"
	ErrorCodeSymlinkEscape    ErrorCode = "SYMLINK_ESCAPE"
	ErrorCodeNotADirectory    ErrorCode = "NOT_A_DIRECTORY"
)

// ToolError represents a structured error with suggestions
type ToolError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Suggestion string                 `json:"suggestion,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// Error implements error interface
func (e *ToolError) Error() string {
	return e.Message
}

// NewToolError creates a new ToolError
func NewToolError(code ErrorCode, message string) *ToolError {
	return &ToolError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithSuggestion adds a suggestion to the error
func (e *ToolError) WithSuggestion(suggestion string) *ToolError {
	e.Suggestion = suggestion
	return e
}

// WithDetail adds a detail to the error
func (e *ToolError) WithDetail(key string, value interface{}) *ToolError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// FileNotFoundError creates a FILE_NOT_FOUND error
func FileNotFoundError(path string) *ToolError {
	return NewToolError(
		ErrorCodeFileNotFound,
		fmt.Sprintf("File not found: %s", path),
	).WithSuggestion(fmt.Sprintf("Check the path is correct. Current: %s", path))
}

// PermissionDeniedError creates a PERMISSION_DENIED error
func PermissionDeniedError(path string) *ToolError {
	return NewToolError(
		ErrorCodePermissionDenied,
		fmt.Sprintf("Permission denied: %s", path),
	).WithSuggestion("Check file permissions with 'ls -la'")
}

// PathTraversalError creates a PATH_TRAVERSAL error
func PathTraversalError(path string, basePath string) *ToolError {
	return NewToolError(
		ErrorCodePathTraversal,
		fmt.Sprintf("Path traversal detected: %s is outside %s", path, basePath),
	).WithSuggestion("Make sure the file path is within the allowed directory")
}

// SymlinkEscapeError creates a SYMLINK_ESCAPE error
func SymlinkEscapeError(path string, realPath string, basePath string) *ToolError {
	return NewToolError(
		ErrorCodeSymlinkEscape,
		fmt.Sprintf("Symlink points outside base directory: %s -> %s (base: %s)", path, realPath, basePath),
	).WithSuggestion("Ensure symlinks point to files within the allowed directory")
}

// InvalidInputError creates an INVALID_INPUT error
func InvalidInputError(message string) *ToolError {
	return NewToolError(
		ErrorCodeInvalidInput,
		fmt.Sprintf("Invalid input: %s", message),
	)
}

// OperationFailedError creates an OPERATION_FAILED error
func OperationFailedError(operation string, err error) *ToolError {
	return NewToolError(
		ErrorCodeOperationFailed,
		fmt.Sprintf("Operation failed: %s - %v", operation, err),
	)
}

// PatchFailedError creates a PATCH_FAILED error
func PatchFailedError(reason string) *ToolError {
	return NewToolError(
		ErrorCodePatchFailed,
		fmt.Sprintf("Patch failed: %s", reason),
	).WithSuggestion("Check that the patch matches the current file content")
}
