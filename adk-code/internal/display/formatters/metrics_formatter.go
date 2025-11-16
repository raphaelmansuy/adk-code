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
// contextWindow is in tokens, or -1 if unknown/not applicable
func (mf *MetricsFormatter) RenderTokenMetrics(promptTokens, cachedTokens, responseTokens, thoughtTokens, totalTokens, contextWindow int64) string {
	isTTY := styles.IsTTY != nil && styles.IsTTY()
	if mf.outputFormat == styles.OutputFormatPlain || !isTTY || totalTokens == 0 {
		return ""
	}

	// Use a muted color for metrics
	metricStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}).
		Italic(true)

	// Calculate meaningful metrics
	// Note: promptTokens from tracker includes cached portion (from Gemini API PromptTokenCount)
	// So we need to subtract cached to get truly new tokens
	newPromptTokens := promptTokens - cachedTokens // New prompt tokens (excluding cached)
	actualTokensUsed := newPromptTokens + responseTokens // New tokens actually processed (what you pay for)
	cacheHitTokens := cachedTokens                      // Tokens served from cache

	// Calculate cache efficiency: percentage of INPUT that was cached
	// (response tokens don't apply to caching, only input does)
	var cacheEfficiency float64
	if promptTokens > 0 {
		cacheEfficiency = (float64(cacheHitTokens) / float64(promptTokens)) * 100
	}

	// Determine cache efficiency indicator
	cacheIndicator := ""
	switch {
	case cacheEfficiency >= 80:
		cacheIndicator = "🚀 excellent"
	case cacheEfficiency >= 50:
		cacheIndicator = "✅ good"
	case cacheEfficiency >= 20:
		cacheIndicator = "⚠️ modest"
	default:
		cacheIndicator = "❌ minimal"
	}

	// Build metrics string with meaningful insights
	// Format: "Session: new:29K tok | cached:26K tok (92% excellent) | context:28K/1M tok (3% ✅ healthy)"
	var parts []string

	// Show new tokens used (cost to the user) - make it clear these are tokens
	if actualTokensUsed > 0 {
		parts = append(parts, fmt.Sprintf("new:%s tok", formatCompactNumber(actualTokensUsed)))
	}

	// Show cache reuse efficiency - make it clear these are tokens
	if cacheHitTokens > 0 {
		parts = append(parts, fmt.Sprintf("cached:%s tok (%.0f%% %s)", formatCompactNumber(cacheHitTokens), cacheEfficiency, cacheIndicator))
	}

	// Show response size only if significant - make it clear these are tokens
	if responseTokens > 0 {
		parts = append(parts, fmt.Sprintf("response:%s tok", formatCompactNumber(responseTokens)))
	}

	// Add session total with context window utilization
	// totalTokens includes ALL tokens: new + cached + thoughts + tool use
	if contextWindow > 0 {
		contextUsagePercent := (float64(totalTokens) / float64(contextWindow)) * 100
		contextIndicator := getContextWindowIndicator(contextUsagePercent)
		
		// Show thought tokens if they're a significant portion (>10% of total)
		thoughtNote := ""
		if thoughtTokens > 0 && float64(thoughtTokens)/float64(totalTokens) > 0.1 {
			thoughtNote = fmt.Sprintf(" incl. %s thoughts", formatCompactNumber(thoughtTokens))
		}
		
		parts = append(parts, fmt.Sprintf("session:%s/%s tok (%.1f%% %s%s)", formatCompactNumber(totalTokens), formatCompactNumber(contextWindow), contextUsagePercent, contextIndicator, thoughtNote))
	}

	metricsStr := fmt.Sprintf("Session: %s", strings.Join(parts, " | "))

	return metricStyle.Render(metricsStr)
}

// formatCompactNumber converts large numbers to compact form (e.g., 28029 -> 28K)
func formatCompactNumber(n int64) string {
	switch {
	case n >= 1000000:
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	case n >= 1000:
		return fmt.Sprintf("%.0fK", float64(n)/1000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// getContextWindowIndicator returns a visual indicator for context window usage
func getContextWindowIndicator(usagePercent float64) string {
	switch {
	case usagePercent < 10:
		return "✅ healthy"
	case usagePercent < 25:
		return "🟢 good"
	case usagePercent < 50:
		return "🟡 moderate"
	case usagePercent < 75:
		return "🟠 high"
	default:
		return "🔴 critical"
	}
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
