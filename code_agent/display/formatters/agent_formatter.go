package formatters

import (
	"fmt"
	"strings"

	"code_agent/display/styles"

	"github.com/charmbracelet/lipgloss"
)

// AgentFormatter formats agent-related messages
type AgentFormatter struct {
	styles       *styles.Styles
	formatter    *styles.Formatter
	outputFormat string
	mdRenderer   MarkdownRenderer
}

// MarkdownRenderer interface for rendering markdown
type MarkdownRenderer interface {
	Render(string) (string, error)
}

// NewAgentFormatter creates a new agent formatter
func NewAgentFormatter(outputFormat string, s *styles.Styles, f *styles.Formatter, mdRenderer MarkdownRenderer) *AgentFormatter {
	return &AgentFormatter{
		styles:       s,
		formatter:    f,
		outputFormat: outputFormat,
		mdRenderer:   mdRenderer,
	}
}

// RenderAgentThinking renders the "agent is thinking" message
func (af *AgentFormatter) RenderAgentThinking() string {
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if af.outputFormat == styles.OutputFormatPlain || !isTTY {
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

// RenderAgentWorking renders an explicit "working" message for when the model is processing
func (af *AgentFormatter) RenderAgentWorking(action string) string {
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if af.outputFormat == styles.OutputFormatPlain || !isTTY {
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

// RenderAgentResponse renders an agent's text response
func (af *AgentFormatter) RenderAgentResponse(text string) string {
	// Agent responses are typically markdown
	rendered := text
	isTTY := styles.IsTTY != nil && styles.IsTTY()

	// Try to render markdown if we have a renderer
	if af.mdRenderer != nil && af.outputFormat != styles.OutputFormatPlain && isTTY {
		if mdRendered, err := af.mdRenderer.Render(text); err == nil {
			rendered = mdRendered
		}
	}

	// Add subtle left border and indentation for better readability
	if af.outputFormat != styles.OutputFormatPlain && isTTY {
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
