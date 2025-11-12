package components

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderBanner renders the application banner with version, model, and working directory
func RenderBanner(version, model, workdir string) string {
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
		shortPath := ShortenPath(workdir, 45)
		lines = append(lines, dimStyle.Render(shortPath))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Render(content)
}

// ShortenPath shortens a filesystem path to fit within maxLen
func ShortenPath(path string, maxLen int) string {
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
