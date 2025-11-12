// Package main - OpenAI model definitions and registration
package main

// RegisterOpenAIModels registers all OpenAI models with the registry
// This function should be called during registry initialization
func RegisterOpenAIModels(registry *ModelRegistry) {
	// Frontier models - GPT-5 series
	registry.RegisterModel(ModelConfig{
		ID:            "gpt-5",
		Name:          "GPT-5",
		DisplayName:   "GPT-5",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Latest frontier model. Best for coding and agentic tasks across all domains.",
		RecommendedFor: []string{"coding", "agentic tasks", "complex reasoning", "advanced analysis"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-5-mini",
		Name:          "GPT-5 Mini",
		DisplayName:   "GPT-5 Mini",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Faster, cost-efficient version of GPT-5 for well-defined tasks.",
		RecommendedFor: []string{"coding", "task completion", "prototyping", "high volume"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-5-nano",
		Name:          "GPT-5 Nano",
		DisplayName:   "GPT-5 Nano",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Fastest and most cost-efficient version of GPT-5. Great for summarization and classification.",
		RecommendedFor: []string{"summarization", "classification", "rapid iteration", "cost-sensitive"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-5-pro",
		Name:          "GPT-5 Pro",
		DisplayName:   "GPT-5 Pro",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "The smartest and most precise model. Produces the most accurate responses.",
		RecommendedFor: []string{"precision tasks", "complex analysis", "critical applications"},
		IsDefault:      false,
	})

	// GPT-4.1 series
	registry.RegisterModel(ModelConfig{
		ID:            "gpt-4.1",
		Name:          "GPT-4.1",
		DisplayName:   "GPT-4.1",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Smartest non-reasoning model. High intelligence for general tasks.",
		RecommendedFor: []string{"coding", "analysis", "reasoning", "general intelligence"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-4.1-mini",
		Name:          "GPT-4.1 Mini",
		DisplayName:   "GPT-4.1 Mini",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Smaller and faster version of GPT-4.1 for focused tasks.",
		RecommendedFor: []string{"rapid tasks", "cost-effective coding", "prototyping"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-4.1-nano",
		Name:          "GPT-4.1 Nano",
		DisplayName:   "GPT-4.1 Nano",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Very small and fast model for simple, focused tasks.",
		RecommendedFor: []string{"simple tasks", "low cost", "high volume"},
		IsDefault:      false,
	})

	// Reasoning models - O-series
	registry.RegisterModel(ModelConfig{
		ID:            "gpt-5-codex",
		Name:          "GPT-5 Codex",
		DisplayName:   "GPT-5 Codex",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Specialized version of GPT-5 optimized for agentic coding.",
		RecommendedFor: []string{"coding", "code generation", "programming agents"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "o4-mini",
		Name:          "o4-mini",
		DisplayName:   "o4-mini (Fast Reasoning)",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Fast, cost-efficient reasoning model. Successor to o3-mini.",
		RecommendedFor: []string{"reasoning", "problem solving", "quick inference"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "o3",
		Name:          "o3",
		DisplayName:   "o3 (Deep Reasoning)",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Reasoning model for complex tasks. Predecessor to GPT-5.",
		RecommendedFor: []string{"complex reasoning", "mathematics", "deep analysis"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "o3-mini",
		Name:          "o3-mini",
		DisplayName:   "o3-mini (Lightweight Reasoning)",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Small reasoning model alternative to o3.",
		RecommendedFor: []string{"reasoning", "coding", "efficient inference"},
		IsDefault:      false,
	})

	// Vision and older models
	registry.RegisterModel(ModelConfig{
		ID:            "gpt-4o",
		Name:          "GPT-4o",
		DisplayName:   "GPT-4o",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Fast, intelligent, flexible model. Multimodal with vision support.",
		RecommendedFor: []string{"coding", "vision", "analysis", "general tasks"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gpt-4o-mini",
		Name:          "GPT-4o Mini",
		DisplayName:   "GPT-4o Mini",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Fast, affordable small model for focused tasks.",
		RecommendedFor: []string{"rapid prototyping", "high volume", "cost-effective"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "o1",
		Name:          "o1",
		DisplayName:   "o1 (Previous Reasoning)",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Previous full o-series reasoning model. Solid for complex tasks.",
		RecommendedFor: []string{"reasoning", "mathematics", "complex problem solving"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "o1-mini",
		Name:          "o1-mini",
		DisplayName:   "o1-mini (Deprecated)",
		Backend:       "openai",
		ContextWindow: 128000,
		Capabilities: ModelCapabilities{
			VisionSupport:     false,
			ToolUseSupport:    false,
			LongContextWindow: true,
			CostTier:          "standard",
		},
		Description:    "Small reasoning model alternative to o1 (Deprecated, use o4-mini instead).",
		RecommendedFor: []string{"reasoning", "cost-effective inference"},
		IsDefault:      false,
	})

	// Register OpenAI models for provider
	registry.RegisterModelForProvider(
		"openai",
		"gpt-5",
		[]string{"5", "latest", "best", "frontier"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-5-mini",
		[]string{"5-mini", "5m"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-5-nano",
		[]string{"5-nano", "5n"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-5-pro",
		[]string{"5-pro", "5p"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-4.1",
		[]string{"4.1"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-4.1-mini",
		[]string{"4.1-mini", "4.1m"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-4.1-nano",
		[]string{"4.1-nano", "4.1n"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-5-codex",
		[]string{"codex", "5-codex"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"o4-mini",
		[]string{"o4-mini", "o4m"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"o3",
		[]string{"o3", "reasoning"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"o3-mini",
		[]string{"o3-mini", "o3m"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-4o",
		[]string{"4o"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"gpt-4o-mini",
		[]string{"4o-mini", "mini", "fast"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"o1",
		[]string{"o1"},
	)
	registry.RegisterModelForProvider(
		"openai",
		"o1-mini",
		[]string{"o1-mini"},
	)
}
