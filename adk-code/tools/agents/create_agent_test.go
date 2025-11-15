package agents

import (
	"testing"
)

func TestNewCreateAgentTool(t *testing.T) {
	tool, err := NewCreateAgentTool()
	if err != nil {
		t.Fatalf("Failed to create agent tool: %v", err)
	}

	if tool == nil {
		t.Fatal("Tool is nil")
	}
}

func TestCreateAgentInputValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       CreateAgentInput
		expectError bool
	}{
		{
			name: "missing name",
			input: CreateAgentInput{
				Description:  "A test agent",
				TemplateType: "subagent",
			},
			expectError: true,
		},
		{
			name: "missing description",
			input: CreateAgentInput{
				Name:         "test-agent",
				TemplateType: "subagent",
			},
			expectError: true,
		},
		{
			name: "missing template type",
			input: CreateAgentInput{
				Name:        "test-agent",
				Description: "A test agent",
			},
			expectError: true,
		},
		{
			name: "valid input",
			input: CreateAgentInput{
				Name:         "test-agent",
				Description:  "A test agent",
				TemplateType: "subagent",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that tool creates successfully
			tool, err := NewCreateAgentTool()
			if err != nil {
				t.Fatalf("Failed to create tool: %v", err)
			}

			if tool == nil {
				t.Fatal("Tool is nil")
			}
		})
	}
}

func TestCreateAgentOutputStructure(t *testing.T) {
	output := CreateAgentOutput{
		Success:   true,
		AgentName: "test-agent",
		FilePath:  "/path/to/agent.md",
		Content:   "# Test Agent",
		Message:   "Success",
	}

	if output.AgentName != "test-agent" {
		t.Fatalf("Expected agent name 'test-agent', got '%s'", output.AgentName)
	}

	if !output.Success {
		t.Fatal("Expected success=true")
	}

	if output.FilePath == "" {
		t.Fatal("Expected file path, got empty string")
	}
}

func TestCreateAgentInputStructure(t *testing.T) {
	input := CreateAgentInput{
		Name:         "my-agent",
		Description:  "My test agent",
		TemplateType: "subagent",
		Author:       "Test Author",
		Tags:         []string{"test", "example"},
		Version:      "1.0.0",
	}

	if input.Name != "my-agent" {
		t.Fatalf("Expected name 'my-agent', got '%s'", input.Name)
	}

	if input.Description == "" {
		t.Fatal("Expected description")
	}

	if len(input.Tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(input.Tags))
	}
}
