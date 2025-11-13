// Package backends - LLM provider backend adapters
package backends

import (
	"context"
	"fmt"

	adkmodel "google.golang.org/adk/model"

	"adk-code/pkg/models"
)

// VertexAIProvider implements the ProviderBackend interface for Vertex AI
type VertexAIProvider struct {
	metadata models.ProviderMetadata
}

// NewVertexAIProvider creates a new Vertex AI provider instance
func NewVertexAIProvider() *VertexAIProvider {
	return &VertexAIProvider{
		metadata: models.GetProviderMetadata(models.ProviderVertexAI),
	}
}

// Create implements ProviderBackend.Create for Vertex AI
func (p *VertexAIProvider) Create(ctx context.Context, config any) (adkmodel.LLM, error) {
	cfg, ok := config.(models.VertexAIConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for Vertex AI provider: expected VertexAIConfig, got %T", config)
	}

	// Delegate to pkg/models factory function
	return models.CreateVertexAIModel(ctx, cfg)
}

// Validate implements ProviderBackend.Validate for Vertex AI
func (p *VertexAIProvider) Validate(config any) error {
	cfg, ok := config.(models.VertexAIConfig)
	if !ok {
		return fmt.Errorf("invalid config type for Vertex AI provider: expected VertexAIConfig, got %T", config)
	}

	if cfg.Project == "" {
		return fmt.Errorf("vertex AI project is required")
	}
	if cfg.Location == "" {
		return fmt.Errorf("vertex AI location is required")
	}
	if cfg.ModelName == "" {
		return fmt.Errorf("model name is required")
	}

	return nil
}

// GetMetadata implements ProviderBackend.GetMetadata for Vertex AI
func (p *VertexAIProvider) GetMetadata() models.ProviderMetadata {
	return p.metadata
}

// Name implements ProviderBackend.Name for Vertex AI
func (p *VertexAIProvider) Name() string {
	return "vertexai"
}
