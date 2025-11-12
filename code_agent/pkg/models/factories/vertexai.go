package factories

import (
	"context"
	"fmt"

	"code_agent/pkg/models"

	"google.golang.org/adk/model"
)

// VertexAIFactory creates Vertex AI models
type VertexAIFactory struct{}

// Create builds a Vertex AI model with the provided configuration
func (f *VertexAIFactory) Create(ctx context.Context, config ModelConfig) (model.LLM, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, err
	}

	// Delegate to the existing Vertex AI model creation logic
	cfg := models.VertexAIConfig{
		Project:   config.Project,
		Location:  config.Location,
		ModelName: config.ModelName,
	}

	llm, err := models.CreateVertexAIModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI model: %w", err)
	}

	return llm, nil
}

// ValidateConfig checks if the configuration is valid for Vertex AI
func (f *VertexAIFactory) ValidateConfig(config ModelConfig) error {
	if config.Project == "" {
		return fmt.Errorf("Vertex AI project is required")
	}
	if config.Location == "" {
		return fmt.Errorf("Vertex AI location is required")
	}
	if config.ModelName == "" {
		return fmt.Errorf("model name is required")
	}
	return nil
}

// Info returns metadata about the Vertex AI factory
func (f *VertexAIFactory) Info() FactoryInfo {
	return FactoryInfo{
		Provider:      "VertexAI",
		Description:   "Google Cloud Vertex AI models",
		RequiredField: []string{"Project", "Location", "ModelName"},
	}
}

// NewVertexAIFactory creates a new Vertex AI factory instance
func NewVertexAIFactory() *VertexAIFactory {
	return &VertexAIFactory{}
}
