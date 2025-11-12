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

package factories

import (
	"context"

	"google.golang.org/adk/model"
)

// ModelFactory defines the interface for creating LLM models
// Each provider (Gemini, OpenAI, Vertex AI) implements this interface
type ModelFactory interface {
	// Create builds an LLM model instance with the provided configuration
	Create(ctx context.Context, config ModelConfig) (model.LLM, error)

	// ValidateConfig checks if the provided configuration is valid for this factory
	ValidateConfig(config ModelConfig) error

	// Info returns metadata about this factory
	Info() FactoryInfo
}

// ModelConfig holds configuration for model creation
// Different providers may use different fields
type ModelConfig struct {
	// Common fields
	ModelName string

	// Gemini API specific
	APIKey string

	// Vertex AI specific
	Project  string
	Location string

	// OpenAI specific
	// Uses APIKey field
}

// FactoryInfo contains metadata about a factory
type FactoryInfo struct {
	Provider      string // "Gemini", "OpenAI", "VertexAI"
	Description   string
	RequiredField []string // Required fields for this provider
}
