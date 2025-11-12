// Package models - Model registry implementation
package models

import (
	"fmt"
	"sort"
	"strings"
)

// Registry manages available models and configurations
type Registry struct {
	models           map[string]Config   // model ID → config
	aliases          map[string]string   // "provider/shorthand" → model ID
	modelsByProvider map[string][]string // provider → list of model IDs
}

// NewRegistry creates and initializes a new model registry with all models registered
func NewRegistry() *Registry {
	registry := &Registry{
		models:           make(map[string]Config),
		aliases:          make(map[string]string),
		modelsByProvider: make(map[string][]string),
	}

	// Register all available models
	RegisterGeminiAndVertexAIModels(registry)
	RegisterOpenAIModels(registry)

	return registry
}

// RegisterModel adds a model to the registry
func (mr *Registry) RegisterModel(model Config) {
	mr.models[model.ID] = model
}

// GetModel retrieves a model by ID
func (mr *Registry) GetModel(id string) (Config, error) {
	model, exists := mr.models[id]
	if !exists {
		return Config{}, fmt.Errorf("model %q not found in registry", id)
	}
	return model, nil
}

// GetModelByName retrieves a model by display name (case-insensitive)
func (mr *Registry) GetModelByName(name string) (Config, error) {
	name = strings.ToLower(name)
	for _, model := range mr.models {
		if strings.ToLower(model.Name) == name {
			return model, nil
		}
	}
	return Config{}, fmt.Errorf("model %q not found", name)
}

// GetDefaultModel returns the default model
func (mr *Registry) GetDefaultModel() Config {
	for _, model := range mr.models {
		if model.IsDefault {
			return model
		}
	}
	// Fallback to gemini-2.5-flash if no default found
	if model, err := mr.GetModel("gemini-2.5-flash"); err == nil {
		return model
	}
	// This should never happen with proper initialization
	panic("no models registered")
}

// ListModels returns all available models
func (mr *Registry) ListModels() []Config {
	models := make([]Config, 0, len(mr.models))
	for _, model := range mr.models {
		models = append(models, model)
	}
	return models
}

// ListModelsByBackend returns models for a specific backend
func (mr *Registry) ListModelsByBackend(backend string) []Config {
	var models []Config
	for _, model := range mr.models {
		if model.Backend == backend {
			models = append(models, model)
		}
	}
	return models
}

// ResolveModel determines which model to use based on user input and context
// Priority: explicit model ID > explicit backend > defaults
func (mr *Registry) ResolveModel(modelID string, backend string) (Config, error) {
	// If model ID is specified, use it
	if modelID != "" {
		return mr.GetModel(modelID)
	}

	// If backend is specified, find the default model for that backend
	if backend != "" {
		for _, model := range mr.ListModelsByBackend(backend) {
			if model.IsDefault || model.ID == "gemini-2.5-flash" || model.ID == "gemini-2.5-flash-vertex" {
				return model, nil
			}
		}
		// If no default for backend, return the first model for that backend
		backendModels := mr.ListModelsByBackend(backend)
		if len(backendModels) > 0 {
			return backendModels[0], nil
		}
	}

	// Otherwise use global default
	return mr.GetDefaultModel(), nil
}

// ExtractModelIDFromGemini converts gemini-2.5-flash to the actual model ID for API
// This is because the API model name is just "gemini-2.5-flash", not an ID
func ExtractModelIDFromGemini(modelID string) string {
	// For Gemini API, strip the -vertex suffix if present, but keep the base model
	if strings.HasSuffix(modelID, "-vertex") {
		return strings.TrimSuffix(modelID, "-vertex")
	}
	return modelID
}

// RegisterModelForProvider registers a base model for a specific provider with optional shorthands
// This avoids duplicating model definitions across providers
func (mr *Registry) RegisterModelForProvider(
	provider string,
	baseModelID string,
	shorthands []string,
) error {
	// Verify base model exists
	if _, exists := mr.models[baseModelID]; !exists {
		return fmt.Errorf("base model %q not found", baseModelID)
	}

	// Register provider/fullid → baseModelID alias
	key := fmt.Sprintf("%s/%s", provider, baseModelID)
	mr.aliases[key] = baseModelID

	// Register provider/shorthand → baseModelID aliases
	for _, shorthand := range shorthands {
		key := fmt.Sprintf("%s/%s", provider, shorthand)
		mr.aliases[key] = baseModelID
	}

	// Track models by provider
	mr.modelsByProvider[provider] = append(
		mr.modelsByProvider[provider],
		baseModelID,
	)

	return nil
}

// GetProviderModels returns all models available for a specific provider
func (mr *Registry) GetProviderModels(provider string) []Config {
	modelIDs := mr.modelsByProvider[provider]
	result := make([]Config, 0, len(modelIDs))
	for _, id := range modelIDs {
		if model, err := mr.GetModel(id); err == nil {
			result = append(result, model)
		}
	}
	return result
}

// ListProviders returns a list of all available providers
func (mr *Registry) ListProviders() []string {
	providers := make([]string, 0, len(mr.modelsByProvider))
	for p := range mr.modelsByProvider {
		providers = append(providers, p)
	}
	// Sort for consistent output
	sort.Strings(providers)
	return providers
}

// ResolveFromProviderSyntax resolves a model using provider/model syntax
// Returns the resolved Config based on provider and model identifier
// providerName: explicit provider, or empty string for shorthand
// modelIdentifier: model ID or shorthand (e.g., "flash", "2.5-flash", "gemini-2.5-flash")
// defaultProvider: fallback provider if not specified (e.g., "gemini")
func (mr *Registry) ResolveFromProviderSyntax(
	providerName string,
	modelIdentifier string,
	defaultProvider string,
) (Config, error) {
	// If provider not specified, use default
	if providerName == "" {
		providerName = defaultProvider
	}

	// Try to resolve using alias first (provider/modelid or provider/shorthand)
	aliasKey := fmt.Sprintf("%s/%s", providerName, modelIdentifier)
	if baseModelID, exists := mr.aliases[aliasKey]; exists {
		return mr.GetModel(baseModelID)
	}

	// Try exact model ID lookup if provided
	if model, err := mr.GetModel(modelIdentifier); err == nil {
		// Verify it's available for the requested provider
		providerModels := mr.GetProviderModels(providerName)
		for _, m := range providerModels {
			if m.ID == modelIdentifier {
				return model, nil
			}
		}
	}

	// Generate helpful error message
	return Config{}, fmt.Errorf(
		"model %q not found for provider %q", modelIdentifier, providerName)
}
