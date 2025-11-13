package factories

import (
	"context"
	"fmt"

	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// GeminiFactory creates Gemini API models
type GeminiFactory struct{}

// Create builds a Gemini model with the provided configuration
func (f *GeminiFactory) Create(ctx context.Context, config ModelConfig) (model.LLM, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, err
	}

	llm, err := gemini.NewModel(ctx, config.ModelName, &genai.ClientConfig{
		Backend: genai.BackendGeminiAPI,
		APIKey:  config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini model: %w", err)
	}

	return llm, nil
}

// ValidateConfig checks if the configuration is valid for Gemini
func (f *GeminiFactory) ValidateConfig(config ModelConfig) error {
	return ValidateAllRequiredFields(
		NewFieldCheck("Gemini API key", config.APIKey),
		NewFieldCheck("model name", config.ModelName),
	)
}

// Info returns metadata about the Gemini factory
func (f *GeminiFactory) Info() FactoryInfo {
	return FactoryInfo{
		Provider:      "Gemini",
		Description:   "Google Gemini API models",
		RequiredField: []string{"APIKey", "ModelName"},
	}
}

// NewGeminiFactory creates a new Gemini factory instance
func NewGeminiFactory() *GeminiFactory {
	return &GeminiFactory{}
}
