// Package commands provides CLI command handlers organized by functionality.
package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"adk-code/internal/display"
	"adk-code/internal/llm/backends"
	"adk-code/internal/session/compaction"
	"adk-code/pkg/agents"
	"adk-code/pkg/models"

	"google.golang.org/adk/session"
)

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
	lines = append(lines, "   • "+renderer.Bold("/agents")+" - List available ADK Code agents")
	lines = append(lines, "   • "+renderer.Bold("/run-agent <name>")+" - Show agent details or execute agent (preview)")
	lines = append(lines, "   • "+renderer.Bold("/prompt")+" - Display the system prompt")
	lines = append(lines, "   • "+renderer.Bold("/tokens")+" - Show token usage statistics")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("📊 Session Management (REPL commands):"))
	lines = append(lines, "   • "+renderer.Bold("/session")+" - Display current session history and event timeline")
	lines = append(lines, "   • "+renderer.Bold("/session event <index>")+" - View full content of a specific event")
	lines = append(lines, "   • "+renderer.Bold("/session <id>")+" - Display a specific session by ID")
	lines = append(lines, "   • "+renderer.Bold("/show-session <id>")+" - Display a session by ID (alias for /session <id>)")
	lines = append(lines, "   • "+renderer.Bold("/list-sessions")+" - List all available sessions")
	lines = append(lines, "   • "+renderer.Bold("/new-session")+" - Create a new session with auto-generated ID (session-YYYYMMDD-HHMMSS)")
	lines = append(lines, "   • "+renderer.Bold("/new-session <name>")+" - Create a new session with specified name")
	lines = append(lines, "   • "+renderer.Bold("/switch-session <id>")+" - Switch to a different session")
	lines = append(lines, "   • "+renderer.Bold("/delete-session <name>")+" - Delete a session (with confirmation)")
	lines = append(lines, "   • "+renderer.Bold("/compaction")+" - Show session history compaction configuration")
	lines = append(lines, "   • "+renderer.Bold("/mcp")+" - Manage MCP servers (list, status, tools)")
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

	lines = append(lines, renderer.Bold("📦 Session History Compaction:"))
	lines = append(lines, "   Automatically summarize old conversation history to save tokens:")
	lines = append(lines, "   • "+renderer.Dim("./code-agent --compaction")+" (enable with defaults)")
	lines = append(lines, "   • "+renderer.Dim("./code-agent --compaction --compaction-threshold 5")+" (customize)")
	lines = append(lines, "   Use "+renderer.Cyan("'/compaction'")+" command in REPL to see current settings")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("📚 Session Management (CLI commands):"))
	lines = append(lines, "   These commands are run from the shell (not in the REPL):")
	lines = append(lines, "   • "+renderer.Dim("./code-agent new-session <name>")+" - Create a new session")
	lines = append(lines, "   • "+renderer.Dim("./code-agent list-sessions")+" - List all sessions")
	lines = append(lines, "   • "+renderer.Dim("./code-agent delete-session <name>")+" - Delete a session")
	lines = append(lines, "   • "+renderer.Dim("./code-agent --session <name>")+" - Resume a specific session")
	lines = append(lines, "")
	lines = append(lines, renderer.Dim("   Note: Use /list-sessions, /new-session, /delete-session in REPL instead"))

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
	lines = append(lines, "   ✓ "+renderer.Bold("apply_v4a_patch")+" - Apply V4A semantic patches")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("🔍 Discovery & Search Tools:"))
	lines = append(lines, "   ✓ "+renderer.Bold("list_files")+" - Explore directory structure")
	lines = append(lines, "   ✓ "+renderer.Bold("search_files")+" - Find files by pattern (*.go, test_*.py)")
	lines = append(lines, "   ✓ "+renderer.Bold("grep_search")+" - Search text in files (with line numbers)")
	lines = append(lines, "   ✓ "+renderer.Bold("preview_replace")+" - Preview search/replace results before applying")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("🌐 Web Tools:"))
	lines = append(lines, "   ✓ "+renderer.Bold("fetch_web")+" - Fetch and parse web content from URLs")
	lines = append(lines, "   ✓ "+renderer.Bold("google_search")+" - Search the web with Google (real-time)")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("⚡ Execution Tools:"))
	lines = append(lines, "   ✓ "+renderer.Bold("execute_command")+" - Run shell commands (pipes, redirects)")
	lines = append(lines, "   ✓ "+renderer.Bold("execute_program")+" - Run programs directly (no quoting issues)")
	lines = append(lines, "")

	// Agent management is not an LLM tool; it is controlled via CLI commands and
	// the REPL command '/agents'. Keep agent management functionality implemented
	// in the `tools/agents` package (used by CLI tools), but do not advertise it
	// in the model-accessible tools list. This prevents accidental LLM tool use.

	// Also show a short pointer to the REPL/CLI commands so users know how to
	// interact with agent functionality without exposing it as a model-callable
	// tool.
	lines = append(lines, renderer.Bold("🛠 Agent Management (CLI-only):"))
	lines = append(lines, "   • "+renderer.Bold("/agents")+" - Discover and list ADK Code agents (REPL command; not model-callable)")
	lines = append(lines, "   • "+renderer.Bold("adk-code agents")+" - CLI agent helper commands (not model-callable)")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("📊 Discovery & Info Tools:"))
	lines = append(lines, "   ✓ "+renderer.Bold("list_models")+" - List available AI models and capabilities")
	lines = append(lines, "   ✓ "+renderer.Bold("model_info")+" - Get detailed info about a specific model")
	lines = append(lines, "")

	lines = append(lines, renderer.Bold("🎨 Display & UI Tools:"))
	lines = append(lines, "   ✓ "+renderer.Bold("display_message")+" - Display formatted messages to user")
	lines = append(lines, "   ✓ "+renderer.Bold("update_task_list")+" - Show/update task progress in REPL")
	lines = append(lines, "")

	lines = append(lines, renderer.Dim("💡 Total: 18 tools across 6 categories | Type /help for commands"))
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
func buildProvidersListLines(ctx context.Context, renderer *display.Renderer, registry *models.Registry) []string {
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
		var modelsCfg []models.Config

		// For Ollama, try to dynamically discover models from the server
		if providerName == "ollama" {
			ollamaProvider := backends.NewOllamaProvider()
			if dynamicModels, err := ollamaProvider.ListModels(ctx, false); err == nil && len(dynamicModels) > 0 {
				// Display dynamic models from Ollama
				lines = append(lines, renderer.Dim("   Dynamic models from Ollama server:"))
				lines = append(lines, "")

				for _, modelInfo := range dynamicModels {
					icon := "○"
					costIcon := "💰"
					modelSyntax := fmt.Sprintf("ollama/%s", modelInfo.Name)
					description := ""
					if modelInfo.Description != "" {
						description = fmt.Sprintf(" - %s", modelInfo.Description)
					}
					lines = append(lines, fmt.Sprintf("   %s %s %s%s", icon, costIcon, renderer.Bold(modelSyntax), description))
				}
				lines = append(lines, "")
				continue
			}
			// Fall back to static models if dynamic discovery fails
		}

		// Use static models from registry (fallback)
		modelsCfg = registry.GetProviderModels(providerName)
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

// buildAgentsListLines builds the agents list as an array of lines for pagination
func buildAgentsListLines(renderer *display.Renderer, result *agents.DiscoveryResult) []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, renderer.Cyan("                    Available Agents"))
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, "")

	if result.IsEmpty() {
		lines = append(lines, renderer.Yellow("⚠ No agents found in .adk/agents/"))
		lines = append(lines, "")
		lines = append(lines, "To create an agent, create a .adk/agents/my-agent.md file with:")
		lines = append(lines, "")
		lines = append(lines, renderer.Dim("---"))
		lines = append(lines, renderer.Dim("name: my-agent"))
		lines = append(lines, renderer.Dim("description: What this agent does"))
		lines = append(lines, renderer.Dim("---"))
		lines = append(lines, renderer.Dim("# My Agent"))
		lines = append(lines, renderer.Dim("[agent content here]"))
		lines = append(lines, "")
		return lines
	}

	lines = append(lines, renderer.Bold("ADK Code Agents:"))
	lines = append(lines, "")

	for _, agent := range result.Agents {
		// Main agent entry
		lines = append(lines, "  • "+renderer.Bold(agent.Name))

		// Description is indented and styled
		lines = append(lines, "    "+agent.Description)

		// Show optional metadata
		var metadata []string
		if agent.Version != "" {
			metadata = append(metadata, "v"+agent.Version)
		}
		if agent.Author != "" {
			metadata = append(metadata, "by "+agent.Author)
		}
		if len(agent.Tags) > 0 {
			metadata = append(metadata, "Tags: "+strings.Join(agent.Tags, ", "))
		}

		if len(metadata) > 0 {
			lines = append(lines, renderer.Dim("    ("+strings.Join(metadata, " • ")+")"))
		}

		lines = append(lines, "")
	}

	lines = append(lines, renderer.Dim(fmt.Sprintf("Total: %d agent(s) discovered", result.Total)))
	lines = append(lines, "")
	lines = append(lines, renderer.Bold("💡 How Agents Work:"))
	lines = append(lines, "")
	lines = append(lines, "Agents are automatically available as specialist tools. Simply ask")
	lines = append(lines, "the main agent to perform tasks, and it will delegate to the appropriate")
	lines = append(lines, "specialist agent when needed.")
	lines = append(lines, "")
	lines = append(lines, renderer.Bold("Examples:"))
	lines = append(lines, renderer.Dim("  ❯ Review the security in auth.go"))
	lines = append(lines, renderer.Dim("    → Automatically delegates to code-reviewer"))
	lines = append(lines, "")
	lines = append(lines, renderer.Dim("  ❯ Write tests for the API handlers"))
	lines = append(lines, renderer.Dim("    → Automatically delegates to test-engineer"))
	lines = append(lines, "")
	lines = append(lines, renderer.Dim("  ❯ Why is the database connection failing?"))
	lines = append(lines, renderer.Dim("    → Automatically delegates to debugger"))
	lines = append(lines, "")
	lines = append(lines, renderer.Bold("Commands:"))
	lines = append(lines, "  • "+renderer.Cyan("/run-agent <name>")+" - View agent details and examples")
	lines = append(lines, "")

	return lines
}

