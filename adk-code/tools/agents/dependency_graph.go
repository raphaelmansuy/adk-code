// Package agents provides tools for agent definition discovery and management
package agents

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/pkg/agents"
	common "adk-code/tools/base"
)

// DependencyGraphInput defines input for the dependency_graph tool.
type DependencyGraphInput struct {
	// Format is the output format (graphviz, json, text)
	Format string `json:"format,omitempty" jsonschema:"Output format: 'graphviz', 'json', or 'text' (default: 'text')"`

	// MaxDepth limits the depth of the dependency tree
	MaxDepth int `json:"max_depth,omitempty" jsonschema:"Maximum depth to traverse (default: unlimited)"`

	// IncludeVersions includes version information in the graph
	IncludeVersions bool `json:"include_versions,omitempty" jsonschema:"Include version information"`

	// HighlightCycles highlights any circular dependencies found
	HighlightCycles bool `json:"highlight_cycles,omitempty" jsonschema:"Highlight circular dependencies"`
}

// DependencyGraphOutput defines output for the dependency_graph tool.
type DependencyGraphOutput struct {
	// Format is the output format used
	Format string `json:"format"`

	// GraphData contains the graph representation
	GraphData string `json:"graph_data"`

	// JSONData contains structured graph data
	JSONData GraphDataJSON `json:"json_data,omitempty"`

	// Summary contains graph statistics
	Summary GraphSummary `json:"summary"`

	// Cycles contains detected circular dependencies
	Cycles [][]string `json:"cycles,omitempty"`

	// Error is any error that occurred
	Error string `json:"error,omitempty"`
}

// GraphDataJSON represents the graph in JSON format.
type GraphDataJSON struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// GraphNode represents a node in the graph.
type GraphNode struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// GraphEdge represents an edge in the graph.
type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// GraphSummary contains graph statistics.
type GraphSummary struct {
	TotalNodes        int `json:"total_nodes"`
	TotalEdges        int `json:"total_edges"`
	MaxDepth          int `json:"max_depth"`
	CircularDepCount  int `json:"circular_dependency_count"`
	DisconnectedNodes int `json:"disconnected_nodes"`
}

// NewDependencyGraphTool creates a new dependency_graph tool.
func NewDependencyGraphTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input DependencyGraphInput) DependencyGraphOutput {
		output := DependencyGraphOutput{
			Format: input.Format,
			Cycles: [][]string{},
		}

		// Set default format
		if input.Format == "" {
			output.Format = "text"
		}

		// Discover all agents
		discoverer := agents.NewDiscoverer(".")
		result, err := discoverer.DiscoverAll()
		if err != nil {
			output.Error = fmt.Sprintf("discovery failed: %v", err)
			return output
		}

		// Build dependency graph
		graph, err := agents.BuildGraphFromDiscovery(result)
		if err != nil {
			output.Error = fmt.Sprintf("failed to build dependency graph: %v", err)
			return output
		}

		// Generate graph data in requested format
		switch output.Format {
		case "graphviz":
			output.GraphData = generateGraphvizFormat(graph, input.IncludeVersions)

		case "json":
			jsonData := generateJSONFormat(graph)
			output.JSONData = jsonData
			jsonBytes, _ := json.MarshalIndent(jsonData, "", "  ")
			output.GraphData = string(jsonBytes)

		case "text", "":
			output.GraphData = generateTextFormat(graph, input.MaxDepth, input.IncludeVersions)

		default:
			output.Error = fmt.Sprintf("unsupported format: %q", output.Format)
			return output
		}

		// Generate summary
		output.Summary = generateGraphSummary(graph)

		// Detect cycles if requested
		if input.HighlightCycles {
			output.Cycles = detectCycles(graph)
		}

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "dependency_graph",
		Description: "Generate and visualize agent dependency graphs in multiple formats (graphviz, json, text).",
	}, handler)

	if err != nil {
		return nil, fmt.Errorf("failed to create dependency_graph tool: %w", err)
	}

	// Register the tool
	common.Register(common.ToolMetadata{
		Tool:      t,
		Category:  common.CategorySearchDiscovery,
		Priority:  7,
		UsageHint: "Visualize and analyze agent dependency relationships",
	})

	return t, nil
}

