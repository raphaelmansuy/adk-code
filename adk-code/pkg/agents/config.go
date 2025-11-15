// Package agents provides agent definition discovery and management for adk-code.
// This file implements the configuration system for Phase 1, supporting
// multi-path agent discovery with environment variable overrides.
package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the agent discovery configuration.
// It specifies where and how agents are discovered across the system.
type Config struct {
	// ProjectPath is the project-level agent directory (typically .adk/agents/)
	ProjectPath string

	// UserPath is the user-level agent directory (typically ~/.adk/agents/)
	UserPath string

	// PluginPaths are the plugin directory paths where agents can be located
	PluginPaths []string

	// SearchOrder determines the priority order for agent discovery
	// Valid values: "project", "user", "plugin"
	SearchOrder []string

	// SkipMissing if true, continues discovery if a path doesn't exist
	SkipMissing bool
}

// configFile represents the structure of .adk/config.yaml
type configFile struct {
	Agent struct {
		ProjectPath string   `yaml:"project_path"`
		UserPath    string   `yaml:"user_path"`
		PluginPaths []string `yaml:"plugin_paths"`
		SearchOrder []string `yaml:"search_order"`
		SkipMissing *bool    `yaml:"skip_missing"`
	} `yaml:"agent"`
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		ProjectPath: ".adk/agents",
		UserPath:    "~/.adk/agents",
		PluginPaths: []string{},
		SearchOrder: []string{"project", "user", "plugin"},
		SkipMissing: true,
	}
}

// LoadConfig loads agent configuration from .adk/config.yaml in the project root.
// It merges defaults with file-based config and environment variables.
// Environment variables take precedence over file configuration.
func LoadConfig(projectRoot string) (*Config, error) {
	// Start with defaults
	cfg := NewConfig()

	// Try to load from .adk/config.yaml
	configPath := filepath.Join(projectRoot, ".adk", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		fileData, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		var cfgFile configFile
		if err := yaml.Unmarshal(fileData, &cfgFile); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}

		// Apply file configuration
		if cfgFile.Agent.ProjectPath != "" {
			cfg.ProjectPath = cfgFile.Agent.ProjectPath
		}
		if cfgFile.Agent.UserPath != "" {
			cfg.UserPath = cfgFile.Agent.UserPath
		}
		if len(cfgFile.Agent.PluginPaths) > 0 {
			cfg.PluginPaths = cfgFile.Agent.PluginPaths
		}
		if len(cfgFile.Agent.SearchOrder) > 0 {
			cfg.SearchOrder = cfgFile.Agent.SearchOrder
		}
		if cfgFile.Agent.SkipMissing != nil {
			cfg.SkipMissing = *cfgFile.Agent.SkipMissing
		}
	}

	// Apply environment variable overrides
	if projectPath := os.Getenv("ADK_AGENT_PROJECT_PATH"); projectPath != "" {
		cfg.ProjectPath = projectPath
	}
	if userPath := os.Getenv("ADK_AGENT_USER_PATH"); userPath != "" {
		cfg.UserPath = userPath
	}
	if pluginPaths := os.Getenv("ADK_AGENT_PLUGIN_PATHS"); pluginPaths != "" {
		cfg.PluginPaths = strings.Split(pluginPaths, ":")
	}
	if searchOrder := os.Getenv("ADK_AGENT_SEARCH_ORDER"); searchOrder != "" {
		cfg.SearchOrder = strings.Split(searchOrder, ",")
	}
	if skipMissing := os.Getenv("ADK_AGENT_SKIP_MISSING"); skipMissing != "" {
		cfg.SkipMissing = skipMissing == "true" || skipMissing == "1"
	}

	// Expand paths
	if err := cfg.ExpandPaths(); err != nil {
		return nil, err
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ExpandPaths expands ~ and other path variables
func (c *Config) ExpandPaths() error {
	var err error

	// Expand project path (relative to project root, typically stays as-is)
	// but we should support absolute paths
	if strings.HasPrefix(c.ProjectPath, "~") {
		c.ProjectPath, err = expandUserPath(c.ProjectPath)
		if err != nil {
			return err
		}
	}

	// Expand user path
	if strings.HasPrefix(c.UserPath, "~") {
		c.UserPath, err = expandUserPath(c.UserPath)
		if err != nil {
			return err
		}
	}

	// Expand plugin paths
	expandedPlugins := make([]string, 0, len(c.PluginPaths))
	for _, path := range c.PluginPaths {
		expanded := path
		if strings.HasPrefix(path, "~") {
			var err error
			expanded, err = expandUserPath(path)
			if err != nil {
				return err
			}
		}
		expandedPlugins = append(expandedPlugins, expanded)
	}
	c.PluginPaths = expandedPlugins

	return nil
}

// Validate checks the configuration for validity
func (c *Config) Validate() error {
	// Validate search order
	validSources := map[string]bool{
		"project": true,
		"user":    true,
		"plugin":  true,
	}

	for _, source := range c.SearchOrder {
		if !validSources[source] {
			return fmt.Errorf("invalid search order source: %s (must be project, user, or plugin)", source)
		}
	}

	// Ensure at least one search source is specified
	if len(c.SearchOrder) == 0 {
		return fmt.Errorf("search_order must not be empty")
	}

	return nil
}

// GetAllPaths returns all configured paths in search order
func (c *Config) GetAllPaths() []string {
	paths := []string{}

	for _, source := range c.SearchOrder {
		switch source {
		case "project":
			if c.ProjectPath != "" {
				paths = append(paths, c.ProjectPath)
			}
		case "user":
			if c.UserPath != "" {
				paths = append(paths, c.UserPath)
			}
		case "plugin":
			paths = append(paths, c.PluginPaths...)
		}
	}

	return paths
}

// GetSourceForPath determines the source (project/user/plugin) for a given path
func (c *Config) GetSourceForPath(path string) AgentSource {
	absPath, _ := filepath.Abs(path)
	projPath, _ := filepath.Abs(c.ProjectPath)
	userPath, _ := filepath.Abs(c.UserPath)

	if strings.HasPrefix(absPath, projPath) {
		return SourceProject
	}
	if strings.HasPrefix(absPath, userPath) {
		return SourceUser
	}

	// Check plugin paths
	for _, pluginPath := range c.PluginPaths {
		absPluginPath, _ := filepath.Abs(pluginPath)
		if strings.HasPrefix(absPath, absPluginPath) {
			return SourcePlugin
		}
	}

	return SourceProject // default fallback
}

// expandUserPath expands ~ to the user's home directory
func expandUserPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return home, nil
	}
	return path, nil
}
