// Package models - Ollama model definitions and registration
package models

// RegisterOllamaModels registers common Ollama models with the registry
// This function should be called during registry initialization
// Note: Ollama is highly flexible and supports any model that can be pulled from ollama.com
// These are just commonly used examples
func RegisterOllamaModels(registry *Registry) {
	// Popular open-source models available on ollama.com
	// Users can pull any model they want with: ollama pull <model-name>

	registry.RegisterModel(Config{
		ID:            "llama2",
		Name:          "Llama 2",
		DisplayName:   "Llama 2 (7B)",
		Backend:       "ollama",
		ContextWindow: 4096,
		Capabilities: Capabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: false,
			CostTier:          "free",
		},
		Description:    "Meta's Llama 2 model. Fast, efficient, open-source.",
		RecommendedFor: []string{"general purpose", "coding"},
		IsDefault:      false,
	})

	registry.RegisterModel(Config{
		ID:            "neural-chat",
		Name:          "Neural Chat",
		DisplayName:   "Neural Chat (7B)",
		Backend:       "ollama",
		ContextWindow: 4096,
		Capabilities: Capabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: false,
			CostTier:          "free",
		},
		Description:    "Intel's Neural Chat model optimized for conversational tasks.",
		RecommendedFor: []string{"conversation", "chat", "instruction following"},
		IsDefault:      false,
	})

	registry.RegisterModel(Config{
		ID:            "mistral",
		Name:          "Mistral",
		DisplayName:   "Mistral (7B)",
		Backend:       "ollama",
		ContextWindow: 32000,
		Capabilities: Capabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "free",
		},
		Description:    "Mistral 7B - High-quality open-source model with 32K context.",
		RecommendedFor: []string{"coding", "analysis", "extended context"},
		IsDefault:      false,
	})

	registry.RegisterModel(Config{
		ID:            "dolphin-mixtral",
		Name:          "Dolphin Mixtral",
		DisplayName:   "Dolphin Mixtral (8x7B)",
		Backend:       "ollama",
		ContextWindow: 32000,
		Capabilities: Capabilities{
			VisionSupport:     false,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "free",
		},
		Description:    "Dolphin Mixtral - High-quality model with function calling support.",
		RecommendedFor: []string{"coding", "tool use", "complex reasoning"},
		IsDefault:      false,
	})

	registry.RegisterModel(Config{
		ID:            "llama2-uncensored",
		Name:          "Llama 2 Uncensored",
		DisplayName:   "Llama 2 Uncensored (7B)",
		Backend:       "ollama",
		ContextWindow: 4096,
		Capabilities: Capabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: false,
			CostTier:          "free",
		},
		Description:    "Uncensored variant of Llama 2 for unrestricted generation.",
		RecommendedFor: []string{"creative writing", "unconstrained tasks"},
		IsDefault:      false,
	})

	registry.RegisterModel(Config{
		ID:            "openhermes",
		Name:          "OpenHermes 2.5",
		DisplayName:   "OpenHermes 2.5 (7B)",
		Backend:       "ollama",
		ContextWindow: 4096,
		Capabilities: Capabilities{
			VisionSupport:     false,
			ToolUseSupport:    true,
			LongContextWindow: false,
			CostTier:          "free",
		},
		Description:    "OpenHermes 2.5 - Function calling capable model.",
		RecommendedFor: []string{"instruction following", "tool use"},
		IsDefault:      false,
	})

	// Register aliases for models
	registry.RegisterModelForProvider("ollama", "llama2", []string{"llama2", "llama-2"})
	registry.RegisterModelForProvider("ollama", "neural-chat", []string{"neural-chat"})
	registry.RegisterModelForProvider("ollama", "mistral", []string{"mistral"})
	registry.RegisterModelForProvider("ollama", "dolphin-mixtral", []string{"dolphin-mixtral"})
	registry.RegisterModelForProvider("ollama", "llama2-uncensored", []string{"llama2-uncensored"})
	registry.RegisterModelForProvider("ollama", "openhermes", []string{"openhermes", "openhermes-2.5"})
}
