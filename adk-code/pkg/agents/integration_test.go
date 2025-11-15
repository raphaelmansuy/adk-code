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

// TestIntegrationMetadataDiscovery tests that metadata is properly parsed and retrieved
func TestIntegrationMetadataDiscovery(t *testing.T) {
	tmpDir := t.TempDir()

	// Create agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create agent with full metadata
	agentWithMetadata := `---
name: comprehensive-agent
description: An agent with full metadata
version: 1.2.3
author: test@example.com
tags: [refactoring, python, testing]
dependencies: [base-agent, config-manager]
---
# Comprehensive Agent

This agent has complete metadata.
`
	if err := os.WriteFile(filepath.Join(agentsDir, "comprehensive.md"), []byte(agentWithMetadata), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Create agent with minimal metadata
	minimalAgent := `---
name: minimal-agent
description: An agent with no extra metadata
---
# Minimal Agent

This agent has only required fields.
`
	if err := os.WriteFile(filepath.Join(agentsDir, "minimal.md"), []byte(minimalAgent), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Discover agents
	cfg := NewConfig()
	cfg.ProjectPath = agentsDir

	discoverer := NewDiscovererWithConfig(tmpDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("Expected 2 agents, got %d", result.Total)
	}

	// Find comprehensive agent and verify metadata
	var comprehensiveAgent *Agent
	for _, agent := range result.Agents {
		if agent.Name == "comprehensive-agent" {
			comprehensiveAgent = agent
			break
		}
	}

	if comprehensiveAgent == nil {
		t.Fatal("Expected to find comprehensive-agent")
	}

	// Verify version
	if comprehensiveAgent.Version != "1.2.3" {
		t.Errorf("Expected version 1.2.3, got %s", comprehensiveAgent.Version)
	}

	// Verify author
	if comprehensiveAgent.Author != "test@example.com" {
		t.Errorf("Expected author test@example.com, got %s", comprehensiveAgent.Author)
	}

	// Verify tags
	if len(comprehensiveAgent.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(comprehensiveAgent.Tags))
	}

	expectedTags := map[string]bool{"refactoring": true, "python": true, "testing": true}
	for _, tag := range comprehensiveAgent.Tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}

	// Verify dependencies
	if len(comprehensiveAgent.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(comprehensiveAgent.Dependencies))
	}

	expectedDeps := map[string]bool{"base-agent": true, "config-manager": true}
	for _, dep := range comprehensiveAgent.Dependencies {
		if !expectedDeps[dep] {
			t.Errorf("Unexpected dependency: %s", dep)
		}
	}

	// Verify minimal agent has empty metadata fields
	var minimalAgentFound *Agent
	for _, agent := range result.Agents {
		if agent.Name == "minimal-agent" {
			minimalAgentFound = agent
			break
		}
	}

	if minimalAgentFound == nil {
		t.Fatal("Expected to find minimal-agent")
	}

	if minimalAgentFound.Version != "" {
		t.Error("Expected empty version for minimal agent")
	}

	if minimalAgentFound.Author != "" {
		t.Error("Expected empty author for minimal agent")
	}

	if len(minimalAgentFound.Tags) > 0 {
		t.Error("Expected no tags for minimal agent")
	}

	if len(minimalAgentFound.Dependencies) > 0 {
		t.Error("Expected no dependencies for minimal agent")
	}
}

