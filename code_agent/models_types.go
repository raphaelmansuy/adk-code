// Package main - Model type definitions
package main

// ModelCapabilities represents the capabilities of a model
type ModelCapabilities struct {
	VisionSupport     bool   // Can process images
	ToolUseSupport    bool   // Can use tools/functions
	LongContextWindow bool   // Has extended context length
	CostTier          string // "economy", "standard", "premium"
}

// ModelConfig holds configuration for a specific model
type ModelConfig struct {
	ID             string
	Name           string
	DisplayName    string
	Backend        string // "gemini" or "vertexai"
	ContextWindow  int    // tokens
	Capabilities   ModelCapabilities
	Description    string
	RecommendedFor []string // Use cases: "coding", "analysis", "creative", etc.
	IsDefault      bool
}
