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

		fmt.Print(renderer.Yellow("\n=== System Prompt (XML-Structured) ===\n\n"))
		fmt.Print(renderer.Dim(cleanedPrompt))
		fmt.Print(renderer.Yellow("\n\n=== End of Prompt ===\n\n"))
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
		return false
	}
}

// printHelpMessage displays the help message
func printHelpMessage(renderer *display.Renderer) {
	fmt.Print("\n" + renderer.Cyan("════════════════════════════════════════════════════════════════\n"))
	fmt.Print(renderer.Cyan("                       Code Agent Help\n"))
	fmt.Print(renderer.Cyan("════════════════════════════════════════════════════════════════\n") + "\n")

	fmt.Print(renderer.Bold("🤖 Natural Language Requests:\n"))
	fmt.Print("   Just type what you want in plain English!\n\n")

	fmt.Print(renderer.Bold("⌨️  Built-in Commands:\n"))
	fmt.Print("   • " + renderer.Bold("/help") + " - Show this help message\n")
	fmt.Print("   • " + renderer.Bold("/tools") + " - List all available tools\n")
	fmt.Print("   • " + renderer.Bold("/models") + " - Show all available AI models\n")
	fmt.Print("   • " + renderer.Bold("/providers") + " - Show available providers and their models\n")
	fmt.Print("   • " + renderer.Bold("/current-model") + " - Show details about the current model\n")
	fmt.Print("   • " + renderer.Bold("/prompt") + " - Display the system prompt\n")
	fmt.Print("   • " + renderer.Bold("/tokens") + " - Show token usage statistics\n")
	fmt.Print("   • " + renderer.Bold("/exit") + " - Exit the agent\n")

	fmt.Print(renderer.Bold("\n📚 Model Selection:\n"))
	fmt.Print("   Start the agent with --model flag using provider/model syntax:\n")
	fmt.Print("   • " + renderer.Dim("./code-agent --model gemini/2.5-flash") + "\n")
	fmt.Print("   • " + renderer.Dim("./code-agent --model gemini/flash") + " (shorthand)\n")
	fmt.Print("   • " + renderer.Dim("./code-agent --model vertexai/1.5-pro") + "\n")
	fmt.Print("   Use " + renderer.Cyan("'/providers'") + " command to see all available options\n")

	fmt.Print(renderer.Bold("\n📚 Session Management (CLI commands):\n"))
	fmt.Print("   • " + renderer.Bold("./code-agent new-session <name>") + " - Create a new session\n")
	fmt.Print("   • " + renderer.Bold("./code-agent list-sessions") + " - List all sessions\n")
	fmt.Print("   • " + renderer.Bold("./code-agent delete-session <name>") + " - Delete a session\n")
	fmt.Print("   • " + renderer.Bold("./code-agent --session <name>") + " - Resume a specific session\n")

	fmt.Print(renderer.Bold("\n💡 Example Requests:\n"))
	fmt.Print("   ❯ Add error handling to main.go\n")
	fmt.Print("   ❯ Create a README.md with project overview\n")
	fmt.Print("   ❯ Refactor the calculate function\n")
	fmt.Print("   ❯ Run tests and fix any failures\n")
	fmt.Print("   ❯ Add comments to all Python files\n\n")

	fmt.Print(renderer.Yellow("📖 More info: ") + "See USER_GUIDE.md for detailed documentation\n\n")
}

// printToolsList displays the available tools
func printToolsList(renderer *display.Renderer) {
	fmt.Print("\n" + renderer.Cyan("════════════════════════════════════════════════════════════════\n"))
	fmt.Print(renderer.Cyan("                    Available Tools\n"))
	fmt.Print(renderer.Cyan("════════════════════════════════════════════════════════════════\n") + "\n")

	fmt.Print(renderer.Bold("📝 Core Editing Tools:\n"))
	fmt.Print("   ✓ " + renderer.Bold("read_file") + " - Read file contents (supports line ranges)\n")
	fmt.Print("   ✓ " + renderer.Bold("write_file") + " - Create or overwrite files (atomic, safe)\n")
	fmt.Print("   ✓ " + renderer.Bold("search_replace") + " - Make targeted changes (RECOMMENDED)\n")
	fmt.Print("   ✓ " + renderer.Bold("edit_lines") + " - Edit by line number (structural changes)\n")
	fmt.Print("   ✓ " + renderer.Bold("apply_patch") + " - Apply unified diff patches (standard)\n")
	fmt.Print("   ✓ " + renderer.Bold("apply_v4a_patch") + " - Apply V4A semantic patches (NEW!)\n")

	fmt.Print(renderer.Bold("\n🔍 Discovery Tools:\n"))
	fmt.Print("   ✓ " + renderer.Bold("list_files") + " - Explore directory structure\n")
	fmt.Print("   ✓ " + renderer.Bold("search_files") + " - Find files by pattern (*.go, test_*.py)\n")
	fmt.Print("   ✓ " + renderer.Bold("grep_search") + " - Search text in files (with line numbers)\n")

	fmt.Print(renderer.Bold("\n⚡ Execution Tools:\n"))
	fmt.Print("   ✓ " + renderer.Bold("execute_command") + " - Run shell commands (pipes, redirects)\n")
	fmt.Print("   ✓ " + renderer.Bold("execute_program") + " - Run programs directly (no quoting issues)\n\n")

	fmt.Print("💡 Tip: Type " + renderer.Cyan("'/help'") + " for usage examples and patterns\n\n")
}

