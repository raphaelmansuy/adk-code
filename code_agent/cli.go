// Package main - CLI flag parsing and command handling
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	codingagent "code_agent/agent"
	"code_agent/display"
	"code_agent/tools"
	"code_agent/tracking"
)

// ParseProviderModelSyntax parses a "provider/model" string into components
// Examples:
//
//	"gemini/2.5-flash" → ("gemini", "2.5-flash", nil)
//	"gemini/flash" → ("gemini", "flash", nil)
//	"flash" → ("", "flash", nil)
//	"/flash" → ("", "", error)
//	"a/b/c" → ("", "", error)
func ParseProviderModelSyntax(input string) (string, string, error) {
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

// CLIConfig holds parsed command-line flags
type CLIConfig struct {
	OutputFormat      string
	TypewriterEnabled bool
	SessionName       string
	DBPath            string
	WorkingDirectory  string
	// Backend configuration
	Backend          string // "gemini" or "vertexai"
	APIKey           string // For Gemini API
	VertexAIProject  string // For Vertex AI
	VertexAILocation string // For Vertex AI
	// Model selection
	Model string // Specific model ID (e.g., "gemini-2.5-flash", "gemini-1.5-pro")
	// Thinking configuration
	EnableThinking bool  // Enable model thinking/reasoning output
	ThinkingBudget int32 // Token budget for thinking
}

// ParseCLIFlags parses command-line arguments and returns config and remaining args
func ParseCLIFlags() (CLIConfig, []string) {
	outputFormat := flag.String("output-format", "rich", "Output format: rich, plain, or json")
	typewriterEnabled := flag.Bool("typewriter", false, "Enable typewriter effect for text output")
	sessionName := flag.String("session", "", "Session name (optional, defaults to 'default')")
	dbPath := flag.String("db", "", "Database path for sessions (optional, defaults to ~/.code_agent/sessions.db)")
	workingDirectory := flag.String("working-directory", "", "Working directory for the agent (optional, defaults to current directory)")

	// Model selection flags
	model := flag.String("model", "", "Model to use with provider/model syntax. Examples:\n"+
		"  --model gemini/2.5-flash     (explicit provider)\n"+
		"  --model gemini/flash          (shorthand, means 2.5-flash)\n"+
		"  --model vertexai/1.5-pro      (Vertex AI model)\n"+
		"Use '/providers' command to list all available models.")

	// Backend selection flags
	backend := flag.String("backend", "", "Backend to use: 'gemini' or 'vertexai' (default: auto-detect from env vars)")
	apiKey := flag.String("api-key", os.Getenv("GOOGLE_API_KEY"), "API key for Gemini (default: GOOGLE_API_KEY env var)")
	vertexAIProject := flag.String("project", os.Getenv("GOOGLE_CLOUD_PROJECT"), "GCP Project ID for Vertex AI (default: GOOGLE_CLOUD_PROJECT env var)")
	vertexAILocation := flag.String("location", os.Getenv("GOOGLE_CLOUD_LOCATION"), "GCP Location for Vertex AI (default: GOOGLE_CLOUD_LOCATION env var)")

	// Thinking configuration flags
	enableThinking := flag.Bool("enable-thinking", true, "Enable model thinking/reasoning output (default: true)")
	thinkingBudget := flag.Int("thinking-budget", 1024, "Token budget for thinking when enabled (default: 1024)")

	flag.Parse()

	// Auto-detect backend from environment if not specified
	selectedBackend := *backend
	if selectedBackend == "" {
		if os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "true" || os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "1" {
			selectedBackend = "vertexai"
		} else if *apiKey != "" {
			selectedBackend = "gemini"
		} else if *vertexAIProject != "" {
			// If project is set but backend not specified, assume vertexai
			selectedBackend = "vertexai"
		} else {
			// Default to gemini if nothing is set (existing behavior)
			selectedBackend = "gemini"
		}
	}

	return CLIConfig{
		OutputFormat:      *outputFormat,
		TypewriterEnabled: *typewriterEnabled,
		SessionName:       *sessionName,
		DBPath:            *dbPath,
		WorkingDirectory:  *workingDirectory,
		Backend:           selectedBackend,
		APIKey:            *apiKey,
		VertexAIProject:   *vertexAIProject,
		VertexAILocation:  *vertexAILocation,
		Model:             *model,
		EnableThinking:    *enableThinking,
		ThinkingBudget:    int32(*thinkingBudget),
	}, flag.Args()
}

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

// printHelpMessage displays the help message with pagination
func printHelpMessage(renderer *display.Renderer) {
	lines := buildHelpMessageLines(renderer)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// buildHelpMessageLines builds the help message as an array of lines for pagination
func buildHelpMessageLines(renderer *display.Renderer) []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, renderer.Cyan("                       Code Agent Help"))
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("🤖 Natural Language Requests:"))
	lines = append(lines, "   Just type what you want in plain English!")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("⌨️  Built-in Commands:"))
	lines = append(lines, "   • "+renderer.Bold("/help")+" - Show this help message")
	lines = append(lines, "   • "+renderer.Bold("/tools")+" - List all available tools")
	lines = append(lines, "   • "+renderer.Bold("/models")+" - Show all available AI models")
	lines = append(lines, "   • "+renderer.Bold("/providers")+" - Show available providers and their models")
	lines = append(lines, "   • "+renderer.Bold("/current-model")+" - Show details about the current model")
	lines = append(lines, "   • "+renderer.Bold("/set-model <provider/model>")+" - Validate and plan to switch models")
	lines = append(lines, "   • "+renderer.Bold("/prompt")+" - Display the system prompt")
	lines = append(lines, "   • "+renderer.Bold("/tokens")+" - Show token usage statistics")
	lines = append(lines, "   • "+renderer.Bold("/exit")+" - Exit the agent")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("📚 Model Selection:"))
	lines = append(lines, "   Start the agent with --model flag using provider/model syntax:")
	lines = append(lines, "   • "+renderer.Dim("./code-agent --model gemini/2.5-flash"))
	lines = append(lines, "   • "+renderer.Dim("./code-agent --model gemini/flash")+" (shorthand)")
	lines = append(lines, "   • "+renderer.Dim("./code-agent --model vertexai/1.5-pro"))
	lines = append(lines, "   Use "+renderer.Cyan("'/providers'")+" command to see all available options")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("🧠 Thinking Configuration:"))
	lines = append(lines, "   Control the model's reasoning/thinking output:")
	lines = append(lines, "   • "+renderer.Dim("./code-agent --enable-thinking=true")+" (enabled by default)")
	lines = append(lines, "   • "+renderer.Dim("./code-agent --enable-thinking=false")+" (disable thinking)")
	lines = append(lines, "   • "+renderer.Dim("./code-agent --thinking-budget 2048")+" (set token budget)")
	lines = append(lines, "   Thinking helps with debugging and transparency at a small token cost")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("📚 Session Management (CLI commands):"))
	lines = append(lines, "   • "+renderer.Bold("./code-agent new-session <name>")+" - Create a new session")
	lines = append(lines, "   • "+renderer.Bold("./code-agent list-sessions")+" - List all sessions")
	lines = append(lines, "   • "+renderer.Bold("./code-agent delete-session <name>")+" - Delete a session")
	lines = append(lines, "   • "+renderer.Bold("./code-agent --session <name>")+" - Resume a specific session")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("💡 Example Requests:"))
	lines = append(lines, "   ❯ Add error handling to main.go")
	lines = append(lines, "   ❯ Create a README.md with project overview")
	lines = append(lines, "   ❯ Refactor the calculate function")
	lines = append(lines, "   ❯ Run tests and fix any failures")
	lines = append(lines, "   ❯ Add comments to all Python files")
	lines = append(lines, "")

	lines = append(lines, renderer.Yellow("📖 More info: ")+"See USER_GUIDE.md for detailed documentation")
	lines = append(lines, "")

	return lines
}

