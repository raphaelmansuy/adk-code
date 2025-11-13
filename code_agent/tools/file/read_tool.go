// Package file provides file operation tools for the coding agent.
package file

import (
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"code_agent/pkg/errors"
	common "code_agent/tools/base"
)

// ReadFileInput defines the input parameters for reading a file.
type ReadFileInput struct {
	// Path is the absolute or relative path to the file to read.
	Path string `json:"path" jsonschema:"Path to the file to read"`
	// Offset is the starting line number (1-indexed, optional).
	Offset *int `json:"offset,omitempty" jsonschema:"Start line number (1-indexed, default: 1)"`
	// Limit is the maximum number of lines to read (optional, default: 1000).
	Limit *int `json:"limit,omitempty" jsonschema:"Number of lines to read (default: 1000)"`
}

// ReadFileOutput defines the output of reading a file.
type ReadFileOutput struct {
	// Content is the content of the file.
	Content string `json:"content"`
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`
	// Error contains error message if the operation failed.
	Error string `json:"error,omitempty"`
	// TotalLines is the total number of lines in the file.
	TotalLines int `json:"total_lines"`
	// ReturnedLines is the number of lines returned.
	ReturnedLines int `json:"returned_lines"`
	// StartLine is the starting line number returned.
	StartLine int `json:"start_line"`
	// FilePath is the absolute path to the file.
	FilePath string `json:"file_path"`
	// DateCreated is the creation time of the file (RFC3339 format).
	DateCreated string `json:"date_created,omitempty"`
	// DateModified is the last modification time of the file (RFC3339 format).
	DateModified string `json:"date_modified"`
}

// NewReadFileTool creates a tool for reading files.
func NewReadFileTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ReadFileInput) ReadFileOutput {
		content, err := os.ReadFile(input.Path)
		if err != nil {
			return ReadFileOutput{
				Success: false,
				Error:   errors.FileNotFoundError(input.Path).Error(),
			}
		}

		lines := strings.Split(string(content), "\n")
		totalLines := len(lines)

		// Handle offsets and limits
		offset := 1
		if input.Offset != nil && *input.Offset > 1 {
			offset = *input.Offset
		}

		// Default limit is 1000 lines to prevent reading excessively large files
		limit := 1000
		if totalLines < 1000 {
			limit = totalLines
		}
		if input.Limit != nil && *input.Limit > 0 {
			limit = *input.Limit
		}

		endIdx := offset + limit - 1
		if endIdx > totalLines {
			endIdx = totalLines
		}

		var selectedLines []string
		if offset <= totalLines {
			selectedLines = lines[offset-1 : endIdx]
		}

		// Get file stats for path, creation time, and modification time
		absPath, _ := filepath.Abs(input.Path)
		dateModified := ""
		dateCreated := ""

		if fileInfo, err := os.Stat(input.Path); err == nil {
			dateModified = fileInfo.ModTime().Format("2006-01-02T15:04:05Z07:00")
			// Note: On Unix systems, birth time is not readily available.
			// On macOS, we would need system-specific code to get it.
			// For now, we set it empty or use ModTime as fallback.
			dateCreated = dateModified
		}

		return ReadFileOutput{
			Content:       strings.Join(selectedLines, "\n"),
			Success:       true,
			TotalLines:    totalLines,
			ReturnedLines: len(selectedLines),
			StartLine:     offset,
			FilePath:      absPath,
			DateCreated:   dateCreated,
			DateModified:  dateModified,
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "builtin_read_file",
		Description: "Reads the content of a file from the filesystem with optional line range support. By default, returns up to 1000 lines from the file. Use offset to start at a specific line number (1-indexed, default: 1) and limit to control the maximum number of lines returned (omit to use default of 1000). Use this to examine code, configuration files, or any text files.",
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategoryFileOperations,
			Priority:  0,
			UsageHint: "Examine code, read configs, supports line ranges (offset/limit) for large files",
		})
	}

	return t, err
}

// init registers the read file tool automatically at package initialization.
func init() {
	_, _ = NewReadFileTool()
}
