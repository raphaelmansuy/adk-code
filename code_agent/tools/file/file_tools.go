// Package file provides file operation tools for the coding agent.
package file

import (
	"fmt"
	"os"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"code_agent/tools/common"
)

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