// TestIntegrationTagFiltering tests filtering agents by tags
func TestIntegrationTagFiltering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create agents with different tags
	agents := []struct {
		name string
		file string
		tags string
	}{
		{"python-linter", "python_linter.md", "python, linting, code-quality"},
		{"go-formatter", "go_formatter.md", "golang, formatting, code-quality"},
		{"docs-generator", "docs_gen.md", "documentation, markdown"},
	}

	for _, a := range agents {
		content := `---
name: ` + a.name + `
description: Test agent
tags: [` + a.tags + `]
---
`
		if err := os.WriteFile(filepath.Join(agentsDir, a.file), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create agent: %v", err)
		}
	}

	// Discover all agents
	cfg := NewConfig()
	cfg.ProjectPath = agentsDir

	discoverer := NewDiscovererWithConfig(tmpDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected 3 agents, got %d", result.Total)
	}

	// Count agents with "code-quality" tag
	codeQualityCount := 0
	for _, agent := range result.Agents {
		for _, tag := range agent.Tags {
			if tag == "code-quality" {
				codeQualityCount++
				break
			}
		}
	}

	if codeQualityCount != 2 {
		t.Errorf("Expected 2 agents with code-quality tag, got %d", codeQualityCount)
	}

	// Count agents with "documentation" tag
	docCount := 0
	for _, agent := range result.Agents {
		for _, tag := range agent.Tags {
			if tag == "documentation" {
				docCount++
				break
			}
		}
	}

	if docCount != 1 {
		t.Errorf("Expected 1 agent with documentation tag, got %d", docCount)
	}
}

// TestIntegrationAuthorFiltering tests filtering agents by author
func TestIntegrationAuthorFiltering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create agents with different authors
	agents := []struct {
		name   string
		file   string
		author string
	}{
		{"alice-agent", "alice.md", "alice@example.com"},
		{"bob-agent", "bob.md", "bob@example.com"},
		{"alice-tool", "alice_tool.md", "alice@example.com"},
	}

	for _, a := range agents {
		content := `---
name: ` + a.name + `
description: Test agent
author: ` + a.author + `
---
`
		if err := os.WriteFile(filepath.Join(agentsDir, a.file), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create agent: %v", err)
		}
	}

	// Discover all agents
	cfg := NewConfig()
	cfg.ProjectPath = agentsDir

	discoverer := NewDiscovererWithConfig(tmpDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected 3 agents, got %d", result.Total)
	}

	// Count agents by author
	aliceCount := 0
	bobCount := 0

	for _, agent := range result.Agents {
		switch agent.Author {
		case "alice@example.com":
			aliceCount++
		case "bob@example.com":
			bobCount++
		}
	}

	if aliceCount != 2 {
		t.Errorf("Expected 2 agents from alice@example.com, got %d", aliceCount)
	}

	if bobCount != 1 {
		t.Errorf("Expected 1 agent from bob@example.com, got %d", bobCount)
	}
}

