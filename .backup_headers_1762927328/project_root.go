// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetProjectRoot traverses to find the project root,
// identified by the presence of a "go.mod" file.
// It searches: current path, immediate subdirectories, and parent directories.
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
			// Reached the root of the filesystem
			return "", fmt.Errorf("go.mod not found in %s, its subdirectories, or any parent directories", startPath)
		}
		currentPath = parentPath

		goModPath := filepath.Join(currentPath, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentPath, nil
		}
	}
}
