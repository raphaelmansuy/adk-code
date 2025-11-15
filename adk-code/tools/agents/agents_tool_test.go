package agents

import (
	"testing"
)

// TestListAgentsTool verifies the list_agents tool is created correctly
func TestListAgentsTool(t *testing.T) {
	tool, err := NewListAgentsTool()
	if err != nil {
		t.Fatalf("Failed to create list_agents tool: %v", err)
	}
	if tool == nil {
		t.Fatal("list_agents tool is nil")
	}
	if tool.Name() != "list_agents" {
		t.Errorf("Expected tool name 'list_agents', got '%s'", tool.Name())
	}
}

// TestFormatSummaryEmpty tests summary generation for empty results
func TestFormatSummaryEmpty(t *testing.T) {
	summary := formatSummary(0, ListAgentsInput{}, 0)
	if summary != "No agents found in the project" {
		t.Errorf("Unexpected summary: %s", summary)
	}
}

// TestFormatSummarySingle tests summary with one agent
func TestFormatSummarySingle(t *testing.T) {
	summary := formatSummary(1, ListAgentsInput{}, 0)
	if summary != "Found 1 agent(s)" {
		t.Errorf("Expected 'Found 1 agent(s)', got '%s'", summary)
	}
}

// TestFormatSummaryMultiple tests summary with multiple agents
func TestFormatSummaryMultiple(t *testing.T) {
	summary := formatSummary(3, ListAgentsInput{}, 0)
	if summary != "Found 3 agent(s)" {
		t.Errorf("Expected 'Found 3 agent(s)', got '%s'", summary)
	}
}

// TestFormatSummaryWithErrors tests summary with errors
func TestFormatSummaryWithErrors(t *testing.T) {
	summary := formatSummary(2, ListAgentsInput{}, 1)
	expected := "Found 2 agent(s) (1 error(s))"
	if summary != expected {
		t.Errorf("Expected '%s', got '%s'", expected, summary)
	}
}
