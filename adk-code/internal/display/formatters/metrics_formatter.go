package formatters

import (
	"fmt"
	"strings"

	"adk-code/internal/display/styles"

	"github.com/charmbracelet/lipgloss"
)

// MetricsFormatter formats metrics and API usage information
type MetricsFormatter struct {
	styles       *styles.Styles
	formatter    *styles.Formatter
	outputFormat string
	mdRenderer   MarkdownRenderer
}

// NewMetricsFormatter creates a new metrics formatter
func NewMetricsFormatter(outputFormat string, s *styles.Styles, f *styles.Formatter, mdRenderer MarkdownRenderer) *MetricsFormatter {
	return &MetricsFormatter{
		styles:       s,
		formatter:    f,
		outputFormat: outputFormat,
		mdRenderer:   mdRenderer,
	}
}

// RenderTokenMetrics renders compact token usage metrics for display
func (mf *MetricsFormatter) RenderTokenMetrics(promptTokens, cachedTokens, responseTokens, totalTokens int64) string {
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if mf.outputFormat == styles.OutputFormatPlain || !isTTY || totalTokens == 0 {
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

// FormatNumber formats numbers with k/m abbreviations
func FormatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fm", float64(n)/1000000.0)
	} else if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%d", n)
}

// RenderAPIUsage renders API usage information
func (mf *MetricsFormatter) RenderAPIUsage(status string, usage *APIUsageInfo) string {
	if usage == nil || usage.Cost < 0 {
		return ""
	}

	parts := make([]string, 0, 4)

	if usage.TokensIn > 0 {
		parts = append(parts, fmt.Sprintf("↑ %s", FormatNumber(usage.TokensIn)))
	}
	if usage.TokensOut > 0 {
		parts = append(parts, fmt.Sprintf("↓ %s", FormatNumber(usage.TokensOut)))
	}
	if usage.CacheReads > 0 {
		parts = append(parts, fmt.Sprintf("→ %s", FormatNumber(usage.CacheReads)))
	}
	if usage.CacheWrites > 0 {
		parts = append(parts, fmt.Sprintf("← %s", FormatNumber(usage.CacheWrites)))
	}

	var usageInfo string
	if len(parts) > 0 {
		usageInfo = fmt.Sprintf("%s $%.4f", strings.Join(parts, " "), usage.Cost)
	} else {
		usageInfo = fmt.Sprintf("$%.4f", usage.Cost)
	}

	markdown := fmt.Sprintf("## API %s `%s`", status, usageInfo)

	// Try to render as markdown
	rendered := markdown
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if mf.mdRenderer != nil && mf.outputFormat != styles.OutputFormatPlain && isTTY {
		if mdRendered, err := mf.mdRenderer.Render(markdown); err == nil {
			rendered = mdRendered
		}
	}

	return "\n" + rendered + "\n"
}

// RenderTaskComplete renders the task completion message
func (mf *MetricsFormatter) RenderTaskComplete() string {
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if mf.outputFormat == styles.OutputFormatPlain || !isTTY {
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
	// We'll hardcode a reasonable width since we can't import display package
	width := 80 // Default terminal width
	if width > 100 {
		width = 100 // Cap at 100 chars
	}

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "252", Dark: "240"})

	separator := separatorStyle.Render(strings.Repeat("─", width))
	return "\n" + checkmark + "\n" + separator + "\n\n"
}

// RenderTaskFailed renders the task failure message
func (mf *MetricsFormatter) RenderTaskFailed() string {
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if mf.outputFormat == styles.OutputFormatPlain || !isTTY {
		return "\nFailed.\n\n"
	}

	return "\n" + mf.formatter.ErrorX("Task failed") + "\n\n"
}
