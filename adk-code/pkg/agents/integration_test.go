// Package agents integration tests
// These tests verify the entire agent discovery system works end-to-end
package agents

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIntegrationFullDiscoveryWorkflow tests the complete discovery workflow
// Project → User → Plugin paths with configuration and deduplication
func TestIntegrationFullDiscoveryWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup: Create project structure with agents at multiple levels
	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create .adk directory in project
	adkDir := filepath.Join(projectDir, ".adk")
	if err := os.MkdirAll(adkDir, 0755); err != nil {
		t.Fatalf("Failed to create .adk directory: %v", err)
	}

	// Create project-level agents
	projectAgentsDir := filepath.Join(adkDir, "agents")
	if err := os.MkdirAll(projectAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create project agents directory: %v", err)
	}

	// Create user-level agents directory
	userAgentsDir := filepath.Join(tmpDir, "user-agents")
	if err := os.MkdirAll(userAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create user agents directory: %v", err)
	}

	// Create plugin-level agents directory
	pluginAgentsDir := filepath.Join(tmpDir, "plugin-agents")
	if err := os.MkdirAll(pluginAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create plugin agents directory: %v", err)
	}

	// Define agents at each level
	agents := []struct {
		name        string
		file        string // Filename (relative to level directory)
		description string
		level       string // "project", "user", "plugin"
	}{
		{"project-core", "project_core.md", "Core project agent", "project"},
		{"project-utils", "project_utils.md", "Project utilities agent", "project"},
		{"shared-agent", "shared_agent.md", "This exists at multiple levels", "project"}, // Will be overridden by user/plugin
		{"project-nested", "subdir/project_nested.md", "Nested project agent", "project"},

		{"user-shared", "user_shared.md", "Shared user agent", "user"},
		{"shared-agent", "shared_agent.md", "User version of shared agent", "user"}, // Duplicate name
		{"user-nested", "subdir/user_nested.md", "Nested user agent", "user"},

		{"plugin-special", "plugin_special.md", "Special plugin agent", "plugin"},
		{"shared-agent", "shared_agent.md", "Plugin version of shared agent", "plugin"}, // Duplicate name
	}

	// Create actual agent files
	for _, a := range agents {
		content := `---
name: ` + a.name + `
description: ` + a.description + `
---
# ` + a.name + `

This is a test agent.
`

		var dir string
		switch a.level {
		case "project":
			dir = projectAgentsDir
		case "user":
			dir = userAgentsDir
		case "plugin":
			dir = pluginAgentsDir
		}

		// Create any subdirectories
		filePath := filepath.Join(dir, a.file)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create agent file: %v", err)
		}
	}

	// Create .adk/config.yaml in project
	configContent := `agent:
  skip_missing: true
search_order:
  - project
  - user
  - plugin
`
	configPath := filepath.Join(adkDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config.yaml: %v", err)
	}

	// Load configuration
	cfg, err := LoadConfig(projectDir)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Manually set paths since config file doesn't contain them
	cfg.ProjectPath = projectAgentsDir
	cfg.UserPath = userAgentsDir
	cfg.PluginPaths = []string{pluginAgentsDir}

	// Discover agents
	discoverer := NewDiscovererWithConfig(projectDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	// Verify results
	// Should have: project-core, project-utils, shared-agent (from project),
	//             project-nested, user-shared, user-nested, plugin-special
	// Total: 7 agents (shared-agent deduplicated, exists 3 times but counted once)
	expectedCount := 7
	if result.Total != expectedCount {
		t.Errorf("Expected %d agents, got %d", expectedCount, result.Total)
		t.Logf("Discovered agents:")
		for i, agent := range result.Agents {
			t.Logf("  %d. %s (source: %s)", i+1, agent.Name, agent.Source)
		}
	}

	// Verify shared-agent came from project source (first in search order)
	var sharedAgent *Agent
	for _, agent := range result.Agents {
		if agent.Name == "shared-agent" {
			sharedAgent = agent
			break
		}
	}

	if sharedAgent == nil {
		t.Error("Expected to find shared-agent")
	} else if sharedAgent.Source != SourceProject {
		t.Errorf("Expected shared-agent from SourceProject, got %s", sharedAgent.Source)
	}

	// Verify no errors (SkipMissing should be true from config)
	if result.HasErrors() {
		t.Errorf("Expected no errors, got %d", result.ErrorCount)
		for _, err := range result.Errors {
			t.Logf("  Error: %v", err)
		}
	}
}

