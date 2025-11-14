// Package backends - LLM provider backend adapters
package backends

import (
	"context"
	"fmt"

	adkmodel "google.golang.org/adk/model"

	"adk-code/pkg/models"
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

// ListModels implements ModelDiscovery.ListModels for OpenAI
// Returns a list of available OpenAI models with their capabilities
func (p *OpenAIProvider) ListModels(ctx context.Context, forceRefresh bool) ([]models.ModelInfo, error) {
	// For V1, return a static list of known OpenAI models
	// Future versions can query the OpenAI API dynamically
	return []models.ModelInfo{
		{
			Name:        "gpt-4o",
			DisplayName: "GPT-4o",
			Provider:    "openai",
			Family:      "gpt-4",
			Capabilities: models.ModelCapabilities{
				VisionSupport: true,
				ToolCalling:   true,
				Streaming:     true,
				JSONMode:      true,
				ContextWindow: 128000,
			},
			Description: "High-intelligence flagship model for complex tasks",
		},
		{
			Name:        "gpt-4o-mini",
			DisplayName: "GPT-4o Mini",
			Provider:    "openai",
			Family:      "gpt-4",
			Capabilities: models.ModelCapabilities{
				VisionSupport: true,
				ToolCalling:   true,
				Streaming:     true,
				JSONMode:      true,
				ContextWindow: 128000,
			},
			Description: "Affordable small model for fast, lightweight tasks",
		},
		{
			Name:        "gpt-4-turbo",
			DisplayName: "GPT-4 Turbo",
			Provider:    "openai",
			Family:      "gpt-4",
			Capabilities: models.ModelCapabilities{
				VisionSupport: true,
				ToolCalling:   true,
				Streaming:     true,
				JSONMode:      true,
				ContextWindow: 128000,
			},
			Description: "Previous generation high-intelligence model",
		},
		{
			Name:        "gpt-3.5-turbo",
			DisplayName: "GPT-3.5 Turbo",
			Provider:    "openai",
			Family:      "gpt-3.5",
			Capabilities: models.ModelCapabilities{
				VisionSupport: false,
				ToolCalling:   true,
				Streaming:     true,
				JSONMode:      true,
				ContextWindow: 16385,
			},
			Description: "Fast, inexpensive model for simple tasks",
		},
	}, nil
}

// GetModelInfo implements ModelDiscovery.GetModelInfo for OpenAI
func (p *OpenAIProvider) GetModelInfo(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	models, err := p.ListModels(ctx, false)
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		if model.Name == modelName {
			return &model, nil
		}
	}

	return nil, fmt.Errorf("model not found: %s", modelName)
}
