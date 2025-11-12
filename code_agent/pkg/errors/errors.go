// Package errors provides unified error handling for the code agent.
// It defines standard error types and codes used across all components.
package errors

import (
	"fmt"
)

// ErrorCode represents standard error categories
type ErrorCode string

const (
	// File operation errors
	CodeFileNotFound  ErrorCode = "FILE_NOT_FOUND"
	CodePermission    ErrorCode = "PERMISSION_DENIED"
	CodePathTraversal ErrorCode = "PATH_TRAVERSAL"
	CodeSymlinkEscape ErrorCode = "SYMLINK_ESCAPE"
	CodeNotADirectory ErrorCode = "NOT_A_DIRECTORY"

	// Execution errors
	CodeExecution ErrorCode = "EXECUTION_FAILED"
	CodeTimeout   ErrorCode = "TIMEOUT"

	// Input/validation errors
	CodeInvalidInput ErrorCode = "INVALID_INPUT"
	CodeValidation   ErrorCode = "VALIDATION_ERROR"

	// Model/Provider errors
	CodeModelNotFound ErrorCode = "MODEL_NOT_FOUND"
	CodeProviderError ErrorCode = "PROVIDER_ERROR"
	CodeAPIKey        ErrorCode = "API_KEY_ERROR"

	// Patch/edit errors
	CodePatchFailed ErrorCode = "PATCH_FAILED"

	// General errors
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
	CodeNotSupported ErrorCode = "NOT_SUPPORTED"
)

// AgentError is the standard error type for the code agent
type AgentError struct {
	Code    ErrorCode
	Message string
	Wrapped error
	Context map[string]string
}

// Error implements the error interface
func (e *AgentError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Wrapped)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error for error chain inspection
func (e *AgentError) Unwrap() error {
	return e.Wrapped
}

// WithContext adds context information to an error
func (e *AgentError) WithContext(key string, value string) *AgentError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// New creates a new AgentError with the given code and message
func New(code ErrorCode, message string) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
		Context: make(map[string]string),
	}
}

// Wrap creates a new AgentError wrapping an existing error
func Wrap(code ErrorCode, message string, err error) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
		Wrapped: err,
		Context: make(map[string]string),
	}
}

// Is checks if an error matches the given error code
func Is(err error, code ErrorCode) bool {
	if agentErr, ok := err.(*AgentError); ok {
		return agentErr.Code == code
	}
	return false
}

// Helper functions for common error patterns

// FileNotFoundError creates a FILE_NOT_FOUND error
func FileNotFoundError(path string) *AgentError {
	return New(CodeFileNotFound, fmt.Sprintf("file not found: %s", path)).
		WithContext("path", path)
}

// PermissionDeniedError creates a PERMISSION_DENIED error
func PermissionDeniedError(path string) *AgentError {
	return New(CodePermission, fmt.Sprintf("permission denied: %s", path)).
		WithContext("path", path)
}

// PathTraversalError creates a PATH_TRAVERSAL error
func PathTraversalError(path, basePath string) *AgentError {
	return New(CodePathTraversal, fmt.Sprintf("path traversal detected: %s is outside %s", path, basePath)).
		WithContext("path", path).
		WithContext("base_path", basePath)
}

// SymlinkEscapeError creates a SYMLINK_ESCAPE error
func SymlinkEscapeError(path, realPath, basePath string) *AgentError {
	return New(CodeSymlinkEscape, fmt.Sprintf("symlink points outside base directory: %s -> %s (base: %s)", path, realPath, basePath)).
		WithContext("path", path).
		WithContext("real_path", realPath).
		WithContext("base_path", basePath)
}

// InvalidInputError creates an INVALID_INPUT error
func InvalidInputError(message string) *AgentError {
	return New(CodeInvalidInput, fmt.Sprintf("invalid input: %s", message))
}

// ExecutionError creates an EXECUTION_FAILED error
func ExecutionError(command string, err error) *AgentError {
	return Wrap(CodeExecution, fmt.Sprintf("execution failed: %s", command), err).
		WithContext("command", command)
}

// TimeoutError creates a TIMEOUT error
func TimeoutError(operation string) *AgentError {
	return New(CodeTimeout, fmt.Sprintf("operation timed out: %s", operation)).
		WithContext("operation", operation)
}

// APIKeyError creates an API_KEY_ERROR
func APIKeyError(provider string) *AgentError {
	return New(CodeAPIKey, fmt.Sprintf("%s API key not configured or invalid", provider)).
		WithContext("provider", provider)
}

// ModelNotFoundError creates a MODEL_NOT_FOUND error
func ModelNotFoundError(modelID string) *AgentError {
	return New(CodeModelNotFound, fmt.Sprintf("model not found: %s", modelID)).
		WithContext("model_id", modelID)
}

// ProviderError creates a PROVIDER_ERROR
func ProviderError(provider string, err error) *AgentError {
	return Wrap(CodeProviderError, fmt.Sprintf("%s provider error", provider), err).
		WithContext("provider", provider)
}

// PatchFailedError creates a PATCH_FAILED error
func PatchFailedError(reason string) *AgentError {
	return New(CodePatchFailed, fmt.Sprintf("patch failed: %s", reason))
}

// InternalError creates an INTERNAL_ERROR
func InternalError(message string) *AgentError {
	return New(CodeInternal, fmt.Sprintf("internal error: %s", message))
}

// NotSupportedError creates a NOT_SUPPORTED error
func NotSupportedError(feature string) *AgentError {
	return New(CodeNotSupported, fmt.Sprintf("not supported: %s", feature)).
		WithContext("feature", feature)
}