// TestIntegrationConfigurationOverrides tests that environment variables override config file
func TestIntegrationConfigurationOverrides(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project directory
	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create .adk/agents directory
	projectAgentsDir := filepath.Join(projectDir, ".adk", "agents")
	if err := os.MkdirAll(projectAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create project agents directory: %v", err)
	}

	// Create test agent
	agentContent := `---
name: test-agent
description: Test agent
---
`
	if err := os.WriteFile(filepath.Join(projectAgentsDir, "test.md"), []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Create config file with one path
	configContent := `agent:
  skip_missing: false
search_order:
  - project
`
	configPath := filepath.Join(projectDir, ".adk", "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Test 1: Load without environment overrides
	cfg, err := LoadConfig(projectDir)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	if cfg.SkipMissing {
		t.Error("Expected SkipMissing to be false from config file")
	}

	// Test 2: Simulate environment variable override
	t.Setenv("ADK_AGENT_SKIP_MISSING", "true")
	cfg, err = LoadConfig(projectDir)
	if err != nil {
		t.Fatalf("Failed to load configuration with env override: %v", err)
	}

	if !cfg.SkipMissing {
		t.Error("Expected SkipMissing to be true (overridden by env var)")
	}
}

// TestIntegrationNestedDirectoryStructure tests discovery with nested agent directories
func TestIntegrationNestedDirectoryStructure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create agents directory with nested structure
	agentsDir := filepath.Join(tmpDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create nested directories
	subdirs := []string{
		"core",
		"utils/helpers",
		"plugins/data",
	}

	agentFiles := []struct {
		relPath string
		name    string
		desc    string
	}{
		{"agent1.md", "agent-1", "Root level agent"},
		{"core/agent2.md", "agent-2", "Core agent"},
		{"utils/helpers/agent3.md", "agent-3", "Helper agent"},
		{"plugins/data/agent4.md", "agent-4", "Data plugin"},
	}

	// Create subdirectories
	for _, subdir := range subdirs {
		fullPath := filepath.Join(agentsDir, subdir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			t.Fatalf("Failed to create subdirectory %s: %v", subdir, err)
		}
	}

	// Create agent files
	for _, af := range agentFiles {
		content := `---
name: ` + af.name + `
description: ` + af.desc + `
---
`
		filePath := filepath.Join(agentsDir, af.relPath)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create agent file: %v", err)
		}
	}

	// Discover agents
	cfg := NewConfig()
	cfg.ProjectPath = agentsDir
	cfg.SearchOrder = []string{"project"}

	discoverer := NewDiscovererWithConfig(tmpDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	// Should discover all 4 agents regardless of nesting
	if result.Total != 4 {
		t.Errorf("Expected 4 agents, got %d", result.Total)
	}

	if result.ErrorCount > 0 {
		t.Errorf("Unexpected errors during discovery: %d", result.ErrorCount)
	}

	// Verify all agents were found
	names := make(map[string]bool)
	for _, agent := range result.Agents {
		names[agent.Name] = true
	}

	expectedNames := []string{"agent-1", "agent-2", "agent-3", "agent-4"}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("Expected to find %s", name)
		}
	}
}

// TestIntegrationErrorHandling tests error handling across multiple paths
func TestIntegrationErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project directory
	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create project-level agents
	projectAgentsDir := filepath.Join(projectDir, ".adk", "agents")
	if err := os.MkdirAll(projectAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create project agents directory: %v", err)
	}

	// Create user-level agents
	userAgentsDir := filepath.Join(tmpDir, "user-agents")
	if err := os.MkdirAll(userAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create user agents directory: %v", err)
	}

	// Create valid agent in project
	validAgent := `---
name: valid-agent
description: A valid agent
---
`
	if err := os.WriteFile(filepath.Join(projectAgentsDir, "valid.md"), []byte(validAgent), 0644); err != nil {
		t.Fatalf("Failed to create valid agent: %v", err)
	}

	// Create invalid agent in user directory (missing name)
	invalidAgent := `---
description: Missing name field
---
`
	if err := os.WriteFile(filepath.Join(userAgentsDir, "invalid.md"), []byte(invalidAgent), 0644); err != nil {
		t.Fatalf("Failed to create invalid agent: %v", err)
	}

	// Create configuration
	cfg := NewConfig()
	cfg.ProjectPath = projectAgentsDir
	cfg.UserPath = userAgentsDir
	cfg.SkipMissing = true
	cfg.SearchOrder = []string{"project", "user"}

	// Discover agents
	discoverer := NewDiscovererWithConfig(projectDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	// Should have found valid agent but recorded error for invalid one
	if result.Total != 1 {
		t.Errorf("Expected 1 valid agent, got %d", result.Total)
	}

	if result.ErrorCount != 1 {
		t.Errorf("Expected 1 error for invalid agent, got %d", result.ErrorCount)
	}

	if !result.HasErrors() {
		t.Error("Expected HasErrors() to return true")
	}

	if result.Agents[0].Name != "valid-agent" {
		t.Errorf("Expected valid-agent to be discovered")
	}
}
