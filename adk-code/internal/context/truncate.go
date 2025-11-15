package context

import (
	"fmt"
	"strings"
)

// truncateHeadTail keeps beginning and end, omits verbose middle
// Result format:
// [First N lines of content]
// [... omitted X of Y lines ...]
// [Last N lines of content]
func truncateHeadTail(
	content string,
	maxLines int,
	headLines int,
	tailLines int,
	maxBytes int,
) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// If already under limits, return as-is
	if len(content) <= maxBytes && totalLines <= maxLines {
		return content
	}

	// Take head and tail segments
	headSegment := take(lines, headLines)
	tailSegment := takeLast(lines, tailLines)

	omittedLines := totalLines - len(headSegment) - len(tailSegment)
	if omittedLines < 0 {
		omittedLines = 0
	}

	// Build result with elision marker
	head := strings.Join(headSegment, "\n")
	tail := strings.Join(tailSegment, "\n")

	var result string
	if omittedLines > 0 {
		marker := fmt.Sprintf(
			"\n[... omitted %d of %d lines ...]\n\n",
			omittedLines, totalLines,
		)
		result = head + marker + tail
	} else {
		result = head + "\n" + tail
	}

	// If still over byte limit, truncate from end
	if len(result) > maxBytes {
		result = result[:maxBytes] + "\n[... truncated for length ...]"
	}

	return result
}

func take(lines []string, n int) []string {
	if n > len(lines) {
		n = len(lines)
	}
	return lines[:n]
}

func takeLast(lines []string, n int) []string {
	if n > len(lines) {
		n = len(lines)
	}
	return lines[len(lines)-n:]
}

// FormatOutputForModel formats output with line count information
func FormatOutputForModel(content string, totalLines int) string {
	return fmt.Sprintf("Total output lines: %d\n\n%s", totalLines, content)
}
