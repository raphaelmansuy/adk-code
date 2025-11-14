// Package backends - Ollama provider backend
package backends

import (
	"context"
	"fmt"

	"github.com/ollama/ollama/api"
	adkmodel "google.golang.org/adk/model"

	"adk-code/pkg/models"
)

// OllamaProvider implements the ProviderBackend interface for Ollama
type OllamaProvider struct {
	metadata models.ProviderMetadata
	registry *models.OllamaModelRegistry
}

// NewOllamaProvider creates a new Ollama provider instance
func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{
		metadata: models.GetProviderMetadata(models.ProviderOllama),
		// Registry will be initialized lazily on first use
		registry: nil,
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

// ListModels implements ModelDiscovery.ListModels for Ollama
func (p *OllamaProvider) ListModels(ctx context.Context, forceRefresh bool) ([]models.ModelInfo, error) {
	// Initialize registry if needed
	if p.registry == nil {
		client, err := p.getClient()
		if err != nil {
			return nil, err
		}
		p.registry = models.NewOllamaModelRegistry(client, 0) // Use default 5-minute TTL
	}

	return p.registry.ListModels(ctx, forceRefresh)
}

// GetModelInfo implements ModelDiscovery.GetModelInfo for Ollama
func (p *OllamaProvider) GetModelInfo(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	// Initialize registry if needed
	if p.registry == nil {
		client, err := p.getClient()
		if err != nil {
			return nil, err
		}
		p.registry = models.NewOllamaModelRegistry(client, 0) // Use default 5-minute TTL
	}

	return p.registry.GetModelInfo(ctx, modelName)
}

// getClient creates or retrieves an Ollama API client
func (p *OllamaProvider) getClient() (*api.Client, error) {
	// Use default client from environment (OLLAMA_HOST env var or default local endpoint)
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}
	return client, nil
}
