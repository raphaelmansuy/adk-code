// Package tools provides file operation tools for the coding agent.
package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathSecurityError represents a path validation error
type PathSecurityError struct {
	Code    string // e.g., "DIRECTORY_TRAVERSAL", "OUTSIDE_BASE", "FILE_NOT_FOUND"
	Path    string
	Message string
}

// Error implements error interface
func (e *PathSecurityError) Error() string {
	return e.Message
}

// ValidateFilePath validates a file path for security and existence
// It checks for:
// - Invalid path syntax
// - Directory traversal attacks (../../etc/passwd)
// - Symlink escapes
// - Base path boundary enforcement (if basePath is provided)
// - File existence (if requireExist is true)
func ValidateFilePath(basePath, requestedPath string, requireExist bool) error {
	// 1. Resolve absolute paths
	absRequested, err := filepath.Abs(requestedPath)
	if err != nil {
		return &PathSecurityError{
			Code:    "INVALID_PATH",
			Path:    requestedPath,
			Message: fmt.Sprintf("Invalid path: %v", err),
		}
	}

	// 2. If basePath is specified, validate it
	if basePath != "" {
		absBase, err := filepath.Abs(basePath)
		if err != nil {
			return &PathSecurityError{
				Code:    "INVALID_BASE",
				Path:    basePath,
				Message: fmt.Sprintf("Invalid base path: %v", err),
			}
		}

		// 3. Check for directory traversal (basic check)
		if !strings.HasPrefix(absRequested, absBase) {
			return &PathSecurityError{
				Code:    "DIRECTORY_TRAVERSAL",
				Path:    requestedPath,
				Message: fmt.Sprintf("Path traversal detected: %s is outside %s", absRequested, absBase),
			}
		}

		// 4. Resolve symlinks and verify again
		realPath, err := filepath.EvalSymlinks(absRequested)
		if err == nil {
			// Only check if symlink resolution succeeded
			realBase, _ := filepath.EvalSymlinks(absBase)
			if realBase == "" {
				realBase = absBase
			}
			if !strings.HasPrefix(realPath, realBase) {
				return &PathSecurityError{
					Code:    "SYMLINK_ESCAPE",
					Path:    requestedPath,
					Message: fmt.Sprintf("Symlink points outside base directory: %s -> %s (base: %s)", absRequested, realPath, realBase),
				}
			}
		}
	}

	// 5. Check if file exists (if required)
	if requireExist {
		if _, err := os.Stat(absRequested); os.IsNotExist(err) {
			return &PathSecurityError{
				Code:    "FILE_NOT_FOUND",
				Path:    requestedPath,
				Message: fmt.Sprintf("File not found: %s", absRequested),
			}
		}
	}

	return nil
}

// ValidateDirPath validates a directory path for security
func ValidateDirPath(basePath, requestedPath string, requireExist bool) error {
	// 1. Resolve absolute paths
	absRequested, err := filepath.Abs(requestedPath)
	if err != nil {
		return &PathSecurityError{
			Code:    "INVALID_PATH",
			Path:    requestedPath,
			Message: fmt.Sprintf("Invalid path: %v", err),
		}
	}

	// 2. If basePath is specified, validate it
	if basePath != "" {
		absBase, err := filepath.Abs(basePath)
		if err != nil {
			return &PathSecurityError{
				Code:    "INVALID_BASE",
				Path:    basePath,
				Message: fmt.Sprintf("Invalid base path: %v", err),
			}
		}

		// 3. Check for directory traversal
		if !strings.HasPrefix(absRequested, absBase) {
			return &PathSecurityError{
				Code:    "DIRECTORY_TRAVERSAL",
				Path:    requestedPath,
				Message: fmt.Sprintf("Path traversal detected: %s is outside %s", absRequested, absBase),
			}
		}

		// 4. Resolve symlinks and verify again
		realPath, err := filepath.EvalSymlinks(absRequested)
		if err == nil {
			realBase, _ := filepath.EvalSymlinks(absBase)
			if realBase == "" {
				realBase = absBase
			}
			if !strings.HasPrefix(realPath, realBase) {
				return &PathSecurityError{
					Code:    "SYMLINK_ESCAPE",
					Path:    requestedPath,
					Message: fmt.Sprintf("Symlink points outside base directory: %s -> %s (base: %s)", absRequested, realPath, realBase),
				}
			}
		}
	}

	// 5. Check if directory exists (if required)
	if requireExist {
		info, err := os.Stat(absRequested)
		if err != nil {
			if os.IsNotExist(err) {
				return &PathSecurityError{
					Code:    "PATH_NOT_FOUND",
					Path:    requestedPath,
					Message: fmt.Sprintf("Path not found: %s", absRequested),
				}
			}
			return &PathSecurityError{
				Code:    "PATH_ERROR",
				Path:    requestedPath,
				Message: fmt.Sprintf("Error accessing path: %v", err),
			}
		}
		if !info.IsDir() {
			return &PathSecurityError{
				Code:    "NOT_A_DIRECTORY",
				Path:    requestedPath,
				Message: fmt.Sprintf("Path is not a directory: %s", absRequested),
			}
		}
	}

	return nil
}
