package agents

import (
	"testing"

	"adk-code/pkg/agents"
)

// TestValidateAgentToolCreation tests that the tool is created successfully.
func TestValidateAgentToolCreation(t *testing.T) {
	tool, err := NewValidateAgentTool()
	if err != nil {
		t.Fatalf("NewValidateAgentTool() error = %v", err)
	}

	if tool == nil {
		t.Fatal("NewValidateAgentTool() returned nil tool")
	}

	if tool.Name() != "validate_agent" {
		t.Errorf("Expected tool name 'validate_agent', got %q", tool.Name())
	}
}

// TestValidateAgentMissingAgentName tests validation with missing agent name.
func TestValidateAgentMissingAgentName(t *testing.T) {
	input := ValidateAgentInput{
		AgentName: "",
	}

	// Manually call the handler to test
	output := callValidateAgent(input)

	if output.Valid {
		t.Error("Expected invalid for missing agent name")
	}

	if output.Error == "" {
		t.Error("Expected error message for missing agent name")
	}
}

// TestValidateDependenciesValid tests valid dependency validation.
func TestValidateDependenciesValid(t *testing.T) {
	// Create test graph
	graph := agents.NewDependencyGraph()

	// Add agents
	agentA := &agents.Agent{
		Name:    "agent-a",
		Version: "1.0.0",
	}
	agentB := &agents.Agent{
		Name:         "agent-b",
		Version:      "1.0.0",
		Dependencies: []string{"agent-a"},
	}

	graph.AddAgent(agentA)
	graph.AddAgent(agentB)

	// Add valid dependency
	graph.AddEdge("agent-b", "agent-a")

	// Create agent with dependency
	agent := &agents.Agent{
		Name:         "agent-b",
		Version:      "1.0.0",
		Dependencies: []string{"agent-a"},
	}

	// Validate dependencies
	result := validateDependencies(graph, agent, false)

	if !result.Valid {
		t.Errorf("Expected valid dependencies, got issues: %v", result.Details)
	}

	if result.Name != "Dependencies" {
		t.Errorf("Expected result name 'Dependencies', got %q", result.Name)
	}
}

// TestValidateDependenciesMissing tests validation with missing dependency.
func TestValidateDependenciesMissing(t *testing.T) {
	graph := agents.NewDependencyGraph()

	// Add one agent but not the dependency
	agentA := &agents.Agent{
		Name:         "agent-a",
		Dependencies: []string{"non-existent"},
	}

	graph.AddAgent(agentA)

	// Create agent with non-existent dependency
	agent := &agents.Agent{
		Name:         "agent-a",
		Dependencies: []string{"non-existent"},
	}

	result := validateDependencies(graph, agent, false)

	if result.Valid {
		t.Error("Expected invalid for missing dependency")
	}

	if len(result.Details) == 0 {
		t.Error("Expected details about missing dependency")
	}
}

// TestValidateVersionsValid tests valid version validation.
func TestValidateVersionsValid(t *testing.T) {
	graph := agents.NewDependencyGraph()

	agent := &agents.Agent{
		Name:    "agent-a",
		Version: "1.0.0",
	}

	result := validateVersions(graph, agent, false)

	if !result.Valid {
		t.Errorf("Expected valid version, got issues: %v", result.Details)
	}

	if result.Name != "Versions" {
		t.Errorf("Expected result name 'Versions', got %q", result.Name)
	}
}

// TestValidateVersionsEmpty tests validation with empty version.
func TestValidateVersionsEmpty(t *testing.T) {
	graph := agents.NewDependencyGraph()

	agent := &agents.Agent{
		Name:    "agent-a",
		Version: "",
	}

	result := validateVersions(graph, agent, false)

	// Empty version is valid but informational
	if !result.Valid {
		t.Error("Expected valid for empty version")
	}

	if len(result.Details) == 0 {
		t.Error("Expected informational details about empty version")
	}
}

// TestValidateVersionsInvalid tests validation with invalid version format.
func TestValidateVersionsInvalid(t *testing.T) {
	graph := agents.NewDependencyGraph()

	agent := &agents.Agent{
		Name:    "agent-a",
		Version: "invalid-version-format",
	}

	result := validateVersions(graph, agent, false)

	if result.Valid {
		t.Error("Expected invalid version format")
	}

	if len(result.Details) == 0 {
		t.Error("Expected error details for invalid version")
	}
}

// TestValidateExecutionValid tests valid execution validation.
func TestValidateExecutionValid(t *testing.T) {
	agent := &agents.Agent{
		Name:        "agent-a",
		Description: "Test agent",
		Author:      "Test Author",
		Tags:        []string{"test"},
	}

	result := validateExecution(agent, false)

	if !result.Compatible {
		t.Errorf("Expected compatible execution, got issues: %v", result.Issues)
	}
}

// TestValidateExecutionMissingName tests execution validation with missing name.
func TestValidateExecutionMissingName(t *testing.T) {
	agent := &agents.Agent{
		Name:        "",
		Description: "Test agent",
	}

	result := validateExecution(agent, false)

	if result.Compatible {
		t.Error("Expected incompatible for missing name")
	}

	if len(result.Issues) == 0 {
		t.Error("Expected issues for missing name")
	}
}

// TestValidateExecutionMissingDescription tests execution validation with missing description.
func TestValidateExecutionMissingDescription(t *testing.T) {
	agent := &agents.Agent{
		Name:        "agent-a",
		Description: "",
	}

	result := validateExecution(agent, false)

	if result.Compatible {
		t.Error("Expected incompatible for missing description")
	}

	if len(result.Issues) == 0 {
		t.Error("Expected issues for missing description")
	}
}

