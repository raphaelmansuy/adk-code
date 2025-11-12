// Package main - CLI command handlers
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"code_agent/display"
	"code_agent/tools"
	"code_agent/tracking"

	codingagent "code_agent/agent"
)

// HandleCLICommands processes special CLI commands (new-session, list-sessions, etc.)
// Returns true if a command was handled (and program should exit)
func HandleCLICommands(ctx context.Context, args []string, dbPath string) bool {
	if len(args) == 0 {
		return false
	}

	cmd := args[0]

	switch cmd {
	case "new-session":
		if len(args) < 2 {
			fmt.Println("Usage: code-agent new-session <session-name>")
			os.Exit(1)
		}
		handleNewSession(ctx, args[1], dbPath)
		return true

	case "list-sessions":
		handleListSessions(ctx, dbPath)
		return true

	case "delete-session":
		if len(args) < 2 {
			fmt.Println("Usage: code-agent delete-session <session-name>")
			os.Exit(1)
		}
		handleDeleteSession(ctx, args[1], dbPath)
		return true

	default:
		return false
	}
}

// handleBuiltinCommand handles built-in REPL commands like /help, /tools, etc.
// Returns true if a command was handled, false if input should be sent to agent
// Note: /exit and /quit are handled separately in main.go to break the loop
func handleBuiltinCommand(input string, renderer *display.Renderer, sessionTokens *tracking.SessionTokens, modelRegistry *ModelRegistry, currentModel ModelConfig) bool {
	switch input {
	case "/prompt":
		// Show the XML-structured prompt with minimal context
		registry := tools.GetRegistry()
		ctx := codingagent.PromptContext{
			HasWorkspace:         false,
			WorkspaceRoot:        "",
			WorkspaceSummary:     "(Context not available in REPL)",
			EnvironmentMetadata:  "",
			EnableMultiWorkspace: false,
		}
		xmlPrompt := codingagent.BuildEnhancedPromptWithContext(registry, ctx)

		// Clean up excessive blank lines in the output
		cleanedPrompt := cleanupPromptOutput(xmlPrompt)

		// Build paginated output with header and footer
		lines := buildPromptLines(renderer, cleanedPrompt)
		paginator := display.NewPaginator(renderer)
		paginator.DisplayPaged(lines)
		return true

	case "/help":
		printHelpMessage(renderer)
		return true

	case "/tools":
		printToolsList(renderer)
		return true

	case "/models":
		printModelsList(renderer, modelRegistry)
		return true

	case "/current-model":
		printCurrentModelInfo(renderer, currentModel)
		return true

	case "/providers":
		printProvidersList(renderer, modelRegistry)
		return true

	case "/tokens":
		summary := sessionTokens.GetSummary()
		fmt.Print(tracking.FormatSessionSummary(summary))
		return true

	default:
		// Check if it's a /set-model command
		if strings.HasPrefix(input, "/set-model ") {
			modelSpec := strings.TrimPrefix(input, "/set-model ")
			handleSetModel(renderer, modelRegistry, modelSpec)
			return true
		}
		return false
	}
}

// handleSetModel validates and displays information about switching to a new model
func handleSetModel(renderer *display.Renderer, registry *ModelRegistry, modelSpec string) {
	modelSpec = strings.TrimSpace(modelSpec)
	if modelSpec == "" {
		fmt.Println(renderer.Red("Error: Please specify a model using provider/model syntax"))
		fmt.Println("Example: /set-model gemini/2.5-flash")
		fmt.Println(renderer.Dim("\nUse /providers to see all available models"))
		return
	}

	// Parse the provider/model syntax
	parsedProvider, parsedModel, parseErr := ParseProviderModelSyntax(modelSpec)
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
			models := registry.GetProviderModels(providerName)
			fmt.Printf("\n%s:\n", renderer.Bold(strings.ToUpper(providerName[:1])+strings.ToLower(providerName[1:])))
			for _, m := range models {
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
