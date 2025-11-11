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
	models map[string]ModelConfig
}

// NewModelRegistry creates a new model registry with default models
func NewModelRegistry() *ModelRegistry {
	registry := &ModelRegistry{
		models: make(map[string]ModelConfig),
	}

	// Register default models
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

	// Vertex AI versions (same models, different backend)
	registry.RegisterModel(ModelConfig{
		ID:            "gemini-2.5-flash-vertex",
		Name:          "Gemini 2.5 Flash",
		DisplayName:   "Gemini 2.5 Flash (Vertex AI)",
		Backend:       "vertexai",
		ContextWindow: 1000000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Fast, affordable model via Vertex AI. Same capabilities as Gemini API.",
		RecommendedFor: []string{"coding", "analysis", "enterprise"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gemini-1.5-pro-vertex",
		Name:          "Gemini 1.5 Pro",
		DisplayName:   "Gemini 1.5 Pro (Vertex AI)",
		Backend:       "vertexai",
		ContextWindow: 2000000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Advanced model via Vertex AI. Enterprise deployment option.",
		RecommendedFor: []string{"complex reasoning", "enterprise", "regulated"},
		IsDefault:      false,
	})

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
