// Package agents provides tools for agent definition discovery and management
package agents

import (
	"fmt"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/pkg/agents"
	common "adk-code/tools/base"
)

// ResolveDepInput defines input for the resolve_deps tool.
type ResolveDepInput struct {
	// AgentName is the agent to resolve dependencies for
	AgentName string `json:"agent_name" jsonschema:"Name of the agent to resolve dependencies for (required)"`

	// ShowTransitive includes transitive dependencies
	ShowTransitive bool `json:"show_transitive,omitempty" jsonschema:"Include transitive dependencies"`

	// CheckVersions performs version constraint validation
	CheckVersions bool `json:"check_versions,omitempty" jsonschema:"Validate version constraints"`

	// Format output format: "list", "tree", or "json"
	Format string `json:"format,omitempty" jsonschema:"Output format: list, tree, or json"`
}

// ResolveDepOutput defines output for the resolve_deps tool.
type ResolveDepOutput struct {
	// AgentName is the requested agent
	AgentName string `json:"agent_name"`

	// Dependencies are the resolved dependencies in execution order
	Dependencies []DependencyInfo `json:"dependencies"`

	// TransitiveDependencies includes all transitive dependencies
	TransitiveDependencies []string `json:"transitive_dependencies,omitempty"`

	// VersionIssues contains any version constraint violations
	VersionIssues []string `json:"version_issues,omitempty"`

	// Summary is a human-readable summary
	Summary string `json:"summary"`

	// Error is any error that occurred
	Error string `json:"error,omitempty"`
}

// DependencyInfo contains information about a single dependency.
type DependencyInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Order   int    `json:"order"`
}

// NewResolveDependenciesTool creates a new resolve_deps tool.
func NewResolveDependenciesTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ResolveDepInput) ResolveDepOutput {
		output := ResolveDepOutput{
			AgentName:     input.AgentName,
			Dependencies:  []DependencyInfo{},
			VersionIssues: []string{},
		}

		// Validate input
		if input.AgentName == "" {
			output.Error = "agent_name is required"
			return output
		}

		if input.Format == "" {
			input.Format = "list"
		}

		// Create discoverer
		discoverer := agents.NewDiscoverer(".")
		result, err := discoverer.DiscoverAll()
		if err != nil {
			output.Error = fmt.Sprintf("discovery failed: %v", err)
			return output
		}

		// Build dependency graph
		graph, err := agents.BuildGraphFromDiscovery(result)
		if err != nil {
			output.Error = fmt.Sprintf("failed to build graph: %v", err)
			return output
		}

		// Resolve dependencies
		resolved, err := graph.ResolveDependencies(input.AgentName)
		if err != nil {
			output.Error = fmt.Sprintf("dependency resolution failed: %v", err)
			return output
		}

		// Build dependency info
		for i, dep := range resolved {
			info := DependencyInfo{
				Name:    dep.Name,
				Version: dep.Version,
				Order:   i + 1,
			}
			output.Dependencies = append(output.Dependencies, info)
		}

		// Get transitive dependencies if requested
		if input.ShowTransitive {
			transDeps, err := graph.GetTransitiveDeps(input.AgentName)
			if err == nil {
				output.TransitiveDependencies = transDeps
			}
		}

		// Check versions if requested
		if input.CheckVersions {
			validator := agents.NewAgentMetadataValidator()

			// Add all agents to validator
			for _, agent := range result.Agents {
				validator.AddAgent(agent, "")
			}

			// Add dependency edges
			for _, agent := range result.Agents {
				for _, depName := range agent.Dependencies {
					if _, exists := validator.Graph.Agents[depName]; exists {
						validator.AddDependency(agent.Name, depName)
					}
				}
			}

			// Validate
			report, err := validator.ValidateAgent(input.AgentName)
			if err == nil && !report.Valid {
				output.VersionIssues = report.Issues
			}
		}

		// Format output summary
		output.Summary = formatDependencySummary(input.Format, input.AgentName, output.Dependencies)

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "resolve_deps",
		Description: "Resolve agent dependencies and display dependency graphs. Shows which agents must be executed before a target agent, validates version constraints, and detects circular dependencies.",
	}, handler)

	if err != nil {
		return nil, fmt.Errorf("failed to create resolve_deps tool: %w", err)
	}

	// Register the tool
	common.Register(common.ToolMetadata{
		Tool:      t,
		Category:  common.CategorySearchDiscovery,
		Priority:  7,
		UsageHint: "Analyze agent dependencies and resolve execution order",
	})

	return t, nil
}

// formatDependencySummary formats dependency information based on format type.
func formatDependencySummary(format, agentName string, deps []DependencyInfo) string {
	switch format {
	case "tree":
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s\n", agentName))
		for _, dep := range deps {
			if dep.Name != agentName {
				sb.WriteString(fmt.Sprintf("  └─ %s", dep.Name))
				if dep.Version != "" {
					sb.WriteString(fmt.Sprintf(" (v%s)", dep.Version))
				}
				sb.WriteString("\n")
			}
		}
		return sb.String()

	case "json":
		// JSON is already handled by output marshaling
		return fmt.Sprintf("Resolved %d dependencies for agent %q", len(deps)-1, agentName)

	default: // list
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Dependencies for %s (execution order):\n", agentName))
		for i, dep := range deps {
			if dep.Name != agentName {
				sb.WriteString(fmt.Sprintf("%d. %s", i, dep.Name))
				if dep.Version != "" {
					sb.WriteString(fmt.Sprintf(" v%s", dep.Version))
				}
				sb.WriteString("\n")
			}
		}
		return sb.String()
	}
}

func init() {
	// Register the tool
	if _, err := NewResolveDependenciesTool(); err != nil {
		// Log error if needed
		_ = err
	}
}