// printToolsList displays the available tools with pagination
func printToolsList(renderer *display.Renderer) {
	lines := buildToolsListLines(renderer)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// buildToolsListLines builds the tools list as an array of lines for pagination
func buildToolsListLines(renderer *display.Renderer) []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, renderer.Cyan("                    Available Tools"))
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("📝 Core Editing Tools:"))
	lines = append(lines, "   ✓ "+renderer.Bold("read_file")+" - Read file contents (supports line ranges)")
	lines = append(lines, "   ✓ "+renderer.Bold("write_file")+" - Create or overwrite files (atomic, safe)")
	lines = append(lines, "   ✓ "+renderer.Bold("search_replace")+" - Make targeted changes (RECOMMENDED)")
	lines = append(lines, "   ✓ "+renderer.Bold("edit_lines")+" - Edit by line number (structural changes)")
	lines = append(lines, "   ✓ "+renderer.Bold("apply_patch")+" - Apply unified diff patches (standard)")
	lines = append(lines, "   ✓ "+renderer.Bold("apply_v4a_patch")+" - Apply V4A semantic patches (NEW!)")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("🔍 Discovery Tools:"))
	lines = append(lines, "   ✓ "+renderer.Bold("list_files")+" - Explore directory structure")
	lines = append(lines, "   ✓ "+renderer.Bold("search_files")+" - Find files by pattern (*.go, test_*.py)")
	lines = append(lines, "   ✓ "+renderer.Bold("grep_search")+" - Search text in files (with line numbers)")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("⚡ Execution Tools:"))
	lines = append(lines, "   ✓ "+renderer.Bold("execute_command")+" - Run shell commands (pipes, redirects)")
	lines = append(lines, "   ✓ "+renderer.Bold("execute_program")+" - Run programs directly (no quoting issues)")
	lines = append(lines, "")

	lines = append(lines, "💡 Tip: Type "+renderer.Cyan("'/help'")+" for usage examples and patterns")
	lines = append(lines, "")

	return lines
}

