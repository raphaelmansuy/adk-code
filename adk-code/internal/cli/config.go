// Package cli - CLI configuration
package cli

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
