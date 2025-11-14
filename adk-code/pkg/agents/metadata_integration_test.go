package agents

import (
	"testing"
)

// TestMetadataValidatorNew tests creating a new metadata validator.
func TestMetadataValidatorNew(t *testing.T) {
	v := NewAgentMetadataValidator()
	if v == nil {
		t.Fatal("Expected non-nil validator")
	}

	if v.Graph == nil {
		t.Error("Expected non-nil graph")
	}

	if len(v.AgentVersions) != 0 {
		t.Error("Expected empty versions map")
	}

	if len(v.Constraints) != 0 {
		t.Error("Expected empty constraints map")
	}
}

// TestMetadataValidatorAddAgent tests adding agents to the validator.
func TestMetadataValidatorAddAgent(t *testing.T) {
	v := NewAgentMetadataValidator()
	agent := &Agent{
		Name:        "test-agent",
		Description: "Test agent",
		Version:     "1.0.0",
	}

	err := v.AddAgent(agent, ">=1.0.0")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if v.Graph.AgentCount() != 1 {
		t.Errorf("Expected 1 agent in graph, got %d", v.Graph.AgentCount())
	}

	if len(v.AgentVersions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(v.AgentVersions))
	}

	if len(v.Constraints) != 1 {
		t.Errorf("Expected 1 constraint, got %d", len(v.Constraints))
	}
}

// TestMetadataValidatorAddAgentNil tests adding nil agent.
func TestMetadataValidatorAddAgentNil(t *testing.T) {
	v := NewAgentMetadataValidator()
	err := v.AddAgent(nil, "")
	if err == nil {
		t.Error("Expected error for nil agent")
	}
}

// TestMetadataValidatorAddAgentInvalidVersion tests adding agent with invalid version.
func TestMetadataValidatorAddAgentInvalidVersion(t *testing.T) {
	v := NewAgentMetadataValidator()
	agent := &Agent{
		Name:        "test",
		Description: "Test",
		Version:     "invalid",
	}

	err := v.AddAgent(agent, "")
	if err == nil {
		t.Error("Expected error for invalid version")
	}
}

// TestMetadataValidatorAddAgentInvalidConstraint tests adding agent with invalid constraint.
func TestMetadataValidatorAddAgentInvalidConstraint(t *testing.T) {
	v := NewAgentMetadataValidator()
	agent := &Agent{
		Name:        "test",
		Description: "Test",
		Version:     "1.0.0",
	}

	err := v.AddAgent(agent, "invalid-constraint")
	if err == nil {
		t.Error("Expected error for invalid constraint")
	}
}

// TestMetadataValidatorAddDependency tests adding dependencies.
func TestMetadataValidatorAddDependency(t *testing.T) {
	v := NewAgentMetadataValidator()

	agent1 := &Agent{Name: "a1", Description: "Agent 1"}
	agent2 := &Agent{Name: "a2", Description: "Agent 2"}

	v.AddAgent(agent1, "")
	v.AddAgent(agent2, "")

	err := v.AddDependency("a1", "a2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if v.Graph.EdgeCount() != 1 {
		t.Errorf("Expected 1 edge, got %d", v.Graph.EdgeCount())
	}
}

