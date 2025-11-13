// Package commands provides CLI command handlers organized by functionality.
package commands

import (
	"fmt"
	"strings"

	"adk-code/internal/display"
	"adk-code/pkg/models"
)

// HandleSetModel validates and displays information about switching to a new model
func HandleSetModel(renderer *display.Renderer, registry *models.Registry, modelSpec string) {
	modelSpec = strings.TrimSpace(modelSpec)
	if modelSpec == "" {
		fmt.Println(renderer.Red("Error: Please specify a model using provider/model syntax"))
		fmt.Println("Example: /set-model gemini/2.5-flash")
		fmt.Println(renderer.Dim("\nUse /providers to see all available models"))
		return
	}

	// Parse the provider/model syntax
	parsedProvider, parsedModel, parseErr := parseProviderModelSyntax(modelSpec)
	if parseErr != nil {
		fmt.Printf("%s\n", renderer.Red(fmt.Sprintf("Invalid model syntax: %v", parseErr)))
		fmt.Println("Use format: provider/model or shorthand/model")
		fmt.Println("Examples:")
		fmt.Println("  /set-model gemini/2.5-flash")
		fmt.Println("  /set-model gemini/flash")
		fmt.Println("  /set-model vertexai/1.5-pro")
		return
	}

	// Determine default provider if not specified
	defaultProvider := "gemini"
	if parsedProvider == "" {
		parsedProvider = defaultProvider
	}

	// Try to resolve the model
	resolvedModel, modelErr := registry.ResolveFromProviderSyntax(
		parsedProvider,
		parsedModel,
		defaultProvider,
	)
	if modelErr != nil {
		fmt.Printf("%s\n", renderer.Red(fmt.Sprintf("Model not found: %v", modelErr)))
		fmt.Println("\n" + renderer.Bold("Available models:"))
		for _, providerName := range registry.ListProviders() {
			modelsCfg := registry.GetProviderModels(providerName)
			fmt.Printf("\n%s:\n", renderer.Bold(strings.ToUpper(providerName[:1])+strings.ToLower(providerName[1:])))
			for _, m := range modelsCfg {
				fmt.Printf("  • %s/%s\n", providerName, m.ID)
			}
		}
		return
	}

	// Display the resolved model information
	fmt.Println("")
	fmt.Println(renderer.Green("✓ Model validation successful!"))
	fmt.Println("")
	fmt.Printf("You selected: %s (%s)\n", renderer.Bold(resolvedModel.DisplayName), resolvedModel.Backend)
	fmt.Printf("Context window: %d tokens\n", resolvedModel.ContextWindow)
	fmt.Printf("Cost tier: %s\n", resolvedModel.Capabilities.CostTier)
	fmt.Println("")
	fmt.Println(renderer.Yellow("ℹ️  Note:"))
	fmt.Println("The model can only be switched at startup. To actually use this model, exit the agent and restart with:")

	// Build the recommended command
	recommendedCommand := ""
	if strings.Contains(modelSpec, "/") {
		// User provided provider/model syntax, use as-is
		recommendedCommand = fmt.Sprintf("--model %s", modelSpec)
	} else {
		// User provided just the model ID or shorthand
		// Try to extract a better shorthand from the model ID
		// For "gemini-2.5-flash" suggest "2.5-flash", etc.
		shorthand := extractShorthandFromModelID(resolvedModel.ID)
		recommendedCommand = fmt.Sprintf("--model %s/%s", parsedProvider, shorthand)
	}

	fmt.Printf("  %s\n", renderer.Dim(fmt.Sprintf("./code-agent %s", recommendedCommand)))
	fmt.Println("")
}

// parseProviderModelSyntax parses a "provider/model" string into components
// Examples:
//
//	"gemini/2.5-flash" → ("gemini", "2.5-flash", nil)
//	"gemini/flash" → ("gemini", "flash", nil)
//	"flash" → ("", "flash", nil)
//	"/flash" → ("", "", error)
//	"a/b/c" → ("", "", error)
func parseProviderModelSyntax(input string) (string, string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", fmt.Errorf("model syntax cannot be empty")
	}

	parts := strings.Split(input, "/")
	switch len(parts) {
	case 1:
		// Shorthand without provider: "flash" → ("", "flash")
		return "", parts[0], nil
	case 2:
		// Full syntax: "provider/model" → ("provider", "model")
		if parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("invalid model syntax: %q (use provider/model)", input)
		}
		return parts[0], parts[1], nil
	default:
		return "", "", fmt.Errorf("invalid model syntax: %q (use provider/model)", input)
	}
}

// extractShorthandFromModelID extracts a shorthand from a full model ID
// Examples: "gemini-2.5-flash" → "2.5-flash", "gemini-1.5-pro" → "1.5-pro"
func extractShorthandFromModelID(modelID string) string {
	// Most model IDs follow the pattern: gemini-VERSION-VARIANT or similar
	// Try to extract everything after "gemini-" or "claude-" etc.
	parts := strings.Split(modelID, "-")
	if len(parts) > 1 {
		// Skip the provider prefix and return the rest
		return strings.Join(parts[1:], "-")
	}
	// Fallback to the full ID if we can't extract a shorthand
	return modelID
}
