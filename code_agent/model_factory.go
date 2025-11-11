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
	"context"
	"fmt"

	"code_agent/model"

	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// VertexAIConfig holds configuration for Vertex AI backend
type VertexAIConfig struct {
	Project   string
	Location  string
	ModelName string
}

// GeminiConfig holds configuration for Gemini API backend
type GeminiConfig struct {
	APIKey    string
	ModelName string
}

// OpenAIConfig holds configuration for OpenAI API backend
type OpenAIConfig struct {
	APIKey    string
	ModelName string
}

// CreateVertexAIModel creates a Gemini model configured to use Vertex AI backend
// This leverages the Gemini SDK's built-in support for Vertex AI backend
func CreateVertexAIModel(ctx context.Context, cfg VertexAIConfig) (adkmodel.LLM, error) {
	if cfg.Project == "" {
		return nil, fmt.Errorf("Vertex AI project is required")
	}
	if cfg.Location == "" {
		return nil, fmt.Errorf("Vertex AI location is required")
	}
	if cfg.ModelName == "" {
		return nil, fmt.Errorf("model name is required")
	}

	// Create a Gemini model with Vertex AI backend configuration
	// The genai SDK automatically handles the backend differences
	llm, err := gemini.NewModel(ctx, cfg.ModelName, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  cfg.Project,
		Location: cfg.Location,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI model: %w", err)
	}

	return llm, nil
}

// CreateGeminiModel creates a model using the Gemini API backend
func CreateGeminiModel(ctx context.Context, cfg GeminiConfig) (adkmodel.LLM, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}
	if cfg.ModelName == "" {
		return nil, fmt.Errorf("model name is required")
	}

	llm, err := gemini.NewModel(ctx, cfg.ModelName, &genai.ClientConfig{
		Backend: genai.BackendGeminiAPI,
		APIKey:  cfg.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini model: %w", err)
	}

	return llm, nil
}

// CreateOpenAIModel creates a model using the official OpenAI API
// Delegates to the model package implementation
func CreateOpenAIModel(ctx context.Context, cfg OpenAIConfig) (adkmodel.LLM, error) {
	modelCfg := model.OpenAIConfig{
		APIKey:    cfg.APIKey,
		ModelName: cfg.ModelName,
	}
	return model.CreateOpenAIModel(ctx, modelCfg)
}
