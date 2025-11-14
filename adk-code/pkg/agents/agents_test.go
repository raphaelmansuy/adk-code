package agents

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestParseAgentFileValid tests parsing a valid agent file
func TestParseAgentFileValid(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create a valid agent file
	agentContent := `---
name: test-agent
description: A test agent for unit testing
---
# Test Agent

This is a test agent markdown content.
`
	filePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(filePath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the agent file
	agent, err := ParseAgentFile(filePath)
	if err != nil {
		t.Fatalf("Failed to parse agent file: %v", err)
	}

	// Verify agent fields
	if agent.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", agent.Name)
	}
	if agent.Description != "A test agent for unit testing" {
		t.Errorf("Expected description 'A test agent for unit testing', got '%s'", agent.Description)
	}
	if agent.Path != filePath {
		t.Errorf("Expected path '%s', got '%s'", filePath, agent.Path)
	}
	if !agent.ModTime.After(time.Now().Add(-1 * time.Minute)) {
		t.Errorf("Expected recent ModTime")
	}
}

// TestParseAgentFileMissingFrontmatter tests that missing frontmatter is caught
func TestParseAgentFileMissingFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file without frontmatter
	content := "# No Frontmatter\n\nJust markdown content"
	filePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseAgentFile(filePath)
	if err != ErrNoFrontmatter {
		t.Errorf("Expected ErrNoFrontmatter, got %v", err)
	}
}

// TestParseAgentFileMissingName tests that missing name field is caught
func TestParseAgentFileMissingName(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file without name field
	content := `---
description: Missing name field
---
Content
`
	filePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseAgentFile(filePath)
	if err != ErrMissingName {
		t.Errorf("Expected ErrMissingName, got %v", err)
	}
}

// TestParseAgentFileMissingDescription tests that missing description is caught
func TestParseAgentFileMissingDescription(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file without description field
	content := `---
name: agent-name
---
Content
`
	filePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseAgentFile(filePath)
	if err != ErrMissingDescription {
		t.Errorf("Expected ErrMissingDescription, got %v", err)
	}
}

// TestParseAgentFileInvalidYAML tests that invalid YAML is caught
func TestParseAgentFileInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with invalid YAML (bad syntax)
	content := `---
name: agent-name
description: test
invalid: : bad yaml
---
Content
`
	filePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseAgentFile(filePath)
	if err != ErrInvalidYAML {
		t.Errorf("Expected ErrInvalidYAML, got %v", err)
	}
}

// TestExtractFrontmatterValid tests valid frontmatter extraction
func TestExtractFrontmatterValid(t *testing.T) {
	content := `---
name: test
description: test desc
---
Markdown content here`

	yaml, markdown, err := extractFrontmatter([]byte(content))
	if err != nil {
		t.Fatalf("Failed to extract frontmatter: %v", err)
	}

	if string(yaml) != "name: test\ndescription: test desc" {
		t.Errorf("YAML not extracted correctly: %s", string(yaml))
	}

	if string(markdown) != "Markdown content here" {
		t.Errorf("Markdown not extracted correctly: %s", string(markdown))
	}
}

// TestExtractFrontmatterUnclosed tests unclosed frontmatter
func TestExtractFrontmatterUnclosed(t *testing.T) {
	content := `---
name: test
description: test

Some content without closing delimiter`

	_, _, err := extractFrontmatter([]byte(content))
	if err != ErrNoFrontmatter {
		t.Errorf("Expected ErrNoFrontmatter, got %v", err)
	}
}

// TestAgentTypeString tests AgentType string representation
func TestAgentTypeString(t *testing.T) {
	tests := map[AgentType]string{
		TypeSubagent: "subagent",
		TypeSkill:    "skill",
		TypeCommand:  "command",
		TypePlugin:   "plugin",
	}

	for atype, expected := range tests {
		if atype.String() != expected {
			t.Errorf("Expected %s, got %s", expected, atype.String())
		}
	}
}

// TestAgentSourceString tests AgentSource string representation
func TestAgentSourceString(t *testing.T) {
	tests := map[AgentSource]string{
		SourceProject: "project",
		SourceUser:    "user",
		SourcePlugin:  "plugin",
		SourceCLI:     "cli",
	}

	for asource, expected := range tests {
		if asource.String() != expected {
			t.Errorf("Expected %s, got %s", expected, asource.String())
		}
	}
}

// TestDiscoveryResultIsEmpty tests the IsEmpty method
func TestDiscoveryResultIsEmpty(t *testing.T) {
	// Empty result
	result := &DiscoveryResult{Agents: make([]*Agent, 0)}
	if !result.IsEmpty() {
		t.Error("Expected IsEmpty() to return true for empty agents")
	}

	// Non-empty result
	result.Agents = append(result.Agents, &Agent{Name: "test"})
	result.Total = 1
	if result.IsEmpty() {
		t.Error("Expected IsEmpty() to return false for non-empty agents")
	}
}