// printModelsList displays all available models
func printModelsList(renderer *display.Renderer, registry *ModelRegistry) {
	fmt.Print("\n" + renderer.Cyan("════════════════════════════════════════════════════════════════\n"))
	fmt.Print(renderer.Cyan("                      Available AI Models\n"))
	fmt.Print(renderer.Cyan("════════════════════════════════════════════════════════════════\n") + "\n")

	// Group models by backend
	geminiBakcend := registry.ListModelsByBackend("gemini")
	vertexAIBackend := registry.ListModelsByBackend("vertexai")

	if len(geminiBakcend) > 0 {
		fmt.Print(renderer.Bold("🔷 Gemini API Models:\n"))
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

			fmt.Printf("   %s %s %s - %s\n", icon, costIcon, renderer.Bold(model.Name), model.Description)
			fmt.Printf("      Context: %d tokens | Tools: %v | Vision: %v\n",
				model.ContextWindow,
				model.Capabilities.ToolUseSupport,
				model.Capabilities.VisionSupport)
		}
		fmt.Print("\n")
	}

	if len(vertexAIBackend) > 0 {
		fmt.Print(renderer.Bold("🔶 Vertex AI Models:\n"))
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

			fmt.Printf("   %s %s %s - %s\n", icon, costIcon, renderer.Bold(model.Name), model.Description)
			fmt.Printf("      Context: %d tokens | Tools: %v | Vision: %v\n",
				model.ContextWindow,
				model.Capabilities.ToolUseSupport,
				model.Capabilities.VisionSupport)
		}
		fmt.Print("\n")
	}

	fmt.Print(renderer.Dim("Use --model flag to select a model (e.g., --model gemini-1.5-pro)\n"))
	fmt.Print(renderer.Dim("Use /current-model command to see details about the active model\n\n"))
}

// printCurrentModelInfo displays detailed information about the current model
func printCurrentModelInfo(renderer *display.Renderer, model ModelConfig) {
	fmt.Print("\n" + renderer.Cyan("════════════════════════════════════════════════════════════════\n"))
	fmt.Print(renderer.Cyan("                 Current Model Information\n"))
	fmt.Print(renderer.Cyan("════════════════════════════════════════════════════════════════\n") + "\n")

	// Model name and backend
	backendIcon := "🔷"
	if model.Backend == "vertexai" {
		backendIcon = "🔶"
	}

	fmt.Print(renderer.Bold("Model: ") + fmt.Sprintf("%s %s (%s)\n", backendIcon, model.Name, model.Backend))
	fmt.Print("\n")

	// Description
	fmt.Print(renderer.Bold("Description:\n"))
	fmt.Print(renderer.Dim("  " + model.Description + "\n\n"))

	// Capabilities
	fmt.Print(renderer.Bold("Capabilities:\n"))
	if model.Capabilities.VisionSupport {
		fmt.Print("  ✓ Vision/Image Processing\n")
	} else {
		fmt.Print("  ✗ Vision/Image Processing\n")
	}
	if model.Capabilities.ToolUseSupport {
		fmt.Print("  ✓ Tool/Function Calling\n")
	} else {
		fmt.Print("  ✗ Tool/Function Calling\n")
	}
	if model.Capabilities.LongContextWindow {
		fmt.Print("  ✓ Long Context Window (1M+ tokens)\n")
	} else {
		fmt.Print("  ✗ Long Context Window\n")
	}
	fmt.Print("\n")

	// Context and Cost
	fmt.Print(renderer.Bold("Technical Details:\n"))
	fmt.Printf("  Context Window: %d tokens\n", model.ContextWindow)
	fmt.Printf("  Cost Tier: %s\n", model.Capabilities.CostTier)
	fmt.Print("\n")

	// Recommended use cases
	if len(model.RecommendedFor) > 0 {
		fmt.Print(renderer.Bold("Recommended For:\n"))
		for _, useCase := range model.RecommendedFor {
			fmt.Print("  • " + useCase + "\n")
		}
		fmt.Print("\n")
	}

	fmt.Print(renderer.Dim("Tip: Use --model flag to switch models when starting the agent\n\n"))
}

// printProvidersList displays available providers and their models
func printProvidersList(renderer *display.Renderer, registry *ModelRegistry) {
	fmt.Print("\n" + renderer.Cyan("════════════════════════════════════════════════════════════════\n"))
	fmt.Print(renderer.Cyan("                  Available Providers & Models\n"))
	fmt.Print(renderer.Cyan("════════════════════════════════════════════════════════════════\n") + "\n")

	// Display each provider
	for _, providerName := range registry.ListProviders() {
		provider := ParseProvider(providerName)
		meta := GetProviderMetadata(provider)

		// Provider header
		fmt.Printf("%s %s\n", meta.Icon, renderer.Bold(meta.DisplayName))
		fmt.Printf("   %s\n\n", meta.Description)

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
			fmt.Printf("   %s %s %s - %s\n", icon, costIcon, renderer.Bold(modelSyntax), model.Description)
		}

		fmt.Print("\n")
	}

	fmt.Print(renderer.Dim("Usage: --model provider/model (e.g., --model gemini/2.5-flash)\n"))
	fmt.Print(renderer.Dim("You can also use shorthands: --model gemini/flash\n"))
	fmt.Print(renderer.Dim("Use /current-model command to see details about the active model\n\n"))
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
