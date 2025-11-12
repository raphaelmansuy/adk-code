// Package models provides model factory implementations and a registry for managing
// LLM models from multiple providers (OpenAI, Gemini, Vertex AI).
//
// The package uses the factory pattern for model creation and the registry pattern
// for model lookup and management. It abstracts away provider-specific details and
// provides a unified interface for creating and configuring language models.
//
// Key components:
// - Registry: Manages available models and their configurations
// - Factory: Creates model instances from configuration
// - Config: Model configuration and metadata
// - ProviderAdapter: Abstracts provider-specific capabilities
//
// Supported providers:
// - OpenAI (GPT-4, GPT-4o, o1, o3)
// - Google Gemini (2.0 Flash, 1.5 Pro/Flash)
// - Google Vertex AI (Gemini models with enterprise features)
//
// Model selection:
// - By ID: "gpt-4", "gemini-2.5-flash", "vertex-gemini-pro"
// - By name: Model display names (case-insensitive)
// - By backend: Find default model for a provider
// - Aliases: "openai/gpt4" resolves to "gpt-4"
//
// Example:
//
//	registry := models.NewRegistry()
//	cfg, err := registry.GetModel("gpt-4")
//	if err != nil {
//		return err
//	}
//
//	factory := factories.GetFactory("openai")
//	llm, err := factory.Create(ctx, cfg)
//	if err != nil {
//		return err
//	}
//
// The package also provides:
// - Provider information (capabilities, token limits)
// - Model validation
// - Backend-specific configurations
package models
