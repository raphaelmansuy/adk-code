package tools

import (
	"fmt"
	"path/filepath"
	"strings"

	"code_agent/workspace"
)

// WorkspaceTools provides workspace-aware file operation tools
type WorkspaceTools struct {
	resolver *workspace.Resolver
}

// NewWorkspaceTools creates a new workspace tools instance
func NewWorkspaceTools(resolver *workspace.Resolver) *WorkspaceTools {
	return &WorkspaceTools{
		resolver: resolver,
	}
}

// ResolvePath resolves a path that may contain workspace hints
// Supports formats like:
// - @workspace:path/to/file
// - path/to/file (uses primary workspace)
// - /absolute/path/to/file
func (wt *WorkspaceTools) ResolvePath(path string) (string, error) {
	if wt.resolver == nil {
		// No workspace resolver, return as-is
		return path, nil
	}

	// Check if path contains workspace hint
	if strings.HasPrefix(path, "@") {
		resolved, err := wt.resolver.ResolvePath(path, nil)
		if err != nil {
			return "", fmt.Errorf("failed to resolve workspace path: %w", err)
		}
		return resolved.AbsolutePath, nil
	}

	// Absolute path - return as-is
	if filepath.IsAbs(path) {
		return path, nil
	}

	// Relative path - use resolver with disambiguation
	resolved, err := wt.resolver.ResolvePathWithDisambiguation(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve relative path: %w", err)
	}

	return resolved.AbsolutePath, nil
}

// FormatPathWithHint formats a path with its workspace hint
func (wt *WorkspaceTools) FormatPathWithHint(absolutePath string) string {
	if wt.resolver == nil {
		return absolutePath
	}

	workspaceName := wt.resolver.GetWorkspaceForPath(absolutePath)
	if workspaceName == "" {
		return absolutePath
	}

	return workspace.FormatPathWithHint(workspaceName, absolutePath)
}

// WorkspaceReadFileInput extends ReadFileInput with workspace support
type WorkspaceReadFileInput struct {
	ReadFileInput
}

// WorkspaceWriteFileInput extends WriteFileInput with workspace support
type WorkspaceWriteFileInput struct {
	WriteFileInput
}

// WorkspaceListDirectoryInput extends ListDirectoryInput with workspace support
type WorkspaceListDirectoryInput struct {
	ListDirectoryInput
}

// ResolveReadFilePath resolves the path for read_file tool
func (wt *WorkspaceTools) ResolveReadFilePath(input ReadFileInput) (ReadFileInput, error) {
	resolvedPath, err := wt.ResolvePath(input.Path)
	if err != nil {
		return input, err
	}

	input.Path = resolvedPath
	return input, nil
}

// ResolveWriteFilePath resolves the path for write_file tool
func (wt *WorkspaceTools) ResolveWriteFilePath(input WriteFileInput) (WriteFileInput, error) {
	resolvedPath, err := wt.ResolvePath(input.Path)
	if err != nil {
		return input, err
	}

	input.Path = resolvedPath
	return input, nil
}

// ResolveListDirectoryPath resolves the path for list_directory tool
func (wt *WorkspaceTools) ResolveListDirectoryPath(input ListDirectoryInput) (ListDirectoryInput, error) {
	resolvedPath, err := wt.ResolvePath(input.Path)
	if err != nil {
		return input, err
	}

	input.Path = resolvedPath
	return input, nil
}

// ParseWorkspaceHint parses a workspace hint from a path
// Returns workspace name and relative path
func ParseWorkspaceHint(path string) (workspaceName string, relativePath string, hasHint bool) {
	if !strings.HasPrefix(path, "@") {
		return "", path, false
	}

	parts := strings.SplitN(path[1:], ":", 2)
	if len(parts) != 2 {
		return "", path, false
	}

	return parts[0], parts[1], true
}

// FormatWorkspaceHint formats a workspace hint
func FormatWorkspaceHint(workspaceName string, relativePath string) string {
	return fmt.Sprintf("@%s:%s", workspaceName, relativePath)
}