// buildSessionDisplayLines builds session event history as paginated lines
func buildSessionDisplayLines(renderer *display.Renderer, sess session.Session) []string {
	var lines []string

	// === HEADER ===
	lines = append(lines, "")
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, renderer.Cyan(fmt.Sprintf("                    Session: %s", sess.ID())))
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, "")

	// === SESSION METADATA ===
	lines = append(lines, renderer.Bold("📋 Session Details:"))
	lines = append(lines, fmt.Sprintf("  %s App:      %s", renderer.Dim("•"), sess.AppName()))
	lines = append(lines, fmt.Sprintf("  %s User:     %s", renderer.Dim("•"), sess.UserID()))
	lines = append(lines, fmt.Sprintf("  %s ID:       %s", renderer.Dim("•"), sess.ID()))

	// Update time
	lastUpdate := sess.LastUpdateTime()
	lines = append(lines, fmt.Sprintf("  %s Updated:  %s (%s ago)",
		renderer.Dim("•"),
		lastUpdate.Format("2006-01-02 15:04:05"),
		formatTimeAgo(lastUpdate),
	))
	lines = append(lines, "")

	// === EVENT SUMMARY ===
	events := sess.Events()
	totalEvents := events.Len()

	eventCounts := countEventsByType(events)
	lines = append(lines, renderer.Bold(fmt.Sprintf("📊 Events: %d total", totalEvents)))

	if eventCounts.userMessages > 0 {
		lines = append(lines, fmt.Sprintf("  %s User inputs:      %d",
			renderer.Dim("•"), eventCounts.userMessages))
	}
	if eventCounts.modelResponses > 0 {
		lines = append(lines, fmt.Sprintf("  %s Model responses:  %d",
			renderer.Dim("•"), eventCounts.modelResponses))
	}
	if eventCounts.toolCalls > 0 {
		lines = append(lines, fmt.Sprintf("  %s Tool calls:       %d",
			renderer.Dim("•"), eventCounts.toolCalls))
	}
	if eventCounts.toolResults > 0 {
		lines = append(lines, fmt.Sprintf("  %s Tool results:     %d",
			renderer.Dim("•"), eventCounts.toolResults))
	}
	if eventCounts.compactions > 0 {
		lines = append(lines, fmt.Sprintf("  %s Compactions:      %d",
			renderer.Green("★"), eventCounts.compactions))
	}
	lines = append(lines, "")

	// === CUMULATIVE TOKENS ===
	totalTokens := countTotalTokens(events)
	if totalTokens > 0 {
		lines = append(lines, renderer.Bold(fmt.Sprintf("🎯 Token Usage: %d total", totalTokens)))
		lines = append(lines, "")
	}

	// === EVENT TIMELINE ===
	if totalEvents == 0 {
		lines = append(lines, renderer.Yellow("⚠ No events in this session yet"))
		return lines
	}

	lines = append(lines, renderer.Bold("📜 Event Timeline:"))
	lines = append(lines, "")

	// Iterate through events
	for i := 0; i < events.Len(); i++ {
		event := events.At(i)
		if event == nil {
			continue
		}

		// Event number and timestamp
		ts := event.Timestamp.Format("15:04:05")
		eventNum := fmt.Sprintf("[%d/%d]", i+1, totalEvents)

		lines = append(lines, renderer.Dim(fmt.Sprintf("  %s %s", eventNum, ts)))

		// Determine event type and format accordingly
		if compactionMeta, err := compaction.GetCompactionMetadata(event); err == nil {
			// === COMPACTION EVENT ===
			lines = append(lines, renderer.Green("    ★ COMPACTION EVENT"))
			lines = append(lines, fmt.Sprintf("      Events compressed:    %d → summary",
				compactionMeta.EventCount))
			lines = append(lines, fmt.Sprintf("      Tokens saved:          %d → %d (%.1f%% compression)",
				compactionMeta.OriginalTokens,
				compactionMeta.CompactedTokens,
				compactionMeta.CompressionRatio*100,
			))
			lines = append(lines, fmt.Sprintf("      Period:                %s to %s",
				compactionMeta.StartTimestamp.Format("15:04:05"),
				compactionMeta.EndTimestamp.Format("15:04:05"),
			))
		} else if event.Content != nil {
			// === REGULAR EVENT ===
			author := event.Author
			if author == "" {
				author = "system"
			}

			// Color-code by author. Support subagents (e.g., "coding_agent")
			// which should be displayed as a model/agent response rather than a "?" fallback.
			var authorStr string
			switch author {
			case "user":
				authorStr = renderer.Blue("👤 USER")
			case "model":
				authorStr = renderer.Green("🤖 MODEL")
			case "system":
				authorStr = renderer.Yellow("⚙️  SYSTEM")
			default:
				if strings.Contains(author, "agent") {
					authorStr = renderer.Green("🤖 AGENT")
				} else {
					authorStr = renderer.Dim(fmt.Sprintf("❓ %s", author))
				}
			}

			lines = append(lines, fmt.Sprintf("    %s (ID: %s)",
				authorStr,
				truncateID(event.ID, 8),
			))

			// Display content preview
			if len(event.Content.Parts) > 0 {
				for _, part := range event.Content.Parts {
					if part == nil {
						continue
					}

					// Handle tool calls (function invocations)
					if part.FunctionCall != nil {
						lines = append(lines, fmt.Sprintf("      %s Tool: %s",
							renderer.Cyan("🔧"),
							renderer.Bold(part.FunctionCall.Name)))
					}

					// Handle tool responses (results)
					if part.FunctionResponse != nil {
						lines = append(lines, fmt.Sprintf("      %s %s",
							renderer.Green("✅ Result:"),
							renderer.Bold(part.FunctionResponse.Name)))
					}

					// Handle text content
					if part.Text != "" {
						preview := truncateText(part.Text, 60)
						lines = append(lines, fmt.Sprintf("      %s", renderer.Dim(preview)))
					}
				}
			}

			// Show token count if available
			if event.UsageMetadata != nil {
				promptTokens := int(event.UsageMetadata.PromptTokenCount)
				outputTokens := int(event.UsageMetadata.CandidatesTokenCount)
				if promptTokens > 0 || outputTokens > 0 {
					lines = append(lines, fmt.Sprintf("      %s",
						renderer.Dim(fmt.Sprintf("Tokens: %d prompt + %d output",
							promptTokens, outputTokens))))
				}
			}
		}

		// Add spacing between events
		if i < events.Len()-1 {
			lines = append(lines, "")
		}
	}

	// === FOOTER ===
	lines = append(lines, "")
	lines = append(lines, renderer.Dim("Press SPACE to continue, Q to quit"))
	lines = append(lines, "")

	return lines
}

