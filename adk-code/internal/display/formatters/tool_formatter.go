package formatters

import (
	"fmt"
	"path/filepath"
	"strings"

	"adk-code/internal/display/styles"

	"github.com/charmbracelet/lipgloss"
)

// ToolFormatter formats tool calls and results
type ToolFormatter struct {
	styles       *styles.Styles
	formatter    *styles.Formatter
	outputFormat string
}

// NewToolFormatter creates a new tool formatter
func NewToolFormatter(outputFormat string, s *styles.Styles, f *styles.Formatter) *ToolFormatter {
	return &ToolFormatter{
		styles:       s,
		formatter:    f,
		outputFormat: outputFormat,
	}
}

// RenderToolCall renders a tool call with contextual formatting
func (tf *ToolFormatter) RenderToolCall(toolName string, args map[string]any) string {
	// Create contextual header based on tool
	header := tf.getToolHeader(toolName, args)
	// Add spacing before tool call for better readability
	return "\n" + header + "\n"
}

// truncatePath smartly truncates long file paths for display
// Shows filename + parent directory for long paths, preserving important context
// Examples:
//
//	/very/long/path/to/project/src/main.go -> .../src/main.go
//	./main.go -> ./main.go
func (tf *ToolFormatter) truncatePath(path string, maxLength int) string {
	if len(path) <= maxLength {
		return path
	}

	// Try to show filename + parent directory
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	parent := filepath.Base(dir)

	shortened := filepath.Join("...", parent, base)
	if len(shortened) <= maxLength {
		return shortened
	}

	// If still too long, just show filename with ellipsis
	if len(base) <= maxLength-4 {
		return ".../" + base
	}

	// Last resort: truncate the filename itself
	return "..." + base[len(base)-(maxLength-3):]
}

// getToolHeader generates a contextual header for tool calls
func (tf *ToolFormatter) getToolHeader(toolName string, args map[string]any) string {
	// Create a subtle tool icon
	toolIcon := "◆"
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if tf.outputFormat == styles.OutputFormatPlain || !isTTY {
		toolIcon = "→"
	}

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "245"})

	toolStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")). // Blue
		Bold(false)

	switch toolName {
	case "read_file":
		if path, ok := args["path"].(string); ok {
			displayPath := tf.truncatePath(path, 60)
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Reading") + " " + tf.formatter.Dim(displayPath)
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Reading file")

	case "write_file":
		if path, ok := args["path"].(string); ok {
			displayPath := tf.truncatePath(path, 60)
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Writing") + " " + tf.formatter.Dim(displayPath)
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Writing file")

	case "replace_in_file", "search_replace":
		if path, ok := args["path"].(string); ok {
			displayPath := tf.truncatePath(path, 60)
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Editing") + " " + tf.formatter.Dim(displayPath)
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Editing file")

	case "list_directory":
		if path, ok := args["path"].(string); ok {
			displayPath := tf.truncatePath(path, 60)
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Listing") + " " + tf.formatter.Dim(displayPath)
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Listing files")

	case "execute_command", "execute_program":
		if command, ok := args["command"].(string); ok {
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Running") + " " + tf.formatter.Dim("`"+command+"`")
		}
		if program, ok := args["program"].(string); ok {
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Running") + " " + tf.formatter.Dim("`"+program+"`")
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Running command")

	case "grep_search":
		if pattern, ok := args["pattern"].(string); ok {
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Searching for") + " " + tf.formatter.Dim("`"+pattern+"`")
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Searching files")

	default:
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render(toolName)
	}
}

// extractError extracts error messages from various error formats in tool results
// Handles: string errors, empty objects {}, objects with message/details fields
func (tf *ToolFormatter) extractError(result map[string]any) string {
	errorValue, hasError := result["error"]
	if !hasError {
		return ""
	}

	// Handle string error (most common case)
	if errStr, ok := errorValue.(string); ok && errStr != "" {
		return errStr
	}

	// Handle empty error object {} (common with MCP tools that fail)
	if errorMap, ok := errorValue.(map[string]any); ok {
		// If error object is empty, return generic message
		if len(errorMap) == 0 {
			// Check if there's any other useful information in the result
			if output, ok := result["output"].(string); ok && output != "" {
				return output
			}
			return "Tool execution failed with no error details provided"
		}

		// Try common error field names in the error object
		if msg, ok := errorMap["message"].(string); ok && msg != "" {
			return msg
		}
		if msg, ok := errorMap["error"].(string); ok && msg != "" {
			return msg
		}
		if msg, ok := errorMap["details"].(string); ok && msg != "" {
			return msg
		}
		if msg, ok := errorMap["text"].(string); ok && msg != "" {
			return msg
		}

		// If error object has fields but none match common patterns, return generic message
		return "Tool execution failed"
	}

	// Fallback: convert any other error type to string
	return fmt.Sprintf("%v", errorValue)
}

// RenderToolResult renders a tool result with contextual formatting
func (tf *ToolFormatter) RenderToolResult(toolName string, result map[string]any) string {
	// Check for errors - handle multiple error formats
	if err := tf.extractError(result); err != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")). // Red
			Bold(false)
		return "  " + errorStyle.Render("✗ "+err) + "\n"
	}

	// Subtle success indicator
	checkmark := "✓"
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if tf.outputFormat == styles.OutputFormatPlain || !isTTY {
		checkmark = "OK"
	}

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"})

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")) // Green

	// Add contextual success message based on tool type
	var message string
	switch toolName {
	case "read_file":
		if content, ok := result["content"].(string); ok {
			lines := len(strings.Split(content, "\n"))
			message = dimStyle.Render(fmt.Sprintf("  %s Read %d lines", successStyle.Render(checkmark), lines))
		} else {
			message = dimStyle.Render("  " + successStyle.Render(checkmark) + " Read complete")
		}
	case "write_file":
		if path, ok := result["path"].(string); ok {
			displayPath := tf.truncatePath(path, 50)
			message = dimStyle.Render("  " + successStyle.Render(checkmark) + " Wrote " + displayPath)
		} else {
			message = dimStyle.Render("  " + successStyle.Render(checkmark) + " Write complete")
		}
	case "replace_in_file", "search_replace":
		message = dimStyle.Render("  " + successStyle.Render(checkmark) + " Edit applied")
	case "list_directory":
		if items, ok := result["items"].([]any); ok {
			message = dimStyle.Render(fmt.Sprintf("  %s Found %d items", successStyle.Render(checkmark), len(items)))
		} else {
			message = dimStyle.Render("  " + successStyle.Render(checkmark) + " List complete")
		}
	case "execute_command", "execute_program":
		if exitCode, ok := result["exit_code"].(int); ok && exitCode == 0 {
			message = dimStyle.Render("  " + successStyle.Render(checkmark) + " Command successful")
		} else if exitCode, ok := result["exit_code"].(float64); ok && exitCode == 0 {
			message = dimStyle.Render("  " + successStyle.Render(checkmark) + " Command successful")
		} else {
			message = dimStyle.Render("  " + successStyle.Render(checkmark) + " Command complete")
		}
	case "grep_search":
		if matches, ok := result["matches"].([]any); ok {
			message = dimStyle.Render(fmt.Sprintf("  %s Found %d matches", successStyle.Render(checkmark), len(matches)))
		} else {
			message = dimStyle.Render("  " + successStyle.Render(checkmark) + " Search complete")
		}
	default:
		message = dimStyle.Render("  " + successStyle.Render(checkmark) + " Complete")
	}

	return message + "\n"
}