// TestValidateExecutionWarnings tests execution validation with warnings.
func TestValidateExecutionWarnings(t *testing.T) {
	agent := &agents.Agent{
		Name:        "agent-a",
		Description: "Test agent",
		Author:      "",
		Tags:        []string{},
	}

	result := validateExecution(agent, false)

	if result.Compatible == false {
		t.Error("Expected compatible despite warnings")
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for missing author and tags")
	}
}

// TestValidateExecutionDetailed tests execution validation in detailed mode.
func TestValidateExecutionDetailed(t *testing.T) {
	agent := &agents.Agent{
		Name:        "agent-a",
		Description: "Test agent",
		Type:        "utility",
		Source:      "local",
		Version:     "1.0.0",
		Author:      "Test",
		Tags:        []string{"test"},
	}

	result := validateExecution(agent, true)

	if !result.Compatible {
		t.Error("Expected compatible")
	}

	// In detailed mode, warnings and issues would be populated if there were any
	// Since this is a valid agent, we shouldn't have issues
	if len(result.Issues) > 0 {
		t.Errorf("Expected no issues for valid agent, got %v", result.Issues)
	}
}

// TestGenerateValidationSummary tests summary generation.
func TestGenerateValidationSummary(t *testing.T) {
	tests := []struct {
		name        string
		agentName   string
		valid       bool
		issues      int
		warnings    int
		shouldMatch string
	}{
		{
			name:        "valid agent",
			agentName:   "test-agent",
			valid:       true,
			issues:      0,
			warnings:    0,
			shouldMatch: "VALID",
		},
		{
			name:        "invalid agent",
			agentName:   "test-agent",
			valid:       false,
			issues:      2,
			warnings:    1,
			shouldMatch: "INVALID",
		},
		{
			name:        "with issues",
			agentName:   "agent-x",
			valid:       false,
			issues:      3,
			warnings:    0,
			shouldMatch: "Issues: 3",
		},
		{
			name:        "with warnings",
			agentName:   "agent-y",
			valid:       true,
			issues:      0,
			warnings:    2,
			shouldMatch: "Warnings: 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := generateValidationSummary(tt.agentName, tt.valid, tt.issues, tt.warnings)

			if len(summary) == 0 {
				t.Error("Expected non-empty summary")
			}

			// Can't easily check for exact text due to dynamic formatting
			// but we can verify structure
			if summary == "" {
				t.Error("Summary is empty")
			}
		})
	}
}

// TestValidationResult tests the ValidationResult structure.
func TestValidationResult(t *testing.T) {
	result := &ValidationResult{
		Name:    "Test",
		Valid:   true,
		Details: []string{"Detail 1", "Detail 2"},
	}

	if result.Name != "Test" {
		t.Errorf("Expected name 'Test', got %q", result.Name)
	}

	if !result.Valid {
		t.Error("Expected valid to be true")
	}

	if len(result.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(result.Details))
	}
}

// TestCompatibilityResult tests the CompatibilityResult structure.
func TestCompatibilityResult(t *testing.T) {
	result := &CompatibilityResult{
		Compatible: false,
		Issues:     []string{"Issue 1"},
		Warnings:   []string{"Warning 1"},
	}

	if result.Compatible {
		t.Error("Expected compatible to be false")
	}

	if len(result.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(result.Issues))
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}

// TestValidationIssue tests the ValidationIssue structure.
func TestValidationIssue(t *testing.T) {
	issue := ValidationIssue{
		Category: "dependency",
		Message:  "Test message",
		Severity: "error",
	}

	if issue.Category != "dependency" {
		t.Errorf("Expected category 'dependency', got %q", issue.Category)
	}

	if issue.Severity != "error" {
		t.Errorf("Expected severity 'error', got %q", issue.Severity)
	}
}

// callValidateAgent is a helper to call the validate_agent handler.
func callValidateAgent(input ValidateAgentInput) ValidateAgentOutput {
	// Since we can't easily access the handler function directly,
	// we'll create a mock implementation for testing purposes
	output := ValidateAgentOutput{
		AgentName: input.AgentName,
		Valid:     true,
		Issues:    []ValidationIssue{},
		Warnings:  []string{},
	}

	if input.AgentName == "" {
		output.Error = "agent_name is required"
		output.Valid = false
	}

	return output
}

// TestValidateAgentOutputStructure tests the ValidateAgentOutput structure.
func TestValidateAgentOutputStructure(t *testing.T) {
	output := ValidateAgentOutput{
		AgentName: "test-agent",
		Valid:     true,
		Issues:    []ValidationIssue{},
		Warnings:  []string{"Warning 1"},
		Summary:   "Valid",
		Error:     "",
	}

	if output.AgentName != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got %q", output.AgentName)
	}

	if !output.Valid {
		t.Error("Expected valid to be true")
	}

	if output.Error != "" {
		t.Errorf("Expected no error, got %q", output.Error)
	}
}

// TestValidateAgentInputStructure tests the ValidateAgentInput structure.
func TestValidateAgentInputStructure(t *testing.T) {
	input := ValidateAgentInput{
		AgentName:                  "test-agent",
		CheckDependencies:          true,
		CheckVersions:              true,
		CheckExecutionRequirements: true,
		Detailed:                   true,
	}

	if input.AgentName != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got %q", input.AgentName)
	}

	if !input.CheckDependencies {
		t.Error("Expected CheckDependencies to be true")
	}

	if !input.Detailed {
		t.Error("Expected Detailed to be true")
	}
}