// eventTypeCounts tracks the breakdown of event types
type eventTypeCounts struct {
	userMessages   int
	modelResponses int
	toolCalls      int
	toolResults    int
	compactions    int
}

// countEventsByType counts events by their type
func countEventsByType(events session.Events) eventTypeCounts {
	counts := eventTypeCounts{}

	for i := 0; i < events.Len(); i++ {
		evt := events.At(i)
		if evt == nil {
			continue
		}

		// Check if it's a compaction event first
		if compaction.IsCompactionEvent(evt) {
			counts.compactions++
		} else if evt.Author == "user" {
			counts.userMessages++
		} else if evt.Author == "model" || strings.Contains(evt.Author, "agent") {
			counts.modelResponses++
		}

		// Count tool calls and results from Content.Parts
		if evt.Content != nil && len(evt.Content.Parts) > 0 {
			for _, part := range evt.Content.Parts {
				if part != nil {
					if part.FunctionCall != nil {
						counts.toolCalls++
					}
					if part.FunctionResponse != nil {
						counts.toolResults++
					}
				}
			}
		}
	}

	return counts
}

// countTotalTokens sums up all tokens across events
func countTotalTokens(events session.Events) int {
	total := 0
	for i := 0; i < events.Len(); i++ {
		evt := events.At(i)
		if evt == nil {
			continue
		}
		if evt.UsageMetadata != nil {
			total += int(evt.UsageMetadata.TotalTokenCount)
		}
	}
	return total
}

