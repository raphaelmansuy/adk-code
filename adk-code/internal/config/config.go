// Package config - Configuration management
package config

import (
	"flag"
	"fmt"
	"os"

	"adk-code/pkg/models"
)

// Config holds all application configuration from CLI flags and environment
type Config struct {
	// CLI Output configuration
	OutputFormat      string
	TypewriterEnabled bool

	// Session configuration
	SessionName string
	DBPath      string

	// Working directory
	WorkingDirectory string

	// Backend configuration
	Backend          string // "gemini" or "vertexai"
	APIKey           string // For Gemini API
	VertexAIProject  string // For Vertex AI
	VertexAILocation string // For Vertex AI

	// Model selection
	Model string // Specific model ID (e.g., "gemini-2.5-flash", "gemini-1.5-pro")

	// Thinking configuration
	EnableThinking bool  // Enable model thinking/reasoning output
	ThinkingBudget int32 // Token budget for thinking

	// MCP configuration
	MCPConfigPath string
	MCPConfig     *MCPConfig
}

// LoadFromEnv loads configuration from environment and CLI flags
// This is the new consolidated factory that replaces ParseCLIFlags
func LoadFromEnv() (Config, []string) {
	outputFormat := flag.String("output-format", "rich", "Output format: rich, plain, or json")
	typewriterEnabled := flag.Bool("typewriter", false, "Enable typewriter effect for text output")
	sessionName := flag.String("session", "", "Session name (optional, defaults to 'default')")
	dbPath := flag.String("db", "", "Database path for sessions (optional, defaults to ~/.code_agent/sessions.db)")
	workingDirectory := flag.String("working-directory", "", "Working directory for the agent (optional, defaults to current directory)")

	// Model selection flags
	model := flag.String("model", "", "Model to use with provider/model syntax. Examples:\n"+
		"  --model gemini/2.5-flash     (explicit provider)\n"+
		"  --model gemini/flash          (shorthand, means 2.5-flash)\n"+
		"  --model vertexai/1.5-pro      (Vertex AI model)\n"+
		"Use '/providers' command to list all available models.")

	// Backend selection flags
	backend := flag.String("backend", "", "Backend to use: 'gemini' or 'vertexai' (default: auto-detect from env vars)")

	// Read API keys from environment without exposing in help
	envAPIKey := os.Getenv("GOOGLE_API_KEY")
	apiKey := flag.String("api-key", "", "API key for Gemini (default: GOOGLE_API_KEY env var)")

	envVertexAIProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	vertexAIProject := flag.String("project", "", "GCP Project ID for Vertex AI (default: GOOGLE_CLOUD_PROJECT env var)")

	envVertexAILocation := os.Getenv("GOOGLE_CLOUD_LOCATION")
	vertexAILocation := flag.String("location", "", "GCP Location for Vertex AI (default: GOOGLE_CLOUD_LOCATION env var)")

	// Thinking configuration flags
	enableThinking := flag.Bool("enable-thinking", true, "Enable model thinking/reasoning output (default: true)")
	thinkingBudget := flag.Int("thinking-budget", 1024, "Token budget for thinking when enabled (default: 1024)")

	// MCP configuration flags
	mcpConfigPath := flag.String("mcp-config", "", "Path to MCP config file (optional)")

	flag.Parse()

	// Use provided flags or fall back to environment
	apiKeyValue := *apiKey
	if apiKeyValue == "" {
		apiKeyValue = envAPIKey
	}

	projectValue := *vertexAIProject
	if projectValue == "" {
		projectValue = envVertexAIProject
	}

	locationValue := *vertexAILocation
	if locationValue == "" {
		locationValue = envVertexAILocation
	}

	// Auto-detect backend from environment if not specified
	selectedBackend := *backend
	if selectedBackend == "" {
		if os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "true" || os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "1" {
			selectedBackend = "vertexai"
		} else if apiKeyValue != "" {
			selectedBackend = "gemini"
		} else if projectValue != "" {
			// If project is set but backend not specified, assume vertexai
			selectedBackend = "vertexai"
		} else {
			// Default to gemini if nothing is set (existing behavior)
			selectedBackend = "gemini"
		}
	}

	// Load MCP config if path specified
	var mcpConfig *MCPConfig
	if *mcpConfigPath != "" {
		loadedConfig, err := LoadMCP(*mcpConfigPath)
		if err != nil {
			// Log error but don't fail - MCP is optional
			fmt.Fprintf(os.Stderr, "Warning: Failed to load MCP config from %s: %v\n", *mcpConfigPath, err)
		} else {
			mcpConfig = loadedConfig
		}
	}

	return Config{
		OutputFormat:      *outputFormat,
		TypewriterEnabled: *typewriterEnabled,
		SessionName:       *sessionName,
		DBPath:            *dbPath,
		WorkingDirectory:  *workingDirectory,
		Backend:           selectedBackend,
		APIKey:            apiKeyValue,
		VertexAIProject:   projectValue,
		VertexAILocation:  locationValue,
		Model:             *model,
		EnableThinking:    *enableThinking,
		ThinkingBudget:    int32(*thinkingBudget),
		MCPConfigPath:     *mcpConfigPath,
		MCPConfig:         mcpConfig,
	}, flag.Args()
}

// GetModelRegistry creates a model registry for use with config
// This is a helper to resolve which model to use
func (c *Config) GetModelRegistry() *models.Registry {
	return models.NewRegistry()
}

// ResolveModel returns the resolved model configuration based on CLI input
func (c *Config) ResolveModel(registry *models.Registry) (models.Config, error) {
	return registry.ResolveModel(c.Model, c.Backend)
}
