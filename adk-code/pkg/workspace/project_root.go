package workspace

import (
	"os"
	"path/filepath"
)

// GetProjectRoot traverses to find the project root,
// identified by the presence of a "go.mod" file.
// It searches: current path, immediate subdirectories, and parent directories.
// If no go.mod is found, it returns the start path as a fallback to support
// usage in non-Go projects or when installed as a global CLI tool.
func GetProjectRoot(startPath string) (string, error) {
	// First, check if go.mod exists in the start path
	if _, err := os.Stat(filepath.Join(startPath, "go.mod")); err == nil {
		return startPath, nil
	}

	// Check if go.mod exists in immediate subdirectories (e.g., code_agent/)
	entries, err := os.ReadDir(startPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subdir := filepath.Join(startPath, entry.Name())
				if _, err := os.Stat(filepath.Join(subdir, "go.mod")); err == nil {
					return subdir, nil
				}
			}
		}
	}

	// Then traverse upwards to find go.mod in parent directories
	currentPath := startPath
	for {
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached the root of the filesystem without finding go.mod
			// Fall back to using startPath to support non-Go projects
			return startPath, nil
		}
		currentPath = parentPath

		goModPath := filepath.Join(currentPath, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentPath, nil
		}
	}
}
