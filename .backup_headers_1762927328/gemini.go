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

	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// GeminiFactory creates Gemini API models
type GeminiFactory struct{}

// Create builds a Gemini model with the provided configuration
func (f *GeminiFactory) Create(ctx context.Context, config ModelConfig) (model.LLM, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, err
	}

	llm, err := gemini.NewModel(ctx, config.ModelName, &genai.ClientConfig{
		Backend: genai.BackendGeminiAPI,
		APIKey:  config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini model: %w", err)
	}

	return llm, nil
}

// ValidateConfig checks if the configuration is valid for Gemini
func (f *GeminiFactory) ValidateConfig(config ModelConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("Gemini API key is required")
	}
	if config.ModelName == "" {
		return fmt.Errorf("model name is required")
	}
	return nil
}

// Info returns metadata about the Gemini factory
func (f *GeminiFactory) Info() FactoryInfo {
	return FactoryInfo{
		Provider:      "Gemini",
		Description:   "Google Gemini API models",
		RequiredField: []string{"APIKey", "ModelName"},
	}
}

// NewGeminiFactory creates a new Gemini factory instance
func NewGeminiFactory() *GeminiFactory {
	return &GeminiFactory{}
}
