package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MCPConfig represents the overall MCP configuration
type MCPConfig struct {
	Enabled bool                    `json:"enabled"`
	Servers map[string]ServerConfig `json:"servers"`
}

// ServerConfig represents a single MCP server configuration
type ServerConfig struct {
	Name    string            `json:"-"`                 // Set from map key
	Type    string            `json:"type"`              // "stdio", "sse", "streamable"
	Command string            `json:"command,omitempty"` // For stdio
	Args    []string          `json:"args,omitempty"`
	URL     string            `json:"url,omitempty"` // For sse/streamable
	Headers map[string]string `json:"headers,omitempty"`
	Env     map[string]string `json:"env,omitempty"`     // Environment variables for stdio
	Cwd     string            `json:"cwd,omitempty"`     // Working directory for stdio
	Timeout int               `json:"timeout,omitempty"` // milliseconds, default 30000
}

// LoadMCP loads MCP config from file or returns disabled config
// Supports relative paths, absolute paths, and tilde (~) expansion
func LoadMCP(configPath string) (*MCPConfig, error) {
	if configPath == "" {
		return &MCPConfig{Enabled: false}, nil
	}

	// Resolve the path (handle tilde expansion and relative paths)
	resolvedPath, err := resolvePath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &MCPConfig{Enabled: false}, nil
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", resolvedPath, err)
	}

	var cfg MCPConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set server names from map keys
	for name, srv := range cfg.Servers {
		s := srv
		s.Name = name
		cfg.Servers[name] = s
	}

	// Validate all servers
	for name, srv := range cfg.Servers {
		if err := srv.validate(); err != nil {
			return nil, fmt.Errorf("server '%s': %w", name, err)
		}
	}

	return &cfg, nil
}

// resolvePath resolves a file path, handling:
// - Tilde (~) expansion to home directory
// - Relative paths (resolved relative to current working directory)
// - Absolute paths (returned as-is)
func resolvePath(path string) (string, error) {
	// Handle tilde expansion
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		// Replace ~ with home directory
		path = filepath.Join(homeDir, path[1:])
	}

	// Handle relative paths - resolve relative to current working directory
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		path = filepath.Join(cwd, path)
	}

	return path, nil
}

// validate checks if server configuration is valid
func (s ServerConfig) validate() error {
	if s.Type == "" {
		return fmt.Errorf("type required")
	}
	switch s.Type {
	case "stdio":
		if s.Command == "" {
			return fmt.Errorf("command required for stdio")
		}
	case "sse", "streamable":
		if s.URL == "" {
			return fmt.Errorf("url required for %s", s.Type)
		}
	default:
		return fmt.Errorf("unsupported type: %s", s.Type)
	}
	return nil
}
