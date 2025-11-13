package factories

import (
	"context"

	"google.golang.org/adk/model"
)

// ModelFactory defines the interface for creating LLM models
// Each provider (Gemini, OpenAI, Vertex AI) implements this interface
type ModelFactory interface {
	// Create builds an LLM model instance with the provided configuration
	Create(ctx context.Context, config ModelConfig) (model.LLM, error)

	// ValidateConfig checks if the provided configuration is valid for this factory
	ValidateConfig(config ModelConfig) error

	// Info returns metadata about this factory
	Info() FactoryInfo
}

// ModelConfig holds configuration for model creation
// Different providers may use different fields
type ModelConfig struct {
	// Common fields
	ModelName string

	// Gemini API specific
	APIKey string

	// Vertex AI specific
	Project  string
	Location string

	// OpenAI specific
	// Uses APIKey field
}

// FactoryInfo contains metadata about a factory
type FactoryInfo struct {
	Provider      string // "Gemini", "OpenAI", "VertexAI"
	Description   string
	RequiredField []string // Required fields for this provider
}
