package tools

import (
	"strings"
	"testing"
)

// TestExtractError_EmptyObject tests handling of empty error objects
func TestExtractError_EmptyObject(t *testing.T) {
	parser := NewToolResultParser(nil)

	tests := []struct {
		name     string
		result   map[string]any
		wantErr  bool
		contains string
	}{
		{
			name:     "empty error object",
			result:   map[string]any{"error": map[string]any{}},
			wantErr:  true,
			contains: "failed with no error details",
		},
		{
			name:     "error with message field",
			result:   map[string]any{"error": map[string]any{"message": "Connection timeout"}},
			wantErr:  true,
			contains: "Connection timeout",
		},
		{
			name:     "error string",
			result:   map[string]any{"error": "Something went wrong"},
			wantErr:  true,
			contains: "Something went wrong",
		},
		{
			name:     "empty error object with output",
			result:   map[string]any{"error": map[string]any{}, "output": "Failed to navigate"},
			wantErr:  true,
			contains: "Failed to navigate",
		},
		{
			name:     "error with details field",
			result:   map[string]any{"error": map[string]any{"details": "Network unreachable"}},
			wantErr:  true,
			contains: "Network unreachable",
		},
		{
			name:    "no error",
			result:  map[string]any{"output": "Success"},
			wantErr: false,
		},
		{
			name:    "empty result",
			result:  map[string]any{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := parser.extractError(tt.result)

			if tt.wantErr {
				if errMsg == "" {
					t.Errorf("extractError() expected error but got empty string")
				}
				if tt.contains != "" && !strings.Contains(errMsg, tt.contains) {
					t.Errorf("extractError() = %q, want to contain %q", errMsg, tt.contains)
				}
			} else {
				if errMsg != "" {
					t.Errorf("extractError() = %q, want empty string", errMsg)
				}
			}
		})
	}
}

// TestParseToolResult_WithEmptyError tests the full parsing flow
func TestParseToolResult_WithEmptyError(t *testing.T) {
	parser := NewToolResultParser(nil)

	// Test empty error object scenario (common with MCP tools)
	result := map[string]any{
		"error": map[string]any{},
	}

	output := parser.ParseToolResult("scraping_browser_navigate", result)

	if !strings.Contains(output, "❌ Error") {
		t.Errorf("ParseToolResult() expected error indicator, got: %s", output)
	}

	if !strings.Contains(output, "failed with no error details") {
		t.Errorf("ParseToolResult() expected generic error message, got: %s", output)
	}
}

// TestParseToolResult_MCPPlaywrightScenarios tests Playwright MCP-specific scenarios
func TestParseToolResult_MCPPlaywrightScenarios(t *testing.T) {
	parser := NewToolResultParser(nil)

	scenarios := []struct {
		name     string
		toolName string
		result   map[string]any
		wantErr  string
	}{
		{
			name:     "scraping_browser_navigate empty error",
			toolName: "scraping_browser_navigate",
			result:   map[string]any{"error": map[string]any{}},
			wantErr:  "failed with no error details",
		},
		{
			name:     "browser_console_messages empty error",
			toolName: "browser_console_messages",
			result:   map[string]any{"error": map[string]any{}},
			wantErr:  "failed with no error details",
		},
		{
			name:     "scrape_as_markdown empty error",
			toolName: "scrape_as_markdown",
			result:   map[string]any{"error": map[string]any{}},
			wantErr:  "failed with no error details",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			output := parser.ParseToolResult(sc.toolName, sc.result)

			if !strings.Contains(output, "❌ Error") {
				t.Errorf("Expected error indicator in output, got: %s", output)
			}

			if !strings.Contains(output, sc.wantErr) {
				t.Errorf("Expected error message to contain %q, got: %s", sc.wantErr, output)
			}
		})
	}
}
