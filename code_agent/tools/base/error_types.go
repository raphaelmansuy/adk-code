// Package common provides shared utilities and error types for tools.
package common

import (
	"code_agent/pkg/errors"
)

// ErrorCode is a facade for pkg/errors.ErrorCode
// It maintains backward compatibility with code that imports tools/common
type ErrorCode = errors.ErrorCode

// Backward-compatible error code constants (re-export from pkg/errors)
const (
	ErrorCodeFileNotFound     ErrorCode = errors.CodeFileNotFound
	ErrorCodePermissionDenied ErrorCode = errors.CodePermission
	ErrorCodePathTraversal    ErrorCode = errors.CodePathTraversal
	ErrorCodeInvalidInput     ErrorCode = errors.CodeInvalidInput
	ErrorCodeOperationFailed  ErrorCode = errors.CodeExecution // OPERATION_FAILED maps to EXECUTION_FAILED in pkg/errors
	ErrorCodePatchFailed      ErrorCode = errors.CodePatchFailed
	ErrorCodeSymlinkEscape    ErrorCode = errors.CodeSymlinkEscape
	ErrorCodeNotADirectory    ErrorCode = errors.CodeNotADirectory
)

// ToolError is a facade for pkg/errors.AgentError
// It maintains backward compatibility with code that imports tools/common
// The fields match AgentError, allowing seamless interoperability
type ToolError = errors.AgentError

// NewToolError creates a new ToolError (re-exports pkg/errors.New)
// Signature: NewToolError(code ErrorCode, message string) *ToolError
func NewToolError(code ErrorCode, message string) *ToolError {
	return errors.New(code, message)
}

// FileNotFoundError creates a FILE_NOT_FOUND error
func FileNotFoundError(path string) *ToolError {
	return errors.FileNotFoundError(path).
		WithSuggestion("Check the path is correct. Current: " + path)
}

// PermissionDeniedError creates a PERMISSION_DENIED error
func PermissionDeniedError(path string) *ToolError {
	return errors.PermissionDeniedError(path).
		WithSuggestion("Check file permissions with 'ls -la'")
}

// PathTraversalError creates a PATH_TRAVERSAL error
func PathTraversalError(path string, basePath string) *ToolError {
	return errors.PathTraversalError(path, basePath).
		WithSuggestion("Make sure the file path is within the allowed directory")
}

// SymlinkEscapeError creates a SYMLINK_ESCAPE error
func SymlinkEscapeError(path string, realPath string, basePath string) *ToolError {
	return errors.SymlinkEscapeError(path, realPath, basePath).
		WithSuggestion("Ensure symlinks point to files within the allowed directory")
}

// InvalidInputError creates an INVALID_INPUT error
func InvalidInputError(message string) *ToolError {
	return errors.InvalidInputError(message)
}

// OperationFailedError creates an OPERATION_FAILED error
func OperationFailedError(operation string, err error) *ToolError {
	return errors.OperationFailedError(operation, err)
}

// PatchFailedError creates a PATCH_FAILED error
func PatchFailedError(reason string) *ToolError {
	return errors.PatchFailedError(reason).
		WithSuggestion("Check that the patch matches the current file content")
}
