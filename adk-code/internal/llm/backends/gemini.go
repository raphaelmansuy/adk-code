// Package backends - LLM provider backend adapters
package backends

import (
	"context"
	"fmt"

	adkmodel "google.golang.org/adk/model"

	"adk-code/pkg/models"
)

// GeminiProvider implements the ProviderBackend interface for Gemini API
type GeminiProvider struct {
	metadata models.ProviderMetadata
}

// NewGeminiProvider creates a new Gemini provider instance
func NewGeminiProvider() *GeminiProvider {
	return &GeminiProvider{
		metadata: models.GetProviderMetadata(models.ProviderGemini),
	}
}

// Create implements ProviderBackend.Create for Gemini
func (p *GeminiProvider) Create(ctx context.Context, config any) (adkmodel.LLM, error) {
	cfg, ok := config.(models.GeminiConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for Gemini provider: expected GeminiConfig, got %T", config)
	}

	// Delegate to pkg/models factory function
	return models.CreateGeminiModel(ctx, cfg)
}

// Validate implements ProviderBackend.Validate for Gemini
func (p *GeminiProvider) Validate(config any) error {
	cfg, ok := config.(models.GeminiConfig)
	if !ok {
		return fmt.Errorf("invalid config type for Gemini provider: expected GeminiConfig, got %T", config)
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("gemini API key is required")
	}
	if cfg.ModelName == "" {
		return fmt.Errorf("model name is required")
	}

	return nil
}

// GetMetadata implements ProviderBackend.GetMetadata for Gemini
func (p *GeminiProvider) GetMetadata() models.ProviderMetadata {
	return p.metadata
}

// Name implements ProviderBackend.Name for Gemini
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// ListModels implements ModelDiscovery.ListModels for Gemini
// Returns a list of available Gemini models with their capabilities
func (p *GeminiProvider) ListModels(ctx context.Context, forceRefresh bool) ([]models.ModelInfo, error) {
	// For V1, return a static list of known Gemini models
	// Future versions can query the Gemini API dynamically
	return []models.ModelInfo{
		{
			Name:        "gemini-2.0-flash-exp",
			DisplayName: "Gemini 2.0 Flash (Experimental)",
			Provider:    "gemini",
			Family:      "gemini",
			Capabilities: models.ModelCapabilities{
				VisionSupport: true,
				ToolCalling:   true,
				Streaming:     true,
				JSONMode:      true,
				ContextWindow: 1048576,
			},
			Description: "Fastest multimodal model with 1M token context window",
		},
		{
			Name:        "gemini-1.5-flash",
			DisplayName: "Gemini 1.5 Flash",
			Provider:    "gemini",
			Family:      "gemini",
			Capabilities: models.ModelCapabilities{
				VisionSupport: true,
				ToolCalling:   true,
				Streaming:     true,
				JSONMode:      true,
				ContextWindow: 1048576,
			},
			Description: "Fast and versatile multimodal model",
		},
		{
			Name:        "gemini-1.5-pro",
			DisplayName: "Gemini 1.5 Pro",
			Provider:    "gemini",
			Family:      "gemini",
			Capabilities: models.ModelCapabilities{
				VisionSupport: true,
				ToolCalling:   true,
				Streaming:     true,
				JSONMode:      true,
				ContextWindow: 2097152,
			},
			Description: "Most capable model with 2M token context window",
		},
	}, nil
}

// GetModelInfo implements ModelDiscovery.GetModelInfo for Gemini
func (p *GeminiProvider) GetModelInfo(ctx context.Context, modelName string) (*models.ModelInfo, error) {
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
