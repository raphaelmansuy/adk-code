package agents

import (
	"testing"

	"google.golang.org/adk/tool"

	agentspkg "adk-code/pkg/agents"
)

func TestNewLintAgentTool(t *testing.T) {
	toolResult, err := NewLintAgentTool()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if toolResult == nil {
		t.Fatal("expected tool, got nil")
	}
}

func TestLintAgentInput(t *testing.T) {
	input := &LintAgentInput{
		AgentName:       "test-agent",
		IncludeWarnings: true,
		IncludeInfo:     true,
	}

	if input.AgentName != "test-agent" {
		t.Errorf("expected agent name 'test-agent', got %q", input.AgentName)
	}

	if !input.IncludeWarnings {
		t.Error("expected IncludeWarnings=true")
	}

	if !input.IncludeInfo {
		t.Error("expected IncludeInfo=true")
	}
}

func TestLintAgentOutput_Empty(t *testing.T) {
	output := &LintAgentOutput{}

	if output.Success {
		t.Error("expected Success=false for empty output")
	}

	if output.Passed {
		t.Error("expected Passed=false for empty output")
	}

	// In the handler, we initialize these slices
	// But in zero value, they're nil, which is fine
	if output.AgentName != "" {
		t.Errorf("expected empty agent name, got %q", output.AgentName)
	}
}

func TestLintAgentOutput_WithData(t *testing.T) {
	output := &LintAgentOutput{
		Success:   true,
		Passed:    true,
		AgentName: "test",
		Summary:   "No issues",
		Errors:    []LintIssueOutput{},
		Warnings:  []LintIssueOutput{},
		Info:      []LintIssueOutput{},
		Total:     0,
	}

	if !output.Success {
		t.Error("expected Success=true")
	}

	if !output.Passed {
		t.Error("expected Passed=true")
	}

	if output.AgentName != "test" {
		t.Errorf("expected agent name 'test', got %q", output.AgentName)
	}

	if output.Total != 0 {
		t.Errorf("expected Total=0, got %d", output.Total)
	}
}

func TestLintIssueOutput(t *testing.T) {
	issue := LintIssueOutput{
		Rule:       "test-rule",
		Message:    "test message",
		Field:      "name",
		Suggestion: "fix this",
	}

	if issue.Rule != "test-rule" {
		t.Errorf("expected rule 'test-rule', got %q", issue.Rule)
	}

	if issue.Message != "test message" {
		t.Errorf("expected message 'test message', got %q", issue.Message)
	}

	if issue.Field != "name" {
		t.Errorf("expected field 'name', got %q", issue.Field)
	}

	if issue.Suggestion != "fix this" {
		t.Errorf("expected suggestion 'fix this', got %q", issue.Suggestion)
	}
}

func TestLintAgentIntegration(t *testing.T) {
	// Create an agent with issues
	agent := &agentspkg.Agent{
		Name:        "BadAgent", // Should be kebab-case
		Description: "short",    // Too short
		Author:      "",         // Missing
		Version:     "",         // Missing
		Tags:        []string{}, // Empty
	}

	linter := agentspkg.NewLinter()
	result := linter.Lint(agent)

	if result.ErrorCount == 0 {
		t.Error("expected error count > 0 for agent with issues")
	}

	if result.Passed {
		t.Error("expected Passed=false for agent with errors")
	}
}

func TestLintAgentToolInterface(t *testing.T) {
	toolInst, err := NewLintAgentTool()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the tool implements the Tool interface
	var _ tool.Tool = toolInst
}
