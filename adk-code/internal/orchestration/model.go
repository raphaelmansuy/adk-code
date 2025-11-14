package orchestration

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/adk/model"

	"adk-code/internal/cli"
	"adk-code/internal/config"
	"adk-code/internal/llm"
	"adk-code/pkg/models"
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
			// For Ollama, try dynamic discovery before giving up
			resolvedProvider := parsedProvider
			if resolvedProvider == "" {
				resolvedProvider = defaultProvider
			}

			if resolvedProvider == "ollama" {
				// Try dynamic Ollama discovery
				dynamicModel, discoveryErr := tryOllamaDynamicDiscovery(ctx, parsedModel)
				if dynamicModel != nil {
					initializer.selected = *dynamicModel
					err = nil
				} else {
					// Still failed, show both static and dynamic available models
					if discoveryErr != nil {
						fmt.Printf("❌ Error: %v (discovery error: %v)\n\nAvailable models:\n", err, discoveryErr)
					} else {
						fmt.Printf("❌ Error: %v\n\nAvailable models:\n", err)
					}
					printAvailableModels(initializer.registry, ctx, resolvedProvider)
					return nil, fmt.Errorf("model resolution failed")
				}
			} else {
				// For other providers, show registry models
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

	case "ollama":
		initializer.llm, err = models.CreateOllamaModel(ctx, models.OllamaConfig{
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

// tryOllamaDynamicDiscovery attempts to resolve a model using Ollama's dynamic discovery
// Returns a dynamically created Config if the model exists on the Ollama server
func tryOllamaDynamicDiscovery(ctx context.Context, modelName string) (*models.Config, error) {
	// Get the Ollama provider backend from the LLM registry
	llmRegistry := llm.NewRegistry()
	providerBackend, err := llmRegistry.Get("ollama")
	if err != nil {
		return nil, err
	}

	// Check if provider supports model discovery
	discoverable, ok := providerBackend.(llm.ModelDiscovery)
	if !ok {
		return nil, fmt.Errorf("ollama provider does not support model discovery")
	}

	// Query for the model
	_, err = discoverable.GetModelInfo(ctx, modelName)
	if err != nil {
		return nil, err
	}

	// Model exists! Create a dynamic Config for it
	dynamicModel := models.Config{
		ID:            modelName,
		Name:          modelName,
		DisplayName:   modelName,
		Backend:       "ollama",
		ContextWindow: 4096, // Default context window
		Capabilities: models.Capabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: false,
			CostTier:          "free",
		},
		Description:    fmt.Sprintf("Ollama model: %s (dynamically discovered)", modelName),
		RecommendedFor: []string{"general purpose"},
		IsDefault:      false,
	}

	return &dynamicModel, nil
}

// printAvailableModels prints available models from registry and dynamic sources
func printAvailableModels(registry *models.Registry, ctx context.Context, preferredProvider string) {
	// For Ollama, show actual models from the server
	if preferredProvider == "ollama" {
		// Try to get dynamic models from Ollama
		llmRegistry := llm.NewRegistry()
		providerBackend, err := llmRegistry.Get("ollama")
		if err == nil {
			if discoverable, ok := providerBackend.(llm.ModelDiscovery); ok {
				dynamicModels, err := discoverable.ListModels(ctx, false)
				if err == nil && len(dynamicModels) > 0 {
					fmt.Printf("\nOllama (from running server):\n")
					for _, m := range dynamicModels {
						fmt.Printf("  • ollama/%s\n", m.Name)
					}
					return
				}
			}
		}
	}

	// Fallback to static registry models
	for _, providerName := range registry.ListProviders() {
		registryModels := registry.GetProviderModels(providerName)
		fmt.Printf("\n%s:\n", strings.ToUpper(providerName[:1])+strings.ToLower(providerName[1:]))
		for _, m := range registryModels {
			fmt.Printf("  • %s/%s\n", providerName, m.ID)
		}
	}
}
