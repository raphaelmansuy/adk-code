// Package agents provides tools for agent definition discovery and management
package agents

import (
	"fmt"
	"sort"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/pkg/agents"
	common "adk-code/tools/base"
)

// ListAgentsInput defines input parameters for listing agents
type ListAgentsInput struct {
	AgentType string `json:"agent_type,omitempty" jsonschema:"Filter by agent type (subagent, skill, command, plugin)"`
	Source    string `json:"source,omitempty" jsonschema:"Filter by agent source (project, user, plugin, cli)"`
	Detailed  bool   `json:"detailed,omitempty" jsonschema:"Include detailed metadata for each agent"`
}

// AgentEntry represents a single agent in output
type AgentEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Path        string `json:"path,omitempty"`
	Modified    string `json:"modified,omitempty"`
}

// ListAgentsOutput defines the output of listing agents
type ListAgentsOutput struct {
	Agents     []AgentEntry `json:"agents"`
	Count      int          `json:"count"`
	Success    bool         `json:"success"`
	Error      string       `json:"error,omitempty"`
	Summary    string       `json:"summary"`
	ErrorCount int          `json:"error_count,omitempty"`
}

// NewListAgentsTool creates a tool for listing discovered agents
// Uses current working directory as project root for discovery
// Automatically loads configuration from .adk/config.yaml if present
func NewListAgentsTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ListAgentsInput) ListAgentsOutput {
		// Use current working directory as project root
		projectRoot := "."

		// Load configuration (Phase 1 feature)
		cfg, err := agents.LoadConfig(projectRoot)
		if err != nil {
			// Fall back to default config if loading fails
			cfg = agents.NewConfig()
			cfg.ProjectPath = ".adk/agents"
		}

		// Create discoverer with configuration
		discoverer := agents.NewDiscovererWithConfig(projectRoot, cfg)
		result, err := discoverer.DiscoverAll()

		output := ListAgentsOutput{
			Agents:     make([]AgentEntry, 0),
			ErrorCount: result.ErrorCount,
		}

		if err != nil {
			output.Success = false
			output.Error = fmt.Sprintf("Discovery failed: %v", err)
			output.Summary = output.Error
			return output
		}

		// Filter agents
		filtered := filterAgents(result.Agents, input)

		// Convert to output format
		for _, agent := range filtered {
			entry := AgentEntry{
				Name:        agent.Name,
				Description: agent.Description,
				Type:        agent.Type.String(),
				Source:      agent.Source.String(),
			}

			if input.Detailed {
				entry.Path = agent.Path
				entry.Modified = agent.ModTime.Format("2006-01-02 15:04:05")
			}

			output.Agents = append(output.Agents, entry)
		}

		// Sort by name
		sort.Slice(output.Agents, func(i, j int) bool {
			return output.Agents[i].Name < output.Agents[j].Name
		})

		output.Count = len(output.Agents)
		output.Success = true
		output.Summary = formatSummary(output.Count, input, result.ErrorCount)

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "list_agents",
		Description: "Discovers and lists agent definitions in the project. Supports filtering by type and source.",
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategorySearchDiscovery,
			Priority:  8,
			UsageHint: "Discover and list available agents in the project",
		})
	}

	return t, err
}

// filterAgents applies filters to a list of agents
func filterAgents(agentList []*agents.Agent, input ListAgentsInput) []*agents.Agent {
	var filtered []*agents.Agent

	for _, agent := range agentList {
		if input.AgentType != "" && agent.Type.String() != input.AgentType {
			continue
		}
		if input.Source != "" && agent.Source.String() != input.Source {
			continue
		}
		filtered = append(filtered, agent)
	}

	return filtered
}

// formatSummary creates a human-readable summary
func formatSummary(count int, input ListAgentsInput, errorCount int) string {
	if count == 0 && errorCount == 0 {
		return "No agents found in the project"
	}
	if count == 0 {
		return fmt.Sprintf("No valid agents found (%d parsing error(s))", errorCount)
	}
	msg := fmt.Sprintf("Found %d agent(s)", count)
	if errorCount > 0 {
		msg = fmt.Sprintf("%s (%d error(s))", msg, errorCount)
	}
	return msg
}

// RegisterAgentTools registers all agent-related tools
func RegisterAgentTools() error {
	_, err := NewListAgentsTool()
	return err
}

// init registers tools when package is imported
func init() {
	_ = RegisterAgentTools()
}
