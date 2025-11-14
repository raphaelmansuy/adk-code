// Package file provides file operation tools for the coding agent.
package file

import (
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/pkg/errors"
	common "adk-code/tools/base"
)

// WriteFileInput defines the input parameters for writing a file.
type WriteFileInput struct {
	// Path is the absolute or relative path to the file to write.
	Path string `json:"path" jsonschema:"Path to the file to write"`
	// Content is the content to write to the file.
	Content string `json:"content" jsonschema:"Content to write to the file"`
	// CreateDirs indicates whether to create parent directories if they don't exist.
	CreateDirs *bool `json:"create_dirs,omitempty" jsonschema:"Create parent directories if they don't exist (default: true)"`
	// Atomic indicates whether to use atomic write (default: true)
	Atomic *bool `json:"atomic,omitempty" jsonschema:"Use atomic write for safety (default: true)"`
	// AllowSizeReduce allows writing much smaller content than the current file size (default: false)
	AllowSizeReduce *bool `json:"allow_size_reduce,omitempty" jsonschema:"Allow writing content that is <10% of current file size (default: false)"`
}

// WriteFileOutput defines the output of writing a file.
type WriteFileOutput struct {
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`
	// Message contains a success message.
	Message string `json:"message,omitempty"`
	// Error contains error message if the operation failed.
	Error string `json:"error,omitempty"`
}

// NewWriteFileTool creates a tool for writing files.
func NewWriteFileTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input WriteFileInput) WriteFileOutput {
		// SAFEGUARD: Check for suspicious size reduction
		if info, err := os.Stat(input.Path); err == nil {
			currentSize := info.Size()
			newSize := int64(len(input.Content))

			// Detect dangerous size reduction (>90% reduction, file >1KB)
			if currentSize > 1000 && newSize < currentSize/10 {
				allowSizeReduce := false
				if input.AllowSizeReduce != nil {
					allowSizeReduce = *input.AllowSizeReduce
				}

				if !allowSizeReduce {
					return WriteFileOutput{
						Success: false,
						Error: fmt.Sprintf(
							"SAFETY CHECK FAILED: Refusing to reduce file size from %d to %d bytes (%.1f%% reduction).\n"+
								"This might be accidental data loss. If this is intentional, set allow_size_reduce=true.\n"+
								"TIP: Use read_file first to verify you have the complete content, or use edit_lines for targeted changes.",
							currentSize, newSize, float64(currentSize-newSize)/float64(currentSize)*100,
						),
					}
				}
			}
		}

		// Default to creating directories
		createDirs := true
		if input.CreateDirs != nil {
			createDirs = *input.CreateDirs
		}
		if createDirs {
			dir := filepath.Dir(input.Path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return WriteFileOutput{
					Success: false,
					Error:   errors.Wrap(errors.CodeExecution, "failed to create directories", err).Error(),
				}
			}
		}

		// Default to using atomic write
		useAtomic := true
		if input.Atomic != nil {
			useAtomic = *input.Atomic
		}

		var err error
		if useAtomic {
			err = AtomicWrite(input.Path, []byte(input.Content), 0644)
		} else {
			err = os.WriteFile(input.Path, []byte(input.Content), 0644)
		}

		if err != nil {
			return WriteFileOutput{
				Success: false,
				Error:   errors.Wrap(errors.CodeExecution, "failed to write file", err).Error(),
			}
		}

		return WriteFileOutput{
			Success: true,
			Message: fmt.Sprintf("Successfully wrote %d bytes to %s", len(input.Content), input.Path),
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name: "builtin_write_file",
		Description: `Write content to a file with built-in safety checks.

**Parameters:**
- path (required): Path to file to create or overwrite
- content (required): Complete content to write
- create_dirs (optional): Create parent directories if missing (default: true)
- atomic (optional): Use atomic write for safety (default: true)
- allow_size_reduce (optional): Allow write if new size is <10% of old size (default: false)

**Key Features:**
- Creates file if it doesn't exist
- Automatically creates parent directories
- Atomic write prevents file corruption (writes to temp file, then renames)
- Size validation: Rejects if new content is <10% of current file size (prevents accidental data loss)

**Important:** Always provide COMPLETE file content. Never truncate or omit sections.

**Example:** Create new file with defaults:
  path="src/app.go", content="package main\n..."

**Example:** Overwrite with explicit size reduction allowed:
  path="config.txt", content="new content", allow_size_reduce=true`,
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategoryFileOperations,
			Priority:  1,
			UsageHint: "Create or overwrite files with safety checks, atomic writes prevent corruption",
		})
	}

	return t, err
}

// init registers the write file tool automatically at package initialization.
func init() {
	_, _ = NewWriteFileTool()
}
