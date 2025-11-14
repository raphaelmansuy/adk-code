// Package discovery provides model discovery and introspection tools
package discovery

import (
	"fmt"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/internal/llm"
	common "adk-code/tools/base"
)

// ModelInfoInput defines the input parameters for getting model information
type ModelInfoInput struct {
	// Model is the model identifier in the format "provider/model" or just "model"
	Model string `json:"model" jsonschema:"Model identifier (e.g., 'ollama/llama2', 'gemini-1.5-flash')"`
	// Provider is the explicit provider name (optional, can be inferred from model string)
	Provider string `json:"provider,omitempty" jsonschema:"Provider name (gemini, ollama, openai, vertexai)"`
}

// ModelInfoOutput defines the output of getting model information
type ModelInfoOutput struct {
	// Name is the model identifier
	Name string `json:"name"`
	// DisplayName is the human-readable name
	DisplayName string `json:"display_name"`
	// Provider is the backend provider
	Provider string `json:"provider"`
	// Family is the model family
	Family string `json:"family,omitempty"`
	// Size is the model size
	Size string `json:"size,omitempty"`
	// Quantization is the quantization level (for Ollama models)
	Quantization string `json:"quantization,omitempty"`
	// ParameterCount is the number of parameters
	ParameterCount int64 `json:"parameter_count,omitempty"`
	// Capabilities describes what the model can do
	Capabilities CapabilitiesInfo `json:"capabilities"`
	// Description provides additional information
	Description string `json:"description,omitempty"`
	// Modified is the last modification time (for Ollama models)
	Modified string `json:"modified,omitempty"`
	// Success indicates whether the operation was successful
	Success bool `json:"success"`
	// Error contains error message if the operation failed
	Error string `json:"error,omitempty"`
}

// CapabilitiesInfo describes model capabilities
type CapabilitiesInfo struct {
	VisionSupport bool `json:"vision_support"`
	ToolCalling   bool `json:"tool_calling"`
	Streaming     bool `json:"streaming"`
	JSONMode      bool `json:"json_mode"`
	ContextWindow int  `json:"context_window"`
}

// NewModelInfoTool creates a tool for getting detailed model information
func NewModelInfoTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ModelInfoInput) ModelInfoOutput {
		// Parse model string to extract provider and model name
		provider, modelName := parseModelString(input.Model, input.Provider)

		if provider == "" {
			return ModelInfoOutput{
				Success: false,
				Error:   "provider not specified or could not be inferred from model string",
			}
		}

		if modelName == "" {
			return ModelInfoOutput{
				Success: false,
				Error:   "model name is required",
			}
		}

		// Get provider from registry
		registry := llm.NewRegistry()
		providerBackend, err := registry.Get(provider)
		if err != nil {
			return ModelInfoOutput{
				Success: false,
				Error:   fmt.Sprintf("provider not found: %s", provider),
			}
		}

		// Check if provider supports model discovery
		discoverable, ok := providerBackend.(llm.ModelDiscovery)
		if !ok {
			return ModelInfoOutput{
				Success: false,
				Error:   fmt.Sprintf("provider %s does not support model discovery", provider),
			}
		}

		// Get model info
		modelInfo, err := discoverable.GetModelInfo(ctx, modelName)
		if err != nil {
			return ModelInfoOutput{
				Success: false,
				Error:   fmt.Sprintf("failed to get model info: %v", err),
			}
		}

		// Convert to output format
		return ModelInfoOutput{
			Name:           modelInfo.Name,
			DisplayName:    modelInfo.DisplayName,
			Provider:       modelInfo.Provider,
			Family:         modelInfo.Family,
			Size:           modelInfo.Size,
			Quantization:   modelInfo.Quantization,
			ParameterCount: modelInfo.ParameterCount,
			Capabilities: CapabilitiesInfo{
				VisionSupport: modelInfo.Capabilities.VisionSupport,
				ToolCalling:   modelInfo.Capabilities.ToolCalling,
				Streaming:     modelInfo.Capabilities.Streaming,
				JSONMode:      modelInfo.Capabilities.JSONMode,
				ContextWindow: modelInfo.Capabilities.ContextWindow,
			},
			Description: modelInfo.Description,
			Modified:    modelInfo.Modified,
			Success:     true,
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name: "model_info",
		Description: "Gets detailed information about a specific model including capabilities, " +
			"size, family, and other metadata. Works across all providers (Gemini, OpenAI, Ollama). " +
			"Provide model name in the format 'provider/model' (e.g., 'ollama/llama2') or just the model name " +
			"with an explicit provider parameter.",
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategorySearchDiscovery,
			Priority:  11,
			UsageHint: "Get detailed information about a specific model including capabilities",
		})
	}

	return t, err
}

// parseModelString parses a model string in the format "provider/model" or just "model"
// Returns (provider, modelName)
func parseModelString(modelStr, explicitProvider string) (string, string) {
	// If explicit provider is given, use it
	if explicitProvider != "" {
		return explicitProvider, modelStr
	}

	// Try to parse provider/model format
	parts := strings.Split(modelStr, "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	// Try to infer provider from model name
	provider := inferProvider(modelStr)
	return provider, modelStr
}

// inferProvider tries to infer the provider from the model name
func inferProvider(modelName string) string {
	modelLower := strings.ToLower(modelName)

	// Check for known patterns
	if strings.HasPrefix(modelLower, "gemini-") {
		return "gemini"
	}
	if strings.HasPrefix(modelLower, "gpt-") {
		return "openai"
	}

	// Check for Ollama-specific model patterns
	ollamaPatterns := []string{"llama", "mistral", "neural-chat", "dolphin", "openhermes"}
	for _, pattern := range ollamaPatterns {
		if strings.Contains(modelLower, pattern) {
			return "ollama"
		}
	}

	// Default to empty string if cannot infer
	return ""
}

// init registers the model info tool automatically at package initialization
func init() {
	_, _ = NewModelInfoTool()
}
