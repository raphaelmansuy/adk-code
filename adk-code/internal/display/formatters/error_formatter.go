package formatters

import (
	"fmt"
	"strings"

	"adk-code/internal/display/styles"

	"github.com/charmbracelet/lipgloss"
)

// ErrorFormatter formats error, warning, and info messages
type ErrorFormatter struct {
	styles       *styles.Styles
	formatter    *styles.Formatter
	outputFormat string
	mdRenderer   MarkdownRenderer
}

// NewErrorFormatter creates a new error formatter
func NewErrorFormatter(outputFormat string, s *styles.Styles, f *styles.Formatter, mdRenderer MarkdownRenderer) *ErrorFormatter {
	return &ErrorFormatter{
		styles:       s,
		formatter:    f,
		outputFormat: outputFormat,
		mdRenderer:   mdRenderer,
	}
}

// RenderWarning renders a warning message with subtle styling
func (ef *ErrorFormatter) RenderWarning(message string) string {
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if ef.outputFormat == styles.OutputFormatPlain || !isTTY {
		return "Warning: " + message + "\n"
	}

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")). // Yellow
		Bold(false)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "245"})

	icon := "⚠"
	return "\n" + warningStyle.Render(icon+" Warning") + ": " + dimStyle.Render(message) + "\n"
}

// RenderInfo renders an informational message with subtle styling
func (ef *ErrorFormatter) RenderInfo(message string) string {
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if ef.outputFormat == styles.OutputFormatPlain || !isTTY {
		return "Info: " + message + "\n"
	}

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")). // Blue
		Bold(false)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "245"})

	icon := "ℹ"
	return "\n" + infoStyle.Render(icon+" Info") + ": " + dimStyle.Render(message) + "\n"
}

// RenderError renders an error message with actionable suggestions
func (ef *ErrorFormatter) RenderError(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Detect error type and provide suggestions
	suggestions := GetErrorSuggestions(errMsg)

	// Format the main error
	markdown := fmt.Sprintf("### %s\n\n%s", ef.formatter.ErrorX("Error"), errMsg)

	// Try to render as markdown
	rendered := markdown
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if ef.mdRenderer != nil && ef.outputFormat != styles.OutputFormatPlain && isTTY {
		if mdRendered, mdErr := ef.mdRenderer.Render(markdown); mdErr == nil {
			rendered = mdRendered
		}
	}

	output := "\n" + rendered

	// Add suggestions if available
	if len(suggestions) > 0 {
		suggestionsStr := "\n💡 **Suggestions:**\n"
		for _, suggestion := range suggestions {
			suggestionsStr += fmt.Sprintf("• %s\n", suggestion)
		}
		output += "\n" + suggestionsStr
	}

	return output + "\n"
}

// GetErrorSuggestions returns context-aware suggestions for common errors
func GetErrorSuggestions(errMsg string) []string {
	errLower := strings.ToLower(errMsg)
	var suggestions []string

	// File not found
	if strings.Contains(errLower, "not found") || strings.Contains(errLower, "no such file") {
		suggestions = append(suggestions, "Check the file path spelling and capitalization")
		suggestions = append(suggestions, "Verify the file exists in the specified directory")
		suggestions = append(suggestions, "Try using '/list' to explore available files")
	}

	// Permission denied
	if strings.Contains(errLower, "permission denied") || strings.Contains(errLower, "access denied") {
		suggestions = append(suggestions, "Check if you have read/write permissions for the file")
		suggestions = append(suggestions, "Try changing the file permissions or location")
	}

	// Network/connection errors
	if strings.Contains(errLower, "connection") || strings.Contains(errLower, "timeout") {
		suggestions = append(suggestions, "Check your internet connection")
		suggestions = append(suggestions, "Verify the API key is valid and not rate-limited")
		suggestions = append(suggestions, "Try again in a few moments if rate-limited")
	}

	// Tool/command errors
	if strings.Contains(errLower, "tool") || strings.Contains(errLower, "command") {
		suggestions = append(suggestions, "Verify the tool/command is installed and available")
		suggestions = append(suggestions, "Check the tool arguments and syntax")
		suggestions = append(suggestions, "Run '/tools list' to see available tools")
	}

	// Generic fallback suggestions
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Review the error message for clues")
		suggestions = append(suggestions, "Try a different approach or tool")
		suggestions = append(suggestions, "Type '/help' for available commands")
	}

	// Limit to 3 suggestions for readability
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	return suggestions
}
