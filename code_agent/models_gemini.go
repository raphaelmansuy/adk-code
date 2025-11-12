// Package main - Gemini model definitions and registration
package main

// RegisterGeminiAndVertexAIModels registers all Gemini and Vertex AI models with the registry
// This function should be called during registry initialization
func RegisterGeminiAndVertexAIModels(registry *ModelRegistry) {
	// Define base models ONCE (no more -vertex duplicates!)
	registry.RegisterModel(ModelConfig{
		ID:            "gemini-2.5-flash",
		Name:          "Gemini 2.5 Flash",
		DisplayName:   "Gemini 2.5 Flash",
		Backend:       "gemini",
		ContextWindow: 1000000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Fast, affordable multimodal model. Best for real-time applications.",
		RecommendedFor: []string{"coding", "analysis", "rapid iteration"},
		IsDefault:      true,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gemini-2.0-flash",
		Name:          "Gemini 2.0 Flash",
		DisplayName:   "Gemini 2.0 Flash",
		Backend:       "gemini",
		ContextWindow: 1000000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Previous generation fast model. Still powerful and cost-effective.",
		RecommendedFor: []string{"coding", "prototyping"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gemini-1.5-flash",
		Name:          "Gemini 1.5 Flash",
		DisplayName:   "Gemini 1.5 Flash",
		Backend:       "gemini",
		ContextWindow: 1000000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "economy",
		},
		Description:    "Earlier flash model with large context window.",
		RecommendedFor: []string{"coding", "document processing"},
		IsDefault:      false,
	})

	registry.RegisterModel(ModelConfig{
		ID:            "gemini-1.5-pro",
		Name:          "Gemini 1.5 Pro",
		DisplayName:   "Gemini 1.5 Pro",
		Backend:       "gemini",
		ContextWindow: 2000000,
		Capabilities: ModelCapabilities{
			VisionSupport:     true,
			ToolUseSupport:    true,
			LongContextWindow: true,
			CostTier:          "premium",
		},
		Description:    "Advanced reasoning model. Best for complex tasks.",
		RecommendedFor: []string{"complex reasoning", "analysis", "creative"},
		IsDefault:      false,
	})

	// Register each base model for both providers with shorthands
	// This eliminates the need for -vertex duplicate entries
	registry.RegisterModelForProvider(
		"gemini",
		"gemini-2.5-flash",
		[]string{"2.5-flash", "flash", "latest"},
	)
	registry.RegisterModelForProvider(
		"gemini",
		"gemini-2.0-flash",
		[]string{"2.0-flash"},
	)
	registry.RegisterModelForProvider(
		"gemini",
		"gemini-1.5-flash",
		[]string{"1.5-flash"},
	)
	registry.RegisterModelForProvider(
		"gemini",
		"gemini-1.5-pro",
		[]string{"1.5-pro", "pro"},
	)

	// Register same models for Vertex AI
	registry.RegisterModelForProvider(
		"vertexai",
		"gemini-2.5-flash",
		[]string{"2.5-flash", "flash", "latest"},
	)
	registry.RegisterModelForProvider(
		"vertexai",
		"gemini-2.0-flash",
		[]string{"2.0-flash"},
	)
	registry.RegisterModelForProvider(
		"vertexai",
		"gemini-1.5-flash",
		[]string{"1.5-flash"},
	)
	registry.RegisterModelForProvider(
		"vertexai",
		"gemini-1.5-pro",
		[]string{"1.5-pro", "pro"},
	)
}
