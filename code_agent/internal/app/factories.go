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

package app

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/adk/model"

	"code_agent/display"
	"code_agent/pkg/cli"
	"code_agent/pkg/models"
)

// DisplayComponentFactory creates display components with consistent configuration
type DisplayComponentFactory struct {
	config *cli.CLIConfig
}

// NewDisplayComponentFactory creates a new display component factory
func NewDisplayComponentFactory(config *cli.CLIConfig) *DisplayComponentFactory {
	return &DisplayComponentFactory{
		config: config,
	}
}

// Create builds all display components
func (f *DisplayComponentFactory) Create() (*DisplayComponents, error) {
	renderer, err := display.NewRenderer(f.config.OutputFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	typewriter := display.NewTypewriterPrinter(display.DefaultTypewriterConfig())
	typewriter.SetEnabled(f.config.TypewriterEnabled)
	streamDisplay := display.NewStreamingDisplay(renderer, typewriter)

	return &DisplayComponents{
		Renderer:       renderer,
		BannerRenderer: display.NewBannerRenderer(renderer),
		Typewriter:     typewriter,
		StreamDisplay:  streamDisplay,
	}, nil
}

// ModelComponentFactory creates model-related components
type ModelComponentFactory struct {
	config *cli.CLIConfig
}

// NewModelComponentFactory creates a new model component factory
func NewModelComponentFactory(config *cli.CLIConfig) *ModelComponentFactory {
	return &ModelComponentFactory{
		config: config,
	}
}

// Create builds model components and creates the LLM instance
func (f *ModelComponentFactory) Create(ctx context.Context, displayComponents *DisplayComponents) (*ModelComponents, error) {
	registry := models.NewRegistry()

	// Resolve which model to use
	var selectedModel models.Config
	var err error
	if f.config.Model == "" {
		selectedModel = registry.GetDefaultModel()
	} else {
		parsedProvider, parsedModel, parseErr := cli.ParseProviderModelSyntax(f.config.Model)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid model syntax: %w\nUse format: provider/model (e.g., gemini/2.5-flash)", parseErr)
		}

		defaultProvider := f.config.Backend
		if defaultProvider == "" {
			defaultProvider = "gemini"
		}

		selectedModel, err = registry.ResolveFromProviderSyntax(parsedProvider, parsedModel, defaultProvider)
		if err != nil {
			// Print available models and return error
			fmt.Printf("❌ Error: %v\n\nAvailable models:\n", err)
			for _, providerName := range registry.ListProviders() {
				models := registry.GetProviderModels(providerName)
				fmt.Printf("\n%s:\n", providerName)
				for _, m := range models {
					fmt.Printf("  • %s/%s\n", providerName, m.ID)
				}
			}
			return nil, fmt.Errorf("model resolution failed")
		}
	}

	// Get API key
	apiKey := f.config.APIKey
	if apiKey == "" && selectedModel.Backend == "gemini" {
		return nil, fmt.Errorf("gemini API backend requires GOOGLE_API_KEY environment variable or --api-key flag")
	}

	// Resolve working directory
	workingDir := f.resolveWorkingDirectory()

	// Print welcome banner
	displayName := selectedModel.DisplayName
	banner := displayComponents.BannerRenderer.RenderStartBanner("1.0.0", displayName, workingDir)
	fmt.Print(banner)

	// Create LLM model based on backend
	actualModelID := models.ExtractModelIDFromGemini(selectedModel.ID)
	var llm model.LLM

	switch selectedModel.Backend {
	case "vertexai":
		if f.config.VertexAIProject == "" {
			return nil, fmt.Errorf("vertex AI backend requires GOOGLE_CLOUD_PROJECT environment variable or --project flag")
		}
		if f.config.VertexAILocation == "" {
			return nil, fmt.Errorf("vertex AI backend requires GOOGLE_CLOUD_LOCATION environment variable or --location flag")
		}
		vertexModel, err := models.CreateVertexAIModel(ctx, models.VertexAIConfig{
			Project:   f.config.VertexAIProject,
			Location:  f.config.VertexAILocation,
			ModelName: actualModelID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Vertex AI model: %w", err)
		}
		llm = vertexModel

	case "openai":
		openaiKey := os.Getenv("OPENAI_API_KEY")
		if openaiKey == "" {
			return nil, fmt.Errorf("OpenAI backend requires OPENAI_API_KEY environment variable")
		}
		openaiModel, err := models.CreateOpenAIModel(ctx, models.OpenAIConfig{
			APIKey:    openaiKey,
			ModelName: actualModelID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI model: %w", err)
		}
		llm = openaiModel

	case "gemini":
		fallthrough
	default:
		geminiModel, err := models.CreateGeminiModel(ctx, models.GeminiConfig{
			APIKey:    apiKey,
			ModelName: actualModelID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini model: %w", err)
		}
		llm = geminiModel
	}

	return &ModelComponents{
		Registry: registry,
		Selected: selectedModel,
		LLM:      llm,
	}, nil
}

// resolveWorkingDirectory resolves and validates the working directory
func (f *ModelComponentFactory) resolveWorkingDirectory() string {
	workingDir := f.config.WorkingDirectory
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			workingDir = "."
		}
	}

	// Expand ~ in the path
	if len(workingDir) > 0 && workingDir[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return workingDir
		}
		if len(workingDir) > 1 {
			workingDir = homeDir + workingDir[1:]
		} else {
			workingDir = homeDir
		}
	}

	return workingDir
}
