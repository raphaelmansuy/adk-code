// Package models - Model type definitions
package models

// Capabilities represents the capabilities of a model
type Capabilities struct {
	VisionSupport     bool   // Can process images
	ToolUseSupport    bool   // Can use tools/functions
	LongContextWindow bool   // Has extended context length
	CostTier          string // "economy", "standard", "premium"
}

// Config holds configuration for a specific model
type Config struct {
	ID             string
	Name           string
	DisplayName    string
	Backend        string // "gemini" or "vertexai"
	ContextWindow  int    // tokens
	Capabilities   Capabilities
	Description    string
	RecommendedFor []string // Use cases: "coding", "analysis", "creative", etc.
	IsDefault      bool
}
