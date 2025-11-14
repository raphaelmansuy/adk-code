// Package discovery provides model discovery and introspection tools
package discovery

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/internal/llm"
	"adk-code/pkg/models"
	common "adk-code/tools/base"
)

// ListModelsInput defines the input parameters for listing models
type ListModelsInput struct {
	// Provider filters models by provider (optional: gemini, ollama, openai, vertexai)
	Provider string `json:"provider,omitempty" jsonschema:"Filter by provider (gemini, ollama, openai, vertexai)"`
	// Family filters models by family (optional: e.g., llama, mistral, gemini, gpt-4)
	Family string `json:"family,omitempty" jsonschema:"Filter by model family (e.g., llama, mistral, gemini, gpt-4)"`
	// HasVision filters to only models with vision support
	HasVision *bool `json:"has_vision,omitempty" jsonschema:"Filter to models with vision support"`
	// HasTools filters to only models with tool calling support
	HasTools *bool `json:"has_tools,omitempty" jsonschema:"Filter to models with tool calling support"`
	// RefreshCache forces a refresh of cached model data (default: false)
	RefreshCache bool `json:"refresh_cache,omitempty" jsonschema:"Force refresh of cached model data"`
}

// ModelEntry represents a single model in the output list
type ModelEntry struct {
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	Provider      string `json:"provider"`
	Family        string `json:"family,omitempty"`
	Size          string `json:"size,omitempty"`
	VisionSupport bool   `json:"vision_support"`
	ToolCalling   bool   `json:"tool_calling"`
	ContextWindow int    `json:"context_window"`
	Description   string `json:"description,omitempty"`
}

// ListModelsOutput defines the output of listing models
type ListModelsOutput struct {
	// Models is the list of available models matching the filters
	Models []ModelEntry `json:"models"`
	// Count is the total number of models returned
	Count int `json:"count"`
	// Success indicates whether the operation was successful
	Success bool `json:"success"`
	// Error contains error message if the operation failed
	Error string `json:"error,omitempty"`
	// Summary provides a human-readable summary
	Summary string `json:"summary"`
}

// NewListModelsTool creates a tool for listing available models across providers
func NewListModelsTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ListModelsInput) ListModelsOutput {
		registry := llm.NewRegistry()

		var allModels []models.ModelInfo
		var errors []string

		// Determine which providers to query
		providers := []string{"gemini", "openai", "ollama"}
		if input.Provider != "" {
			providers = []string{input.Provider}
		}

		// Query each provider
		for _, providerName := range providers {
			provider, err := registry.Get(providerName)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", providerName, err))
				continue
			}

			// Check if provider supports model discovery
			discoverable, ok := provider.(llm.ModelDiscovery)
			if !ok {
				continue
			}

			// List models from this provider
			providerModels, err := discoverable.ListModels(ctx, input.RefreshCache)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", providerName, err))
				continue
			}

			allModels = append(allModels, providerModels...)
		}

		// Apply filters
		filtered := filterModels(allModels, input)

		// Convert to output format
		entries := make([]ModelEntry, 0, len(filtered))
		for _, model := range filtered {
			entries = append(entries, ModelEntry{
				Name:          model.Name,
				DisplayName:   model.DisplayName,
				Provider:      model.Provider,
				Family:        model.Family,
				Size:          model.Size,
				VisionSupport: model.Capabilities.VisionSupport,
				ToolCalling:   model.Capabilities.ToolCalling,
				ContextWindow: model.Capabilities.ContextWindow,
				Description:   model.Description,
			})
		}

		// Sort by provider, then by name
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Provider != entries[j].Provider {
				return entries[i].Provider < entries[j].Provider
			}
			return entries[i].Name < entries[j].Name
		})

		// Generate summary
		summary := generateSummary(entries, input, errors)

		output := ListModelsOutput{
			Models:  entries,
			Count:   len(entries),
			Success: len(errors) == 0 || len(entries) > 0,
			Summary: summary,
		}

		if len(errors) > 0 {
			output.Error = strings.Join(errors, "; ")
		}

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name: "list_models",
		Description: "Lists available models across all providers (Gemini, OpenAI, Ollama). " +
			"Supports filtering by provider, family, vision support, and tool calling capability. " +
			"Returns model names, capabilities, context windows, and descriptions. " +
			"Use refresh_cache=true to bypass caching and fetch fresh model data.",
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategorySearchDiscovery,
			Priority:  10,
			UsageHint: "Discover available models across providers with filtering options",
		})
	}

	return t, err
}

// filterModels applies filters to a list of models
func filterModels(modelList []models.ModelInfo, input ListModelsInput) []models.ModelInfo {
	var filtered []models.ModelInfo

	for _, model := range modelList {
		// Filter by provider
		if input.Provider != "" && !strings.EqualFold(model.Provider, input.Provider) {
			continue
		}

		// Filter by family
		if input.Family != "" && !strings.Contains(strings.ToLower(model.Family), strings.ToLower(input.Family)) {
			continue
		}

		// Filter by vision support
		if input.HasVision != nil && model.Capabilities.VisionSupport != *input.HasVision {
			continue
		}

		// Filter by tool calling support
		if input.HasTools != nil && model.Capabilities.ToolCalling != *input.HasTools {
			continue
		}

		filtered = append(filtered, model)
	}

	return filtered
}

// generateSummary creates a human-readable summary of the results
func generateSummary(entries []ModelEntry, input ListModelsInput, errors []string) string {
	var parts []string

	// Count by provider
	providerCounts := make(map[string]int)
	for _, entry := range entries {
		providerCounts[entry.Provider]++
	}

	if len(entries) == 0 {
		return "No models found matching the specified filters"
	}

	// Main summary
	parts = append(parts, fmt.Sprintf("Found %d model(s)", len(entries)))

	// Provider breakdown
	if len(providerCounts) > 1 {
		var providerParts []string
		for provider, count := range providerCounts {
			providerParts = append(providerParts, fmt.Sprintf("%s: %d", provider, count))
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(providerParts, ", ")))
	}

	// Applied filters
	var filters []string
	if input.Provider != "" {
		filters = append(filters, fmt.Sprintf("provider=%s", input.Provider))
	}
	if input.Family != "" {
		filters = append(filters, fmt.Sprintf("family=%s", input.Family))
	}
	if input.HasVision != nil {
		filters = append(filters, fmt.Sprintf("vision=%t", *input.HasVision))
	}
	if input.HasTools != nil {
		filters = append(filters, fmt.Sprintf("tools=%t", *input.HasTools))
	}

	if len(filters) > 0 {
		parts = append(parts, fmt.Sprintf("with filters: %s", strings.Join(filters, ", ")))
	}

	// Errors
	if len(errors) > 0 {
		parts = append(parts, fmt.Sprintf("(with %d error(s))", len(errors)))
	}

	return strings.Join(parts, " ")
}

// init registers the list models tool automatically at package initialization
func init() {
	_, _ = NewListModelsTool()
}
