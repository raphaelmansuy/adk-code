// Package file provides file operation tools for the coding agent.
package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"code_agent/tools/common"
)

// ReadFileInput defines the input parameters for reading a file.
type ReadFileInput struct {
	// Path is the absolute or relative path to the file to read.
	Path string `json:"path" jsonschema:"Path to the file to read"`
	// Offset is the starting line number (1-indexed, optional).
	Offset *int `json:"offset,omitempty" jsonschema:"Start line number (1-indexed, default: 1)"`
	// Limit is the maximum number of lines to read (optional, 0 = all).
	Limit *int `json:"limit,omitempty" jsonschema:"Number of lines to read (default: all)"`
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
				Error:   fmt.Sprintf("Failed to read file: %v", err),
			}
		}

		lines := strings.Split(string(content), "\n")
		totalLines := len(lines)

		// Handle offsets and limits
		offset := 1
		if input.Offset != nil && *input.Offset > 1 {
			offset = *input.Offset
		}

		limit := totalLines
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
		Name:        "read_file",
		Description: "Reads the content of a file from the filesystem with optional line range support. Use this to examine code, configuration files, or any text files.",
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
					Error:   fmt.Sprintf("Failed to create directories: %v", err),
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
				Error:   fmt.Sprintf("Failed to write file: %v", err),
			}
		}

		return WriteFileOutput{
			Success: true,
			Message: fmt.Sprintf("Successfully wrote %d bytes to %s", len(input.Content), input.Path),
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "write_file",
		Description: "Writes content to a file with atomic write support and size validation for safety. Creates the file if it doesn't exist, or overwrites it if it does. Automatically creates parent directories. Prevents accidental data loss by rejecting writes that reduce file size by >90% (override with allow_size_reduce=true).",
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

// ReplaceInFileInput defines the input parameters for replacing text in a file.
type ReplaceInFileInput struct {
	// Path is the path to the file to modify.
	Path string `json:"path" jsonschema:"Path to the file to modify"`
	// OldText is the text to find and replace (must match exactly).
	OldText string `json:"old_text" jsonschema:"Text to find and replace (must match exactly)"`
	// NewText is the text to replace with.
	NewText string `json:"new_text" jsonschema:"Text to replace with"`
	// MaxReplacements is the maximum number of replacements to make (optional, 0 = unlimited)
	MaxReplacements *int `json:"max_replacements,omitempty" jsonschema:"Maximum number of replacements (default: unlimited)"`
}

// ReplaceInFileOutput defines the output of replacing text in a file.
type ReplaceInFileOutput struct {
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`
	// ReplacementCount is the number of replacements made.
	ReplacementCount int `json:"replacement_count"`
	// Message contains a success message.
	Message string `json:"message,omitempty"`
	// Error contains error message if the operation failed.
	Error string `json:"error,omitempty"`
}

// NewReplaceInFileTool creates a tool for replacing text in files.
func NewReplaceInFileTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ReplaceInFileInput) ReplaceInFileOutput {
		// SAFEGUARD: Reject dangerous empty replacements
		if input.NewText == "" {
			return ReplaceInFileOutput{
				Success: false,
				Error: "Refusing to replace with empty text (would delete lines). " +
					"Use edit_lines tool with mode='delete' for intentional deletions, or ensure new_text is not empty.",
			}
		}

		// SAFEGUARD: Normalize whitespace in old_text for better matching
		normalizedOldText := normalizeText(input.OldText)

		content, err := os.ReadFile(input.Path)
		if err != nil {
			return ReplaceInFileOutput{
				Success: false,
				Error:   fmt.Sprintf("Failed to read file: %v", err),
			}
		}

		originalContent := string(content)
		if !strings.Contains(originalContent, normalizedOldText) && !strings.Contains(originalContent, input.OldText) {
			return ReplaceInFileOutput{
				Success: false,
				Error: "Text to replace not found in file. Make sure the old_text matches exactly. " +
					"Note: whitespace (spaces, tabs, newlines) must match exactly.",
			}
		}

		newContent := strings.ReplaceAll(originalContent, normalizedOldText, input.NewText)
		if newContent == originalContent {
			// Try with original text if normalized didn't work
			newContent = strings.ReplaceAll(originalContent, input.OldText, input.NewText)
		}
		replacementCount := strings.Count(originalContent, normalizedOldText)
		if replacementCount == 0 {
			replacementCount = strings.Count(originalContent, input.OldText)
		}

		// SAFEGUARD: Validate replacement count against max_replacements
		if input.MaxReplacements != nil && *input.MaxReplacements > 0 {
			if replacementCount > *input.MaxReplacements {
				return ReplaceInFileOutput{
					Success: false,
					Error: fmt.Sprintf(
						"Too many replacements would occur (%d found, max %d allowed). "+
							"Refusing to apply. Use preview_replace_in_file first to inspect changes.",
						replacementCount,
						*input.MaxReplacements,
					),
				}
			}
		}

		err = os.WriteFile(input.Path, []byte(newContent), 0644)
		if err != nil {
			return ReplaceInFileOutput{
				Success: false,
				Error:   fmt.Sprintf("Failed to write file: %v", err),
			}
		}

		return ReplaceInFileOutput{
			Success:          true,
			ReplacementCount: replacementCount,
			Message:          fmt.Sprintf("Successfully replaced %d occurrence(s) in %s", replacementCount, input.Path),
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "replace_in_file",
		Description: "Finds and replaces text in a file with safety guards. The old_text must match exactly (including whitespace). Useful for making targeted edits to existing files.",
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategoryFileOperations,
			Priority:  2,
			UsageHint: "Simple text replacement (exact match), has max_replacements safety",
		})
	}

	return t, err
}

// normalizeText normalizes whitespace in text for better matching
func normalizeText(text string) string {
	// Convert escaped newlines to actual newlines
	text = strings.ReplaceAll(text, "\\n", "\n")
	text = strings.ReplaceAll(text, "\\t", "\t")
	text = strings.ReplaceAll(text, "\\r", "\r")
	return text
}

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
		Name:        "list_directory",
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