// truncateText truncates text to maxLen characters with ellipsis
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}

// truncateID truncates an ID string to maxLen characters
func truncateID(id string, maxLen int) string {
	if len(id) <= maxLen {
		return id
	}
	return id[:maxLen]
}

// formatTimeAgo formats a time in human-readable "ago" format
func formatTimeAgo(t time.Time) string {
	elapsed := time.Since(t)
	if elapsed < time.Minute {
		return "just now"
	} else if elapsed < time.Hour {
		minutes := int(elapsed.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	} else if elapsed < 24*time.Hour {
		hours := int(elapsed.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}
	days := int(elapsed.Hours()) / 24
	return fmt.Sprintf("%dd ago", days)
}

// buildEventDisplayLines builds a detailed display of a single event without truncation
func buildEventDisplayLines(renderer *display.Renderer, evt *session.Event, sessionID string) []string {
	var lines []string

	// === HEADER ===
	lines = append(lines, "")
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, renderer.Cyan(fmt.Sprintf("                    Event: %s", evt.ID)))
	lines = append(lines, renderer.Cyan("════════════════════════════════════════════════════════════════"))
	lines = append(lines, "")

	// === EVENT METADATA ===
	lines = append(lines, renderer.Bold("📋 Event Details:"))
	lines = append(lines, fmt.Sprintf("  %s Session:    %s", renderer.Dim("•"), sessionID))
	lines = append(lines, fmt.Sprintf("  %s Event ID:   %s", renderer.Dim("•"), evt.ID))
	lines = append(lines, fmt.Sprintf("  %s Timestamp:  %s", renderer.Dim("•"), evt.Timestamp.Format("2006-01-02 15:04:05")))
	if evt.InvocationID != "" {
		lines = append(lines, fmt.Sprintf("  %s Invocation: %s", renderer.Dim("•"), evt.InvocationID))
	}
	if evt.Author != "" {
		author := evt.Author
		var authorLabel string
		switch author {
		case "user":
			authorLabel = renderer.Blue("👤 USER")
		case "model":
			authorLabel = renderer.Green("🤖 MODEL")
		case "system":
			authorLabel = renderer.Yellow("⚙️  SYSTEM")
		default:
			if strings.Contains(author, "agent") {
				authorLabel = renderer.Green("🤖 AGENT")
			} else {
				authorLabel = evt.Author
			}
		}
		lines = append(lines, fmt.Sprintf("  %s Author:     %s", renderer.Dim("•"), authorLabel))
	}
	if evt.Branch != "" {
		lines = append(lines, fmt.Sprintf("  %s Branch:     %s", renderer.Dim("•"), evt.Branch))
	}
	lines = append(lines, "")

	// === TOKEN USAGE ===
	if evt.UsageMetadata != nil {
		lines = append(lines, renderer.Bold("🎯 Token Usage:"))
		lines = append(lines, fmt.Sprintf("  %s Prompt Tokens:    %d", renderer.Dim("•"), evt.UsageMetadata.PromptTokenCount))
		lines = append(lines, fmt.Sprintf("  %s Output Tokens:    %d", renderer.Dim("•"), evt.UsageMetadata.CandidatesTokenCount))
		lines = append(lines, fmt.Sprintf("  %s Total Tokens:     %d", renderer.Dim("•"), evt.UsageMetadata.TotalTokenCount))
		lines = append(lines, "")
	}

	// === CONTENT ===
	if evt.Content != nil && len(evt.Content.Parts) > 0 {
		lines = append(lines, renderer.Bold("📝 Content:"))
		lines = append(lines, "")

		for partIdx, part := range evt.Content.Parts {
			if part == nil {
				continue
			}

			// Display tool calls with icon
			if part.FunctionCall != nil {
				if partIdx > 0 {
					lines = append(lines, "")
				}
				lines = append(lines, renderer.Cyan("🔧 Tool Call:"))
				lines = append(lines, fmt.Sprintf("  Tool: %s", renderer.Bold(part.FunctionCall.Name)))
				if len(part.FunctionCall.Args) > 0 {
					lines = append(lines, "  Arguments:")
					for k, v := range part.FunctionCall.Args {
						lines = append(lines, fmt.Sprintf("    %s: %v", k, v))
					}
				}
			}

			// Display tool responses with icon
			if part.FunctionResponse != nil {
				if partIdx > 0 {
					lines = append(lines, "")
				}
				lines = append(lines, renderer.Green("✅ Tool Result:"))
				lines = append(lines, fmt.Sprintf("  Tool: %s", renderer.Bold(part.FunctionResponse.Name)))
				if len(part.FunctionResponse.Response) > 0 {
					lines = append(lines, "  Response:")
					for k, v := range part.FunctionResponse.Response {
						lines = append(lines, fmt.Sprintf("    %s: %v", k, v))
					}
				}
			}

			if part.Text != "" {
				// Display full text without truncation
				if partIdx > 0 {
					lines = append(lines, "")
				}
				contentLines := strings.Split(part.Text, "\n")
				for _, contentLine := range contentLines {
					lines = append(lines, "  "+contentLine)
				}
			}
		}
		lines = append(lines, "")
	}

	// === FOOTER ===
	lines = append(lines, renderer.Dim("Press SPACE to continue, Q to quit"))
	lines = append(lines, "")

	return lines
}
