package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigFileName is the name of the workspace configuration file
const ConfigFileName = ".workspace.json"

// Config represents the persistent workspace configuration
type Config struct {
	// Version of the config format for future compatibility
	Version int `json:"version"`

	// Roots are the workspace roots
	Roots []WorkspaceRoot `json:"roots"`

	// PrimaryIndex is the index of the primary workspace
	PrimaryIndex int `json:"primaryIndex"`

	// Preferences for workspace behavior
	Preferences Preferences `json:"preferences,omitempty"`
}

// Preferences contains user preferences for workspace behavior
type Preferences struct {
	// AutoDetectWorkspaces enables automatic workspace detection
	AutoDetectWorkspaces bool `json:"autoDetectWorkspaces,omitempty"`

	// MaxWorkspaces limits the number of workspaces to manage
	MaxWorkspaces int `json:"maxWorkspaces,omitempty"`

	// PreferVCSRoots prioritizes VCS repository roots when detecting workspaces
	PreferVCSRoots bool `json:"preferVCSRoots,omitempty"`

	// IncludeHidden includes hidden directories in workspace detection
	IncludeHidden bool `json:"includeHidden,omitempty"`
}

// DefaultPreferences returns the default workspace preferences
func DefaultPreferences() Preferences {
	return Preferences{
		AutoDetectWorkspaces: false,
		MaxWorkspaces:        10,
		PreferVCSRoots:       true,
		IncludeHidden:        false,
	}
}

// LoadConfig loads workspace configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	if config.Version == 0 {
		config.Version = 1
	}

	if config.PrimaryIndex < 0 || config.PrimaryIndex >= len(config.Roots) {
		config.PrimaryIndex = 0
	}

	// Set default preferences if not specified
	if config.Preferences.MaxWorkspaces == 0 {
		config.Preferences = DefaultPreferences()
	}

	return &config, nil
}

// SaveConfig saves workspace configuration to a file
func SaveConfig(path string, config *Config) error {
	// Ensure version is set
	if config.Version == 0 {
		config.Version = 1
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with proper permissions
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfigFromDirectory loads config from a directory
// Looks for .workspace.json in the given directory
func LoadConfigFromDirectory(dir string) (*Config, error) {
	configPath := filepath.Join(dir, ConfigFileName)
	return LoadConfig(configPath)
}

// SaveConfigToDirectory saves config to a directory
func SaveConfigToDirectory(dir string, config *Config) error {
	configPath := filepath.Join(dir, ConfigFileName)
	return SaveConfig(configPath, config)
}

// ConfigExists checks if a workspace config file exists in the directory
func ConfigExists(dir string) bool {
	configPath := filepath.Join(dir, ConfigFileName)
	_, err := os.Stat(configPath)
	return err == nil
}

// ManagerFromConfig creates a workspace manager from a config
func ManagerFromConfig(config *Config) *Manager {
	return NewManager(config.Roots, config.PrimaryIndex)
}

// ManagerToConfig converts a workspace manager to a config
func ManagerToConfig(manager *Manager, preferences *Preferences) *Config {
	config := &Config{
		Version:      1,
		Roots:        manager.GetRoots(),
		PrimaryIndex: manager.GetPrimaryIndex(),
	}

	if preferences != nil {
		config.Preferences = *preferences
	} else {
		config.Preferences = DefaultPreferences()
	}

	return config
}

// LoadManagerFromDirectory attempts to load a workspace manager from a config file
// If no config exists, returns nil without error
func LoadManagerFromDirectory(dir string) (*Manager, *Preferences, error) {
	if !ConfigExists(dir) {
		return nil, nil, nil
	}

	config, err := LoadConfigFromDirectory(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	manager := ManagerFromConfig(config)
	return manager, &config.Preferences, nil
}

// SaveManagerToDirectory saves a workspace manager to a config file
func SaveManagerToDirectory(dir string, manager *Manager, preferences *Preferences) error {
	config := ManagerToConfig(manager, preferences)
	return SaveConfigToDirectory(dir, config)
}

// MigrateConfig migrates an older config to the current version
func MigrateConfig(config *Config) (*Config, error) {
	// Currently only version 1 exists, so no migration needed
	if config.Version == 1 {
		return config, nil
	}

	return nil, fmt.Errorf("unsupported config version: %d", config.Version)
}

// ValidateConfig validates a workspace configuration
func ValidateConfig(config *Config) error {
	if config.Version != 1 {
		return fmt.Errorf("unsupported config version: %d", config.Version)
	}

	if len(config.Roots) == 0 {
		return fmt.Errorf("config must have at least one workspace root")
	}

	if config.PrimaryIndex < 0 || config.PrimaryIndex >= len(config.Roots) {
		return fmt.Errorf("invalid primary index: %d (have %d roots)", config.PrimaryIndex, len(config.Roots))
	}

	// Validate each root
	for i, root := range config.Roots {
		if root.Path == "" {
			return fmt.Errorf("root %d has empty path", i)
		}

		if root.Name == "" {
			return fmt.Errorf("root %d has empty name", i)
		}

		// Check if path exists
		info, err := os.Stat(root.Path)
		if err != nil {
			return fmt.Errorf("root %d path does not exist: %s", i, root.Path)
		}

		if !info.IsDir() {
			return fmt.Errorf("root %d path is not a directory: %s", i, root.Path)
		}
	}

	// Validate preferences
	if config.Preferences.MaxWorkspaces < 1 {
		return fmt.Errorf("maxWorkspaces must be at least 1")
	}

	return nil
}
