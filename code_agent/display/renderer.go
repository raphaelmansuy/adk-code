package display

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"google.golang.org/genai"
)

// Renderer is the main facade for all display operations.
type Renderer struct {
	mdRenderer   *MarkdownRenderer
	outputFormat string

	// Lipgloss styles
	dimStyle     lipgloss.Style
	greenStyle   lipgloss.Style
	redStyle     lipgloss.Style
	yellowStyle  lipgloss.Style
	blueStyle    lipgloss.Style
	cyanStyle    lipgloss.Style
	whiteStyle   lipgloss.Style
	boldStyle    lipgloss.Style
	successStyle lipgloss.Style
}

// OutputFormat constants
const (
	OutputFormatRich  = "rich"
	OutputFormatPlain = "plain"
	OutputFormatJSON  = "json"
)

// EventType represents different types of events in the agent execution
type EventType string

// Event types
const (
	EventTypeThinking  EventType = "thinking"
	EventTypeExecuting EventType = "executing"
	EventTypeResult    EventType = "result"
	EventTypeSuccess   EventType = "success"
	EventTypeWarning   EventType = "warning"
	EventTypeError     EventType = "error"
	EventTypeProgress  EventType = "progress"
)

// EventTypeIcon returns the emoji icon for an event type
func EventTypeIcon(eventType EventType) string {
	switch eventType {
	case EventTypeThinking:
		return "🧠"
	case EventTypeExecuting:
		return "🔧"
	case EventTypeResult:
		return "📊"
	case EventTypeSuccess:
		return "✓"
	case EventTypeWarning:
		return "⚠️"
	case EventTypeError:
		return "❌"
	case EventTypeProgress:
		return "📍"
	default:
		return "•"
	}
}

// TimelineEvent represents a single event in the operation timeline
type TimelineEvent struct {
	ToolName string
	Status   string // "pending", "executing", "completed", "failed"
}

// EventTimeline tracks a sequence of operations
type EventTimeline struct {
	events []TimelineEvent
}

// NewEventTimeline creates a new event timeline
func NewEventTimeline() *EventTimeline {
	return &EventTimeline{
		events: make([]TimelineEvent, 0),
	}
}

// AppendEvent adds an operation to the timeline
func (et *EventTimeline) AppendEvent(toolName, status string) {
	et.events = append(et.events, TimelineEvent{
		ToolName: toolName,
		Status:   status,
	})
}

// RenderTimeline returns a formatted timeline string
func (et *EventTimeline) RenderTimeline() string {
	if len(et.events) == 0 {
		return ""
	}

	// Build timeline string like: [read_file] → [grep_search] → [write_file]
	var parts []string
	for i, event := range et.events {
		// Use short tool name (last part after underscore)
		shortName := strings.TrimPrefix(event.ToolName, "list_")
		shortName = strings.TrimPrefix(shortName, "search_")
		shortName = strings.TrimPrefix(shortName, "read_")
		shortName = strings.TrimPrefix(shortName, "write_")
		shortName = strings.TrimPrefix(shortName, "execute_")

		parts = append(parts, fmt.Sprintf("[%s]", shortName))

		// Add arrow between events (except after last)
		if i < len(et.events)-1 {
			parts = append(parts, "→")
		}
	}

	return "Timeline: " + strings.Join(parts, " ")
}

// GetEventCount returns the number of events in the timeline
func (et *EventTimeline) GetEventCount() int {
	return len(et.events)
}

// UpdateLastEventStatus updates the status of the most recent event
func (et *EventTimeline) UpdateLastEventStatus(status string) {
	if len(et.events) > 0 {
		et.events[len(et.events)-1].Status = status
	}
}

// RenderProgress returns a simple progress indicator
func (et *EventTimeline) RenderProgress() string {
	if len(et.events) == 0 {
		return ""
	}

	completed := 0
	for _, event := range et.events {
		if event.Status == "completed" {
			completed++
		}
	}

	total := len(et.events)
	percent := (completed * 100) / total

	// Simple progress bar: [████░░░░░░] 40% (2 of 5)
	barWidth := 10
	filledWidth := (completed * barWidth) / total
	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

	return fmt.Sprintf("Progress: [%s] %d%% (%d of %d operations)", bar, percent, completed, total)
}

