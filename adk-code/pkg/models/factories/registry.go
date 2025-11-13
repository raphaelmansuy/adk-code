package factories

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/adk/model"
)

// Registry manages model factories for different providers
type Registry struct {
	mu        sync.RWMutex
	factories map[string]ModelFactory
}

// defaultRegistry is the global factory registry
var defaultRegistry *Registry
var once sync.Once

// GetRegistry returns the default factory registry, initializing if needed
func GetRegistry() *Registry {
	once.Do(func() {
		defaultRegistry = &Registry{
			factories: make(map[string]ModelFactory),
		}
		// Register default factories
		defaultRegistry.Register("gemini", NewGeminiFactory())
		defaultRegistry.Register("openai", NewOpenAIFactory())
		defaultRegistry.Register("vertexai", NewVertexAIFactory())
	})
	return defaultRegistry
}

// Register adds a factory to the registry
func (r *Registry) Register(provider string, factory ModelFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[provider] = factory
}

// Get retrieves a factory by provider name
func (r *Registry) Get(provider string) (ModelFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[provider]
	if !ok {
		return nil, fmt.Errorf("unknown model provider: %s", provider)
	}
	return factory, nil
}

// List returns all registered providers
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]string, 0, len(r.factories))
	for provider := range r.factories {
		providers = append(providers, provider)
	}
	return providers
}

// CreateModel creates a model using the specified provider
func (r *Registry) CreateModel(ctx context.Context, provider string, config ModelConfig) (model.LLM, error) {
	factory, err := r.Get(provider)
	if err != nil {
		return nil, err
	}

	if err := factory.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration for %s: %w", provider, err)
	}

	return factory.Create(ctx, config)
}
