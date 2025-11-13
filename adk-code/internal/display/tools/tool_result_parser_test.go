package tools

import (
	"strings"
	"testing"
)

// TestParseFileContent_NoContent verifies that file read events only show filename and line count
func TestParseFileContent_NoContent(t *testing.T) {
	parser := NewToolResultParser(nil)

	tests := []struct {
		name        string
		result      map[string]any
		expectPath  string
		expectLines string
		notExpect   string // Should NOT contain file content
	}{
		{
			name: "with_file_path_and_total_lines",
			result: map[string]any{
				"file_path":   "/path/to/main.go",
				"total_lines": 156,
				"content":     "package main\n\nfunc main() { ... }",
			},
			expectPath:  "/path/to/main.go",
			expectLines: "156 lines",
			notExpect:   "package main", // Should not include file content
		},
		{
			name: "fallback_to_content_counting",
			result: map[string]any{
				"file_path": "/test.txt",
				"content":   "line1\nline2\nline3",
				// No total_lines, should count from content
			},
			expectPath:  "/test.txt",
			expectLines: "3 lines",
			notExpect:   "line1",
		},
		{
			name: "with_float_total_lines",
			result: map[string]any{
				"file_path":   "/another/file.go",
				"total_lines": float64(42),
				"content":     "some content here",
			},
			expectPath:  "/another/file.go",
			expectLines: "42 lines",
			notExpect:   "some content",
		},
		{
			name: "no_path",
			result: map[string]any{
				"total_lines": 10,
				"content":     "test",
			},
			expectPath:  "",
			expectLines: "10 lines",
			notExpect:   "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := parser.ParseToolResult("read_file", tt.result)

			// Should contain path if expected
			if tt.expectPath != "" && !strings.Contains(output, tt.expectPath) {
				t.Errorf("Expected output to contain path %q, got: %s", tt.expectPath, output)
			}

			// Should contain line count
			if !strings.Contains(output, tt.expectLines) {
				t.Errorf("Expected output to contain %q, got: %s", tt.expectLines, output)
			}

			// Should NOT contain the file content
			if tt.notExpect != "" && strings.Contains(output, tt.notExpect) {
				t.Errorf("Expected output NOT to contain %q, but it does. Output: %s", tt.notExpect, output)
			}
		})
	}
}

// TestParseFileContent_ContentNotIncluded verifies that even with large content, it's not displayed
func TestParseFileContent_ContentNotIncluded(t *testing.T) {
	parser := NewToolResultParser(nil)

	largeContent := strings.Repeat("This is a line of content\n", 1000)

	result := map[string]any{
		"file_path":   "/large/file.go",
		"total_lines": 1000,
		"content":     largeContent,
	}

	output := parser.ParseToolResult("read_file", result)

	// Should only show summary
	if !strings.Contains(output, "/large/file.go") {
		t.Errorf("Expected output to contain file path")
	}
	if !strings.Contains(output, "1000 lines") {
		t.Errorf("Expected output to contain line count")
	}

	// Should NOT contain the actual file content
	if strings.Contains(output, "This is a line of content") {
		t.Errorf("Output should not contain file content, but it does")
	}

	// Should NOT contain code block markers that would indicate content display
	if strings.Count(output, "```") > 0 {
		t.Errorf("Output should not contain code blocks, but it does")
	}
}