// printModelsList displays all available models with pagination
func printModelsList(renderer *display.Renderer, registry *ModelRegistry) {
	lines := buildModelsListLines(renderer, registry)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// buildModelsListLines builds the models list as an array of lines for pagination
func buildModelsListLines(renderer *display.Renderer, registry *ModelRegistry) []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, renderer.Cyan("                      Available AI Models"))
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, "")

	// Group models by backend
	geminiBakcend := registry.ListModelsByBackend("gemini")
	vertexAIBackend := registry.ListModelsByBackend("vertexai")

	if len(geminiBakcend) > 0 {
		lines = append(lines, renderer.Bold("🔷 Gemini API Models:"))
		for _, model := range geminiBakcend {
			icon := "○"
			if model.IsDefault {
				icon = "✓"
			}
			costIcon := "💰"
			if model.Capabilities.CostTier == "economy" {
				costIcon = "💵"
			} else if model.Capabilities.CostTier == "premium" {
				costIcon = "💎"
			}

			lines = append(lines, fmt.Sprintf("   %s %s %s - %s", icon, costIcon, renderer.Bold(model.Name), model.Description))
			lines = append(lines, fmt.Sprintf("      Context: %d tokens | Tools: %v | Vision: %v",
				model.ContextWindow,
				model.Capabilities.ToolUseSupport,
				model.Capabilities.VisionSupport))
		}
		lines = append(lines, "")
	}

	if len(vertexAIBackend) > 0 {
		lines = append(lines, renderer.Bold("🔶 Vertex AI Models:"))
		for _, model := range vertexAIBackend {
			icon := "○"
			if model.IsDefault {
				icon = "✓"
			}
			costIcon := "💰"
			if model.Capabilities.CostTier == "economy" {
				costIcon = "💵"
			} else if model.Capabilities.CostTier == "premium" {
				costIcon = "💎"
			}

			lines = append(lines, fmt.Sprintf("   %s %s %s - %s", icon, costIcon, renderer.Bold(model.Name), model.Description))
			lines = append(lines, fmt.Sprintf("      Context: %d tokens | Tools: %v | Vision: %v",
				model.ContextWindow,
				model.Capabilities.ToolUseSupport,
				model.Capabilities.VisionSupport))
		}
		lines = append(lines, "")
	}

	lines = append(lines, renderer.Dim("Use --model flag to select a model (e.g., --model gemini-1.5-pro)"))
	lines = append(lines, renderer.Dim("Use /current-model command to see details about the active model"))
	lines = append(lines, "")

	return lines
}

