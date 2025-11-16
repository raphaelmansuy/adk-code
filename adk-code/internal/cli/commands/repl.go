// Package commands provides CLI command handlers organized by functionality.
package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"adk-code/internal/config"
	"adk-code/internal/display"
	"adk-code/internal/mcp"
	agentprompts "adk-code/internal/prompts"
	"adk-code/internal/tracking"
	"adk-code/pkg/agents"
	"adk-code/pkg/models"
	"adk-code/tools"
)

// HandleBuiltinCommand handles built-in REPL commands like /help, /tools, etc.
// Returns true if a command was handled, false if input should be sent to agent
// Note: /exit and /quit are handled separately in repl.go to break the loop
func HandleBuiltinCommand(ctx context.Context, input string, renderer *display.Renderer, sessionTokens *tracking.SessionTokens, modelRegistry *models.Registry, currentModel models.Config, mcpManager *mcp.Manager, appConfig interface{}) bool {
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

	case "/compaction":
		handleCompactionCommand(renderer, appConfig)
		return true

	case "/agents":
		handleAgentsCommand(renderer)
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
		// Check if it's a /run-agent command
		if strings.HasPrefix(input, "/run-agent ") {
			agentRequest := strings.TrimPrefix(input, "/run-agent ")
			handleRunAgentCommand(renderer, agentRequest)
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

// handleCompactionCommand displays the session history compaction configuration
func handleCompactionCommand(renderer *display.Renderer, appConfig interface{}) {
	// Type assert to get the actual config
	cfg, ok := appConfig.(*config.Config)
	if !ok {
		fmt.Println(renderer.Red("Error: Unable to access configuration"))
		return
	}

	fmt.Println()
	fmt.Println(renderer.Bold("Session History Compaction Configuration:"))
	fmt.Println()

	// Display status
	if cfg.CompactionEnabled {
		fmt.Println(renderer.Green("✓ Status: ") + renderer.Cyan("ENABLED"))
	} else {
		fmt.Println(renderer.Yellow("⚠ Status: ") + renderer.Dim("DISABLED"))
	}
	fmt.Println()

	// Display current settings
	fmt.Println(renderer.Bold("Current Settings:"))
	fmt.Printf("  %s Invocation Threshold:  %d invocations\n", renderer.Dim("•"), cfg.CompactionThreshold)
	fmt.Printf("  %s Overlap Window:        %d invocations\n", renderer.Dim("•"), cfg.CompactionOverlap)
	fmt.Printf("  %s Token Threshold:       %d tokens\n", renderer.Dim("•"), cfg.CompactionTokens)
	fmt.Printf("  %s Safety Ratio:          %.1f%%\n", renderer.Dim("•"), cfg.CompactionSafety*100)
	fmt.Println()

	// Display what this means
	fmt.Println(renderer.Bold("What This Means:"))
	fmt.Println()
	fmt.Println(renderer.Dim("  • Invocation Threshold: Compaction triggers after this many agent interactions"))
	fmt.Println(renderer.Dim("  • Overlap Window: How many recent invocations to retain in context"))
	fmt.Println(renderer.Dim("  • Token Threshold: Summarization occurs when session exceeds this token limit"))
	fmt.Println(renderer.Dim("  • Safety Ratio: Buffer below the token limit to prevent exceeding it"))
	fmt.Println()

	// Display usage information
	fmt.Println(renderer.Bold("Enable Compaction:"))
	fmt.Println(renderer.Dim("  To enable compaction, start adk-code with the --compaction flag:"))
	fmt.Println()
	fmt.Println("  " + renderer.Cyan("adk-code --compaction"))
	fmt.Println()
	fmt.Println(renderer.Dim("  Or customize settings:"))
	fmt.Println()
	fmt.Println("  " + renderer.Cyan("adk-code --compaction \\"))
	fmt.Println("           " + renderer.Cyan("--compaction-threshold 5 \\"))
	fmt.Println("           " + renderer.Cyan("--compaction-overlap 2 \\"))
	fmt.Println("           " + renderer.Cyan("--compaction-tokens 700000 \\"))
	fmt.Println("           " + renderer.Cyan("--compaction-safety 0.7"))
	fmt.Println()
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

// handleAgentsCommand displays available agents
func handleAgentsCommand(renderer *display.Renderer) {
	// Get the project root from environment or current directory
	projectRoot := os.Getenv("ADK_PROJECT_ROOT")
	if projectRoot == "" {
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			fmt.Println(renderer.Red("Error: Could not determine project root"))
			return
		}
	}

	// Create a context with timeout to prevent hanging on network mounts or symlink loops
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Channel to receive discovery result
	type discoveryResult struct {
		result *agents.DiscoveryResult
		err    error
	}
	resultChan := make(chan discoveryResult, 1)

	// Run discovery in a goroutine to allow timeout
	go func() {
		discoverer := agents.NewDiscoverer(projectRoot)
		result, err := discoverer.DiscoverAll()
		resultChan <- discoveryResult{
			result: result,
			err:    err,
		}
	}()

	// Wait for discovery or timeout
	var result *agents.DiscoveryResult
	select {
	case res := <-resultChan:
		if res.err != nil {
			fmt.Println(renderer.Red(fmt.Sprintf("Error discovering agents: %v", res.err)))
			return
		}
		result = res.result
	case <-ctx.Done():
		fmt.Println(renderer.Yellow("⚠ Agent discovery timed out (5s). This may indicate slow file system access."))
		fmt.Println(renderer.Yellow("Tip: Check if ~/.adk/agents/ is on a network mount or has symlink issues."))
		return
	}

	lines := buildAgentsListLines(renderer, result)
	paginator := display.NewPaginator(renderer)
	paginator.DisplayPaged(lines)
}

// handleRunAgentCommand attempts to run a specific agent (preview feature)
func handleRunAgentCommand(renderer *display.Renderer, agentRequest string) {
	// Parse agent name from request (format: "agent-name" or "agent-name request")
	parts := strings.SplitN(agentRequest, " ", 2)
	agentName := strings.TrimSpace(parts[0])

	if agentName == "" {
		fmt.Println(renderer.Yellow("⚠ Please specify an agent name: /run-agent agent-name [request]"))
		return
	}

	// Get the project root from environment or current directory
	projectRoot := os.Getenv("ADK_PROJECT_ROOT")
	if projectRoot == "" {
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			fmt.Println(renderer.Red("Error: Could not determine project root"))
			return
		}
	}

	// Create a context with timeout to prevent hanging on network mounts or symlink loops
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Channel to receive discovery result
	type discoveryResult struct {
		agents []*agents.Agent
		err    error
	}
	resultChan := make(chan discoveryResult, 1)

	// Run discovery in a goroutine to allow timeout
	go func() {
		discoverer := agents.NewDiscoverer(projectRoot)
		result, err := discoverer.DiscoverAll()
		resultChan <- discoveryResult{
			agents: result.Agents,
			err:    err,
		}
	}()

	// Wait for discovery or timeout
	var discoveredAgents []*agents.Agent
	select {
	case res := <-resultChan:
		if res.err != nil {
			fmt.Println(renderer.Red(fmt.Sprintf("Error discovering agents: %v", res.err)))
			return
		}
		discoveredAgents = res.agents
	case <-ctx.Done():
		fmt.Println(renderer.Yellow("⚠ Agent discovery timed out (5s). This may indicate slow file system access."))
		fmt.Println(renderer.Yellow("Tip: Check if ~/.adk/agents/ is on a network mount or has symlink issues."))
		return
	}

	// Find the requested agent
	var foundAgent *agents.Agent
	for _, agent := range discoveredAgents {
		if agent.Name == agentName {
			foundAgent = agent
			break
		}
	}

	if foundAgent == nil {
		fmt.Println(renderer.Red(fmt.Sprintf("✗ Agent not found: %s", agentName)))
		fmt.Println()
		fmt.Println(renderer.Yellow("Available agents:"))
		for _, agent := range discoveredAgents {
			fmt.Println(renderer.Cyan("  • ") + renderer.Bold(agent.Name))
		}
		return
	}

	// Display agent info
	fmt.Println()
	fmt.Println(renderer.Green("✓ ") + renderer.Bold(foundAgent.Name))
	fmt.Println(renderer.Dim("  Description: ") + foundAgent.Description)
	if foundAgent.Version != "" {
		fmt.Println(renderer.Dim("  Version: ") + foundAgent.Version)
	}
	if foundAgent.Author != "" {
		fmt.Println(renderer.Dim("  Author: ") + foundAgent.Author)
	}
	if len(foundAgent.Tags) > 0 {
		fmt.Println(renderer.Dim("  Tags: ") + strings.Join(foundAgent.Tags, ", "))
	}
	fmt.Println()

	// Explain how agents actually work
	fmt.Println(renderer.Bold("💡 How to use this agent:"))
	fmt.Println()
	fmt.Println("Agents are automatically available as tools. The main agent will")
	fmt.Println("delegate to specialist agents when appropriate.")
	fmt.Println()

	if len(parts) > 1 {
		request := strings.TrimSpace(parts[1])
		fmt.Println(renderer.Cyan("Example Request: ") + renderer.Dim(request))
		fmt.Println()
		fmt.Println(renderer.Green("To execute this agent, simply ask the main agent:"))
		fmt.Println(renderer.Dim("  ❯ ") + request)
		fmt.Println()
		fmt.Println("The main agent will automatically delegate to " + renderer.Bold(foundAgent.Name) + " if appropriate.")
	} else {
		fmt.Println(renderer.Green("Examples of tasks that will delegate to this agent:"))
		fmt.Println()
		// Show contextual examples based on agent type
		switch foundAgent.Name {
		case "code-reviewer":
			fmt.Println(renderer.Dim("  ❯ Review the code in main.go for security issues"))
			fmt.Println(renderer.Dim("  ❯ Check if there are any bugs in the payment handler"))
		case "debugger":
			fmt.Println(renderer.Dim("  ❯ Why is the API returning 500 errors?"))
			fmt.Println(renderer.Dim("  ❯ Debug the authentication flow"))
		case "test-engineer":
			fmt.Println(renderer.Dim("  ❯ Write tests for the user service"))
			fmt.Println(renderer.Dim("  ❯ Improve test coverage in handlers/"))
		case "documentation-writer":
			fmt.Println(renderer.Dim("  ❯ Write a README for this project"))
			fmt.Println(renderer.Dim("  ❯ Document the API endpoints"))
		case "architect":
			fmt.Println(renderer.Dim("  ❯ Design a data model for user profiles"))
			fmt.Println(renderer.Dim("  ❯ Plan the architecture for a notification system"))
		default:
			fmt.Println(renderer.Dim("  ❯ Ask the main agent to perform tasks related to:"))
			fmt.Println(renderer.Dim("    " + foundAgent.Description))
		}
	}
	fmt.Println()
}

// Helper functions for building display lines

// buildHelpMessageLines builds the help message as an array of lines for pagination

// Builder functions are in repl_builders.go
