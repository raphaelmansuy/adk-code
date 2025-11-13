package orchestration

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"

	"code_agent/internal/config"
	"code_agent/internal/mcp"
	agentprompts "code_agent/internal/prompts"
)

// InitializeAgentComponent creates the coding agent with MCP support
// Returns the agent and MCP components
func InitializeAgentComponent(ctx context.Context, cfg *config.Config, llm model.LLM) (agent.Agent, *MCPComponents, error) {
	// Initialize MCP toolsets
	var mcpToolsets []tool.Toolset
	mcpComponents := &MCPComponents{
		Manager: nil,
		Enabled: false,
	}

	if cfg.MCPConfig != nil && cfg.MCPConfig.Enabled {
		mcpManager := mcp.NewManager()
		if err := mcpManager.LoadServers(ctx, cfg.MCPConfig); err != nil {
			return nil, nil, fmt.Errorf("failed to load MCP servers: %w", err)
		}
		mcpToolsets = mcpManager.Toolsets()
		mcpComponents.Manager = mcpManager
		mcpComponents.Enabled = true
	}

	ag, err := agentprompts.NewCodingAgent(ctx, agentprompts.Config{
		Model:            llm,
		WorkingDirectory: cfg.WorkingDirectory,
		EnableThinking:   cfg.EnableThinking,
		ThinkingBudget:   cfg.ThinkingBudget,
		MCPToolsets:      mcpToolsets,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create coding agent: %w", err)
	}

	return ag, mcpComponents, nil
}
