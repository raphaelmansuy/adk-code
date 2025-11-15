// Package agents provides subagent delegation using ADK's native agent-as-tool pattern
package agents

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"

	adkcontext "adk-code/internal/context"
	"adk-code/pkg/agents"
	"adk-code/pkg/models"
	common "adk-code/tools/base"
)

// SubAgentManager creates and manages subagent tools
// This uses Google ADK's native agent-as-tool pattern via agenttool.New()
//
// Context Management for Sub-Agents:
// Each sub-agent gets its own dedicated ContextManager for independent context tracking:
// - Sub-agents inherit the model (modelLLM) from the main agent
// - Each sub-agent has a separate context budget and compaction management
// - Compaction for sub-agents uses a dedicated ContextManager
// - This allows sub-agents to operate independently without affecting main agent context
type SubAgentManager struct {
	projectRoot string
	modelLLM    model.LLM
	modelConfig models.Config                         // Model configuration for creating context managers
	mcpToolsets []tool.Toolset                        // MCP toolsets to make available to subagents
	contextMgrs map[string]*adkcontext.ContextManager // Dedicated context managers per sub-agent
	mu          sync.RWMutex                          // Protects contextMgrs map
}

// NewSubAgentManager creates a new subagent manager
func NewSubAgentManager(projectRoot string, modelLLM model.LLM, modelConfig models.Config) *SubAgentManager {
	return &SubAgentManager{
		projectRoot: projectRoot,
		modelLLM:    modelLLM,
		modelConfig: modelConfig,
		mcpToolsets: []tool.Toolset{},
		contextMgrs: make(map[string]*adkcontext.ContextManager),
	}
}

// NewSubAgentManagerWithMCP creates a subagent manager with MCP toolsets
func NewSubAgentManagerWithMCP(projectRoot string, modelLLM model.LLM, modelConfig models.Config, mcpToolsets []tool.Toolset) *SubAgentManager {
	return &SubAgentManager{
		projectRoot: projectRoot,
		modelLLM:    modelLLM,
		modelConfig: modelConfig,
		mcpToolsets: mcpToolsets,
		contextMgrs: make(map[string]*adkcontext.ContextManager),
	}
}

// LoadSubAgentTools discovers agent definitions and converts them to tools
// Returns a list of tools that can be registered with the main agent
func (m *SubAgentManager) LoadSubAgentTools(ctx context.Context) ([]tool.Tool, error) {
	// Discover agent definitions
	discoverer := agents.NewDiscoverer(m.projectRoot)
	result, err := discoverer.DiscoverAll()
	if err != nil {
		return nil, fmt.Errorf("failed to discover agents: %w", err)
	}

	if result.IsEmpty() {
		// No subagents found - this is OK, return empty list
		return []tool.Tool{}, nil
	}

	// Convert each agent definition to a tool
	var subagentTools []tool.Tool
	for _, agentDef := range result.Agents {
		// Create an llmagent from the definition
		subAgent, err := m.createSubAgent(agentDef)
		if err != nil {
			// Log error but continue with other agents
			fmt.Fprintf(os.Stderr, "Warning: Failed to create subagent %s: %v\n", agentDef.Name, err)
			continue
		}

		// Convert the agent to a tool using ADK's agenttool
		agentTool := agenttool.New(subAgent, &agenttool.Config{
			SkipSummarization: false, // Let ADK summarize subagent results
		})

		// Register with common tool registry
		common.Register(common.ToolMetadata{
			Tool:      agentTool,
			Category:  common.CategoryExecution,
			Priority:  9, // High priority - delegation is a key capability
			UsageHint: fmt.Sprintf("Delegate to %s: %s", agentDef.Name, agentDef.Description),
		})

		subagentTools = append(subagentTools, agentTool)
	}

	return subagentTools, nil
}

// createSubAgent creates an llmagent from an agent definition with dedicated context management
// Each sub-agent gets its own ContextManager for independent compaction and token tracking.
func (m *SubAgentManager) createSubAgent(agentDef *agents.Agent) (agent.Agent, error) {
	// Parse allowed tools from agent definition
	allowedTools := m.parseAllowedTools(agentDef)

	// Create a dedicated ContextManager for this sub-agent
	// This allows the sub-agent to have independent context management and compaction
	contextManager := adkcontext.NewContextManagerWithModel(m.modelConfig, m.modelLLM)

	// Store the context manager for this sub-agent
	m.mu.Lock()
	m.contextMgrs[agentDef.Name] = contextManager
	m.mu.Unlock()

	// Create the subagent using ADK's llmagent
	// The subagent gets its own isolated context and uses the agent's content as instruction
	// It inherits the model (m.modelLLM) from the main agent, ensuring:
	// 1. Consistent behavior across main agent and sub-agents
	// 2. Independent context management via dedicated ContextManager
	// 3. Separate compaction for each sub-agent
	subAgent, err := llmagent.New(llmagent.Config{
		Name:        agentDef.Name,
		Description: agentDef.Description,
		Model:       m.modelLLM,       // Inherits model from main agent
		Instruction: agentDef.Content, // The markdown content is the system instruction
		Tools:       allowedTools,     // Restricted toolset based on agent definition
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create llmagent: %w", err)
	}

	return subAgent, nil
}

// GetContextManager returns the ContextManager for a specific sub-agent
func (m *SubAgentManager) GetContextManager(agentName string) *adkcontext.ContextManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contextMgrs[agentName]
}

