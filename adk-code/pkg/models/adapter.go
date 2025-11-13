package models

import (
	"adk-code/pkg/errors"
)

// ProviderInfo describes a provider's capabilities and characteristics
type ProviderInfo struct {
	// Name is the human-readable name of the provider (e.g., "OpenAI", "Google Gemini")
	Name string

	// SupportsFunctions indicates whether the provider supports function/tool calling
	SupportsFunctions bool

	// SupportsThinking indicates whether the provider supports extended thinking
	SupportsThinking bool

	// TokenLimits maps token type (e.g., "input", "output", "context") to max values
	TokenLimits map[string]int

	// Description provides additional information about the provider
	Description string
}

// ProviderAdapter abstracts differences between LLM providers.
// This interface allows multiple providers to be plugged in without changing
// the core agent code. Each provider (OpenAI, Gemini, Vertex AI, etc.) can
// implement this interface.
//
// ProviderAdapter should NOT implement model.LLM directly; instead, it wraps
// provider-specific clients and handles protocol conversion.
type ProviderAdapter interface {
	// GetInfo returns metadata about this provider's capabilities and characteristics
	GetInfo() ProviderInfo

	// ValidateConfig checks if the configuration is valid for this provider
	ValidateConfig(config map[string]string) error

	// Name returns a friendly name for this adapter
	Name() string
}

// DefaultProviderInfo returns a default ProviderInfo with all false/empty values
func DefaultProviderInfo() ProviderInfo {
	return ProviderInfo{
		TokenLimits: make(map[string]int),
	}
}

// OpenAIProviderAdapter implements ProviderAdapter for OpenAI
// Note: OpenAIModelAdapter (which implements model.LLM) remains unchanged.
// This adapter provides metadata and validation specific to the OpenAI provider.
type OpenAIProviderAdapter struct {
	info ProviderInfo
}

// NewOpenAIProviderAdapter creates a new OpenAI provider adapter
func NewOpenAIProviderAdapter() *OpenAIProviderAdapter {
	return &OpenAIProviderAdapter{
		info: ProviderInfo{
			Name:              "OpenAI",
			SupportsFunctions: true,
			SupportsThinking:  true, // o1/o3 models support extended thinking
			TokenLimits: map[string]int{
				"input":   128000, // Varies by model, but this is a reasonable default
				"output":  4096,   // Also varies, but reasonable default
				"context": 128000, // Max context window
			},
			Description: "OpenAI provider supporting GPT-4, GPT-4o, o1, o3 models",
		},
	}
}

// GetInfo returns provider information
func (a *OpenAIProviderAdapter) GetInfo() ProviderInfo {
	return a.info
}

// ValidateConfig checks if OpenAI configuration is valid
func (a *OpenAIProviderAdapter) ValidateConfig(config map[string]string) error {
	if apiKey, ok := config["api_key"]; !ok || apiKey == "" {
		return errors.New(errors.CodeAPIKey, "OpenAI API key is required")
	}
	if modelName, ok := config["model_name"]; !ok || modelName == "" {
		return errors.New(errors.CodeValidation, "model name is required for OpenAI")
	}
	return nil
}

// Name returns the adapter name
func (a *OpenAIProviderAdapter) Name() string {
	return a.info.Name
}

// GeminiProviderAdapter implements ProviderAdapter for Google Gemini
type GeminiProviderAdapter struct {
	info ProviderInfo
}

// NewGeminiProviderAdapter creates a new Gemini provider adapter
func NewGeminiProviderAdapter() *GeminiProviderAdapter {
	return &GeminiProviderAdapter{
		info: ProviderInfo{
			Name:              "Google Gemini",
			SupportsFunctions: true,
			SupportsThinking:  true, // Gemini 2.0 Flash has thinking capability
			TokenLimits: map[string]int{
				"input":   1000000, // Gemini has large context windows
				"output":  8192,
				"context": 1000000,
			},
			Description: "Google Gemini provider supporting 2.0 Flash, 1.5 Pro/Flash models",
		},
	}
}

// GetInfo returns provider information
func (a *GeminiProviderAdapter) GetInfo() ProviderInfo {
	return a.info
}

// ValidateConfig checks if Gemini configuration is valid
func (a *GeminiProviderAdapter) ValidateConfig(config map[string]string) error {
	if apiKey, ok := config["api_key"]; !ok || apiKey == "" {
		return errors.New(errors.CodeAPIKey, "Gemini API key is required")
	}
	return nil
}

// Name returns the adapter name
func (a *GeminiProviderAdapter) Name() string {
	return a.info.Name
}

// VertexAIProviderAdapter implements ProviderAdapter for Google Vertex AI
type VertexAIProviderAdapter struct {
	info ProviderInfo
}

// NewVertexAIProviderAdapter creates a new Vertex AI provider adapter
func NewVertexAIProviderAdapter() *VertexAIProviderAdapter {
	return &VertexAIProviderAdapter{
		info: ProviderInfo{
			Name:              "Google Vertex AI",
			SupportsFunctions: true,
			SupportsThinking:  true,
			TokenLimits: map[string]int{
				"input":   1000000,
				"output":  8192,
				"context": 1000000,
			},
			Description: "Google Vertex AI provider supporting Gemini models with enterprise features",
		},
	}
}

// GetInfo returns provider information
func (a *VertexAIProviderAdapter) GetInfo() ProviderInfo {
	return a.info
}

// ValidateConfig checks if Vertex AI configuration is valid
func (a *VertexAIProviderAdapter) ValidateConfig(config map[string]string) error {
	if projectID, ok := config["project_id"]; !ok || projectID == "" {
		return errors.New(errors.CodeValidation, "Vertex AI project ID is required")
	}
	if location, ok := config["location"]; !ok || location == "" {
		return errors.New(errors.CodeValidation, "Vertex AI location is required")
	}
	return nil
}

// Name returns the adapter name
func (a *VertexAIProviderAdapter) Name() string {
	return a.info.Name
}
