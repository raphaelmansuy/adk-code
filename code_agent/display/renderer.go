package display

import (
	"fmt"
	"os"
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
	r.dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))           // Bright black (gray)
	r.greenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))         // Green
	r.redStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))           // Red
	r.yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))        // Yellow
	r.blueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))         // Bright blue
	r.cyanStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))          // Cyan
	r.whiteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))         // White
	r.boldStyle = lipgloss.NewStyle().Bold(true)                               // Bold
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
	var output strings.Builder

	// Create contextual header based on tool
	header := r.getToolHeader(toolName, args)
	rendered := r.RenderMarkdown(header)
	output.WriteString("\n")
	output.WriteString(rendered)
	output.WriteString("\n")

	return output.String()
}

// getToolHeader generates a contextual header for tool calls.
func (r *Renderer) getToolHeader(toolName string, args map[string]any) string {
	switch toolName {
	case "read_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("### Agent is reading `%s`", path)
		}
		return "### Agent is reading a file"

	case "write_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("### Agent is writing `%s`", path)
		}
		return "### Agent is writing a file"

	case "replace_in_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("### Agent is editing `%s`", path)
		}
		return "### Agent is editing a file"

	case "list_directory":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("### Agent is listing files in `%s`", path)
		}
		return "### Agent is listing files"

	case "execute_command":
		if command, ok := args["command"].(string); ok {
			return fmt.Sprintf("### Agent is running command\n\n```shell\n%s\n```", command)
		}
		return "### Agent is running a command"

	case "grep_search":
		if pattern, ok := args["pattern"].(string); ok {
			return fmt.Sprintf("### Agent is searching for `%s`", pattern)
		}
		return "### Agent is searching files"

	default:
		return fmt.Sprintf("### Agent is using tool: %s", toolName)
	}
}

// RenderToolResult renders a tool result.
func (r *Renderer) RenderToolResult(toolName string, result map[string]any) string {
	// For now, just indicate success
	return r.Dim("  ✓ Completed\n")
}

// RenderAgentThinking renders the "agent is thinking" message.
func (r *Renderer) RenderAgentThinking() string {
	header := "### Agent is thinking"
	rendered := r.RenderMarkdown(header)
	return "\n" + rendered + "\n\n"
}

// RenderAgentResponse renders an agent's text response.
func (r *Renderer) RenderAgentResponse(text string) string {
	// Agent responses are typically markdown
	rendered := r.RenderMarkdown(text)
	return rendered + "\n"
}

// RenderError renders an error message.
func (r *Renderer) RenderError(err error) string {
	markdown := fmt.Sprintf("### %s\n\n%s", r.ErrorX("Error"), err.Error())
	rendered := r.RenderMarkdown(markdown)
	return "\n" + rendered + "\n"
}

// RenderTaskComplete renders the task completion message.
func (r *Renderer) RenderTaskComplete() string {
	return "\n" + r.SuccessCheckmark("Task completed") + "\n\n"
}

// RenderTaskFailed renders the task failure message.
func (r *Renderer) RenderTaskFailed() string {
	return "\n" + r.ErrorX("Task failed") + "\n\n"
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
