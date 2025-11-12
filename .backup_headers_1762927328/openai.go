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
	"fmt"

	"code_agent/pkg/models"

	"google.golang.org/adk/model"
)

// OpenAIFactory creates OpenAI models
type OpenAIFactory struct{}

// Create builds an OpenAI model with the provided configuration
func (f *OpenAIFactory) Create(ctx context.Context, config ModelConfig) (model.LLM, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, err
	}

	// Delegate to the existing OpenAI model creation logic
	cfg := models.OpenAIConfig{
		APIKey:    config.APIKey,
		ModelName: config.ModelName,
	}

	llm, err := models.CreateOpenAIModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI model: %w", err)
	}

	return llm, nil
}

// ValidateConfig checks if the configuration is valid for OpenAI
func (f *OpenAIFactory) ValidateConfig(config ModelConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("OpenAI API key is required")
	}
	if config.ModelName == "" {
		return fmt.Errorf("model name is required")
	}
	return nil
}

// Info returns metadata about the OpenAI factory
func (f *OpenAIFactory) Info() FactoryInfo {
	return FactoryInfo{
		Provider:      "OpenAI",
		Description:   "OpenAI API models",
		RequiredField: []string{"APIKey", "ModelName"},
	}
}

// NewOpenAIFactory creates a new OpenAI factory instance
func NewOpenAIFactory() *OpenAIFactory {
	return &OpenAIFactory{}
}
