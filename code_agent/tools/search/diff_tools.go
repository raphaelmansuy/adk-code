// Package search provides search discovery tools for the coding agent.
package search

import (
	"fmt"
	"os"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	common "code_agent/tools/base"
)

// PreviewReplaceInput defines the input for previewing a replace operation.
type PreviewReplaceInput struct {
	// FilePath is the path to the file
	FilePath string `json:"file_path" jsonschema:"Path to the file to preview"`
	// OldText is the text to find and replace
	OldText string `json:"old_text" jsonschema:"Text to find and replace"`
	// NewText is the replacement text
	NewText string `json:"new_text" jsonschema:"Text to replace with"`
	// Context is the number of lines of context to show (default: 3)
	Context *int `json:"context,omitempty" jsonschema:"Lines of context to show (default: 3)"`
}

// PreviewReplaceOutput defines the output of previewing a replace operation.
type PreviewReplaceOutput struct {
	// Success indicates whether the preview was generated successfully
	Success bool `json:"success"`
	// Diff is the unified diff preview
	Diff string `json:"diff"`
	// Changes is the number of changes found
	Changes int `json:"changes"`
	// Preview is a human-readable preview
	Preview string `json:"preview,omitempty"`
	// Error is the error message if the operation failed
	Error string `json:"error,omitempty"`
}

// GenerateDiff creates a simple unified diff between original and modified content
func GenerateDiff(original, modified string, contextLines int) string {
	origLines := strings.Split(original, "\n")
	modLines := strings.Split(modified, "\n")

	var diff strings.Builder
	diff.WriteString("--- original\n")
	diff.WriteString("+++ modified\n")

	// Simple line-by-line diff (not a full LCS-based implementation)
	// For a more robust solution, use external library like github.com/sergi/go-diff
	origIdx := 0
	modIdx := 0

	for origIdx < len(origLines) || modIdx < len(modLines) {
		if origIdx < len(origLines) && modIdx < len(modLines) {
			if origLines[origIdx] == modLines[modIdx] {
				// Same line - show as context
				diff.WriteString(fmt.Sprintf(" %s\n", origLines[origIdx]))
				origIdx++
				modIdx++
			} else {
				// Different line - show removal and addition
				diff.WriteString(fmt.Sprintf("-%s\n", origLines[origIdx]))
				origIdx++
				diff.WriteString(fmt.Sprintf("+%s\n", modLines[modIdx]))
				modIdx++
			}
		} else if origIdx < len(origLines) {
			// Remaining lines in original
			diff.WriteString(fmt.Sprintf("-%s\n", origLines[origIdx]))
			origIdx++
		} else {
			// Remaining lines in modified
			diff.WriteString(fmt.Sprintf("+%s\n", modLines[modIdx]))
			modIdx++
		}
	}

	return diff.String()
}

// GeneratePreviewWithContext generates a preview showing the changes with surrounding context
func GeneratePreviewWithContext(original, modified string, contextLines int) string {
	origLines := strings.Split(original, "\n")
	modLines := strings.Split(modified, "\n")

	var preview strings.Builder
	preview.WriteString("Preview of changes:\n")
	preview.WriteString(strings.Repeat("=", 60) + "\n")

	// Find changed lines
	if len(origLines) != len(modLines) {
		// Simple approach for different lengths
		for i := 0; i < len(origLines) && i < len(modLines); i++ {
			if origLines[i] != modLines[i] {
				start := i - contextLines
				if start < 0 {
					start = 0
				}
				end := i + contextLines + 1
				if end > len(origLines) {
					end = len(origLines)
				}

				preview.WriteString(fmt.Sprintf("Around line %d:\n", i+1))
				for j := start; j < end; j++ {
					if j < len(origLines) {
						if j == i {
							preview.WriteString(fmt.Sprintf("< Line %d: %s\n", j+1, origLines[j]))
						} else {
							preview.WriteString(fmt.Sprintf("  Line %d: %s\n", j+1, origLines[j]))
						}
					}
				}
				preview.WriteString("\n")
				for j := start; j < end; j++ {
					if j < len(modLines) {
						if j == i {
							preview.WriteString(fmt.Sprintf("> Line %d: %s\n", j+1, modLines[j]))
						} else {
							preview.WriteString(fmt.Sprintf("  Line %d: %s\n", j+1, modLines[j]))
						}
					}
				}
				preview.WriteString(strings.Repeat("-", 60) + "\n")
			}
		}
	}

	return preview.String()
}

// NewPreviewReplaceTool creates a tool to preview changes before applying them.
func NewPreviewReplaceTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input PreviewReplaceInput) PreviewReplaceOutput {
		// Read the file
		content, err := os.ReadFile(input.FilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return PreviewReplaceOutput{
					Success: false,
					Error:   fmt.Sprintf("File not found: %s", input.FilePath),
				}
			}
			return PreviewReplaceOutput{
				Success: false,
				Error:   fmt.Sprintf("Failed to read file: %v", err),
			}
		}

		original := string(content)

		// Count occurrences
		changes := strings.Count(original, input.OldText)
		if changes == 0 {
			return PreviewReplaceOutput{
				Success: false,
				Error:   fmt.Sprintf("No matches found for: %s", input.OldText),
			}
		}

		// Generate modified content
		modified := strings.ReplaceAll(original, input.OldText, input.NewText)

		// Get context lines (default 3)
		contextLines := 3
		if input.Context != nil {
			contextLines = *input.Context
			if contextLines < 0 {
				contextLines = 0
			}
		}

		// Generate diff
		diff := GenerateDiff(original, modified, contextLines)
		preview := GeneratePreviewWithContext(original, modified, contextLines)

		return PreviewReplaceOutput{
			Success: true,
			Diff:    diff,
			Changes: changes,
			Preview: preview,
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "builtin_preview_replace_in_file",
		Description: "Preview changes before applying a replace operation. Shows a unified diff and context for the changes.",
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategoryCodeEditing,
			Priority:  5,
			UsageHint: "Preview replace operations before applying, shows unified diff",
		})
	}

	return t, err
}

// GeneratePatchFromReplacement generates a unified diff patch from a replacement operation
func GeneratePatchFromReplacement(filePath, oldText, newText string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	original := string(content)
	modified := strings.ReplaceAll(original, oldText, newText)

	// Generate simple patch
	diff := GenerateDiff(original, modified, 3)
	return diff, nil
}
