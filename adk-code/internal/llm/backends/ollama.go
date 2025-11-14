// Package backends - Ollama provider backend
package backends

import (
	"context"
	"fmt"

	adkmodel "google.golang.org/adk/model"

	"adk-code/pkg/models"
)

// OllamaProvider implements the ProviderBackend interface for Ollama
type OllamaProvider struct {
	metadata models.ProviderMetadata
}

// NewOllamaProvider creates a new Ollama provider instance
func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{
		metadata: models.GetProviderMetadata(models.ProviderOllama),
	}
}

// Create implements ProviderBackend.Create for Ollama
func (p *OllamaProvider) Create(ctx context.Context, config any) (adkmodel.LLM, error) {
	cfg, ok := config.(models.OllamaConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for Ollama provider: expected OllamaConfig, got %T", config)
	}

	// Delegate to pkg/models factory function
	return models.CreateOllamaModel(ctx, cfg)
}

// Validate implements ProviderBackend.Validate for Ollama
func (p *OllamaProvider) Validate(config any) error {
	cfg, ok := config.(models.OllamaConfig)
	if !ok {
		return fmt.Errorf("invalid config type for Ollama provider: expected OllamaConfig, got %T", config)
	}

	if cfg.ModelName == "" {
		return fmt.Errorf("Ollama model name is required")
	}

	// Host is optional - will use OLLAMA_HOST env var or default
	return nil
}

// GetMetadata implements ProviderBackend.GetMetadata for Ollama
func (p *OllamaProvider) GetMetadata() models.ProviderMetadata {
	return p.metadata
}

// Name implements ProviderBackend.Name for Ollama
func (p *OllamaProvider) Name() string {
	return "ollama"
}
