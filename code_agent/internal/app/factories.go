package app

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/adk/model"

	"code_agent/display"
	"code_agent/internal/cli"
	"code_agent/internal/config"
	"code_agent/pkg/models"
	"code_agent/pkg/models/factories"
)

// DisplayComponentFactory creates display components with consistent configuration
type DisplayComponentFactory struct {
	config *config.Config
}

// NewDisplayComponentFactory creates a new display component factory
func NewDisplayComponentFactory(cfg *config.Config) *DisplayComponentFactory {
	return &DisplayComponentFactory{
		config: cfg,
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
	config *config.Config
}

// NewModelComponentFactory creates a new model component factory
func NewModelComponentFactory(cfg *config.Config) *ModelComponentFactory {
	return &ModelComponentFactory{
		config: cfg,
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

	// Create LLM model using factory registry
	actualModelID := models.ExtractModelIDFromGemini(selectedModel.ID)
	llm, err := f.createModelUsingFactory(ctx, selectedModel.Backend, actualModelID, apiKey)
	if err != nil {
		return nil, err
	}

	return &ModelComponents{
		Registry: registry,
		Selected: selectedModel,
		LLM:      llm,
	}, nil
}

// createModelUsingFactory creates an LLM model using the factory registry
// This method consolidates model creation logic and validates provider-specific configurations
func (f *ModelComponentFactory) createModelUsingFactory(ctx context.Context, backend, modelName, apiKey string) (model.LLM, error) {
	factoryReg := factories.GetRegistry()

	// Build factory config based on backend
	factoryConfig := factories.ModelConfig{
		ModelName: modelName,
	}

	switch backend {
	case "vertexai":
		if f.config.VertexAIProject == "" {
			return nil, fmt.Errorf("vertex AI backend requires GOOGLE_CLOUD_PROJECT environment variable or --project flag")
		}
		if f.config.VertexAILocation == "" {
			return nil, fmt.Errorf("vertex AI backend requires GOOGLE_CLOUD_LOCATION environment variable or --location flag")
		}
		factoryConfig.Project = f.config.VertexAIProject
		factoryConfig.Location = f.config.VertexAILocation

	case "openai":
		openaiKey := os.Getenv("OPENAI_API_KEY")
		if openaiKey == "" {
			return nil, fmt.Errorf("OpenAI backend requires OPENAI_API_KEY environment variable")
		}
		factoryConfig.APIKey = openaiKey

	case "gemini":
		fallthrough
	default:
		if apiKey == "" {
			return nil, fmt.Errorf("gemini API backend requires GOOGLE_API_KEY environment variable or --api-key flag")
		}
		factoryConfig.APIKey = apiKey
		backend = "gemini"
	}

	// Use factory registry to create the model
	llm, err := factoryReg.CreateModel(ctx, backend, factoryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s model: %w", backend, err)
	}

	return llm, nil
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
