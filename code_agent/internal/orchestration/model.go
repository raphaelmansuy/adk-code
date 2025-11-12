package orchestration

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/adk/model"

	"code_agent/internal/cli"
	"code_agent/internal/config"
	"code_agent/pkg/models"
)

// modelInitializer handles model and LLM setup
type modelInitializer struct {
	registry *models.Registry
	selected models.Config
	llm      model.LLM
}

// InitializeModelComponents sets up the LLM model and related components
func InitializeModelComponents(ctx context.Context, cfg *config.Config) (*ModelComponents, error) {
	initializer := &modelInitializer{
		registry: models.NewRegistry(),
	}

	// Resolve which model to use
	var err error
	if cfg.Model == "" {
		initializer.selected = initializer.registry.GetDefaultModel()
	} else {
		parsedProvider, parsedModel, parseErr := cli.ParseProviderModelSyntax(cfg.Model)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid model syntax: %w\nUse format: provider/model (e.g., gemini/2.5-flash)", parseErr)
		}

		defaultProvider := cfg.Backend
		if defaultProvider == "" {
			defaultProvider = "gemini"
		}

		initializer.selected, err = initializer.registry.ResolveFromProviderSyntax(parsedProvider, parsedModel, defaultProvider)
		if err != nil {
			// Print available models and return error
			fmt.Printf("❌ Error: %v\n\nAvailable models:\n", err)
			for _, providerName := range initializer.registry.ListProviders() {
				models := initializer.registry.GetProviderModels(providerName)
				fmt.Printf("\n%s:\n", strings.ToUpper(providerName[:1])+strings.ToLower(providerName[1:]))
				for _, m := range models {
					fmt.Printf("  • %s/%s\n", providerName, m.ID)
				}
			}
			return nil, fmt.Errorf("model resolution failed")
		}
	}

	// Get API key
	apiKey := cfg.APIKey
	if apiKey == "" && initializer.selected.Backend == "gemini" {
		return nil, fmt.Errorf("gemini API backend requires GOOGLE_API_KEY environment variable or --api-key flag")
	}

	// Create LLM model
	actualModelID := models.ExtractModelIDFromGemini(initializer.selected.ID)

	switch initializer.selected.Backend {
	case "vertexai":
		if cfg.VertexAIProject == "" {
			return nil, fmt.Errorf("vertex AI backend requires GOOGLE_CLOUD_PROJECT environment variable or --project flag")
		}
		if cfg.VertexAILocation == "" {
			return nil, fmt.Errorf("vertex AI backend requires GOOGLE_CLOUD_LOCATION environment variable or --location flag")
		}
		initializer.llm, err = models.CreateVertexAIModel(ctx, models.VertexAIConfig{
			Project:   cfg.VertexAIProject,
			Location:  cfg.VertexAILocation,
			ModelName: actualModelID,
		})

	case "openai":
		openaiKey := os.Getenv("OPENAI_API_KEY")
		if openaiKey == "" {
			return nil, fmt.Errorf("openAI backend requires OPENAI_API_KEY environment variable")
		}
		initializer.llm, err = models.CreateOpenAIModel(ctx, models.OpenAIConfig{
			APIKey:    openaiKey,
			ModelName: actualModelID,
		})

	case "gemini":
		fallthrough
	default:
		initializer.llm, err = models.CreateGeminiModel(ctx, models.GeminiConfig{
			APIKey:    apiKey,
			ModelName: actualModelID,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create LLM model: %w", err)
	}

	return &ModelComponents{
		Registry: initializer.registry,
		Selected: initializer.selected,
		LLM:      initializer.llm,
	}, nil
}
