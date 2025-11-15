package context

import (
	"strings"
	"testing"
)

func TestTruncateHeadTail_PreservesBeginningAndEnd(t *testing.T) {
	// Create content with 300 lines
	lines := make([]string, 300)
	for i := 0; i < 300; i++ {
		lines[i] = "Line " + string(rune('0'+i%10))
	}
	content := strings.Join(lines, "\n")

	// Truncate with head=50, tail=50, max=256 lines
	result := truncateHeadTail(content, 256, 50, 50, 100*1024)

	// Should contain first line
	if !strings.Contains(result, "Line 0") {
		t.Errorf("Expected result to contain first line")
	}

	// Should contain last line
	if !strings.Contains(result, lines[299]) {
		t.Errorf("Expected result to contain last line")
	}

	// Should contain elision marker
	if !strings.Contains(result, "omitted") {
		t.Errorf("Expected result to contain elision marker")
	}
}

func TestTruncateHeadTail_AddsElisionMarker(t *testing.T) {
	lines := make([]string, 300)
	for i := 0; i < 300; i++ {
		lines[i] = "test line"
	}
	content := strings.Join(lines, "\n")

	result := truncateHeadTail(content, 256, 128, 128, 100*1024)

	if !strings.Contains(result, "omitted") {
		t.Errorf("Expected elision marker")
	}
	if !strings.Contains(result, "of 300 lines") {
		t.Errorf("Expected total line count in marker")
	}
}

func TestTruncateHeadTail_RespectsByteLimit(t *testing.T) {
	// Create very large content
	lines := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		lines[i] = "This is a very long line with lots of text that should exceed the byte limit"
	}
	content := strings.Join(lines, "\n")

	maxBytes := 5000
	result := truncateHeadTail(content, 256, 128, 128, maxBytes)

	if len(result) > maxBytes+100 { // Allow small buffer for marker
		t.Errorf("Expected result to respect byte limit, got %d bytes", len(result))
	}
}

func TestTruncateHeadTail_NoTruncationNeeded(t *testing.T) {
	content := "Short content\nwith few lines"

	result := truncateHeadTail(content, 256, 128, 128, 10*1024)

	if result != content {
		t.Errorf("Expected no truncation for short content")
	}
}

func TestFormatOutputForModel(t *testing.T) {
	content := "line1\nline2\nline3"
	result := FormatOutputForModel(content, 3)

	if !strings.Contains(result, "Total output lines: 3") {
		t.Errorf("Expected result to contain line count")
	}
	if !strings.Contains(result, content) {
		t.Errorf("Expected result to contain original content")
	}
}