// NewRenderer creates a new renderer with the specified output format.
func NewRenderer(outputFormat string) (*Renderer, error) {
	// Create markdown renderer
	mdRenderer, err := NewMarkdownRenderer()
	if err != nil {
		// Non-fatal: we can fall back to plain text
		mdRenderer = nil
	}

	r := &Renderer{
		mdRenderer:   mdRenderer,
		outputFormat: outputFormat,
	}

	// Initialize lipgloss styles (will respect the global color profile)
	r.dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))                // Bright black (gray)
	r.greenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))              // Green
	r.redStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))                // Red
	r.yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))             // Yellow
	r.blueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))              // Bright blue
	r.cyanStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))               // Cyan
	r.whiteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))              // White
	r.boldStyle = lipgloss.NewStyle().Bold(true)                                    // Bold
	r.successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true) // Green + Bold

	return r, nil
}

// RenderMarkdown renders markdown text according to output format.
func (r *Renderer) RenderMarkdown(markdown string) string {
	// Skip markdown rendering if plain mode or not in TTY
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return markdown
	}

	if r.mdRenderer == nil {
		return markdown
	}

	rendered, err := r.mdRenderer.Render(markdown)
	if err != nil {
		return markdown
	}

	return rendered
}

// RenderText renders plain text with optional styling.
func (r *Renderer) RenderText(text string) string {
	return text
}

// Style helper methods

// Dim renders text in dim gray.
func (r *Renderer) Dim(text string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return text
	}
	return r.dimStyle.Render(text)
}

// Green renders text in green.
func (r *Renderer) Green(text string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return text
	}
	return r.greenStyle.Render(text)
}

// Red renders text in red.
func (r *Renderer) Red(text string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return text
	}
	return r.redStyle.Render(text)
}

// Yellow renders text in yellow.
func (r *Renderer) Yellow(text string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return text
	}
	return r.yellowStyle.Render(text)
}

// Blue renders text in blue.
func (r *Renderer) Blue(text string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return text
	}
	return r.blueStyle.Render(text)
}

// Cyan renders text in cyan.
func (r *Renderer) Cyan(text string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return text
	}
	return r.cyanStyle.Render(text)
}

// Bold renders text in bold.
func (r *Renderer) Bold(text string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return text
	}
	return r.boldStyle.Render(text)
}

// Success renders text in green with bold.
func (r *Renderer) Success(text string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return text
	}
	return r.successStyle.Render(text)
}

// SuccessCheckmark renders a checkmark with text in green.
func (r *Renderer) SuccessCheckmark(text string) string {
	return r.Success("✓ " + text)
}

// ErrorX renders an X with text in red.
func (r *Renderer) ErrorX(text string) string {
	return r.Red("✗ " + text)
}

// RenderBanner renders a session banner with version and context info.
func (r *Renderer) RenderBanner(version, model, workdir string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // Bright white
		Bold(true)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "248", Dark: "238"})

	borderColor := lipgloss.Color("39") // Blue

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 4)

	var lines []string

	// Title line
	versionStr := version
	if len(versionStr) > 0 && versionStr[0] >= '0' && versionStr[0] <= '9' {
		versionStr = "v" + versionStr
	}
	lines = append(lines, titleStyle.Render("code_agent")+" "+dimStyle.Render(versionStr))

	// Model line
	if model != "" {
		lines = append(lines, dimStyle.Render(model))
	}

	// Workspace line
	if workdir != "" {
		// Shorten path if needed
		shortPath := shortenPath(workdir, 45)
		lines = append(lines, dimStyle.Render(shortPath))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// shortenPath shortens a filesystem path to fit within maxLen.
func shortenPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}

	// Try to replace home directory with ~
	if homeDir, err := os.UserHomeDir(); err == nil {
		if strings.HasPrefix(path, homeDir) {
			shortened := "~" + path[len(homeDir):]
			if len(shortened) <= maxLen {
				return shortened
			}
			path = shortened
		}
	}

	// If still too long, show last part
	if len(path) > maxLen {
		return "..." + path[len(path)-maxLen+3:]
	}

	return path
}

// RenderToolCall renders a tool call with contextual formatting.
func (r *Renderer) RenderToolCall(toolName string, args map[string]any) string {
	// Create contextual header based on tool
	header := r.getToolHeader(toolName, args)
	// Add spacing before tool call for better readability
	return "\n" + header + "\n"
}

