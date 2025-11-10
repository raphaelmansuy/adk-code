// Package tools provides file operation tools for the coding agent.
package tools

import (
	"fmt"
	"os"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// EditLinesInput defines the input for line-based file editing.
type EditLinesInput struct {
	// FilePath is the path to the file to edit
	FilePath string `json:"file_path" jsonschema:"Path to the file to edit"`
	// StartLine is the starting line number (1-indexed, inclusive)
	StartLine int `json:"start_line" jsonschema:"Starting line number (1-indexed, inclusive)"`
	// EndLine is the ending line number (1-indexed, inclusive)
	EndLine int `json:"end_line" jsonschema:"Ending line number (1-indexed, inclusive)"`
	// NewLines is the replacement content (for replace mode)
	NewLines string `json:"new_lines,omitempty" jsonschema:"Replacement content (for replace/insert modes)"`
	// Mode specifies the operation: "replace", "insert", "delete"
	Mode string `json:"mode,omitempty" jsonschema:"Operation mode: 'replace' (replace lines), 'insert' (insert before line), 'delete' (remove lines) (default: 'replace')"`
	// Preview indicates whether to show changes without applying them
	Preview *bool `json:"preview,omitempty" jsonschema:"Show preview without applying (default: false)"`
}

// EditLinesOutput defines the output of line-based editing.
type EditLinesOutput struct {
	// Success indicates whether the operation was successful
	Success bool `json:"success"`
	// LinesModified is the number of lines affected
	LinesModified int `json:"lines_modified"`
	// Message describes the result
	Message string `json:"message,omitempty"`
	// Preview shows the changes (in preview mode only)
	Preview string `json:"preview,omitempty"`
	// Error is the error message if operation failed
	Error string `json:"error,omitempty"`
}

// NewEditLinesTool creates a tool for line-based file editing.
func NewEditLinesTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input EditLinesInput) EditLinesOutput {
		// Validate input
		if input.FilePath == "" {
			return EditLinesOutput{
				Success: false,
				Error:   "FilePath is required",
			}
		}

		if input.StartLine < 1 {
			return EditLinesOutput{
				Success: false,
				Error:   "StartLine must be >= 1 (1-indexed)",
			}
		}

		if input.EndLine < input.StartLine {
			return EditLinesOutput{
				Success: false,
				Error:   "EndLine must be >= StartLine",
			}
		}

		// Default mode is replace
		mode := "replace"
		if input.Mode != "" {
			mode = input.Mode
		}

		// Validate mode
		if mode != "replace" && mode != "insert" && mode != "delete" {
			return EditLinesOutput{
				Success: false,
				Error:   "Mode must be 'replace', 'insert', or 'delete'",
			}
		}

		// Read the file
		content, err := os.ReadFile(input.FilePath)
		if err != nil {
			return EditLinesOutput{
				Success: false,
				Error:   fmt.Sprintf("Failed to read file: %v", err),
			}
		}

		lines := strings.Split(string(content), "\n")

		// Validate line numbers
		if input.StartLine > len(lines) {
			return EditLinesOutput{
				Success: false,
				Error:   fmt.Sprintf("StartLine (%d) exceeds file length (%d lines)", input.StartLine, len(lines)),
			}
		}

		// Perform the operation
		var newLines []string
		var linesModified int

		switch mode {
		case "replace":
			// Replace lines from StartLine to EndLine with NewLines
			endLine := input.EndLine
			if endLine > len(lines) {
				endLine = len(lines)
			}

			// Copy lines before the range
			newLines = append(newLines, lines[:input.StartLine-1]...)

			// Add new content
			replacementLines := strings.Split(input.NewLines, "\n")
			// Remove empty last line if it's just a trailing newline
			if len(replacementLines) > 0 && replacementLines[len(replacementLines)-1] == "" {
				replacementLines = replacementLines[:len(replacementLines)-1]
			}
			newLines = append(newLines, replacementLines...)

			// Copy lines after the range
			newLines = append(newLines, lines[endLine:]...)
			linesModified = endLine - input.StartLine + 1

		case "insert":
			// Insert NewLines before StartLine
			newLines = append(newLines, lines[:input.StartLine-1]...)
			insertionLines := strings.Split(input.NewLines, "\n")
			// Remove empty last line if it's just a trailing newline
			if len(insertionLines) > 0 && insertionLines[len(insertionLines)-1] == "" {
				insertionLines = insertionLines[:len(insertionLines)-1]
			}
			newLines = append(newLines, insertionLines...)
			newLines = append(newLines, lines[input.StartLine-1:]...)
			linesModified = len(insertionLines)

		case "delete":
			// Delete lines from StartLine to EndLine
			endLine := input.EndLine
			if endLine > len(lines) {
				endLine = len(lines)
			}
			newLines = append(newLines, lines[:input.StartLine-1]...)
			newLines = append(newLines, lines[endLine:]...)
			linesModified = endLine - input.StartLine + 1
		}

		// Generate preview if requested or in preview mode
		previewMode := false
		if input.Preview != nil {
			previewMode = *input.Preview
		}

		preview := generateEditPreview(lines, newLines, input.StartLine, input.EndLine, mode)

		if previewMode {
			return EditLinesOutput{
				Success:       true,
				LinesModified: linesModified,
				Message:       fmt.Sprintf("Preview of %s operation on lines %d-%d", mode, input.StartLine, input.EndLine),
				Preview:       preview,
			}
		}

		// Write the modified content back to the file
		newContent := strings.Join(newLines, "\n")
		if err := AtomicWrite(input.FilePath, []byte(newContent), 0644); err != nil {
			return EditLinesOutput{
				Success: false,
				Error:   fmt.Sprintf("Failed to write file: %v", err),
			}
		}

		return EditLinesOutput{
			Success:       true,
			LinesModified: linesModified,
			Message:       fmt.Sprintf("Successfully performed %s on lines %d-%d (%d lines affected)", mode, input.StartLine, input.EndLine, linesModified),
			Preview:       preview,
		}
	}

	return functiontool.New(functiontool.Config{
		Name:        "edit_lines",
		Description: "Edit specific lines in a file by line number. Supports replace, insert, and delete operations. More precise than string-based replacement for structural changes.",
	}, handler)
}

