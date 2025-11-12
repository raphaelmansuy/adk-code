// Package file provides file operation tools for the coding agent.
package file

import (
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"code_agent/tools/base"
)

// SearchFilesInput defines the input parameters for searching files.
type SearchFilesInput struct {
	// Path is the root directory to search in.
	Path string `json:"path" jsonschema:"Root directory to search in"`
	// Pattern is the search pattern (supports * and ? wildcards).
	Pattern string `json:"pattern" jsonschema:"Search pattern (supports * and ? wildcards, e.g., '*.go', 'test_*.py')"`
	// MaxResults is the maximum number of results to return.
	MaxResults *int `json:"max_results,omitempty" jsonschema:"Maximum number of results to return (default: 100)"`
}

// SearchFilesOutput defines the output of searching files.
type SearchFilesOutput struct {
	// Matches is the list of matching file paths.
	Matches []string `json:"matches"`
	// Count is the total number of matches found.
	Count int `json:"count"`
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`
	// Error contains error message if the operation failed.
	Error string `json:"error,omitempty"`
}

// NewSearchFilesTool creates a tool for searching files by pattern.
func NewSearchFilesTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input SearchFilesInput) SearchFilesOutput {
		maxResults := 100 // default
		if input.MaxResults != nil {
			maxResults = *input.MaxResults
		}

		// Initialize with empty slice, not nil
		matches := make([]string, 0)
		err := filepath.Walk(input.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors and continue
			}
			if info.IsDir() {
				return nil
			}

			matched, err := filepath.Match(input.Pattern, filepath.Base(path))
			if err != nil {
				return err
			}

			if matched {
				matches = append(matches, path)
				if len(matches) >= maxResults {
					return filepath.SkipAll
				}
			}

			return nil
		})

		if err != nil && err != filepath.SkipAll {
			return SearchFilesOutput{
				Matches: make([]string, 0),
				Count:   0,
				Success: false,
				Error:   fmt.Sprintf("Failed to search files: %v", err),
			}
		}

		return SearchFilesOutput{
			Matches: matches,
			Count:   len(matches),
			Success: true,
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "search_files",
		Description: "Searches for files matching a pattern in a directory tree. Supports wildcards (* for any characters, ? for single character). Example: '*.go' finds all Go files.",
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategorySearchDiscovery,
			Priority:  0,
			UsageHint: "Find files by pattern (*.go, test_*.py), uses wildcard matching",
		})
	}

	return t, err
}

// init registers the search files tool automatically at package initialization.
func init() {
	_, _ = NewSearchFilesTool()
}