// printCurrentModelInfo displays detailed information about the current model with pagination
func printCurrentModelInfo(renderer *display.Renderer, model ModelConfig) {
	lines := buildCurrentModelInfoLines(renderer, model)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// buildCurrentModelInfoLines builds the current model info as an array of lines for pagination
func buildCurrentModelInfoLines(renderer *display.Renderer, model ModelConfig) []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, renderer.Cyan("                 Current Model Information"))
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, "")

	// Model name and backend
	backendIcon := "🔷"
	if model.Backend == "vertexai" {
		backendIcon = "🔶"
	}

	lines = append(lines, renderer.Bold("Model: ")+fmt.Sprintf("%s %s (%s)", backendIcon, model.Name, model.Backend))
	lines = append(lines, "")

	// Description
	lines = append(lines, renderer.Bold("Description:"))
	lines = append(lines, renderer.Dim("  "+model.Description))
	lines = append(lines, "")

	// Capabilities
	lines = append(lines, renderer.Bold("Capabilities:"))
	if model.Capabilities.VisionSupport {
		lines = append(lines, "  ✓ Vision/Image Processing")
	} else {
		lines = append(lines, "  ✗ Vision/Image Processing")
	}
	if model.Capabilities.ToolUseSupport {
		lines = append(lines, "  ✓ Tool/Function Calling")
	} else {
		lines = append(lines, "  ✗ Tool/Function Calling")
	}
	if model.Capabilities.LongContextWindow {
		lines = append(lines, "  ✓ Long Context Window (1M+ tokens)")
	} else {
		lines = append(lines, "  ✗ Long Context Window")
	}
	lines = append(lines, "")

	// Context and Cost
	lines = append(lines, renderer.Bold("Technical Details:"))
	lines = append(lines, fmt.Sprintf("  Context Window: %d tokens", model.ContextWindow))
	lines = append(lines, fmt.Sprintf("  Cost Tier: %s", model.Capabilities.CostTier))
	lines = append(lines, "")

	// Recommended use cases
	if len(model.RecommendedFor) > 0 {
		lines = append(lines, renderer.Bold("Recommended For:"))
		for _, useCase := range model.RecommendedFor {
			lines = append(lines, "  • "+useCase)
		}
		lines = append(lines, "")
	}

	lines = append(lines, renderer.Dim("Tip: Use --model flag to switch models when starting the agent"))
	lines = append(lines, "")

	return lines
}

// printProvidersList displays available providers and their models with pagination
func printProvidersList(renderer *display.Renderer, registry *ModelRegistry) {
	lines := buildProvidersListLines(renderer, registry)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// buildProvidersListLines builds the providers list as an array of lines for pagination
func buildProvidersListLines(renderer *display.Renderer, registry *ModelRegistry) []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, renderer.Cyan("                  Available Providers & Models"))
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, "")

	// Display each provider
	for _, providerName := range registry.ListProviders() {
		provider := ParseProvider(providerName)
		meta := GetProviderMetadata(provider)

		// Provider header
		lines = append(lines, fmt.Sprintf("%s %s", meta.Icon, renderer.Bold(meta.DisplayName)))
		lines = append(lines, fmt.Sprintf("   %s", meta.Description))
		lines = append(lines, "")

		// List models for this provider
		models := registry.GetProviderModels(providerName)
		for _, model := range models {
			icon := "○"
			if model.IsDefault {
				icon = "✓"
			}
			costIcon := "💰"
			if model.Capabilities.CostTier == "economy" {
				costIcon = "💵"
			} else if model.Capabilities.CostTier == "premium" {
				costIcon = "💎"
			}

			// Display model with provider syntax
			modelSyntax := fmt.Sprintf("%s/%s", providerName, model.ID)
			lines = append(lines, fmt.Sprintf("   %s %s %s - %s", icon, costIcon, renderer.Bold(modelSyntax), model.Description))
		}

		lines = append(lines, "")
	}

	lines = append(lines, renderer.Dim("Usage: --model provider/model (e.g., --model gemini/2.5-flash)"))
	lines = append(lines, renderer.Dim("You can also use shorthands: --model gemini/flash"))
	lines = append(lines, renderer.Dim("Use /current-model command to see details about the active model"))
	lines = append(lines, "")

	return lines
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

// buildPromptLines builds the system prompt as an array of lines for pagination
func buildPromptLines(renderer *display.Renderer, cleanedPrompt string) []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, renderer.Yellow("=== System Prompt (XML-Structured) ==="))
	lines = append(lines, "")

	// Split the prompt by newlines and add each line
	promptLines := strings.Split(cleanedPrompt, "\n")
	for _, line := range promptLines {
		lines = append(lines, renderer.Dim(line))
	}

	lines = append(lines, "")
	lines = append(lines, renderer.Yellow("=== End of Prompt ==="))
	lines = append(lines, "")

	return lines
}

// cleanupPromptOutput removes excessive blank lines while preserving readability
// This prevents visual clutter when displaying the XML prompt
func cleanupPromptOutput(prompt string) string {
	lines := strings.Split(prompt, "\n")
	var result []string
	blankLineCount := 0

	for _, line := range lines {
		// Check if line is blank (only whitespace)
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			blankLineCount++
			// Allow up to 2 consecutive blank lines for readability, skip more
			if blankLineCount <= 2 {
				result = append(result, line)
			}
		} else {
			blankLineCount = 0
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
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
