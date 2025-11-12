// Package commands provides CLI command handlers organized by functionality.
package commands

import (
	"fmt"
	"strings"

	codingagent "code_agent/agent"
	"code_agent/display"
	"code_agent/pkg/models"
	"code_agent/tools"
	"code_agent/tracking"
)

// HandleBuiltinCommand handles built-in REPL commands like /help, /tools, etc.
// Returns true if a command was handled, false if input should be sent to agent
// Note: /exit and /quit are handled separately in repl.go to break the loop
func HandleBuiltinCommand(input string, renderer *display.Renderer, sessionTokens *tracking.SessionTokens, modelRegistry *models.Registry, currentModel models.Config) bool {
	switch input {
	case "/prompt":
		handlePromptCommand(renderer)
		return true

	case "/help":
		handleHelpCommand(renderer)
		return true

	case "/tools":
		handleToolsCommand(renderer)
		return true

	case "/models":
		handleModelsCommand(renderer, modelRegistry)
		return true

	case "/current-model":
		handleCurrentModelCommand(renderer, currentModel)
		return true

	case "/providers":
		handleProvidersCommand(renderer, modelRegistry)
		return true

	case "/tokens":
		handleTokensCommand(sessionTokens)
		return true

	default:
		// Check if it's a /set-model command
		if strings.HasPrefix(input, "/set-model ") {
			modelSpec := strings.TrimPrefix(input, "/set-model ")
			HandleSetModel(renderer, modelRegistry, modelSpec)
			return true
		}
		return false
	}
}

// handlePromptCommand displays the XML-structured prompt
func handlePromptCommand(renderer *display.Renderer) {
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
}

// handleHelpCommand displays the help message
func handleHelpCommand(renderer *display.Renderer) {
	lines := buildHelpMessageLines(renderer)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// handleToolsCommand displays the available tools
func handleToolsCommand(renderer *display.Renderer) {
	lines := buildToolsListLines(renderer)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// handleModelsCommand displays all available models
func handleModelsCommand(renderer *display.Renderer, registry *models.Registry) {
	lines := buildModelsListLines(renderer, registry)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// handleCurrentModelCommand displays detailed information about the current model
func handleCurrentModelCommand(renderer *display.Renderer, model models.Config) {
	lines := buildCurrentModelInfoLines(renderer, model)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// handleProvidersCommand displays available providers and their models
func handleProvidersCommand(renderer *display.Renderer, registry *models.Registry) {
	lines := buildProvidersListLines(renderer, registry)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// handleTokensCommand displays token usage statistics
func handleTokensCommand(sessionTokens *tracking.SessionTokens) {
	summary := sessionTokens.GetSummary()
	fmt.Print(tracking.FormatSessionSummary(summary))
}

// Helper functions for building display lines

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

// buildModelsListLines builds the models list as an array of lines for pagination
func buildModelsListLines(renderer *display.Renderer, registry *models.Registry) []string {
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

// buildCurrentModelInfoLines builds the current model info as an array of lines for pagination
func buildCurrentModelInfoLines(renderer *display.Renderer, model models.Config) []string {
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

// buildProvidersListLines builds the providers list as an array of lines for pagination
func buildProvidersListLines(renderer *display.Renderer, registry *models.Registry) []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, renderer.Cyan("                  Available Providers & Models"))
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, "")

	// Display each provider
	for _, providerName := range registry.ListProviders() {
		provider := models.ParseProvider(providerName)
		meta := models.GetProviderMetadata(provider)

		// Provider header
		lines = append(lines, fmt.Sprintf("%s %s", meta.Icon, renderer.Bold(meta.DisplayName)))
		lines = append(lines, fmt.Sprintf("   %s", meta.Description))
		lines = append(lines, "")

		// List models for this provider
		modelsCfg := registry.GetProviderModels(providerName)
		for _, model := range modelsCfg {
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
