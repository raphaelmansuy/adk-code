package display

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BannerRenderer provides specialized rendering for session banners and separators.
type BannerRenderer struct {
	renderer *Renderer
}

// NewBannerRenderer creates a new banner renderer.
func NewBannerRenderer(renderer *Renderer) *BannerRenderer {
	return &BannerRenderer{
		renderer: renderer,
	}
}

// RenderStartBanner renders the session start banner.
func (br *BannerRenderer) RenderStartBanner(version, model, workdir string) string {
	if br.renderer.outputFormat == OutputFormatPlain || !IsTTY() {
		// Plain text banner
		var lines []string
		lines = append(lines, "===========================================")
		lines = append(lines, fmt.Sprintf("code_agent %s", version))
		if model != "" {
			lines = append(lines, fmt.Sprintf("Model: %s", model))
		}
		if workdir != "" {
			lines = append(lines, fmt.Sprintf("Working directory: %s", workdir))
		}
		lines = append(lines, "===========================================")
		return strings.Join(lines, "\n") + "\n\n"
	}

	// Rich banner using the renderer's built-in method
	banner := br.renderer.RenderBanner(version, model, workdir)
	return "\n" + banner + "\n\n"
}

// RenderSessionInfo renders session information.
func (br *BannerRenderer) RenderSessionInfo(sessionID, startTime string, toolCount int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "248", Dark: "238"})

	var lines []string
	lines = append(lines, titleStyle.Render("Session Information"))
	lines = append(lines, "")
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Session ID: %s", sessionID)))
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Started: %s", startTime)))
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Available tools: %d", toolCount)))
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Runtime: %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)))

	return strings.Join(lines, "\n") + "\n"
}

// RenderSeparator renders a horizontal separator line.
func (br *BannerRenderer) RenderSeparator(char string, width int) string {
	if width <= 0 {
		width = GetTerminalWidth()
		if width <= 0 {
			width = 80
		}
	}

	if br.renderer.outputFormat == OutputFormatPlain || !IsTTY() {
		return strings.Repeat(char, width) + "\n"
	}

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "248", Dark: "238"})

	return dimStyle.Render(strings.Repeat(char, width)) + "\n"
}

// RenderThickSeparator renders a thick separator.
func (br *BannerRenderer) RenderThickSeparator() string {
	return br.RenderSeparator("═", GetTerminalWidth())
}

// RenderThinSeparator renders a thin separator.
func (br *BannerRenderer) RenderThinSeparator() string {
	return br.RenderSeparator("─", GetTerminalWidth())
}

// RenderSection renders a section header.
func (br *BannerRenderer) RenderSection(title string) string {
	if br.renderer.outputFormat == OutputFormatPlain || !IsTTY() {
		return fmt.Sprintf("\n=== %s ===\n", title)
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")). // Blue
		Bold(true)

	return "\n" + titleStyle.Render("▸ "+title) + "\n"
}

// RenderBox renders text in a box.
func (br *BannerRenderer) RenderBox(title, content string) string {
	if br.renderer.outputFormat == OutputFormatPlain || !IsTTY() {
		var lines []string
		lines = append(lines, fmt.Sprintf("┌─ %s ─┐", title))
		for _, line := range strings.Split(content, "\n") {
			lines = append(lines, "│ "+line)
		}
		lines = append(lines, "└─────────┘")
		return strings.Join(lines, "\n") + "\n"
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")). // Blue
		Bold(true)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	contentWithTitle := titleStyle.Render(title) + "\n\n" + content
	return "\n" + boxStyle.Render(contentWithTitle) + "\n"
}

// RenderCompletionBanner renders a task completion banner.
func (br *BannerRenderer) RenderCompletionBanner(success bool, duration string, message string) string {
	if br.renderer.outputFormat == OutputFormatPlain || !IsTTY() {
		status := "SUCCESS"
		if !success {
			status = "FAILED"
		}
		return fmt.Sprintf("\n=== Task %s (%s) ===\n%s\n", status, duration, message)
	}

	var statusText, statusColor string
	if success {
		statusText = "✓ SUCCESS"
		statusColor = "2" // Green
	} else {
		statusText = "✗ FAILED"
		statusColor = "1" // Red
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Bold(true)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "248", Dark: "238"})

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(statusColor)).
		Padding(1, 2).
		Width(60)

	var lines []string
	lines = append(lines, statusStyle.Render(statusText))
	lines = append(lines, "")
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Duration: %s", duration)))
	if message != "" {
		lines = append(lines, "")
		lines = append(lines, message)
	}

	content := strings.Join(lines, "\n")
	return "\n" + boxStyle.Render(content) + "\n\n"
}

// RenderWelcome renders a welcome message with usage hints.
func (br *BannerRenderer) RenderWelcome() string {
	if br.renderer.outputFormat == OutputFormatPlain || !IsTTY() {
		return "\nReady! Type your request or 'exit' to quit.\n\n"
	}

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "245"})

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")). // Blue
		Bold(true)

	var lines []string
	lines = append(lines, dimStyle.Render("Ready to assist with your coding tasks."))
	lines = append(lines, dimStyle.Render("Type 'exit' or press Ctrl+C to quit."))
	lines = append(lines, "")
	lines = append(lines, promptStyle.Render("What would you like me to help you with?"))
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

// RenderError renders an error banner.
func (br *BannerRenderer) RenderError(title, message string) string {
	if br.renderer.outputFormat == OutputFormatPlain || !IsTTY() {
		return fmt.Sprintf("\n!!! ERROR: %s !!!\n%s\n", title, message)
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")). // Red
		Bold(true)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		Padding(1, 2)

	contentWithTitle := titleStyle.Render("✗ "+title) + "\n\n" + message
	return "\n" + boxStyle.Render(contentWithTitle) + "\n"
}

// RenderWarning renders a warning banner.
func (br *BannerRenderer) RenderWarning(title, message string) string {
	if br.renderer.outputFormat == OutputFormatPlain || !IsTTY() {
		return fmt.Sprintf("\n!!! WARNING: %s !!!\n%s\n", title, message)
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")). // Yellow
		Bold(true)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("3")).
		Padding(1, 2)

	contentWithTitle := titleStyle.Render("⚠ "+title) + "\n\n" + message
	return "\n" + boxStyle.Render(contentWithTitle) + "\n"
}

// RenderInfo renders an info banner.
func (br *BannerRenderer) RenderInfo(title, message string) string {
	if br.renderer.outputFormat == OutputFormatPlain || !IsTTY() {
		return fmt.Sprintf("\n--- %s ---\n%s\n", title, message)
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")). // Blue
		Bold(true)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	contentWithTitle := titleStyle.Render("ℹ "+title) + "\n\n" + message
	return "\n" + boxStyle.Render(contentWithTitle) + "\n"
}
