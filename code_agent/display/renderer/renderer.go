package renderer

import (
	"fmt"
	"strings"

	"code_agent/display/components"
	"code_agent/display/formatters"
	"code_agent/display/styles"
	"code_agent/display/terminal"

	"github.com/charmbracelet/lipgloss"
	"google.golang.org/genai"
)

// Renderer is the main facade for all display operations
// It delegates to specialized formatters while maintaining backward compatibility
type Renderer struct {
	mdRenderer       *MarkdownRenderer
	outputFormat     string
	styleFormatter   *styles.Formatter
	toolFormatter    *formatters.ToolFormatter
	agentFormatter   *formatters.AgentFormatter
	errorFormatter   *formatters.ErrorFormatter
	metricsFormatter *formatters.MetricsFormatter
}

// NewRenderer creates a new renderer with the specified output format
func NewRenderer(outputFormat string) (*Renderer, error) {
	// Create markdown renderer
	mdRenderer, err := NewMarkdownRenderer()
	if err != nil {
		// Non-fatal: we can fall back to plain text
		mdRenderer = nil
	}

	// Initialize styles module with IsTTY function
	styles.SetTTYCheck(terminal.IsTTY)

	// Create styles and formatters
	s := styles.NewStyles()
	styleFormatter := styles.NewFormatter(outputFormat, s)

	// Create specialized formatters
	toolFormatter := formatters.NewToolFormatter(outputFormat, s, styleFormatter)
	agentFormatter := formatters.NewAgentFormatter(outputFormat, s, styleFormatter, mdRenderer)
	errorFormatter := formatters.NewErrorFormatter(outputFormat, s, styleFormatter, mdRenderer)
	metricsFormatter := formatters.NewMetricsFormatter(outputFormat, s, styleFormatter, mdRenderer)

	r := &Renderer{
		mdRenderer:       mdRenderer,
		outputFormat:     outputFormat,
		styleFormatter:   styleFormatter,
		toolFormatter:    toolFormatter,
		agentFormatter:   agentFormatter,
		errorFormatter:   errorFormatter,
		metricsFormatter: metricsFormatter,
	}

	return r, nil
}

// RenderMarkdown renders markdown text according to output format.
func (r *Renderer) RenderMarkdown(markdown string) string {
	// Skip markdown rendering if plain mode or not in TTY
	if r.outputFormat == styles.OutputFormatPlain || !terminal.IsTTY() {
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

// Style helper methods - delegate to styleFormatter
func (r *Renderer) Dim(text string) string {
	return r.styleFormatter.Dim(text)
}

func (r *Renderer) Green(text string) string {
	return r.styleFormatter.Green(text)
}

func (r *Renderer) Red(text string) string {
	return r.styleFormatter.Red(text)
}

func (r *Renderer) Yellow(text string) string {
	return r.styleFormatter.Yellow(text)
}

func (r *Renderer) Blue(text string) string {
	return r.styleFormatter.Blue(text)
}

func (r *Renderer) Cyan(text string) string {
	return r.styleFormatter.Cyan(text)
}

func (r *Renderer) Bold(text string) string {
	return r.styleFormatter.Bold(text)
}

func (r *Renderer) Success(text string) string {
	return r.styleFormatter.Success(text)
}

func (r *Renderer) SuccessCheckmark(text string) string {
	return r.styleFormatter.SuccessCheckmark(text)
}

func (r *Renderer) ErrorX(text string) string {
	return r.styleFormatter.ErrorX(text)
}

// RenderBanner renders a session banner with version and context info
func (r *Renderer) RenderBanner(version, model, workdir string) string {
	return components.RenderBanner(version, model, workdir)
}

// Tool formatting methods - delegate to toolFormatter
func (r *Renderer) RenderToolCall(toolName string, args map[string]any) string {
	return r.toolFormatter.RenderToolCall(toolName, args)
}

func (r *Renderer) RenderToolResult(toolName string, result map[string]any) string {
	return r.toolFormatter.RenderToolResult(toolName, result)
}

// Agent formatting methods - delegate to agentFormatter
func (r *Renderer) RenderAgentThinking() string {
	return r.agentFormatter.RenderAgentThinking()
}

func (r *Renderer) RenderAgentWorking(action string) string {
	return r.agentFormatter.RenderAgentWorking(action)
}

func (r *Renderer) RenderAgentResponse(text string) string {
	return r.agentFormatter.RenderAgentResponse(text)
}

// Error formatting methods - delegate to errorFormatter
func (r *Renderer) RenderWarning(message string) string {
	return r.errorFormatter.RenderWarning(message)
}

func (r *Renderer) RenderInfo(message string) string {
	return r.errorFormatter.RenderInfo(message)
}

func (r *Renderer) RenderError(err error) string {
	return r.errorFormatter.RenderError(err)
}

// Metrics formatting methods - delegate to metricsFormatter
func (r *Renderer) RenderTaskComplete() string {
	return r.metricsFormatter.RenderTaskComplete()
}

func (r *Renderer) RenderTaskFailed() string {
	return r.metricsFormatter.RenderTaskFailed()
}

func (r *Renderer) RenderTokenMetrics(promptTokens, cachedTokens, responseTokens, totalTokens int64) string {
	return r.metricsFormatter.RenderTokenMetrics(promptTokens, cachedTokens, responseTokens, totalTokens)
}

func (r *Renderer) RenderAPIUsage(status string, usage *formatters.APIUsageInfo) string {
	return r.metricsFormatter.RenderAPIUsage(status, usage)
}

// RenderBoxedOutput renders output with a subtle border for emphasis
func (r *Renderer) RenderBoxedOutput(title string, content string) string {
	if r.outputFormat == styles.OutputFormatPlain || !terminal.IsTTY() {
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
	width := terminal.GetTerminalWidth()
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

// MarkdownRenderer returns the underlying MarkdownRenderer instance for advanced use.
func (r *Renderer) MarkdownRenderer() *MarkdownRenderer {
	return r.mdRenderer
}

// OutputFormat returns the configured output format for the renderer.
func (r *Renderer) OutputFormat() string {
	return r.outputFormat
}