// truncatePath smartly truncates long file paths for display.
// Shows filename + parent directory for long paths, preserving important context.
// Examples:
//
//	/very/long/path/to/project/src/main.go -> .../src/main.go
//	./main.go -> ./main.go
func (r *Renderer) truncatePath(path string, maxLength int) string {
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

// getToolHeader generates a contextual header for tool calls.
func (r *Renderer) getToolHeader(toolName string, args map[string]any) string {
	// Create a subtle tool icon
	toolIcon := "◆"
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
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
			displayPath := r.truncatePath(path, 60)
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Reading") + " " + r.Dim(displayPath)
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Reading file")

	case "write_file":
		if path, ok := args["path"].(string); ok {
			displayPath := r.truncatePath(path, 60)
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Writing") + " " + r.Dim(displayPath)
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Writing file")

	case "replace_in_file", "search_replace":
		if path, ok := args["path"].(string); ok {
			displayPath := r.truncatePath(path, 60)
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Editing") + " " + r.Dim(displayPath)
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Editing file")

	case "list_directory":
		if path, ok := args["path"].(string); ok {
			displayPath := r.truncatePath(path, 60)
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Listing") + " " + r.Dim(displayPath)
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Listing files")

	case "execute_command", "execute_program":
		if command, ok := args["command"].(string); ok {
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Running") + " " + r.Dim("`"+command+"`")
		}
		if program, ok := args["program"].(string); ok {
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Running") + " " + r.Dim("`"+program+"`")
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Running command")

	case "grep_search":
		if pattern, ok := args["pattern"].(string); ok {
			return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Searching for") + " " + r.Dim("`"+pattern+"`")
		}
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render("Searching files")

	default:
		return iconStyle.Render(toolIcon) + " " + toolStyle.Render(toolName)
	}
}

// RenderToolResult renders a tool result with contextual information.
func (r *Renderer) RenderToolResult(toolName string, result map[string]any) string {
	// Check for errors
	if errStr, ok := result["error"].(string); ok && errStr != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")). // Red
			Bold(false)
		return "  " + errorStyle.Render("✗ "+errStr) + "\n"
	}

	// Subtle success indicator
	checkmark := "✓"
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
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
			displayPath := r.truncatePath(path, 50)
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

// RenderAgentThinking renders the "agent is thinking" message.
func (r *Renderer) RenderAgentThinking() string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return "\nThinking...\n"
	}

	thinkingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "245"}).
		Italic(true)

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")) // Blue

	icon := "◉"
	message := iconStyle.Render(icon) + " " + thinkingStyle.Render("Thinking...")

	return "\n" + message + "\n"
}

// RenderAgentWorking renders an explicit "working" message for when the model is processing.
func (r *Renderer) RenderAgentWorking(action string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return fmt.Sprintf("\n%s...\n", action)
	}

	workingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "245"}).
		Italic(true)

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")) // Blue

	icon := "◉"
	message := iconStyle.Render(icon) + " " + workingStyle.Render(action+"...")

	return "\n" + message + "\n"
}

// RenderAgentResponse renders an agent's text response.
func (r *Renderer) RenderAgentResponse(text string) string {
	// Agent responses are typically markdown
	rendered := r.RenderMarkdown(text)

	// Add subtle left border and indentation for better readability
	if r.outputFormat != OutputFormatPlain && IsTTY() {
		borderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "236"})

		lines := strings.Split(rendered, "\n")
		var styledLines []string
		for i, line := range lines {
			if line != "" {
				// Add subtle border character and indentation
				border := borderStyle.Render("│")
				styledLines = append(styledLines, border+" "+line)
			} else {
				// Empty lines get just the border
				if i < len(lines)-1 { // Don't add border to trailing empty lines
					styledLines = append(styledLines, borderStyle.Render("│"))
				} else {
					styledLines = append(styledLines, line)
				}
			}
		}
		rendered = strings.Join(styledLines, "\n")
	}

	return rendered + "\n"
}

// RenderWarning renders a warning message with subtle styling.
func (r *Renderer) RenderWarning(message string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
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

// RenderInfo renders an informational message with subtle styling.
func (r *Renderer) RenderInfo(message string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
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

// RenderError renders an error message with actionable suggestions.
func (r *Renderer) RenderError(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Detect error type and provide suggestions
	suggestions := getErrorSuggestions(errMsg)

	// Format the main error
	markdown := fmt.Sprintf("### %s\n\n%s", r.ErrorX("Error"), errMsg)
	rendered := r.RenderMarkdown(markdown)

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

// getErrorSuggestions returns context-aware suggestions for common errors
func getErrorSuggestions(errMsg string) []string {
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

// RenderTaskComplete renders the task completion message.
func (r *Renderer) RenderTaskComplete() string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return "\nDone.\n\n"
	}

	// Add subtle success indicator before separator
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")). // Green
		Bold(false)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"})

	checkmark := successStyle.Render("✓") + " " + dimStyle.Render("Complete")

	// Use a shorter, centered separator
	width := GetTerminalWidth()
	if width > 100 {
		width = 100 // Cap at 100 chars
	}

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "252", Dark: "240"})

	separator := separatorStyle.Render(strings.Repeat("─", width))
	return "\n" + checkmark + "\n" + separator + "\n\n"
}

