package agents

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewConfig tests the default configuration
func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.ProjectPath != ".adk/agents" {
		t.Errorf("Expected project_path '.adk/agents', got '%s'", cfg.ProjectPath)
	}
	if cfg.UserPath != "~/.adk/agents" {
		t.Errorf("Expected user_path '~/.adk/agents', got '%s'", cfg.UserPath)
	}
	if !cfg.SkipMissing {
		t.Error("Expected SkipMissing to be true")
	}
	if len(cfg.SearchOrder) == 0 {
		t.Error("Expected non-empty search order")
	}
}

// TestLoadConfigDefaults tests loading config with no config file
func TestLoadConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.ProjectPath != ".adk/agents" {
		t.Errorf("Expected default project path")
	}
	if len(cfg.SearchOrder) == 0 {
		t.Error("Expected non-empty search order")
	}
}

// TestLoadConfigFromFile tests loading configuration from .adk/config.yaml
func TestLoadConfigFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .adk directory
	adkDir := filepath.Join(tmpDir, ".adk")
	if err := os.MkdirAll(adkDir, 0755); err != nil {
		t.Fatalf("Failed to create .adk directory: %v", err)
	}

	// Create config file
	configContent := `agent:
  project_path: ./agents
  user_path: ~/.agents
  plugin_paths:
    - /opt/plugins
    - /usr/local/plugins
  search_order:
    - project
    - plugin
    - user
  skip_missing: false
`
	configPath := filepath.Join(adkDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.ProjectPath != "./agents" {
		t.Errorf("Expected project_path './agents', got '%s'", cfg.ProjectPath)
	}
	if cfg.SkipMissing != false {
		t.Errorf("Expected skip_missing to be false, got %v", cfg.SkipMissing)
	}
}

// TestLoadConfigEnvironmentOverrides tests that environment variables override config file
func TestLoadConfigEnvironmentOverrides(t *testing.T) {
	tmpDir := t.TempDir()

	// Set environment variables
	oldProjectPath := os.Getenv("ADK_AGENT_PROJECT_PATH")
	oldUserPath := os.Getenv("ADK_AGENT_USER_PATH")
	defer func() {
		os.Setenv("ADK_AGENT_PROJECT_PATH", oldProjectPath)
		os.Setenv("ADK_AGENT_USER_PATH", oldUserPath)
		os.Unsetenv("ADK_AGENT_PROJECT_PATH")
		os.Unsetenv("ADK_AGENT_USER_PATH")
	}()

	os.Setenv("ADK_AGENT_PROJECT_PATH", "/override/agents")
	os.Setenv("ADK_AGENT_USER_PATH", "/override/user/agents")

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.ProjectPath != "/override/agents" {
		t.Errorf("Expected env override for project_path, got '%s'", cfg.ProjectPath)
	}
	if cfg.UserPath != "/override/user/agents" {
		t.Errorf("Expected env override for user_path, got '%s'", cfg.UserPath)
	}
}

// TestConfigValidate tests configuration validation
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		shouldErr bool
	}{
		{
			name:      "valid config",
			cfg:       NewConfig(),
			shouldErr: false,
		},
		{
			name: "invalid search order source",
			cfg: &Config{
				SearchOrder: []string{"invalid"},
			},
			shouldErr: true,
		},
		{
			name: "empty search order",
			cfg: &Config{
				SearchOrder: []string{},
			},
			shouldErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.shouldErr {
				t.Errorf("Expected error: %v, got: %v", tc.shouldErr, err)
			}
		})
	}
}

// TestConfigGetAllPaths tests path retrieval in search order
func TestConfigGetAllPaths(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/project/agents",
		UserPath:    "/user/agents",
		PluginPaths: []string{"/plugin1", "/plugin2"},
		SearchOrder: []string{"project", "user", "plugin"},
	}

	paths := cfg.GetAllPaths()

	if len(paths) != 4 {
		t.Errorf("Expected 4 paths, got %d", len(paths))
	}
	if paths[0] != "/project/agents" {
		t.Errorf("Expected first path to be project path")
	}
	if paths[1] != "/user/agents" {
		t.Errorf("Expected second path to be user path")
	}
}

// TestConfigGetAllPathsCustomOrder tests custom search order
func TestConfigGetAllPathsCustomOrder(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/project/agents",
		UserPath:    "/user/agents",
		PluginPaths: []string{"/plugin1"},
		SearchOrder: []string{"plugin", "project"},
	}

	paths := cfg.GetAllPaths()

	if len(paths) != 2 {
		t.Errorf("Expected 2 paths (user not included), got %d", len(paths))
	}
	if paths[0] != "/plugin1" {
		t.Errorf("Expected first path to be plugin path due to custom order")
	}
}

// TestConfigExpandPaths tests path expansion with ~
func TestConfigExpandPaths(t *testing.T) {
	cfg := &Config{
		UserPath:    "~/.adk/agents",
		PluginPaths: []string{"~/plugins"},
	}

	err := cfg.ExpandPaths()
	if err != nil {
		t.Fatalf("Failed to expand paths: %v", err)
	}

	home, _ := os.UserHomeDir()

	if !filepath.IsAbs(cfg.UserPath) {
		t.Errorf("Expected absolute user path after expansion")
	}
	if len(cfg.PluginPaths) > 0 && !filepath.IsAbs(cfg.PluginPaths[0]) {
		t.Errorf("Expected absolute plugin path after expansion")
	}
	if cfg.PluginPaths[0] != filepath.Join(home, "plugins") {
		t.Errorf("Expected expanded plugin path")
	}
}

// TestConfigGetSourceForPath tests source attribution
func TestConfigGetSourceForPath(t *testing.T) {
	tmpDir := t.TempDir()
	projPath := filepath.Join(tmpDir, ".adk", "agents")
	userPath := filepath.Join(tmpDir, "user", "agents")
	pluginPath := filepath.Join(tmpDir, "plugins", "agents")

	cfg := &Config{
		ProjectPath: projPath,
		UserPath:    userPath,
		PluginPaths: []string{pluginPath},
	}

	tests := []struct {
		path     string
		expected AgentSource
	}{
		{filepath.Join(projPath, "agent.md"), SourceProject},
		{filepath.Join(userPath, "agent.md"), SourceUser},
		{filepath.Join(pluginPath, "agent.md"), SourcePlugin},
	}

	for _, tc := range tests {
		source := cfg.GetSourceForPath(tc.path)
		if source != tc.expected {
			t.Errorf("Expected source %s for path %s, got %s", tc.expected, tc.path, source)
		}
	}
}

// TestExpandUserPath tests home directory expansion
func TestExpandUserPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/.adk/agents", filepath.Join(home, ".adk/agents")},
		{"~/agents", filepath.Join(home, "agents")},
		{"/abs/path", "/abs/path"},
		{"relative/path", "relative/path"},
	}

	for _, tc := range tests {
		result, err := expandUserPath(tc.input)
		if err != nil {
			t.Errorf("Unexpected error for %s: %v", tc.input, err)
			continue
		}
		if result != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, result)
		}
	}
}
