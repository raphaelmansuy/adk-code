// Package commands provides CLI command handlers organized by functionality.
package commands

import (
	"context"
	"fmt"
	"strings"

	"adk-code/internal/display"
	"adk-code/internal/mcp"
	agentprompts "adk-code/internal/prompts"
	"adk-code/internal/tracking"
	"adk-code/pkg/models"
	"adk-code/tools"
)

// HandleBuiltinCommand handles built-in REPL commands like /help, /tools, etc.
// Returns true if a command was handled, false if input should be sent to agent
// Note: /exit and /quit are handled separately in repl.go to break the loop
func HandleBuiltinCommand(ctx context.Context, input string, renderer *display.Renderer, sessionTokens *tracking.SessionTokens, modelRegistry *models.Registry, currentModel models.Config, mcpManager *mcp.Manager) bool {
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
		handleProvidersCommand(ctx, renderer, modelRegistry)
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
		// Check if it's an /mcp command
		if strings.HasPrefix(input, "/mcp") {
			handleMCPCommand(input, renderer, mcpManager)
			return true
		}
		return false
	}
}

// handlePromptCommand displays the XML-structured prompt
func handlePromptCommand(renderer *display.Renderer) {
	// Show the XML-structured prompt with minimal context
	registry := tools.GetRegistry()
	ctx := agentprompts.PromptContext{
		HasWorkspace:         false,
		WorkspaceRoot:        "",
		WorkspaceSummary:     "(Context not available in REPL)",
		EnvironmentMetadata:  "",
		EnableMultiWorkspace: false,
	}
	xmlPrompt := agentprompts.BuildEnhancedPromptWithContext(registry, ctx)

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
func handleProvidersCommand(ctx context.Context, renderer *display.Renderer, registry *models.Registry) {
	lines := buildProvidersListLines(ctx, renderer, registry)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// handleTokensCommand displays token usage statistics
func handleTokensCommand(sessionTokens *tracking.SessionTokens) {
	summary := sessionTokens.GetSummary()
	fmt.Print(tracking.FormatSessionSummary(summary))
}

// handleMCPCommand handles /mcp commands and subcommands
func handleMCPCommand(input string, renderer *display.Renderer, mcpManager *mcp.Manager) {
	// Handle case where MCP is disabled or not available
	if mcpManager == nil {
		fmt.Println(renderer.Yellow("⚠ MCP is not enabled. Use --mcp-config flag to enable MCP support."))
		return
	}

	parts := strings.Fields(input)
	if len(parts) == 1 {
		// Just "/mcp" - show help
		handleMCPHelp(renderer)
		return
	}

	subcommand := parts[1]
	switch subcommand {
	case "list":
		handleMCPList(renderer, mcpManager)
	case "status":
		handleMCPStatus(renderer, mcpManager)
	case "tools":
		handleMCPTools(renderer, mcpManager)
	case "help":
		handleMCPHelp(renderer)
	default:
		fmt.Println(renderer.Yellow(fmt.Sprintf("⚠ Unknown /mcp subcommand: %s", subcommand)))
		handleMCPHelp(renderer)
	}
}

// handleMCPHelp shows MCP command help
func handleMCPHelp(renderer *display.Renderer) {
	fmt.Println()
	fmt.Println(renderer.Bold("MCP Commands:"))
	fmt.Println()
	fmt.Println(renderer.Cyan("  /mcp list") + "     - List all configured MCP servers")
	fmt.Println(renderer.Cyan("  /mcp status") + "   - Show status and errors for MCP servers")
	fmt.Println(renderer.Cyan("  /mcp tools") + "    - List all tools provided by MCP servers")
	fmt.Println(renderer.Cyan("  /mcp help") + "     - Show this help message")
	fmt.Println()
}

// handleMCPList lists all configured MCP servers
func handleMCPList(renderer *display.Renderer, mcpManager *mcp.Manager) {
	servers := mcpManager.List()

	if len(servers) == 0 {
		fmt.Println(renderer.Yellow("⚠ No MCP servers configured"))
		return
	}

	fmt.Println()
	fmt.Println(renderer.Bold("Configured MCP Servers:"))
	fmt.Println()

	for _, serverName := range servers {
		fmt.Println(renderer.Cyan("  • ") + serverName)
	}
	fmt.Println()
	fmt.Println(renderer.Dim(fmt.Sprintf("Total: %d server(s)", len(servers))))
	fmt.Println()
}

// handleMCPStatus shows status and errors for MCP servers
func handleMCPStatus(renderer *display.Renderer, mcpManager *mcp.Manager) {
	status := mcpManager.Status()

	if len(status) == 0 {
		fmt.Println(renderer.Yellow("⚠ No MCP servers configured"))
		return
	}

	fmt.Println()
	fmt.Println(renderer.Bold("MCP Server Status:"))
	fmt.Println()

	hasErrors := false
	for serverName, err := range status {
		if err != nil {
			hasErrors = true
			fmt.Println(renderer.Red("  ✗ ") + renderer.Bold(serverName))
			fmt.Println(renderer.Dim("    Error: ") + err.Error())
		} else {
			fmt.Println(renderer.Green("  ✓ ") + renderer.Bold(serverName))
			fmt.Println(renderer.Dim("    Status: Connected"))
		}
		fmt.Println()
	}

	if !hasErrors {
		fmt.Println(renderer.Green("All servers connected successfully"))
		fmt.Println()
	}
}

// handleMCPTools lists all tools from MCP servers
func handleMCPTools(renderer *display.Renderer, mcpManager *mcp.Manager) {
	toolsets := mcpManager.Toolsets()

	if len(toolsets) == 0 {
		fmt.Println(renderer.Yellow("⚠ No tools available from MCP servers"))
		return
	}

	fmt.Println()
	fmt.Println(renderer.Bold("Tools from MCP Servers:"))
	fmt.Println()

	// Note: Tools() requires an agent.ReadonlyContext which is only available during agent execution
	// For now, we just show the number of toolsets loaded
	fmt.Println(renderer.Green(fmt.Sprintf("  ✓ %d MCP toolset(s) loaded successfully", len(toolsets))))
	fmt.Println()
	fmt.Println(renderer.Dim("  Note: Tool details are only available during agent execution."))
	fmt.Println(renderer.Dim("  The agent will have access to all MCP tools when processing requests."))
	fmt.Println()
}

// Helper functions for building display lines

// buildHelpMessageLines builds the help message as an array of lines for pagination

// Builder functions are in repl_builders.go