// TestIntegrationDependencyChains tests handling of agent dependencies
func TestIntegrationDependencyChains(t *testing.T) {
	tmpDir := t.TempDir()

	// Create agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create agents with dependency chains
	// base-agent (no deps)
	// -> mid-agent (depends on base-agent)
	//    -> top-agent (depends on mid-agent)

	baseAgent := `---
name: base-agent
description: Base agent with no dependencies
---
`
	midAgent := `---
name: mid-agent
description: Middle agent
dependencies: [base-agent]
---
`
	topAgent := `---
name: top-agent
description: Top agent
dependencies: [mid-agent, base-agent]
---
`

	if err := os.WriteFile(filepath.Join(agentsDir, "base.md"), []byte(baseAgent), 0644); err != nil {
		t.Fatalf("Failed to create base agent: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "mid.md"), []byte(midAgent), 0644); err != nil {
		t.Fatalf("Failed to create mid agent: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "top.md"), []byte(topAgent), 0644); err != nil {
		t.Fatalf("Failed to create top agent: %v", err)
	}

	// Discover all agents
	cfg := NewConfig()
	cfg.ProjectPath = agentsDir

	discoverer := NewDiscovererWithConfig(tmpDir, cfg)
	result, err := discoverer.DiscoverAll()

	if err != nil {
		t.Fatalf("DiscoverAll() returned error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected 3 agents, got %d", result.Total)
	}

	// Build dependency map
	depMap := make(map[string][]string)
	for _, agent := range result.Agents {
		depMap[agent.Name] = agent.Dependencies
	}

	// Verify dependency chains
	baseDeps, hasBase := depMap["base-agent"]
	if !hasBase || len(baseDeps) != 0 {
		t.Error("Expected base-agent to have no dependencies")
	}

	midDeps, hasMid := depMap["mid-agent"]
	if !hasMid || len(midDeps) != 1 || midDeps[0] != "base-agent" {
		t.Error("Expected mid-agent to depend on base-agent")
	}

	topDeps, hasTop := depMap["top-agent"]
	if !hasTop || len(topDeps) != 2 {
		t.Error("Expected top-agent to have 2 dependencies")
	}
}

// TestE2EDependencyChainResolution tests dependency resolution for a chain.
func TestE2EDependencyChainResolution(t *testing.T) {
	graph := NewDependencyGraph()

	// Create agent chain: api -> db -> config
	config := &Agent{
		Name:         "config",
		Version:      "1.0.0",
		Dependencies: []string{},
	}

	db := &Agent{
		Name:         "db",
		Version:      "2.0.0",
		Dependencies: []string{"config"},
	}

	api := &Agent{
		Name:         "api",
		Version:      "1.5.0",
		Dependencies: []string{"db"},
	}

	graph.AddAgent(config)
	graph.AddAgent(db)
	graph.AddAgent(api)

	graph.AddEdge("db", "config")
	graph.AddEdge("api", "db")

	// Resolve dependencies
	resolved, err := graph.ResolveDependencies("api")
	if err != nil {
		t.Fatalf("Failed to resolve: %v", err)
	}

	if len(resolved) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(resolved))
	}

	// Verify order: config, db, api
	if resolved[0].Name != "config" {
		t.Errorf("First should be config, got %q", resolved[0].Name)
	}
	if resolved[1].Name != "db" {
		t.Errorf("Second should be db, got %q", resolved[1].Name)
	}
	if resolved[2].Name != "api" {
		t.Errorf("Third should be api, got %q", resolved[2].Name)
	}
}

// TestE2ECyclicDependencyDetection tests cycle detection.
func TestE2ECyclicDependencyDetection(t *testing.T) {
	graph := NewDependencyGraph()

	a := &Agent{Name: "a", Dependencies: []string{"b"}}
	b := &Agent{Name: "b", Dependencies: []string{"c"}}
	c := &Agent{Name: "c", Dependencies: []string{"a"}}

	graph.AddAgent(a)
	graph.AddAgent(b)
	graph.AddAgent(c)

	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("c", "a")

	// Should detect cycle
	_, err := graph.ResolveDependencies("a")
	if err == nil {
		t.Error("Expected cycle detection error")
	}
}

// TestE2EVersionMatching tests semantic version matching.
func TestE2EVersionMatching(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		constraint  string
		shouldMatch bool
	}{
		{"exact", "1.0.0", "==1.0.0", true},
		{"greater", "2.0.0", ">1.0.0", true},
		{"caret", "1.5.0", "^1.0.0", true},
		{"tilde", "1.0.5", "~1.0.0", true},
		{"range", "1.5.0", "1.0.0-2.0.0", true},
		{"mismatch", "0.9.0", ">1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("Parse version failed: %v", err)
			}

			con, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("Parse constraint failed: %v", err)
			}

			matches := con.Matches(ver)
			if matches != tt.shouldMatch {
				t.Errorf("Expected %v, got %v", tt.shouldMatch, matches)
			}
		})
	}
}

// TestE2EValidatorIntegration tests AgentMetadataValidator.
func TestE2EValidatorIntegration(t *testing.T) {
	validator := NewAgentMetadataValidator()

	base := &Agent{
		Name:         "base",
		Description:  "Base agent",
		Version:      "1.0.0",
		Dependencies: []string{},
	}

	derived := &Agent{
		Name:         "derived",
		Description:  "Derived agent",
		Version:      "1.0.0",
		Dependencies: []string{"base"},
	}

	// Add agents with versions
	validator.AddAgent(base, "^1.0.0")
	validator.AddAgent(derived, "^1.0.0")

	// Add dependency
	validator.AddDependency("derived", "base")

	// Validate base agent
	report, err := validator.ValidateAgent("base")
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	if !report.Valid {
		t.Errorf("Expected valid, got issues: %v", report.Issues)
	}

	// Validate derived agent
	report2, err := validator.ValidateAgent("derived")
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	if !report2.Valid {
		t.Errorf("Expected valid, got issues: %v", report2.Issues)
	}
}

