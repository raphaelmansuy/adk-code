// Package backends - LLM provider backend adapters
package backends

import (
	"context"
	"fmt"

	adkmodel "google.golang.org/adk/model"

	"code_agent/pkg/models"
)

// OpenAIProvider implements the ProviderBackend interface for OpenAI
type OpenAIProvider struct {
	metadata models.ProviderMetadata
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		metadata: models.GetProviderMetadata(models.ProviderOpenAI),
	}
}

// Create implements ProviderBackend.Create for OpenAI
func (p *OpenAIProvider) Create(ctx context.Context, config any) (adkmodel.LLM, error) {
	cfg, ok := config.(models.OpenAIConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for OpenAI provider: expected OpenAIConfig, got %T", config)
	}

	// Delegate to pkg/models factory function
	return models.CreateOpenAIModel(ctx, cfg)
}

// Validate implements ProviderBackend.Validate for OpenAI
func (p *OpenAIProvider) Validate(config any) error {
	cfg, ok := config.(models.OpenAIConfig)
	if !ok {
		return fmt.Errorf("invalid config type for OpenAI provider: expected OpenAIConfig, got %T", config)
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("openAI API key is required")
	}
	if cfg.ModelName == "" {
		return fmt.Errorf("model name is required")
	}

	return nil
}

// GetMetadata implements ProviderBackend.GetMetadata for OpenAI
func (p *OpenAIProvider) GetMetadata() models.ProviderMetadata {
	return p.metadata
}

// Name implements ProviderBackend.Name for OpenAI
func (p *OpenAIProvider) Name() string {
	return "openai"
}
