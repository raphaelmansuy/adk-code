// Package main - CLI flag parsing
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// ParseProviderModelSyntax parses a "provider/model" string into components
// Examples:
//
//	"gemini/2.5-flash" → ("gemini", "2.5-flash", nil)
//	"gemini/flash" → ("gemini", "flash", nil)
//	"flash" → ("", "flash", nil)
//	"/flash" → ("", "", error)
//	"a/b/c" → ("", "", error)
func ParseProviderModelSyntax(input string) (string, string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", fmt.Errorf("model syntax cannot be empty")
	}

	parts := strings.Split(input, "/")
	switch len(parts) {
	case 1:
		// Shorthand without provider: "flash" → ("", "flash")
		return "", parts[0], nil
	case 2:
		// Full syntax: "provider/model" → ("provider", "model")
		if parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("invalid model syntax: %q (use provider/model)", input)
		}
		return parts[0], parts[1], nil
	default:
		return "", "", fmt.Errorf("invalid model syntax: %q (use provider/model)", input)
	}
}

// CLIConfig holds parsed command-line flags
type CLIConfig struct {
	OutputFormat      string
	TypewriterEnabled bool
	SessionName       string
	DBPath            string
	WorkingDirectory  string
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
}

// ParseCLIFlags parses command-line arguments and returns config and remaining args
func ParseCLIFlags() (CLIConfig, []string) {
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
	apiKey := flag.String("api-key", os.Getenv("GOOGLE_API_KEY"), "API key for Gemini (default: GOOGLE_API_KEY env var)")
	vertexAIProject := flag.String("project", os.Getenv("GOOGLE_CLOUD_PROJECT"), "GCP Project ID for Vertex AI (default: GOOGLE_CLOUD_PROJECT env var)")
	vertexAILocation := flag.String("location", os.Getenv("GOOGLE_CLOUD_LOCATION"), "GCP Location for Vertex AI (default: GOOGLE_CLOUD_LOCATION env var)")

	// Thinking configuration flags
	enableThinking := flag.Bool("enable-thinking", true, "Enable model thinking/reasoning output (default: true)")
	thinkingBudget := flag.Int("thinking-budget", 1024, "Token budget for thinking when enabled (default: 1024)")

	flag.Parse()

	// Auto-detect backend from environment if not specified
	selectedBackend := *backend
	if selectedBackend == "" {
		if os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "true" || os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "1" {
			selectedBackend = "vertexai"
		} else if *apiKey != "" {
			selectedBackend = "gemini"
		} else if *vertexAIProject != "" {
			// If project is set but backend not specified, assume vertexai
			selectedBackend = "vertexai"
		} else {
			// Default to gemini if nothing is set (existing behavior)
			selectedBackend = "gemini"
		}
	}

	return CLIConfig{
		OutputFormat:      *outputFormat,
		TypewriterEnabled: *typewriterEnabled,
		SessionName:       *sessionName,
		DBPath:            *dbPath,
		WorkingDirectory:  *workingDirectory,
		Backend:           selectedBackend,
		APIKey:            *apiKey,
		VertexAIProject:   *vertexAIProject,
		VertexAILocation:  *vertexAILocation,
		Model:             *model,
		EnableThinking:    *enableThinking,
		ThinkingBudget:    int32(*thinkingBudget),
	}, flag.Args()
}
