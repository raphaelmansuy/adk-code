// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"sort"
	"strings"
)

// ModelCapabilities represents the capabilities of a model
type ModelCapabilities struct {
	VisionSupport     bool   // Can process images
	ToolUseSupport    bool   // Can use tools/functions
	LongContextWindow bool   // Has extended context length
	CostTier          string // "economy", "standard", "premium"
}

// ModelConfig holds configuration for a specific model
type ModelConfig struct {
	ID             string
	Name           string
	DisplayName    string
	Backend        string // "gemini" or "vertexai"
	ContextWindow  int    // tokens
	Capabilities   ModelCapabilities
	Description    string
	RecommendedFor []string // Use cases: "coding", "analysis", "creative", etc.
	IsDefault      bool
}

// ModelRegistry manages available models and configurations
type ModelRegistry struct {
	models           map[string]ModelConfig // model ID → config
	aliases          map[string]string      // "provider/shorthand" → model ID
	modelsByProvider map[string][]string    // provider → list of model IDs
}

// NewModelRegistry creates a new model registry with default models
// Models are registered once and then aliased for each provider to avoid duplication
func NewModelRegistry() *ModelRegistry {
	registry := &ModelRegistry{
		models:           make(map[string]ModelConfig),
		aliases:          make(map[string]string),
		modelsByProvider: make(map[string][]string),
	}

	// Define base models ONCE (no more -vertex duplicates!)
	registry.RegisterModel(ModelConfig{
		ID:            "gemini-2.5-flash",
		Name:          "Gemini 2.5 Flash",
		DisplayName:   "Gemini 2.5 Flash",
		Backend:       "gemini",
		ContextWindow: 1000000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Fast, affordable multimodal model. Best for real-time applications.",
		RecommendedFor: []string{"coding", "analysis", "rapid iteration"},
		IsDefault:      true,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gemini-2.0-flash",
		Name:          "Gemini 2.0 Flash",
		DisplayName:   "Gemini 2.0 Flash",
		Backend:       "gemini",
		ContextWindow: 1000000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Previous generation fast model. Still powerful and cost-effective.",
		RecommendedFor: []string{"coding", "prototyping"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gemini-1.5-flash",
		Name:          "Gemini 1.5 Flash",
		DisplayName:   "Gemini 1.5 Flash",
		Backend:       "gemini",
		ContextWindow: 1000000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Earlier flash model with large context window.",
		RecommendedFor: []string{"coding", "document processing"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gemini-1.5-pro",
		Name:          "Gemini 1.5 Pro",
		DisplayName:   "Gemini 1.5 Pro",
		Backend:       "gemini",
		ContextWindow: 2000000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Advanced reasoning model. Best for complex tasks.",
		RecommendedFor: []string{"complex reasoning", "analysis", "creative"},
		IsDefault:      false,
	})

	// Register each base model for both providers with shorthands
	// This eliminates the need for -vertex duplicate entries
	registry.RegisterModelForProvider(
		"gemini",
		"gemini-2.5-flash",
		[]string{"2.5-flash", "flash", "latest"},
	)
	registry.RegisterModelForProvider(
		"gemini",
		"gemini-2.0-flash",
		[]string{"2.0-flash"},
	)
	registry.RegisterModelForProvider(
		"gemini",
		"gemini-1.5-flash",
		[]string{"1.5-flash"},
	)
	registry.RegisterModelForProvider(
		"gemini",
		"gemini-1.5-pro",
		[]string{"1.5-pro", "pro"},
	)

	// Register same models for Vertex AI
	registry.RegisterModelForProvider(
		"vertexai",
		"gemini-2.5-flash",
		[]string{"2.5-flash", "flash", "latest"},
	)
	registry.RegisterModelForProvider(
		"vertexai",
		"gemini-2.0-flash",
		[]string{"2.0-flash"},
	)
	registry.RegisterModelForProvider(
		"vertexai",
		"gemini-1.5-flash",
		[]string{"1.5-flash"},
	)
	registry.RegisterModelForProvider(
		"vertexai",
		"gemini-1.5-pro",
		[]string{"1.5-pro", "pro"},
	)

	// Register OpenAI models (Latest as of November 2025)
	// Frontier models - GPT-5 series
	registry.RegisterModel(ModelConfig{
		ID:            "gpt-5",
		Name:          "GPT-5",
		DisplayName:   "GPT-5",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Latest frontier model. Best for coding and agentic tasks across all domains.",
		RecommendedFor: []string{"coding", "agentic tasks", "complex reasoning", "advanced analysis"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-5-mini",
		Name:          "GPT-5 Mini",
		DisplayName:   "GPT-5 Mini",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Faster, cost-efficient version of GPT-5 for well-defined tasks.",
		RecommendedFor: []string{"coding", "task completion", "prototyping", "high volume"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-5-nano",
		Name:          "GPT-5 Nano",
		DisplayName:   "GPT-5 Nano",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Fastest and most cost-efficient version of GPT-5. Great for summarization and classification.",
		RecommendedFor: []string{"summarization", "classification", "rapid iteration", "cost-sensitive"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-5-pro",
		Name:          "GPT-5 Pro",
		DisplayName:   "GPT-5 Pro",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "The smartest and most precise model. Produces the most accurate responses.",
		RecommendedFor: []string{"precision tasks", "complex analysis", "critical applications"},
		IsDefault:      false,
	})

	// GPT-4.1 series
	registry.RegisterModel(ModelConfig{
		ID:            "gpt-4.1",
		Name:          "GPT-4.1",
		DisplayName:   "GPT-4.1",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Smartest non-reasoning model. High intelligence for general tasks.",
		RecommendedFor: []string{"coding", "analysis", "reasoning", "general intelligence"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-4.1-mini",
		Name:          "GPT-4.1 Mini",
		DisplayName:   "GPT-4.1 Mini",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Smaller and faster version of GPT-4.1 for focused tasks.",
		RecommendedFor: []string{"rapid tasks", "cost-effective coding", "prototyping"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-4.1-nano",
		Name:          "GPT-4.1 Nano",
		DisplayName:   "GPT-4.1 Nano",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Very small and fast model for simple, focused tasks.",
		RecommendedFor: []string{"simple tasks", "low cost", "high volume"},
		IsDefault:      false,
	})

	// Reasoning models - O-series
	registry.RegisterModel(ModelConfig{
		ID:            "gpt-5-codex",
		Name:          "GPT-5 Codex",
		DisplayName:   "GPT-5 Codex",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Specialized version of GPT-5 optimized for agentic coding.",
		RecommendedFor: []string{"coding", "code generation", "programming agents"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "o4-mini",
		Name:          "o4-mini",
		DisplayName:   "o4-mini (Fast Reasoning)",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Fast, cost-efficient reasoning model. Successor to o3-mini.",
		RecommendedFor: []string{"reasoning", "problem solving", "quick inference"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "o3",
		Name:          "o3",
		DisplayName:   "o3 (Deep Reasoning)",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Reasoning model for complex tasks. Predecessor to GPT-5.",
		RecommendedFor: []string{"complex reasoning", "mathematics", "deep analysis"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "o3-mini",
		Name:          "o3-mini",
		DisplayName:   "o3-mini (Lightweight Reasoning)",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Small reasoning model alternative to o3.",
		RecommendedFor: []string{"reasoning", "coding", "efficient inference"},
		IsDefault:      false,
	})

	// Vision and older models
	registry.RegisterModel(ModelConfig{
		ID:            "gpt-4o",
		Name:          "GPT-4o",
		DisplayName:   "GPT-4o",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Fast, intelligent, flexible model. Multimodal with vision support.",
		RecommendedFor: []string{"coding", "vision", "analysis", "general tasks"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-4o-mini",
		Name:          "GPT-4o Mini",
		DisplayName:   "GPT-4o Mini",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Fast, affordable small model for focused tasks.",
		RecommendedFor: []string{"rapid prototyping", "high volume", "cost-effective"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "o1",
		Name:          "o1",
		DisplayName:   "o1 (Previous Reasoning)",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Previous full o-series reasoning model. Solid for complex tasks.",
		RecommendedFor: []string{"reasoning", "mathematics", "complex problem solving"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "o1-mini",
		Name:          "o1-mini",
		DisplayName:   "o1-mini (Deprecated)",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Small reasoning model alternative to o1 (Deprecated, use o4-mini instead).",
		RecommendedFor: []string{"reasoning", "cost-effective inference"},
		IsDefault:      false,
	})

	// Register OpenAI models for provider
	registry.RegisterModelForProvider(
		"openai",
		"gpt-5",
		[]string{"5", "latest", "best", "frontier"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-5-mini",
		[]string{"5-mini", "5m"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-5-nano",
		[]string{"5-nano", "5n"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-5-pro",
		[]string{"5-pro", "5p"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-4.1",
		[]string{"4.1"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-4.1-mini",
		[]string{"4.1-mini", "4.1m"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-4.1-nano",
		[]string{"4.1-nano", "4.1n"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-5-codex",
		[]string{"codex", "5-codex"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"o4-mini",
		[]string{"o4-mini", "o4m"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"o3",
		[]string{"o3", "reasoning"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"o3-mini",
		[]string{"o3-mini", "o3m"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-4o",
		[]string{"4o"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-4o-mini",
		[]string{"4o-mini", "mini", "fast"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"o1",
		[]string{"o1"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"o1-mini",
		[]string{"o1-mini"},
	)

	return registry
}

// RegisterModel adds a model to the registry
func (mr *ModelRegistry) RegisterModel(model ModelConfig) {
	mr.models[model.ID] = model
}

// GetModel retrieves a model by ID
func (mr *ModelRegistry) GetModel(id string) (ModelConfig, error) {
	model, exists := mr.models[id]
	if !exists {
		return ModelConfig{}, fmt.Errorf("model %q not found in registry", id)
	}
	return model, nil
}

// GetModelByName retrieves a model by display name (case-insensitive)
func (mr *ModelRegistry) GetModelByName(name string) (ModelConfig, error) {
	name = strings.ToLower(name)
	for _, model := range mr.models {
		if strings.ToLower(model.Name) == name {
			return model, nil
		}
	}
	return ModelConfig{}, fmt.Errorf("model %q not found", name)
}

// GetDefaultModel returns the default model
func (mr *ModelRegistry) GetDefaultModel() ModelConfig {
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
func (mr *ModelRegistry) ListModels() []ModelConfig {
	models := make([]ModelConfig, 0, len(mr.models))
	for _, model := range mr.models {
		models = append(models, model)
	}
	return models
}

// ListModelsByBackend returns models for a specific backend
func (mr *ModelRegistry) ListModelsByBackend(backend string) []ModelConfig {
	var models []ModelConfig
	for _, model := range mr.models {
		if model.Backend == backend {
			models = append(models, model)
		}
	}
	return models
}

// ResolveModel determines which model to use based on user input and context
// Priority: explicit model ID > explicit backend > defaults
func (mr *ModelRegistry) ResolveModel(modelID string, backend string) (ModelConfig, error) {
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
func (mr *ModelRegistry) RegisterModelForProvider(
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
func (mr *ModelRegistry) GetProviderModels(provider string) []ModelConfig {
	modelIDs := mr.modelsByProvider[provider]
	result := make([]ModelConfig, 0, len(modelIDs))
	for _, id := range modelIDs {
		if model, err := mr.GetModel(id); err == nil {
			result = append(result, model)
		}
	}
	return result
}

// ListProviders returns a list of all available providers
func (mr *ModelRegistry) ListProviders() []string {
	providers := make([]string, 0, len(mr.modelsByProvider))
	for p := range mr.modelsByProvider {
		providers = append(providers, p)
	}
	// Sort for consistent output
	sort.Strings(providers)
	return providers
}

// ResolveFromProviderSyntax resolves a model using provider/model syntax
// Returns the resolved ModelConfig based on provider and model identifier
// providerName: explicit provider, or empty string for shorthand
// modelIdentifier: model ID or shorthand (e.g., "flash", "2.5-flash", "gemini-2.5-flash")
// defaultProvider: fallback provider if not specified (e.g., "gemini")
func (mr *ModelRegistry) ResolveFromProviderSyntax(
	providerName string,
	modelIdentifier string,
	defaultProvider string,
) (ModelConfig, error) {
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
	return ModelConfig{}, fmt.Errorf(
		"model %q not found for provider %q", modelIdentifier, providerName)
}