// RenderTaskFailed renders the task failure message.
func (r *Renderer) RenderTaskFailed() string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		return "\nFailed.\n\n"
	}

	return "\n" + r.ErrorX("Task failed") + "\n\n"
}

// RenderTokenMetrics renders compact token usage metrics for display
func (r *Renderer) RenderTokenMetrics(promptTokens, cachedTokens, responseTokens, totalTokens int64) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() || totalTokens == 0 {
		return ""
	}

	// Use a muted color for metrics
	metricStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}).
		Italic(true)

	// Build metrics string: "Tokens: 2,341 prompt | 892 cached | 1,205 response | Total: 5,054"
	var parts []string

	if promptTokens > 0 {
		parts = append(parts, fmt.Sprintf("%d prompt", promptTokens))
	}
	if cachedTokens > 0 {
		parts = append(parts, fmt.Sprintf("%d cached", cachedTokens))
	}
	if responseTokens > 0 {
		parts = append(parts, fmt.Sprintf("%d response", responseTokens))
	}

	metricsStr := fmt.Sprintf("Tokens: %s | Total: %d", strings.Join(parts, " | "), totalTokens)

	return metricStyle.Render(metricsStr)
}

// APIUsageInfo holds token usage and cost information
type APIUsageInfo struct {
	TokensIn    int
	TokensOut   int
	CacheReads  int
	CacheWrites int
	Cost        float64
}

// formatNumber formats numbers with k/m abbreviations
func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fm", float64(n)/1000000.0)
	} else if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%d", n)
}

// RenderAPIUsage renders API usage information
func (r *Renderer) RenderAPIUsage(status string, usage *APIUsageInfo) string {
	if usage == nil || usage.Cost < 0 {
		return ""
	}

	parts := make([]string, 0, 4)

	if usage.TokensIn > 0 {
		parts = append(parts, fmt.Sprintf("↑ %s", formatNumber(usage.TokensIn)))
	}
	if usage.TokensOut > 0 {
		parts = append(parts, fmt.Sprintf("↓ %s", formatNumber(usage.TokensOut)))
	}
	if usage.CacheReads > 0 {
		parts = append(parts, fmt.Sprintf("→ %s", formatNumber(usage.CacheReads)))
	}
	if usage.CacheWrites > 0 {
		parts = append(parts, fmt.Sprintf("← %s", formatNumber(usage.CacheWrites)))
	}

	var usageInfo string
	if len(parts) > 0 {
		usageInfo = fmt.Sprintf("%s $%.4f", strings.Join(parts, " "), usage.Cost)
	} else {
		usageInfo = fmt.Sprintf("$%.4f", usage.Cost)
	}

	markdown := fmt.Sprintf("## API %s `%s`", status, usageInfo)
	rendered := r.RenderMarkdown(markdown)
	return "\n" + rendered + "\n"
}

// RenderBoxedOutput renders output with a subtle border for emphasis
func (r *Renderer) RenderBoxedOutput(title string, content string) string {
	if r.outputFormat == OutputFormatPlain || !IsTTY() {
		if title != "" {
			return fmt.Sprintf("\n%s:\n%s\n", title, content)
		}
		return fmt.Sprintf("\n%s\n", content)
	}

	// Use a subtle box style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).
		Padding(0, 1)

	// Limit content width
	width := GetTerminalWidth()
	if width > 80 {
		width = 80
	}

	// Truncate lines that are too long
	lines := strings.Split(strings.TrimSpace(content), "\n")
	for i, line := range lines {
		if len(line) > width-4 {
			lines[i] = line[:width-7] + "..."
		}
	}
	content = strings.Join(lines, "\n")

	boxedContent := boxStyle.Render(content)

	if title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true)
		return "\n" + titleStyle.Render(title) + "\n" + boxedContent + "\n"
	}

	return "\n" + boxedContent + "\n"
}

// RenderPartContent renders a content part from the agent.
func (r *Renderer) RenderPartContent(part *genai.Part) string {
	if part.Text != "" {
		// Check if it looks like tool-related text
		text := part.Text
		if strings.Contains(text, "read_file") || strings.Contains(text, "write_file") ||
			strings.Contains(text, "execute_command") || strings.Contains(text, "list_directory") {
			// It's tool-related, render it dimmed
			return r.Dim(text)
		}
		// Regular agent response
		return r.RenderAgentResponse(text)
	}
	return ""
}