// parseAllowedTools extracts and resolves the allowed tools for a subagent
// Reads the 'tools' field from agent YAML frontmatter
// Supports both built-in tools and MCP tools
func (m *SubAgentManager) parseAllowedTools(agentDef *agents.Agent) []tool.Tool {
	// Parse the tools field from YAML
	// Expected format: "tools: read_file, grep_search, execute_command" (exact tool names) or "tools: *" for all tools
	toolsSpec := m.extractToolsFromYAML(agentDef.RawYAML)
	if toolsSpec == "" {
		// No tools specified - agent is analysis-only
		return []tool.Tool{}
	}

	// Special case: "*" means all available tools (built-in + MCP)
	if strings.TrimSpace(toolsSpec) == "*" {
		return m.getAllAvailableTools()
	}

	// Parse comma-separated tool names
	toolNames := splitAndTrim(toolsSpec)

	// Resolve tools from both built-in registry and MCP toolsets
	// Use exact tool names - no mapping to avoid confusion
	var allowedTools []tool.Tool
	for _, toolName := range toolNames {
		// Normalize to lowercase and trim whitespace
		normalizedName := strings.ToLower(strings.TrimSpace(toolName))

		// Find tool by exact name
		if t := m.findToolByName(normalizedName); t != nil {
			allowedTools = append(allowedTools, t)
		} else {
			// Tool not found - log warning with suggestion to use `/tools` command
			fmt.Fprintf(os.Stderr, "Warning: Tool '%s' not found for agent '%s' (use '/tools' to see available tools)\n",
				toolName, agentDef.Name)
		}
	}

	return allowedTools
}

// getAllAvailableTools returns all built-in and MCP tools
func (m *SubAgentManager) getAllAvailableTools() []tool.Tool {
	// Get built-in tools
	registry := common.GetRegistry()
	allTools := registry.GetAllTools()

	// Add MCP tools from toolsets
	// Note: MCP toolsets require context, but we pass nil for tool enumeration
	for _, toolset := range m.mcpToolsets {
		mcpTools, err := toolset.Tools(nil)
		if err != nil {
			// Log error but continue with other toolsets
			fmt.Fprintf(os.Stderr, "Warning: Failed to get tools from MCP toolset: %v\n", err)
			continue
		}
		allTools = append(allTools, mcpTools...)
	}

	return allTools
}

// findToolByName searches for a tool in both built-in and MCP toolsets
// Supports both exact matches and builtin_ prefix variations
// For example: "read_file" will match "builtin_read_file"
func (m *SubAgentManager) findToolByName(name string) tool.Tool {
	// Search in built-in tools
	registry := common.GetRegistry()
	builtinTools := registry.GetAllTools()

	for _, t := range builtinTools {
		toolName := t.Name()
		// Direct match
		if toolName == name {
			return t
		}
		// Try matching with builtin_ prefix if not already present
		if !strings.HasPrefix(name, "builtin_") && toolName == "builtin_"+name {
			return t
		}
	}

	// Search in MCP toolsets
	// Note: MCP toolsets require context, but we pass nil for tool enumeration
	for _, toolset := range m.mcpToolsets {
		mcpTools, err := toolset.Tools(nil)
		if err != nil {
			// Log error but continue with other toolsets
			fmt.Fprintf(os.Stderr, "Warning: Failed to get tools from MCP toolset: %v\n", err)
			continue
		}
		for _, t := range mcpTools {
			if t.Name() == name {
				return t
			}
		}
	}

	return nil
}

// extractToolsFromYAML parses the 'tools' field from YAML frontmatter
func (m *SubAgentManager) extractToolsFromYAML(yamlContent string) string {
	// Simple YAML parsing for 'tools:' field
	lines := strings.Split(yamlContent, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "tools:") {
			// Extract value after 'tools:'
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// splitAndTrim splits a comma-separated string and trims whitespace
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// InitSubAgentTools is a convenience function to load and return subagent tools
// This should be called during application initialization
func InitSubAgentTools(ctx context.Context, projectRoot string, modelLLM model.LLM, modelConfig models.Config) ([]tool.Tool, error) {
	return InitSubAgentToolsWithMCP(ctx, projectRoot, modelLLM, modelConfig, nil)
}

// InitSubAgentToolsWithMCP is a convenience function that includes MCP toolsets
// This should be called during application initialization when MCP is enabled
func InitSubAgentToolsWithMCP(ctx context.Context, projectRoot string, modelLLM model.LLM, modelConfig models.Config, mcpToolsets []tool.Toolset) ([]tool.Tool, error) {
	if projectRoot == "" {
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	manager := NewSubAgentManagerWithMCP(projectRoot, modelLLM, modelConfig, mcpToolsets)
	return manager.LoadSubAgentTools(ctx)
}
