package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolver provides intelligent path resolution across multiple workspaces
type Resolver struct {
	manager *Manager
}

// NewResolver creates a new path resolver
func NewResolver(manager *Manager) *Resolver {
	return &Resolver{
		manager: manager,
	}
}

// ResolvePath resolves a path (with optional workspace hint) to an absolute path
// Workspace hints use the syntax: @workspaceName:relative/path
func (r *Resolver) ResolvePath(path string, workspaceHint *string) (*ResolvedPath, error) {
	// If workspace hint provided, use it
	if workspaceHint != nil {
		return r.resolveWithHint(path, *workspaceHint)
	}

	// If absolute path, find containing workspace
	if filepath.IsAbs(path) {
		root := r.manager.ResolvePathToRoot(path)
		if root == nil {
			// Path is outside all workspaces, use primary
			root = r.manager.GetPrimaryRoot()
			if root == nil {
				return nil, fmt.Errorf("no workspace roots available")
			}
		}

		relPath, err := r.manager.GetRelativePathFromRoot(path, root)
		if err != nil {
			relPath = path // Fallback to absolute path
		}

		return &ResolvedPath{
			AbsolutePath: path,
			Root:         root,
			RelativePath: relPath,
		}, nil
	}

	// Relative path - resolve against primary workspace
	primary := r.manager.GetPrimaryRoot()
	if primary == nil {
		return nil, fmt.Errorf("no workspace roots available")
	}

	absPath := filepath.Join(primary.Path, path)
	return &ResolvedPath{
		AbsolutePath: absPath,
		Root:         primary,
		RelativePath: path,
	}, nil
}

// resolveWithHint resolves a path using an explicit workspace hint
func (r *Resolver) resolveWithHint(path, hint string) (*ResolvedPath, error) {
	// Find workspace by name
	root := r.manager.GetRootByName(hint)
	if root == nil {
		return nil, fmt.Errorf("workspace '%s' not found", hint)
	}

	// If path is absolute, use as-is
	if filepath.IsAbs(path) {
		relPath, err := r.manager.GetRelativePathFromRoot(path, root)
		if err != nil {
			relPath = path
		}

		return &ResolvedPath{
			AbsolutePath: path,
			Root:         root,
			RelativePath: relPath,
		}, nil
	}

	// Relative path - join with workspace root
	absPath := filepath.Join(root.Path, path)
	return &ResolvedPath{
		AbsolutePath: absPath,
		Root:         root,
		RelativePath: path,
	}, nil
}

// ParseWorkspaceHint parses the workspace hint syntax: @workspaceName:path
// Returns the workspace name (without @) and the path portion
func ParseWorkspaceHint(input string) (workspaceHint *string, path string) {
	// Check if input starts with @
	if !strings.HasPrefix(input, "@") {
		return nil, input
	}

	// Find the colon separator
	colonIndex := strings.Index(input[1:], ":")
	if colonIndex == -1 {
		// No colon found, treat as regular path
		return nil, input
	}

	// Extract workspace hint and path
	hint := input[1 : colonIndex+1]
	pathPart := input[colonIndex+2:]

	return &hint, pathPart
}

// FormatPathWithHint formats a path with a workspace hint
func FormatPathWithHint(workspaceName, path string) string {
	return fmt.Sprintf("@%s:%s", workspaceName, path)
}

// ResolvePathString is a convenience method that resolves a path string
// that may contain workspace hints
func (r *Resolver) ResolvePathString(pathWithHint string) (*ResolvedPath, error) {
	hint, path := ParseWorkspaceHint(pathWithHint)
	return r.ResolvePath(path, hint)
}

// GetWorkspaceForPath returns the workspace name for a given path
// Returns empty string if path is not in any workspace
func (r *Resolver) GetWorkspaceForPath(path string) string {
	// Resolve to absolute path first if needed
	absPath := path
	if !filepath.IsAbs(path) {
		primary := r.manager.GetPrimaryRoot()
		if primary == nil {
			return ""
		}
		absPath = filepath.Join(primary.Path, path)
	}

	// Find containing workspace
	root := r.manager.ResolvePathToRoot(absPath)
	if root == nil {
		return ""
	}

	return root.Name
}

// DisambiguatePath helps resolve ambiguous paths by checking all workspaces
// Returns a list of workspaces that actually contain the given relative path
func (r *Resolver) DisambiguatePath(relativePath string) []string {
	var matches []string

	for _, root := range r.manager.GetRoots() {
		absPath := filepath.Join(root.Path, relativePath)
		// Check if the file or directory exists
		if _, err := os.Stat(absPath); err == nil {
			matches = append(matches, root.Name)
		}
	}

	return matches
}

// ResolvePathWithDisambiguation resolves a relative path, checking file existence
// across multiple workspaces. Returns the best match or the primary workspace.
func (r *Resolver) ResolvePathWithDisambiguation(relativePath string) (*ResolvedPath, error) {
	if filepath.IsAbs(relativePath) {
		return r.ResolvePath(relativePath, nil)
	}

	// Find all workspaces that contain this path
	matches := r.DisambiguatePath(relativePath)

	if len(matches) == 0 {
		// No workspace has this file, use primary workspace
		primary := r.manager.GetPrimaryRoot()
		if primary == nil {
			return nil, fmt.Errorf("no workspace roots available")
		}

		absPath := filepath.Join(primary.Path, relativePath)
		return &ResolvedPath{
			AbsolutePath: absPath,
			Root:         primary,
			RelativePath: relativePath,
		}, nil
	}

	if len(matches) == 1 {
		// Unambiguous - only one workspace has this file
		root := r.manager.GetRootByName(matches[0])
		absPath := filepath.Join(root.Path, relativePath)
		return &ResolvedPath{
			AbsolutePath: absPath,
			Root:         root,
			RelativePath: relativePath,
		}, nil
	}

	// Multiple workspaces have this file
	// Prefer primary workspace if it's in the matches
	primary := r.manager.GetPrimaryRoot()
	for _, match := range matches {
		if match == primary.Name {
			absPath := filepath.Join(primary.Path, relativePath)
			return &ResolvedPath{
				AbsolutePath: absPath,
				Root:         primary,
				RelativePath: relativePath,
			}, nil
		}
	}

	// Primary workspace doesn't have it, use first match
	root := r.manager.GetRootByName(matches[0])
	absPath := filepath.Join(root.Path, relativePath)
	return &ResolvedPath{
		AbsolutePath: absPath,
		Root:         root,
		RelativePath: relativePath,
	}, nil
}

// FileExists checks if a file exists in any workspace
func (r *Resolver) FileExists(path string) bool {
	resolved, err := r.ResolvePathWithDisambiguation(path)
	if err != nil {
		return false
	}

	_, err = os.Stat(resolved.AbsolutePath)
	return err == nil
}
