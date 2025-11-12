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

// Note: The original models.go has been refactored into multiple focused files:
// - models_types.go: Core type definitions (ModelCapabilities, ModelConfig, ModelRegistry)
// - models_registry.go: Registry implementation and lookup methods
// - models_gemini.go: Gemini and Vertex AI model definitions
// - models_openai.go: OpenAI model definitions (GPT-5, GPT-4.1, O-series)
//
// The registry constructor below calls the registration functions from the split files.

// NewModelRegistry creates a new model registry with all default models pre-registered
func NewModelRegistry() *ModelRegistry {
	registry := &ModelRegistry{
		models:           make(map[string]ModelConfig),
		aliases:          make(map[string]string),
		modelsByProvider: make(map[string][]string),
	}

	// Register all model definitions from the split files
	RegisterGeminiAndVertexAIModels(registry)
	RegisterOpenAIModels(registry)

	return registry
}
