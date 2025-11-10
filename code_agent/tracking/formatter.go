package tracking

import (
	"fmt"
	"strings"
)

// FormatTokenMetrics returns a formatted string of token metrics for display.
func FormatTokenMetrics(metric TokenMetrics) string {
	var parts []string

	// Calculate actual tokens used (total - cached)
	usedTokens := metric.TotalTokens - metric.CachedTokens
	if usedTokens > 0 {
		parts = append(parts, fmt.Sprintf("↓used=%d", usedTokens))
	}

	if metric.PromptTokens > 0 {
		parts = append(parts, fmt.Sprintf("prompt=%d", metric.PromptTokens))
	}
	if metric.ResponseTokens > 0 {
		parts = append(parts, fmt.Sprintf("response=%d", metric.ResponseTokens))
	}
	if metric.CachedTokens > 0 {
		parts = append(parts, fmt.Sprintf("cached=%d", metric.CachedTokens))
	}
	if metric.ThoughtTokens > 0 {
		parts = append(parts, fmt.Sprintf("thoughts=%d", metric.ThoughtTokens))
	}
	if metric.ToolUseTokens > 0 {
		parts = append(parts, fmt.Sprintf("tool_use=%d", metric.ToolUseTokens))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("total=%d", metric.TotalTokens)
	}

	return fmt.Sprintf("[%s] (total=%d)", strings.Join(parts, ", "), metric.TotalTokens)
}

// FormatSessionSummary returns a formatted string summary of session tokens.
func FormatSessionSummary(summary *Summary) string {
	lines := []string{
		"\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		"📊 Token Usage Summary",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
	}

	// Calculate used tokens and cache efficiency
	usedTokens := summary.TotalTokens - summary.TotalCachedTokens
	var cacheEfficiency float64
	if summary.TotalTokens > 0 {
		cacheEfficiency = float64(summary.TotalCachedTokens) / float64(summary.TotalTokens) * 100
	}

	lines = append(lines, fmt.Sprintf("Total Tokens:        %d", summary.TotalTokens))
	lines = append(lines, fmt.Sprintf("  ├─ Actually Used:  %d", usedTokens))
	lines = append(lines, fmt.Sprintf("  ├─ Prompt:         %d", summary.TotalPromptTokens))
	lines = append(lines, fmt.Sprintf("  ├─ Response:       %d", summary.TotalResponseTokens))

	if summary.TotalCachedTokens > 0 {
		lines = append(lines, fmt.Sprintf("  ├─ Cached:         %d (%.1f%% saved)", summary.TotalCachedTokens, cacheEfficiency))
	}
	if summary.TotalThoughtTokens > 0 {
		lines = append(lines, fmt.Sprintf("  ├─ Thoughts:       %d", summary.TotalThoughtTokens))
	}
	if summary.TotalToolUseTokens > 0 {
		lines = append(lines, fmt.Sprintf("  └─ Tool Use:       %d", summary.TotalToolUseTokens))
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Requests:            %d", summary.RequestCount))
	lines = append(lines, fmt.Sprintf("Avg Tokens/Request:  %.1f", summary.AvgTokensPerRequest))
	lines = append(lines, fmt.Sprintf("Session Duration:    %s", formatDuration(summary.SessionDuration)))
	lines = append(lines, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return strings.Join(lines, "\n")
}

// FormatGlobalSummary returns a formatted string summary of global token usage.
func FormatGlobalSummary(summary *GlobalSummary) string {
	lines := []string{
		"\n📈 Global Token Usage Report",
		"",
		fmt.Sprintf("Total Tokens Across All Sessions: %d", summary.TotalTokens),
		fmt.Sprintf("Total Requests:                  %d", summary.TotalRequests),
		fmt.Sprintf("Average Tokens per Request:      %.1f", summary.AvgTokensPerRequest),
		"",
	}

	if len(summary.Sessions) > 0 {
		lines = append(lines, "Session Breakdown:")
		lines = append(lines, "")

		for sessionID, sessionSummary := range summary.Sessions {
			usedTokens := sessionSummary.TotalTokens - sessionSummary.TotalCachedTokens
			var cacheEff float64
			if sessionSummary.TotalTokens > 0 {
				cacheEff = float64(sessionSummary.TotalCachedTokens) / float64(sessionSummary.TotalTokens) * 100
			}

			if sessionSummary.TotalCachedTokens > 0 {
				lines = append(lines, fmt.Sprintf("  Session %s:", truncateSessionID(sessionID)))
				lines = append(lines, fmt.Sprintf("    Total: %d | Used: %d | Cached: %d (%.1f%%) | Requests: %d",
					sessionSummary.TotalTokens, usedTokens, sessionSummary.TotalCachedTokens, cacheEff, sessionSummary.RequestCount))
			} else {
				lines = append(lines, fmt.Sprintf("  Session %s:", truncateSessionID(sessionID)))
				lines = append(lines, fmt.Sprintf("    Total: %d | Used: %d | Requests: %d",
					sessionSummary.TotalTokens, usedTokens, sessionSummary.RequestCount))
			}
		}
	}

	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

// FormatRequestMetrics returns a formatted inline metric string for quick reference.
func FormatRequestMetrics(metric TokenMetrics) string {
	if metric.TotalTokens == 0 {
		return ""
	}
	return fmt.Sprintf("(tokens: %d)", metric.TotalTokens)
}

// Helper functions

func formatDuration(d interface{}) string {
	// Handle different input types
	switch v := d.(type) {
	case int64:
		if v < 1000000000 { // Less than 1 second in nanoseconds
			return fmt.Sprintf("%dms", v/1000000)
		}
		seconds := v / 1000000000
		if seconds < 60 {
			return fmt.Sprintf("%ds", seconds)
		}
		minutes := seconds / 60
		secs := seconds % 60
		return fmt.Sprintf("%dm%ds", minutes, secs)
	default:
		return fmt.Sprintf("%v", d)
	}
}

func truncateSessionID(id string) string {
	if len(id) > 8 {
		return id[:8] + "..."
	}
	return id
}