// TestDiscoveryResultHasErrors tests the HasErrors method
func TestDiscoveryResultHasErrors(t *testing.T) {
	result := &DiscoveryResult{
		Agents: make([]*Agent, 0),
		Errors: make([]error, 0),
	}

	// No errors
	if result.HasErrors() {
		t.Error("Expected HasErrors() to return false when no errors")
	}

	// With errors
	result.Errors = append(result.Errors, ErrMissingName)
	result.ErrorCount = 1
	if !result.HasErrors() {
		t.Error("Expected HasErrors() to return true when errors present")
	}
}

// TestDiscovererDiscoverAllEmpty tests discovery in empty directory
func TestDiscovererDiscoverAllEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .adk/agents directory but leave it empty
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	discoverer := NewDiscoverer(tmpDir)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	if !result.IsEmpty() {
		t.Error("Expected empty result for empty agents directory")
	}

	if result.ErrorCount != 0 {
		t.Errorf("Expected 0 errors, got %d", result.ErrorCount)
	}
}

// TestDiscovererDiscoverAllSingleAgent tests discovering a single agent
func TestDiscovererDiscoverAllSingleAgent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .adk/agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create a single agent file
	agentContent := `---
name: agent-one
description: First test agent
---
# Agent One
Content here
`
	agentPath := filepath.Join(agentsDir, "agent-one.md")
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	discoverer := NewDiscoverer(tmpDir)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Expected 1 agent, got %d", result.Total)
	}

	if len(result.Agents) != 1 {
		t.Fatalf("Expected 1 agent in results, got %d", len(result.Agents))
	}

	agent := result.Agents[0]
	if agent.Name != "agent-one" {
		t.Errorf("Expected agent name 'agent-one', got '%s'", agent.Name)
	}

	if agent.Source != SourceProject {
		t.Errorf("Expected source SourceProject, got %s", agent.Source)
	}

	if agent.Type != TypeSubagent {
		t.Errorf("Expected type TypeSubagent, got %s", agent.Type)
	}
}

// TestDiscovererDiscoverAllMultipleAgents tests discovering multiple agents
func TestDiscovererDiscoverAllMultipleAgents(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .adk/agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create multiple agent files
	agents := []struct {
		name string
		file string
		desc string
	}{
		{"agent-one", "agent-one.md", "First agent"},
		{"agent-two", "agent-two.md", "Second agent"},
		{"agent-three", "agent-three.md", "Third agent"},
	}

	for _, a := range agents {
		content := `---
name: ` + a.name + `
description: ` + a.desc + `
---
# ` + a.name + `
`
		agentPath := filepath.Join(agentsDir, a.file)
		if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create agent file: %v", err)
		}
	}

	discoverer := NewDiscoverer(tmpDir)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected 3 agents, got %d", result.Total)
	}

	if len(result.Agents) != 3 {
		t.Fatalf("Expected 3 agents in results, got %d", len(result.Agents))
	}

	// Verify all agents were discovered
	names := make(map[string]bool)
	for _, agent := range result.Agents {
		names[agent.Name] = true
	}

	for _, a := range agents {
		if !names[a.name] {
			t.Errorf("Agent %s was not discovered", a.name)
		}
	}
}

// TestDiscovererDiscoverAllSkipsNonMarkdown tests that non-markdown files are skipped
func TestDiscovererDiscoverAllSkipsNonMarkdown(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .adk/agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create a markdown agent
	validContent := `---
name: valid-agent
description: A valid agent
---
Content
`
	validPath := filepath.Join(agentsDir, "valid.md")
	if err := os.WriteFile(validPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create valid agent: %v", err)
	}

	// Create a non-markdown file
	if err := os.WriteFile(filepath.Join(agentsDir, "readme.txt"), []byte("Not an agent"), 0644); err != nil {
		t.Fatalf("Failed to create non-markdown file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(agentsDir, "config.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create json file: %v", err)
	}

	discoverer := NewDiscoverer(tmpDir)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Expected 1 agent (only markdown), got %d", result.Total)
	}

	if result.Agents[0].Name != "valid-agent" {
		t.Errorf("Expected valid-agent to be discovered")
	}
}

