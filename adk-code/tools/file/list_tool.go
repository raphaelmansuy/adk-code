// Package file provides file operation tools for the coding agent.
package file

import (
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	common "adk-code/tools/base"
)

// ListDirectoryInput defines the input parameters for listing directory contents.
type ListDirectoryInput struct {
	// Path is the path to the directory to list.
	Path string `json:"path" jsonschema:"Path to the directory to list"`
	// Recursive indicates whether to list subdirectories recursively.
	Recursive *bool `json:"recursive,omitempty" jsonschema:"List subdirectories recursively (default: false)"`
}

// FileInfo represents information about a file or directory.
type FileInfo struct {
	// Name is the name of the file or directory.
	Name string `json:"name"`
	// Path is the full path to the file or directory.
	Path string `json:"path"`
	// IsDir indicates whether this is a directory.
	IsDir bool `json:"is_dir"`
	// Size is the size of the file in bytes (0 for directories).
	Size int64 `json:"size"`
}

// ListDirectoryOutput defines the output of listing a directory.
type ListDirectoryOutput struct {
	// Files is the list of files and directories.
	Files []FileInfo `json:"files"`
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`
	// Error contains error message if the operation failed.
	Error string `json:"error,omitempty"`
}

// NewListDirectoryTool creates a tool for listing directory contents.
func NewListDirectoryTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ListDirectoryInput) ListDirectoryOutput {
		// Initialize with empty slice, not nil
		files := make([]FileInfo, 0)

		recursive := false
		if input.Recursive != nil {
			recursive = *input.Recursive
		}

		if recursive {
			err := filepath.Walk(input.Path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				files = append(files, FileInfo{
					Name:  info.Name(),
					Path:  path,
					IsDir: info.IsDir(),
					Size:  info.Size(),
				})
				return nil
			})
			if err != nil {
				return ListDirectoryOutput{
					Files:   make([]FileInfo, 0),
					Success: false,
					Error:   fmt.Sprintf("Failed to list directory: %v", err),
				}
			}
		} else {
			entries, err := os.ReadDir(input.Path)
			if err != nil {
				return ListDirectoryOutput{
					Files:   make([]FileInfo, 0),
					Success: false,
					Error:   fmt.Sprintf("Failed to list directory: %v", err),
				}
			}

			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				files = append(files, FileInfo{
					Name:  entry.Name(),
					Path:  filepath.Join(input.Path, entry.Name()),
					IsDir: entry.IsDir(),
					Size:  info.Size(),
				})
			}
		}

		return ListDirectoryOutput{
			Files:   files,
			Success: true,
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "builtin_list_directory",
		Description: "Lists the contents of a directory. Can list recursively to explore entire directory trees. Use this to understand project structure.",
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategoryFileOperations,
			Priority:  3,
			UsageHint: "Explore directory structure, supports recursive listing",
		})
	}

	return t, err
}

// init registers the list directory tool automatically at package initialization.
func init() {
	_, _ = NewListDirectoryTool()
}
