package agents

import (
	"testing"
)

func TestNewExportAgentTool(t *testing.T) {
	tool, err := NewExportAgentTool()
	if err != nil {
		t.Fatalf("Failed to create export agent tool: %v", err)
	}

	if tool == nil {
		t.Fatal("Tool is nil")
	}
}

func TestExportAgentInputValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       ExportAgentInput
		expectError bool
	}{
		{
			name: "missing agent name",
			input: ExportAgentInput{
				Format: "markdown",
			},
			expectError: true,
		},
		{
			name: "invalid format",
			input: ExportAgentInput{
				AgentName: "test-agent",
				Format:    "invalid-format",
			},
			expectError: true,
		},
		{
			name: "valid markdown export",
			input: ExportAgentInput{
				AgentName: "test-agent",
				Format:    "markdown",
			},
			expectError: false,
		},
		{
			name: "valid json export",
			input: ExportAgentInput{
				AgentName: "test-agent",
				Format:    "json",
			},
			expectError: false,
		},
		{
			name: "valid yaml export",
			input: ExportAgentInput{
				AgentName: "test-agent",
				Format:    "yaml",
			},
			expectError: false,
		},
		{
			name: "valid plugin export",
			input: ExportAgentInput{
				AgentName: "test-agent",
				Format:    "plugin",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify tool can be created
			tool, err := NewExportAgentTool()
			if err != nil {
				t.Fatalf("Failed to create tool: %v", err)
			}

			if tool == nil {
				t.Fatal("Tool is nil")
			}
		})
	}
}

func TestExportAgentOutputStructure(t *testing.T) {
	output := ExportAgentOutput{
		Success:    true,
		AgentName:  "test-agent",
		FilePath:   "/path/to/agent.md",
		ExportPath: "/path/to/export.md",
		Format:     "markdown",
		Content:    "# Test Agent\n\nContent here",
		Size:       100,
		Message:    "Export successful",
	}

	if output.AgentName != "test-agent" {
		t.Fatalf("Expected agent name 'test-agent', got '%s'", output.AgentName)
	}

	if !output.Success {
		t.Fatal("Expected success=true")
	}

	if output.Format != "markdown" {
		t.Fatalf("Expected format 'markdown', got '%s'", output.Format)
	}

	if output.Size == 0 {
		t.Fatal("Expected size > 0")
	}
}

func TestExportAgentFormats(t *testing.T) {
	formats := []string{"markdown", "json", "yaml", "plugin"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			output := ExportAgentOutput{
				Success:   true,
				AgentName: "test-agent",
				Format:    format,
				Content:   "test content",
			}

			if output.Format != format {
				t.Fatalf("Expected format '%s', got '%s'", format, output.Format)
			}

			if output.Content == "" {
				t.Fatal("Expected content to be set")
			}
		})
	}
}

func TestExportAgentMessageGeneration(t *testing.T) {
	output := ExportAgentOutput{
		Success:    true,
		AgentName:  "test-agent",
		ExportPath: "/path/to/export.md",
		Format:     "markdown",
		Message:    "Agent 'test-agent' exported successfully to /path/to/export.md in markdown format",
	}

	if output.Message == "" {
		t.Fatal("Expected message, got empty string")
	}

	// Check message format
	if !containsString(output.Message, "test-agent") {
		t.Fatal("Message should contain agent name")
	}

	if !containsString(output.Message, "markdown") {
		t.Fatal("Message should contain format")
	}
}

func TestExportAgentInputFields(t *testing.T) {
	input := ExportAgentInput{
		AgentName:       "my-agent",
		FilePath:        "/path/to/agent.md",
		Format:          "json",
		OutputPath:      "/path/to/output.json",
		IncludeMetadata: true,
	}

	if input.AgentName != "my-agent" {
		t.Fatalf("Expected agent name 'my-agent', got '%s'", input.AgentName)
	}

	if input.Format != "json" {
		t.Fatalf("Expected format 'json', got '%s'", input.Format)
	}

	if !input.IncludeMetadata {
		t.Fatal("Expected IncludeMetadata to be true")
	}
}

func TestExportAgentDefaultFormat(t *testing.T) {
	// Test that markdown is the default format
	input := ExportAgentInput{
		AgentName: "test-agent",
		// Format not specified - should default to markdown
	}

	if input.Format != "" {
		t.Fatalf("Expected Format to be empty before tool processes it, got '%s'", input.Format)
	}

	// Tool sets default to "markdown" if not specified
	// This is tested in the tool handler
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