// TestDiscovererDiscoverAllWithErrors tests that discovery continues on errors
func TestDiscovererDiscoverAllWithErrors(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .adk/agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create a valid agent
	validContent := `---
name: valid-agent
description: Valid
---
`
	if err := os.WriteFile(filepath.Join(agentsDir, "valid.md"), []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create valid agent: %v", err)
	}

	// Create an invalid agent (missing name)
	invalidContent := `---
description: No name field
---
`
	if err := os.WriteFile(filepath.Join(agentsDir, "invalid.md"), []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create invalid agent: %v", err)
	}

	discoverer := NewDiscoverer(tmpDir)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	// Should discover the valid agent
	if result.Total != 1 {
		t.Errorf("Expected 1 valid agent to be discovered, got %d", result.Total)
	}

	// Should have recorded the error for invalid agent
	if result.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount)
	}

	if !result.HasErrors() {
		t.Error("Expected HasErrors() to return true")
	}
}

// TestDiscovererDiscoverProjectAgents tests project-level agent discovery
func TestDiscovererDiscoverProjectAgents(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .adk/agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create an agent
	content := `---
name: project-agent
description: Agent at project level
---
`
	if err := os.WriteFile(filepath.Join(agentsDir, "project.md"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	discoverer := NewDiscoverer(tmpDir)
	result, err := discoverer.DiscoverProjectAgents()

	if err != nil {
		t.Fatalf("DiscoverProjectAgents() returned error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Expected 1 agent, got %d", result.Total)
	}
}

// TestDiscovererTimingInfo tests that timing information is captured
func TestDiscovererTimingInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .adk/agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create an agent
	content := `---
name: timed-agent
description: For timing test
---
`
	if err := os.WriteFile(filepath.Join(agentsDir, "timed.md"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	discoverer := NewDiscoverer(tmpDir)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	if result.TimeTaken == 0 {
		t.Error("Expected non-zero timing information")
	}

	if result.TimeTaken > 5*time.Second {
		t.Errorf("Discovery took unexpectedly long: %v", result.TimeTaken)
	}
}

// TestDiscovererMultiPathDiscoveryWithConfig tests discovery with explicit configuration
func TestDiscovererMultiPathDiscoveryWithConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project-level agents directory
	projectAgentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(projectAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create project agents directory: %v", err)
	}

	// Create user-level agents directory
	userAgentsDir := filepath.Join(tmpDir, "user-agents")
	if err := os.MkdirAll(userAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create user agents directory: %v", err)
	}

	// Create project agent
	projectAgent := `---
name: project-agent
description: Agent at project level
---
`
	if err := os.WriteFile(filepath.Join(projectAgentsDir, "project.md"), []byte(projectAgent), 0644); err != nil {
		t.Fatalf("Failed to create project agent: %v", err)
	}

	// Create user agent
	userAgent := `---
name: user-agent
description: Agent at user level
---
`
	if err := os.WriteFile(filepath.Join(userAgentsDir, "user.md"), []byte(userAgent), 0644); err != nil {
		t.Fatalf("Failed to create user agent: %v", err)
	}

	// Create configuration
	cfg := NewConfig()
	cfg.ProjectPath = filepath.Join(tmpDir, ".adk", "agents")
	cfg.UserPath = filepath.Join(tmpDir, "user-agents")
	cfg.SearchOrder = []string{"project", "user"}

	// Discover with config
	discoverer := NewDiscovererWithConfig(tmpDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	// Should find both agents
	if result.Total != 2 {
		t.Errorf("Expected 2 agents, got %d", result.Total)
	}

	// Verify agent names
	names := make(map[string]bool)
	for _, agent := range result.Agents {
		names[agent.Name] = true
	}

	if !names["project-agent"] {
		t.Error("Expected to find project-agent")
	}

	if !names["user-agent"] {
		t.Error("Expected to find user-agent")
	}
}

// TestDiscovererMultiPathDeduplication tests that duplicate agent names are handled
func TestDiscovererMultiPathDeduplication(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project-level agents directory
	projectAgentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(projectAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create project agents directory: %v", err)
	}

	// Create user-level agents directory
	userAgentsDir := filepath.Join(tmpDir, "user-agents")
	if err := os.MkdirAll(userAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create user agents directory: %v", err)
	}

	// Create same-named agent in both directories
	agentContent := `---
name: shared-agent
description: This agent exists in both locations
---
`
	if err := os.WriteFile(filepath.Join(projectAgentsDir, "shared.md"), []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create project agent: %v", err)
	}

	if err := os.WriteFile(filepath.Join(userAgentsDir, "shared.md"), []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create user agent: %v", err)
	}

	// Create configuration with project first
	cfg := NewConfig()
	cfg.ProjectPath = filepath.Join(tmpDir, ".adk", "agents")
	cfg.UserPath = filepath.Join(tmpDir, "user-agents")
	cfg.SearchOrder = []string{"project", "user"}

	// Discover with config
	discoverer := NewDiscovererWithConfig(tmpDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	// Should only find one agent (deduplicated)
	if result.Total != 1 {
		t.Errorf("Expected 1 agent (deduplicated), got %d", result.Total)
	}

	// Should be from project source (discovered first)
	if result.Agents[0].Source != SourceProject {
		t.Errorf("Expected agent from SourceProject, got %s", result.Agents[0].Source)
	}
}

// TestDiscovererMissingPathWithSkip tests SkipMissing flag behavior
func TestDiscovererMissingPathWithSkip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only project directory, not user directory
	projectAgentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(projectAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create project agents directory: %v", err)
	}

	// Create project agent
	projectAgent := `---
name: project-agent
description: Agent at project level
---
`
	if err := os.WriteFile(filepath.Join(projectAgentsDir, "project.md"), []byte(projectAgent), 0644); err != nil {
		t.Fatalf("Failed to create project agent: %v", err)
	}

	// Create configuration with non-existent user path and SkipMissing=true
	cfg := NewConfig()
	cfg.ProjectPath = filepath.Join(tmpDir, ".adk", "agents")
	cfg.UserPath = filepath.Join(tmpDir, "nonexistent-user-agents")
	cfg.SkipMissing = true
	cfg.SearchOrder = []string{"project", "user"}

	// Discover with config
	discoverer := NewDiscovererWithConfig(tmpDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	// Should find project agent and have no errors
	if result.Total != 1 {
		t.Errorf("Expected 1 agent, got %d", result.Total)
	}

	if result.ErrorCount != 0 {
		t.Errorf("Expected 0 errors with SkipMissing=true, got %d", result.ErrorCount)
	}
}

// TestDiscovererMissingPathWithoutSkip tests error handling for missing paths
func TestDiscovererMissingPathWithoutSkip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only project directory, not user directory
	projectAgentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(projectAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create project agents directory: %v", err)
	}

	// Create project agent
	projectAgent := `---
name: project-agent
description: Agent at project level
---
`
	if err := os.WriteFile(filepath.Join(projectAgentsDir, "project.md"), []byte(projectAgent), 0644); err != nil {
		t.Fatalf("Failed to create project agent: %v", err)
	}

	// Create configuration with non-existent user path and SkipMissing=false
	cfg := NewConfig()
	cfg.ProjectPath = filepath.Join(tmpDir, ".adk", "agents")
	cfg.UserPath = filepath.Join(tmpDir, "nonexistent-user-agents")
	cfg.SkipMissing = false
	cfg.SearchOrder = []string{"project", "user"}

	// Discover with config
	discoverer := NewDiscovererWithConfig(tmpDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	// Should find project agent but record error for missing user path
	if result.Total != 1 {
		t.Errorf("Expected 1 agent, got %d", result.Total)
	}

	if result.ErrorCount != 1 {
		t.Errorf("Expected 1 error for missing path, got %d", result.ErrorCount)
	}

	if !result.HasErrors() {
		t.Error("Expected HasErrors() to return true")
	}
}

// TestDiscovererSourceAttribution tests that source is correctly assigned
func TestDiscovererSourceAttribution(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project and user directories
	projectAgentsDir := filepath.Join(tmpDir, ".adk", "agents")
	userAgentsDir := filepath.Join(tmpDir, "user-agents")

	if err := os.MkdirAll(projectAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create project agents directory: %v", err)
	}
	if err := os.MkdirAll(userAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create user agents directory: %v", err)
	}

	// Create project agent
	projectAgent := `---
name: project-agent
description: Project level agent
---
`
	if err := os.WriteFile(filepath.Join(projectAgentsDir, "project.md"), []byte(projectAgent), 0644); err != nil {
		t.Fatalf("Failed to create project agent: %v", err)
	}

	// Create user agent
	userAgent := `---
name: user-agent
description: User level agent
---
`
	if err := os.WriteFile(filepath.Join(userAgentsDir, "user.md"), []byte(userAgent), 0644); err != nil {
		t.Fatalf("Failed to create user agent: %v", err)
	}

	// Create configuration
	cfg := NewConfig()
	cfg.ProjectPath = filepath.Join(tmpDir, ".adk", "agents")
	cfg.UserPath = filepath.Join(tmpDir, "user-agents")
	cfg.SearchOrder = []string{"project", "user"}

	// Discover with config
	discoverer := NewDiscovererWithConfig(tmpDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	// Check sources
	sourceMap := make(map[string]AgentSource)
	for _, agent := range result.Agents {
		sourceMap[agent.Name] = agent.Source
	}

	if sourceMap["project-agent"] != SourceProject {
		t.Errorf("Expected project-agent source to be SourceProject, got %s", sourceMap["project-agent"])
	}

	if sourceMap["user-agent"] != SourceUser {
		t.Errorf("Expected user-agent source to be SourceUser, got %s", sourceMap["user-agent"])
	}
}
