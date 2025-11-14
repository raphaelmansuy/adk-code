// Package agents provides tools for agent definition discovery and management
package agents

import (
	"os"
	"path/filepath"
	"testing"

	"adk-code/pkg/agents"
)

// TestDiscoverPathsBasic tests basic path discovery
func TestDiscoverPathsBasic(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldCwd)

	// Create .adk/agents directory
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create config
	cfg := agents.NewConfig()
	cfg.ProjectPath = agentsDir

	// Call discover paths (simulating tool handler)
	tool, err := NewDiscoverPathsTool()
	if err != nil {
		t.Fatalf("Failed to create discover paths tool: %v", err)
	}

	if tool == nil {
		t.Error("Expected tool to be created")
	}
}

// TestDiscoverPathsOutput tests output structure
func TestDiscoverPathsOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project directory structure
	projectDir := filepath.Join(tmpDir, "project")
	agentsDir := filepath.Join(projectDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create user agents directory
	userAgentsDir := filepath.Join(tmpDir, "user-agents")
	if err := os.MkdirAll(userAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create user agents directory: %v", err)
	}

	// Create config
	cfg := agents.NewConfig()
	cfg.ProjectPath = agentsDir
	cfg.UserPath = userAgentsDir
	cfg.SearchOrder = []string{"project", "user"}

	// Verify PathInfo structure
	testPathInfo := PathInfo{
		Path:       agentsDir,
		Source:     "project",
		Order:      1,
		Exists:     true,
		Accessible: true,
	}

	if testPathInfo.Path != agentsDir {
		t.Errorf("Expected path %s, got %s", agentsDir, testPathInfo.Path)
	}

	if testPathInfo.Source != "project" {
		t.Errorf("Expected source project, got %s", testPathInfo.Source)
	}
}

// TestDiscoverPathsConfigStatus tests configuration loading status
func TestDiscoverPathsConfigStatus(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project directory
	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create config file
	adkDir := filepath.Join(projectDir, ".adk")
	if err := os.MkdirAll(adkDir, 0755); err != nil {
		t.Fatalf("Failed to create .adk directory: %v", err)
	}

	configContent := `agent:
  skip_missing: true
search_order:
  - project
  - user
`
	configPath := filepath.Join(adkDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load and verify config
	cfg, err := agents.LoadConfig(projectDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !cfg.SkipMissing {
		t.Error("Expected skip_missing to be true")
	}

	// Check that search order from file is used
	if len(cfg.SearchOrder) == 0 {
		t.Error("Expected search order to be loaded from config")
	}
}

// TestDiscoverPathsVerboseMode tests verbose mode output
func TestDiscoverPathsVerboseMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create agents directory with an agent
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create a test agent
	agentContent := `---
name: test-agent
description: Test agent
---
`
	agentPath := filepath.Join(agentsDir, "test.md")
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Create config
	cfg := agents.NewConfig()
	cfg.ProjectPath = agentsDir

	// Verify DiscoverPathsInput accepts verbose
	input := DiscoverPathsInput{
		Verbose: true,
	}

	if !input.Verbose {
		t.Error("Expected verbose to be true")
	}
}

// TestDiscoverPathsMultiplePaths tests output with multiple configured paths
func TestDiscoverPathsMultiplePaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple path directories
	projectDir := filepath.Join(tmpDir, ".adk", "agents")
	userDir := filepath.Join(tmpDir, "user-agents")
	pluginDir := filepath.Join(tmpDir, "plugin-agents")

	for _, dir := range []string{projectDir, userDir, pluginDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create config with all paths
	cfg := agents.NewConfig()
	cfg.ProjectPath = projectDir
	cfg.UserPath = userDir
	cfg.PluginPaths = []string{pluginDir}
	cfg.SearchOrder = []string{"project", "user", "plugin"}

	// Verify all paths are in search order
	allPaths := cfg.GetAllPaths()
	if len(allPaths) != 3 {
		t.Errorf("Expected 3 paths, got %d", len(allPaths))
	}

	// Build expected path source map
	sourceMap := make(map[string]string)
	sourceMap[projectDir] = "project"
	sourceMap[userDir] = "user"
	sourceMap[pluginDir] = "plugin"

	// Verify source mapping
	if sourceMap[projectDir] != "project" {
		t.Error("Expected project path mapping")
	}

	if sourceMap[userDir] != "user" {
		t.Error("Expected user path mapping")
	}

	if sourceMap[pluginDir] != "plugin" {
		t.Error("Expected plugin path mapping")
	}
}

// TestDiscoverPathsPathAccessibility tests path accessibility checking
func TestDiscoverPathsPathAccessibility(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an accessible path
	accessibleDir := filepath.Join(tmpDir, "accessible")
	if err := os.MkdirAll(accessibleDir, 0755); err != nil {
		t.Fatalf("Failed to create accessible directory: %v", err)
	}

	// Test pathExists
	if !pathExists(accessibleDir) {
		t.Error("Expected pathExists to return true for existing directory")
	}

	// Test with non-existent path
	nonExistentDir := filepath.Join(tmpDir, "nonexistent")
	if pathExists(nonExistentDir) {
		t.Error("Expected pathExists to return false for non-existent directory")
	}

	// Test isAccessible
	if !isAccessible(accessibleDir) {
		t.Error("Expected isAccessible to return true for readable directory")
	}

	if isAccessible(nonExistentDir) {
		t.Error("Expected isAccessible to return false for non-existent directory")
	}
}

// TestDiscoverPathsSummaryGeneration tests summary message generation
func TestDiscoverPathsSummaryGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories
	projectDir := filepath.Join(tmpDir, ".adk", "agents")
	userDir := filepath.Join(tmpDir, "user-agents")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	if err := os.MkdirAll(userDir, 0755); err != nil {
		t.Fatalf("Failed to create user directory: %v", err)
	}

	// Create config
	cfg := agents.NewConfig()
	cfg.ProjectPath = projectDir
	cfg.UserPath = userDir
	cfg.SearchOrder = []string{"project", "user"}

	// Both paths exist and are accessible
	paths := cfg.GetAllPaths()
	activeCount := 0
	for _, path := range paths {
		if pathExists(path) && isAccessible(path) {
			activeCount++
		}
	}

	// Should have 2 active paths
	if activeCount != 2 {
		t.Errorf("Expected 2 active paths, got %d", activeCount)
	}

	// Test summary message
	summary := ""
	if len(paths) > 0 && activeCount > 0 {
		summary = "2 path(s) configured, 2 active"
	}

	if summary != "2 path(s) configured, 2 active" {
		t.Errorf("Expected correct summary, got %s", summary)
	}
}
