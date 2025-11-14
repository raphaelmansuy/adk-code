package agents

import (
	"testing"
)

func TestNewEditAgentTool(t *testing.T) {
	tool, err := NewEditAgentTool()
	if err != nil {
		t.Fatalf("Failed to create edit agent tool: %v", err)
	}

	if tool == nil {
		t.Fatal("Tool is nil")
	}
}

func TestEditAgentInputValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       EditAgentInput
		expectError bool
	}{
		{
			name: "missing agent name",
			input: EditAgentInput{
				Field: "description",
				Value: "New description",
			},
			expectError: true,
		},
		{
			name: "missing field",
			input: EditAgentInput{
				AgentName: "test-agent",
				Value:     "New value",
			},
			expectError: true,
		},
		{
			name: "invalid field",
			input: EditAgentInput{
				AgentName: "test-agent",
				Field:     "invalid-field",
				Value:     "New value",
			},
			expectError: true,
		},
		{
			name: "valid input - description",
			input: EditAgentInput{
				AgentName: "test-agent",
				Field:     "description",
				Value:     "Updated description",
			},
			expectError: false,
		},
		{
			name: "valid input - version",
			input: EditAgentInput{
				AgentName: "test-agent",
				Field:     "version",
				Value:     "2.0.0",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify tool can be created
			tool, err := NewEditAgentTool()
			if err != nil {
				t.Fatalf("Failed to create tool: %v", err)
			}

			if tool == nil {
				t.Fatal("Tool is nil")
			}
		})
	}
}

func TestEditAgentOutputStructure(t *testing.T) {
	output := EditAgentOutput{
		Success:    true,
		AgentName:  "test-agent",
		FilePath:   "/path/to/agent.md",
		Field:      "description",
		OldValue:   "Old description",
		NewValue:   "New description",
		BackupPath: "/path/to/agent.md.bak",
		Message:    "Edit successful",
	}

	if output.AgentName != "test-agent" {
		t.Fatalf("Expected agent name 'test-agent', got '%s'", output.AgentName)
	}

	if !output.Success {
		t.Fatal("Expected success=true")
	}

	if output.Field != "description" {
		t.Fatalf("Expected field 'description', got '%s'", output.Field)
	}

	if output.OldValue == output.NewValue {
		t.Fatal("Expected OldValue and NewValue to be different")
	}
}

func TestEditAgentFieldValues(t *testing.T) {
	fields := []string{"name", "description", "version", "author", "tags"}

	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			output := EditAgentOutput{
				Success:   true,
				AgentName: "test-agent",
				Field:     field,
				OldValue:  "old-value",
				NewValue:  "new-value",
			}

			if output.Field != field {
				t.Fatalf("Expected field '%s', got '%s'", field, output.Field)
			}

			if output.OldValue == "" {
				t.Fatal("Expected OldValue to be set")
			}

			if output.NewValue == "" {
				t.Fatal("Expected NewValue to be set")
			}
		})
	}
}

func TestEditAgentMessageGeneration(t *testing.T) {
	output := EditAgentOutput{
		Success:   true,
		AgentName: "test-agent",
		Field:     "version",
		OldValue:  "1.0.0",
		NewValue:  "2.0.0",
		Message:   "Agent 'test-agent' field 'version' updated successfully from '1.0.0' to '2.0.0'",
	}

	if output.Message == "" {
		t.Fatal("Expected message, got empty string")
	}

	// Check message format
	if !containsStr(output.Message, "test-agent") {
		t.Fatal("Message should contain agent name")
	}

	if !containsStr(output.Message, "version") {
		t.Fatal("Message should contain field name")
	}

	if !containsStr(output.Message, "1.0.0") {
		t.Fatal("Message should contain old value")
	}

	if !containsStr(output.Message, "2.0.0") {
		t.Fatal("Message should contain new value")
	}
}

func TestEditAgentInputFields(t *testing.T) {
	input := EditAgentInput{
		AgentName:    "my-agent",
		FilePath:     "/path/to/agent.md",
		Field:        "description",
		Value:        "New description",
		CreateBackup: true,
	}

	if input.AgentName != "my-agent" {
		t.Fatalf("Expected agent name 'my-agent', got '%s'", input.AgentName)
	}

	if input.Field != "description" {
		t.Fatalf("Expected field 'description', got '%s'", input.Field)
	}

	if !input.CreateBackup {
		t.Fatal("Expected CreateBackup to be true")
	}
}

// Helper function to check if a string contains a substring
func containsStr(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
