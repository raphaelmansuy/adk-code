// Package models - Model capability detection and metadata
package models

// ModelCapabilities represents the capabilities of a model
type ModelCapabilities struct {
	// VisionSupport indicates if the model supports image/vision inputs
	VisionSupport bool `json:"vision_support"`
	// ToolCalling indicates if the model supports function/tool calling
	ToolCalling bool `json:"tool_calling"`
	// Streaming indicates if the model supports streaming responses
	Streaming bool `json:"streaming"`
	// JSONMode indicates if the model supports JSON output mode
	JSONMode bool `json:"json_mode"`
	// ContextWindow is the maximum context window size in tokens
	ContextWindow int `json:"context_window"`
}

// ModelInfo represents detailed information about a model
type ModelInfo struct {
	// Name is the model identifier
	Name string `json:"name"`
	// DisplayName is the human-readable name
	DisplayName string `json:"display_name"`
	// Provider is the backend provider (gemini, ollama, openai, vertexai)
	Provider string `json:"provider"`
	// Family is the model family (e.g., llama, mistral, gemini)
	Family string `json:"family,omitempty"`
	// Size is the model size (e.g., "7B", "13B", "70B")
	Size string `json:"size,omitempty"`
	// Quantization is the quantization level (e.g., "Q4_0", "Q8_0")
	Quantization string `json:"quantization,omitempty"`
	// ParameterCount is the number of parameters (e.g., 7000000000 for 7B)
	ParameterCount int64 `json:"parameter_count,omitempty"`
	// Capabilities describes what the model can do
	Capabilities ModelCapabilities `json:"capabilities"`
	// Description provides additional information about the model
	Description string `json:"description,omitempty"`
	// Modified is the last modification time (for Ollama models)
	Modified string `json:"modified,omitempty"`
}

// DetectCapabilitiesFromName infers model capabilities from the model name
// This is a heuristic approach for Ollama models where metadata may be limited
func DetectCapabilitiesFromName(name, family string) ModelCapabilities {
	caps := ModelCapabilities{
		// Default values
		ToolCalling:   true,  // V1: Assume all models support tool calling
		Streaming:     true,  // All Ollama models support streaming
		JSONMode:      false, // Conservative default
		ContextWindow: 4096,  // Conservative default
	}

	// Detect vision support based on model name/family
	visionKeywords := []string{"vision", "llava", "minichat", "bakllava"}
	for _, keyword := range visionKeywords {
		if contains(name, keyword) || contains(family, keyword) {
			caps.VisionSupport = true
			break
		}
	}

	// Detect larger context windows based on model family
	if contains(name, "mistral") || contains(family, "mistral") {
		caps.ContextWindow = 32000
	} else if contains(name, "gpt-oss") || contains(family, "gpt-oss") {
		caps.ContextWindow = 128000
	} else if contains(name, "claude") || contains(family, "claude") {
		caps.ContextWindow = 200000
	}

	return caps
}

// contains checks if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
	str = toLower(str)
	substr = toLower(substr)
	return len(str) >= len(substr) && indexOfSubstring(str, substr) >= 0
}

// toLower converts a string to lowercase
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + ('a' - 'A')
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// indexOfSubstring finds the index of a substring in a string
func indexOfSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
