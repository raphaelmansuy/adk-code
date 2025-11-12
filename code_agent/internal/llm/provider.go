// Package llm - Internal LLM provider abstraction layer
package llm

import (
	"context"

	adkmodel "google.golang.org/adk/model"

	"code_agent/internal/llm/backends"
	"code_agent/pkg/models"
)

// Provider represents a supported LLM backend provider
type Provider string

const (
	ProviderGemini   Provider = "gemini"
	ProviderVertexAI Provider = "vertexai"
	ProviderOpenAI   Provider = "openai"
)

// ProviderBackend is the interface that all LLM provider implementations must satisfy
// This abstracts the creation and validation logic for different LLM backends
type ProviderBackend interface {
	// Create instantiates a new LLM model with the given context and configuration
	// Returns an ADK-compatible LLM instance or an error
	Create(ctx context.Context, config any) (adkmodel.LLM, error)

	// Validate checks if the provided configuration is valid for this provider
	// Returns nil if valid, or an error describing what's missing/wrong
	Validate(config any) error

	// GetMetadata returns information about this provider
	GetMetadata() models.ProviderMetadata

	// Name returns the provider name (e.g., "gemini", "vertexai", "openai")
	Name() string
}

// Registry manages available LLM providers
type Registry struct {
	providers map[string]ProviderBackend
}

// NewRegistry creates a new provider registry with all built-in providers
func NewRegistry() *Registry {
	r := &Registry{
		providers: make(map[string]ProviderBackend),
	}

	// Register all built-in providers
	r.Register(backends.NewGeminiProvider())
	r.Register(backends.NewVertexAIProvider())
	r.Register(backends.NewOpenAIProvider())

	return r
}

// Register adds a provider to the registry
func (r *Registry) Register(backend ProviderBackend) {
	r.providers[backend.Name()] = backend
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (ProviderBackend, error) {
	if provider, ok := r.providers[name]; ok {
		return provider, nil
	}
	return nil, &ProviderNotFoundError{Name: name}
}

// GetMetadata retrieves metadata for a provider
func (r *Registry) GetMetadata(name string) (models.ProviderMetadata, error) {
	provider, err := r.Get(name)
	if err != nil {
		return models.ProviderMetadata{}, err
	}
	return provider.GetMetadata(), nil
}

// ProviderNotFoundError is returned when a requested provider is not found
type ProviderNotFoundError struct {
	Name string
}

func (e *ProviderNotFoundError) Error() string {
	return "provider not found: " + e.Name
}

// CreateLLMFromConfig creates an LLM instance using the appropriate provider
// based on the backend name in the config
func CreateLLMFromConfig(ctx context.Context, backend string, cfg any) (adkmodel.LLM, error) {
	registry := NewRegistry()
	provider, err := registry.Get(backend)
	if err != nil {
		return nil, err
	}

	if err := provider.Validate(cfg); err != nil {
		return nil, err
	}

	return provider.Create(ctx, cfg)
}