// TestE2EComplexGraphResolution tests complex multi-level dependencies.
func TestE2EComplexGraphResolution(t *testing.T) {
	graph := NewDependencyGraph()

	// Create complex graph
	//        api
	//       /   \
	//      /     \
	//   cache    db
	//     |     / \
	//     |    /   \
	//   logging config

	agents := map[string]*Agent{
		"logging": {Name: "logging", Version: "1.0.0", Dependencies: []string{}},
		"config":  {Name: "config", Version: "1.0.0", Dependencies: []string{}},
		"cache":   {Name: "cache", Version: "1.0.0", Dependencies: []string{"logging"}},
		"db":      {Name: "db", Version: "1.0.0", Dependencies: []string{"config"}},
		"api":     {Name: "api", Version: "1.0.0", Dependencies: []string{"cache", "db"}},
	}

	for _, a := range agents {
		graph.AddAgent(a)
	}

	// Add edges
	edges := [][2]string{
		{"cache", "logging"},
		{"db", "config"},
		{"api", "cache"},
		{"api", "db"},
	}

	for _, e := range edges {
		graph.AddEdge(e[0], e[1])
	}

	// Resolve
	resolved, err := graph.ResolveDependencies("api")
	if err != nil {
		t.Fatalf("Resolution failed: %v", err)
	}

	if len(resolved) != 5 {
		t.Errorf("Expected 5 agents, got %d", len(resolved))
	}

	// Verify ordering
	positions := make(map[string]int)
	for i, a := range resolved {
		positions[a.Name] = i
	}

	// Check each agent comes after its dependencies
	for _, a := range resolved {
		for _, dep := range a.Dependencies {
			if positions[dep] >= positions[a.Name] {
				t.Errorf("Ordering violation: %q depends on %q", a.Name, dep)
			}
		}
	}
}

// TestE2ETransitiveDependencies tests transitive dependency collection.
func TestE2ETransitiveDependencies(t *testing.T) {
	graph := NewDependencyGraph()

	// Chain: d -> c -> b -> a
	a := &Agent{Name: "a"}
	b := &Agent{Name: "b"}
	c := &Agent{Name: "c"}
	d := &Agent{Name: "d"}

	graph.AddAgent(a)
	graph.AddAgent(b)
	graph.AddAgent(c)
	graph.AddAgent(d)

	graph.AddEdge("b", "a")
	graph.AddEdge("c", "b")
	graph.AddEdge("d", "c")

	// Get transitive of d
	trans, err := graph.GetTransitiveDeps("d")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if len(trans) != 3 {
		t.Errorf("Expected 3 transitive deps, got %d: %v", len(trans), trans)
	}

	expected := map[string]bool{"a": true, "b": true, "c": true}
	for _, dep := range trans {
		if !expected[dep] {
			t.Errorf("Unexpected dep: %q", dep)
		}
	}
}

// TestE2EVersionParsing tests version parsing and string representation.
func TestE2EVersionParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "1.0.0", "1.0.0"},
		{"prerelease", "1.0.0-alpha", "1.0.0-alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseVersion(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if v.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, v.String())
			}
		})
	}
}

// TestE2EConstraintParsing tests constraint parsing.
func TestE2EConstraintParsing(t *testing.T) {
	constraints := []string{
		"==1.0.0",
		">1.0.0",
		">=1.0.0",
		"<2.0.0",
		"<=2.0.0",
		"^1.0.0",
		"~1.0.0",
		"1.0.0-2.0.0",
	}

	for _, c := range constraints {
		_, err := ParseConstraint(c)
		if err != nil {
			t.Errorf("Failed to parse constraint %q: %v", c, err)
		}
	}
}
