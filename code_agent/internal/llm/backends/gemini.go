// Package backends - LLM provider backend adapters
package backends

import (
	"context"
	"fmt"

	adkmodel "google.golang.org/adk/model"

	"code_agent/pkg/models"
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
