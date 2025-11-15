// Package agents provides subagent delegation using ADK's native agent-as-tool pattern
package agents

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"

	"adk-code/pkg/agents"
	common "adk-code/tools/base"
)

// SubAgentManager creates and manages subagent tools
// This uses Google ADK's native agent-as-tool pattern via agenttool.New()
type SubAgentManager struct {
	projectRoot string
	modelLLM    model.LLM
}

// NewSubAgentManager creates a new subagent manager
func NewSubAgentManager(projectRoot string, modelLLM model.LLM) *SubAgentManager {
	return &SubAgentManager{
		projectRoot: projectRoot,
		modelLLM:    modelLLM,
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

// createSubAgent creates an llmagent from an agent definition
func (m *SubAgentManager) createSubAgent(agentDef *agents.Agent) (agent.Agent, error) {
	// Create the subagent using ADK's llmagent
	// The subagent gets its own isolated context and uses the agent's content as instruction
	subAgent, err := llmagent.New(llmagent.Config{
		Name:        agentDef.Name,
		Description: agentDef.Description,
		Model:       m.modelLLM,
		Instruction: agentDef.Content, // The markdown content is the system instruction
		Tools:       []tool.Tool{},     // Phase 1: Subagents have no tools (analysis only)
		// Future: Parse allowed tools from agent definition and provide restricted toolset
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create llmagent: %w", err)
	}

	return subAgent, nil
}

// InitSubAgentTools is a convenience function to load and return subagent tools
// This should be called during application initialization
func InitSubAgentTools(ctx context.Context, projectRoot string, modelLLM model.LLM) ([]tool.Tool, error) {
	if projectRoot == "" {
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	manager := NewSubAgentManager(projectRoot, modelLLM)
	return manager.LoadSubAgentTools(ctx)
}
