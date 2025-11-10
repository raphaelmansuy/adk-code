// Package file provides file operation tools for the coding agent.
package file

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWrite performs a safe, atomic file write operation.
// It writes to a temporary file, syncs to disk, and then atomically renames
// to the target path. This ensures the file is either fully written or unchanged.
func AtomicWrite(path string, content []byte, perm os.FileMode) error {
	// 1. Create temp file in the same directory
	dir := filepath.Dir(path)
	if dir == "" {
		dir = "."
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// 2. Write content
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// 3. Set permissions
	if err := tmpFile.Chmod(perm); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// 4. Sync to disk
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to sync: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// 5. Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename: %w", err)
	}

	return nil
}