// TestMetadataValidatorValidateDependencies tests dependency validation.
func TestMetadataValidatorValidateDependencies(t *testing.T) {
	v := NewAgentMetadataValidator()

	agent1 := &Agent{Name: "a1", Description: "Agent 1"}
	agent2 := &Agent{Name: "a2", Description: "Agent 2"}

	v.AddAgent(agent1, "")
	v.AddAgent(agent2, "")
	v.AddDependency("a2", "a1")

	err := v.ValidateDependencies("a2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestMetadataValidatorValidateDependenciesNonexistent tests validation with nonexistent agent.
func TestMetadataValidatorValidateDependenciesNonexistent(t *testing.T) {
	v := NewAgentMetadataValidator()
	err := v.ValidateDependencies("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent agent")
	}
}

// TestMetadataValidatorValidateDependenciesCycle tests validation detects cycles.
func TestMetadataValidatorValidateDependenciesCycle(t *testing.T) {
	v := NewAgentMetadataValidator()

	agent1 := &Agent{Name: "a1", Description: "Agent 1"}
	agent2 := &Agent{Name: "a2", Description: "Agent 2"}

	v.AddAgent(agent1, "")
	v.AddAgent(agent2, "")
	v.AddDependency("a1", "a2")
	v.AddDependency("a2", "a1")

	err := v.ValidateDependencies("a1")
	if err == nil {
		t.Error("Expected error for circular dependency")
	}
}

// TestMetadataValidatorValidateVersionConstraints tests version constraint validation.
func TestMetadataValidatorValidateVersionConstraints(t *testing.T) {
	v := NewAgentMetadataValidator()

	agent := &Agent{
		Name:        "test",
		Description: "Test",
		Version:     "1.5.0",
	}

	v.AddAgent(agent, ">=1.0.0")

	err := v.ValidateVersionConstraints()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestMetadataValidatorValidateVersionConstraintsFail tests version constraint failure.
func TestMetadataValidatorValidateVersionConstraintsFail(t *testing.T) {
	v := NewAgentMetadataValidator()

	agent := &Agent{
		Name:        "test",
		Description: "Test",
		Version:     "0.5.0",
	}

	v.AddAgent(agent, ">=1.0.0")

	err := v.ValidateVersionConstraints()
	if err == nil {
		t.Error("Expected error for constraint violation")
	}
}

// TestMetadataValidatorValidateDependencyVersions tests transitive dependency version validation.
func TestMetadataValidatorValidateDependencyVersions(t *testing.T) {
	v := NewAgentMetadataValidator()

	agent1 := &Agent{Name: "a1", Description: "Agent 1", Version: "1.0.0"}
	agent2 := &Agent{Name: "a2", Description: "Agent 2", Version: "2.0.0"}

	v.AddAgent(agent1, ">=1.0.0")
	v.AddAgent(agent2, ">=2.0.0")
	v.AddDependency("a2", "a1")

	err := v.ValidateDependencyVersions("a2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestMetadataValidatorValidateAgent tests comprehensive agent validation.
func TestMetadataValidatorValidateAgent(t *testing.T) {
	v := NewAgentMetadataValidator()

	agent1 := &Agent{Name: "a1", Description: "Agent 1", Version: "1.0.0"}
	agent2 := &Agent{
		Name:         "a2",
		Description:  "Agent 2",
		Version:      "2.0.0",
		Dependencies: []string{"a1"},
	}

	v.AddAgent(agent1, ">=1.0.0")
	v.AddAgent(agent2, ">=1.0.0")
	v.AddDependency("a2", "a1")

	report, err := v.ValidateAgent("a2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !report.Valid {
		t.Errorf("Expected valid agent, got issues: %v", report.Issues)
	}

	if len(report.ResolvedDependencies) != 2 {
		t.Errorf("Expected 2 resolved deps, got %d", len(report.ResolvedDependencies))
	}
}

// TestValidationReportString tests validation report string formatting.
func TestValidationReportString(t *testing.T) {
	report := &ValidationReport{
		AgentName:            "test-agent",
		Valid:                false,
		Issues:               []string{"Issue 1", "Issue 2"},
		ResolvedDependencies: []string{"dep1", "dep2"},
	}

	str := report.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	if !containsStr(str, "test-agent") {
		t.Error("Expected agent name in string")
	}

	if !containsStr(str, "Issue 1") {
		t.Error("Expected issue in string")
	}

	if !containsStr(str, "dep1") {
		t.Error("Expected dependency in string")
	}
}

// TestBuildGraphFromDiscovery tests building a graph from discovery results.
func TestBuildGraphFromDiscovery(t *testing.T) {
	result := &DiscoveryResult{
		Agents: []*Agent{
			{Name: "a1", Description: "Agent 1", Dependencies: []string{}},
			{Name: "a2", Description: "Agent 2", Dependencies: []string{"a1"}},
		},
		Total: 2,
	}

	graph, err := BuildGraphFromDiscovery(result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if graph.AgentCount() != 2 {
		t.Errorf("Expected 2 agents, got %d", graph.AgentCount())
	}

	if graph.EdgeCount() != 1 {
		t.Errorf("Expected 1 edge, got %d", graph.EdgeCount())
	}
}

// TestBuildGraphFromDiscoveryNil tests with nil discovery result.
func TestBuildGraphFromDiscoveryNil(t *testing.T) {
	_, err := BuildGraphFromDiscovery(nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}
}

// TestBuildGraphFromDiscoveryMissingDependency tests handling missing dependencies.
func TestBuildGraphFromDiscoveryMissingDependency(t *testing.T) {
	result := &DiscoveryResult{
		Agents: []*Agent{
			{Name: "a1", Description: "Agent 1", Dependencies: []string{"missing"}},
		},
		Total: 1,
	}

	graph, err := BuildGraphFromDiscovery(result)
	if err != nil {
		t.Errorf("Expected no error for missing dep, got %v", err)
	}

	// Should still succeed but with edge skipped
	if graph.EdgeCount() != 0 {
		t.Errorf("Expected 0 edges for missing dependency, got %d", graph.EdgeCount())
	}
}

// TestValidateAgentCompatibility tests agent compatibility validation.
func TestValidateAgentCompatibility(t *testing.T) {
	agent := &Agent{
		Name:    "test",
		Version: "1.5.0",
	}

	req := &ExecutionRequirements{
		VersionConstraint: ">=1.0.0",
	}

	report := ValidateAgentCompatibility(agent, req)
	if !report.Compatible {
		t.Errorf("Expected compatible agent, got issues: %v", report.Issues)
	}
}

// TestValidateAgentCompatibilityVersionFail tests version incompatibility.
func TestValidateAgentCompatibilityVersionFail(t *testing.T) {
	agent := &Agent{
		Name:    "test",
		Version: "0.5.0",
	}

	req := &ExecutionRequirements{
		VersionConstraint: ">=1.0.0",
	}

	report := ValidateAgentCompatibility(agent, req)
	if report.Compatible {
		t.Error("Expected incompatible agent")
	}

	if len(report.Issues) == 0 {
		t.Error("Expected issues for incompatible agent")
	}
}

// TestValidateAgentCompatibilityNoVersion tests compatibility with no version.
func TestValidateAgentCompatibilityNoVersion(t *testing.T) {
	agent := &Agent{
		Name: "test",
	}

	req := &ExecutionRequirements{
		VersionConstraint: ">=1.0.0",
	}

	report := ValidateAgentCompatibility(agent, req)
	if report.Compatible {
		t.Error("Expected incompatible when no version specified but constraint required")
	}

	if len(report.Issues) == 0 {
		t.Error("Expected issues for missing version")
	}
}

// TestUpdateExecutionRequirementsWithVersion tests updating requirements.
func TestUpdateExecutionRequirementsWithVersion(t *testing.T) {
	req := &ExecutionRequirements{}

	err := UpdateExecutionRequirementsWithVersion(req, "^1.0.0")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if req.VersionConstraint != "^1.0.0" {
		t.Errorf("Expected constraint to be set, got %q", req.VersionConstraint)
	}
}

// TestUpdateExecutionRequirementsWithInvalidVersion tests with invalid constraint.
func TestUpdateExecutionRequirementsWithInvalidVersion(t *testing.T) {
	req := &ExecutionRequirements{}

	err := UpdateExecutionRequirementsWithVersion(req, "invalid-constraint")
	if err == nil {
		t.Error("Expected error for invalid constraint")
	}
}

// TestUpdateExecutionRequirementsNil tests with nil requirements.
func TestUpdateExecutionRequirementsNil(t *testing.T) {
	err := UpdateExecutionRequirementsWithVersion(nil, "1.0.0")
	if err == nil {
		t.Error("Expected error for nil requirements")
	}
}

// TestGetAgentMetadata tests metadata extraction.
func TestGetAgentMetadata(t *testing.T) {
	agent := &Agent{
		Name:        "test",
		Version:     "1.0.0",
		Author:      "Test Author",
		Tags:        []string{"tag1", "tag2"},
		Description: "Test agent",
	}

	metadata := GetAgentMetadata(agent)
	if metadata == nil {
		t.Fatal("Expected non-nil metadata")
	}

	if metadata["name"] != "test" {
		t.Errorf("Expected name=test, got %v", metadata["name"])
	}

	if metadata["version"] != "1.0.0" {
		t.Errorf("Expected version=1.0.0, got %v", metadata["version"])
	}
}

// TestGetAgentMetadataNil tests with nil agent.
func TestGetAgentMetadataNil(t *testing.T) {
	metadata := GetAgentMetadata(nil)
	if metadata != nil {
		t.Error("Expected nil metadata for nil agent")
	}
}

// TestCompatibilityReportString tests compatibility report string formatting.
func TestCompatibilityReportString(t *testing.T) {
	report := &CompatibilityReport{
		Agent:      "test-agent",
		Compatible: false,
		Issues:     []string{"Issue 1"},
		Warnings:   []string{"Warning 1"},
	}

	str := report.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}

	if !containsStr(str, "test-agent") {
		t.Error("Expected agent name in string")
	}

	if !containsStr(str, "Issue 1") {
		t.Error("Expected issue in string")
	}

	if !containsStr(str, "Warning 1") {
		t.Error("Expected warning in string")
	}
}

// Helper function to check if string contains substring
func containsStr(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
