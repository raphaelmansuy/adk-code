// Package sqlite provides SQLite implementations of data repositories
package sqlite

import (
	"strings"

	"code_agent/pkg/models"
)

// ModelRegistryImpl implements the ModelRegistry interface
type ModelRegistryImpl struct {
	registry *models.Registry
}

// NewModelRegistry creates a new model registry wrapper
func NewModelRegistry() *ModelRegistryImpl {
	return &ModelRegistryImpl{
		registry: models.NewRegistry(),
	}
}

// GetModel retrieves a model by ID
func (m *ModelRegistryImpl) GetModel(id string) (any, error) {
	config, err := m.registry.GetModel(id)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetModelByName retrieves a model by display name (case-insensitive)
func (m *ModelRegistryImpl) GetModelByName(name string) (any, error) {
	config, err := m.registry.GetModelByName(name)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetDefaultModel returns the default model
func (m *ModelRegistryImpl) GetDefaultModel() any {
	return m.registry.GetDefaultModel()
}

// ListModels returns all available models
func (m *ModelRegistryImpl) ListModels() []any {
	configs := m.registry.ListModels()
	results := make([]any, len(configs))
	for i, config := range configs {
		results[i] = config
	}
	return results
}

// ListModelsByBackend returns models filtered by backend provider
func (m *ModelRegistryImpl) ListModelsByBackend(backend string) []any {
	// Normalize backend name (case-insensitive)
	backend = strings.ToLower(backend)
	configs := m.registry.ListModelsByBackend(backend)
	results := make([]any, len(configs))
	for i, config := range configs {
		results[i] = config
	}
	return results
}

// ResolveModel determines which model to use based on user input and context
// Priority: explicit model ID > explicit backend > defaults
func (m *ModelRegistryImpl) ResolveModel(modelID string, backend string) (any, error) {
	config, err := m.registry.ResolveModel(modelID, backend)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// ResolveFromProviderSyntax resolves a model using provider/model syntax
func (m *ModelRegistryImpl) ResolveFromProviderSyntax(
	providerName string,
	modelIdentifier string,
	defaultProvider string,
) (any, error) {
	config, err := m.registry.ResolveFromProviderSyntax(providerName, modelIdentifier, defaultProvider)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetProviderModels returns all models available for a specific provider
func (m *ModelRegistryImpl) GetProviderModels(provider string) []any {
	configs := m.registry.GetProviderModels(provider)
	results := make([]any, len(configs))
	for i, config := range configs {
		results[i] = config
	}
	return results
}

// ListProviders returns a list of all available providers
func (m *ModelRegistryImpl) ListProviders() []string {
	return m.registry.ListProviders()
}