// generateGraphvizFormat generates graph in Graphviz DOT format.
func generateGraphvizFormat(graph *agents.DependencyGraph, includeVersions bool) string {
	var sb strings.Builder
	sb.WriteString("digraph {\n")

	// Add nodes
	for name, agent := range graph.Agents {
		if includeVersions && agent.Version != "" {
			sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n%s\"];\n", name, name, agent.Version))
		} else {
			sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\"];\n", name, name))
		}
	}

	// Add edges
	for from, deps := range graph.Edges {
		for _, to := range deps {
			sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", from, to))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

// generateJSONFormat generates graph in JSON format.
func generateJSONFormat(graph *agents.DependencyGraph) GraphDataJSON {
	jsonData := GraphDataJSON{
		Nodes: []GraphNode{},
		Edges: []GraphEdge{},
	}

	// Add nodes
	for name, agent := range graph.Agents {
		jsonData.Nodes = append(jsonData.Nodes, GraphNode{
			ID:      name,
			Name:    name,
			Version: agent.Version,
		})
	}

	// Add edges
	for from, deps := range graph.Edges {
		for _, to := range deps {
			jsonData.Edges = append(jsonData.Edges, GraphEdge{
				From: from,
				To:   to,
			})
		}
	}

	return jsonData
}

// generateTextFormat generates graph in text tree format.
func generateTextFormat(graph *agents.DependencyGraph, maxDepth int, includeVersions bool) string {
	var sb strings.Builder

	visited := make(map[string]bool)
	for name := range graph.Agents {
		if !visited[name] {
			renderTextTree(&sb, graph, name, "", maxDepth, 0, visited, includeVersions)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// renderTextTree renders a node and its dependencies in tree format.
func renderTextTree(sb *strings.Builder, graph *agents.DependencyGraph, nodeName, prefix string, maxDepth, currentDepth int, visited map[string]bool, includeVersions bool) {
	if maxDepth > 0 && currentDepth >= maxDepth {
		return
	}

	if visited[nodeName] {
		sb.WriteString(prefix + "└── " + nodeName + " (circular)\n")
		return
	}

	visited[nodeName] = true

	version := ""
	if includeVersions {
		if agent, exists := graph.Agents[nodeName]; exists && agent.Version != "" {
			version = " (" + agent.Version + ")"
		}
	}

	sb.WriteString(prefix + "└── " + nodeName + version + "\n")

	// Render children
	if deps, exists := graph.Edges[nodeName]; exists {
		for i, dep := range deps {
			newPrefix := prefix + "    "
			if i == len(deps)-1 {
				newPrefix = prefix + "    "
			}
			renderTextTree(sb, graph, dep, newPrefix, maxDepth, currentDepth+1, visited, includeVersions)
		}
	}
}

// generateGraphSummary generates summary statistics.
func generateGraphSummary(graph *agents.DependencyGraph) GraphSummary {
	summary := GraphSummary{
		TotalNodes: len(graph.Agents),
		TotalEdges: 0,
	}

	// Count edges
	for _, deps := range graph.Edges {
		summary.TotalEdges += len(deps)
	}

	// Find max depth
	maxDepth := 0
	for name := range graph.Agents {
		depth := calculateDepth(graph, name, make(map[string]bool))
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	summary.MaxDepth = maxDepth

	// Detect disconnected nodes
	visited := make(map[string]bool)
	for name := range graph.Agents {
		if !visited[name] {
			markConnected(graph, name, visited)
		}
	}
	summary.DisconnectedNodes = len(graph.Agents) - len(visited)

	return summary
}

// calculateDepth calculates the maximum dependency depth for a node.
func calculateDepth(graph *agents.DependencyGraph, nodeName string, visited map[string]bool) int {
	if visited[nodeName] {
		return 0
	}

	visited[nodeName] = true

	maxDepth := 0
	if deps, exists := graph.Edges[nodeName]; exists {
		for _, dep := range deps {
			depth := 1 + calculateDepth(graph, dep, make(map[string]bool))
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	return maxDepth
}

// markConnected marks all reachable nodes from a starting node.
func markConnected(graph *agents.DependencyGraph, nodeName string, visited map[string]bool) {
	if visited[nodeName] {
		return
	}

	visited[nodeName] = true

	if deps, exists := graph.Edges[nodeName]; exists {
		for _, dep := range deps {
			markConnected(graph, dep, visited)
		}
	}

	// Also mark nodes that depend on this node
	for from, deps := range graph.Edges {
		for _, to := range deps {
			if to == nodeName {
				markConnected(graph, from, visited)
			}
		}
	}
}

// detectCycles detects all circular dependencies in the graph.
func detectCycles(graph *agents.DependencyGraph) [][]string {
	var cycles [][]string

	for name := range graph.Agents {
		visited := make(map[string]bool)
		recStack := make(map[string]bool)
		var path []string

		if findCycle(graph, name, visited, recStack, &path) {
			if len(path) > 0 {
				cycles = append(cycles, path)
			}
		}
	}

	return cycles
}

// findCycle finds a cycle starting from a node using DFS.
func findCycle(graph *agents.DependencyGraph, nodeName string, visited, recStack map[string]bool, path *[]string) bool {
	visited[nodeName] = true
	recStack[nodeName] = true
	*path = append(*path, nodeName)

	if deps, exists := graph.Edges[nodeName]; exists {
		for _, dep := range deps {
			if !visited[dep] {
				if findCycle(graph, dep, visited, recStack, path) {
					return true
				}
			} else if recStack[dep] {
				// Found cycle
				return true
			}
		}
	}

	recStack[nodeName] = false
	*path = (*path)[:len(*path)-1]

	return false
}

func init() {
	// Register the tool
	if _, err := NewDependencyGraphTool(); err != nil {
		// Log error if needed
		_ = err
	}
}