// generateEditPreview creates a human-readable preview of the changes
func generateEditPreview(originalLines []string, newLines []string, startLine, endLine int, mode string) string {
	var preview strings.Builder
	preview.WriteString(fmt.Sprintf("Preview of %s operation on lines %d-%d:\n", mode, startLine, endLine))
	preview.WriteString(strings.Repeat("=", 60) + "\n\n")

	// Show context before
	contextStart := startLine - 3
	if contextStart < 1 {
		contextStart = 1
	}

	preview.WriteString("BEFORE:\n")
	preview.WriteString(strings.Repeat("-", 60) + "\n")
	for i := contextStart - 1; i < endLine && i < len(originalLines); i++ {
		if i >= 0 {
			prefix := " "
			if i+1 >= startLine && i+1 <= endLine {
				prefix = "-"
			}
			lineNum := i + 1
			preview.WriteString(fmt.Sprintf("%s %3d: %s\n", prefix, lineNum, originalLines[i]))
		}
	}

	preview.WriteString("\nAFTER:\n")
	preview.WriteString(strings.Repeat("-", 60) + "\n")

	// Show context after
	contextEnd := endLine + 3
	newStartIdx := startLine - 1
	newEndIdx := newStartIdx + (endLine - startLine + 1)

	for i := contextStart - 1; i < contextEnd && i < len(newLines); i++ {
		if i >= 0 {
			prefix := " "
			if mode == "replace" && i >= newStartIdx && i < newEndIdx {
				prefix = "+"
			} else if mode == "delete" && i >= newStartIdx && i < newEndIdx {
				prefix = "-"
			} else if mode == "insert" && i >= newStartIdx && i < newEndIdx {
				prefix = "+"
			}
			lineNum := i + 1
			preview.WriteString(fmt.Sprintf("%s %3d: %s\n", prefix, lineNum, newLines[i]))
		}
	}

	return preview.String()
}
