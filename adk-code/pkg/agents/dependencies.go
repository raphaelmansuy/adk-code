package agents

import (
	"fmt"
	"sort"
)

// DependencyGraph represents a directed graph of agent dependencies.
type DependencyGraph struct {
	// Agents maps agent name to agent struct
	Agents map[string]*Agent

	// Edges maps agent name to list of dependent agent names
	Edges map[string][]string
}

// NewDependencyGraph creates a new empty dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Agents: make(map[string]*Agent),
		Edges:  make(map[string][]string),
	}
}

// AddAgent adds an agent to the graph.
func (dg *DependencyGraph) AddAgent(agent *Agent) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}

	if agent.Name == "" {
		return fmt.Errorf("agent name is empty")
	}

	if _, exists := dg.Agents[agent.Name]; exists {
		return fmt.Errorf("agent %q already exists in graph", agent.Name)
	}

	dg.Agents[agent.Name] = agent

	// Initialize edges if not present
	if _, exists := dg.Edges[agent.Name]; !exists {
		dg.Edges[agent.Name] = []string{}
	}

	return nil
}

// AddEdge adds a dependency edge from one agent to another.
func (dg *DependencyGraph) AddEdge(from, to string) error {
	// Validate both agents exist
	if _, exists := dg.Agents[from]; !exists {
		return fmt.Errorf("agent %q not found in graph", from)
	}

	if _, exists := dg.Agents[to]; !exists {
		return fmt.Errorf("agent %q not found in graph", to)
	}

	// Check if edge already exists
	for _, dep := range dg.Edges[from] {
		if dep == to {
			return fmt.Errorf("edge from %q to %q already exists", from, to)
		}
	}

	dg.Edges[from] = append(dg.Edges[from], to)
	return nil
}

// ResolveDependencies returns a topologically sorted list of agents
// that must be executed before the given agent.
func (dg *DependencyGraph) ResolveDependencies(agentName string) ([]*Agent, error) {
	if _, exists := dg.Agents[agentName]; !exists {
		return nil, fmt.Errorf("agent %q not found in graph", agentName)
	}

	// Detect cycles first
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	if dg.hasCycleDFS(agentName, visited, recStack) {
		return nil, fmt.Errorf("circular dependency detected involving agent %q", agentName)
	}

	// Perform topological sort using DFS
	visited = make(map[string]bool)
	var sorted []*Agent

	dg.topologicalSort(agentName, visited, &sorted)

	return sorted, nil
}

// GetTransitiveDeps returns all transitive dependencies of an agent.
func (dg *DependencyGraph) GetTransitiveDeps(agentName string) ([]string, error) {
	if _, exists := dg.Agents[agentName]; !exists {
		return nil, fmt.Errorf("agent %q not found in graph", agentName)
	}

	visited := make(map[string]bool)
	var deps []string

	dg.getTransitiveDepsHelper(agentName, visited, &deps)

	// Remove duplicates and sort
	depSet := make(map[string]bool)
	var unique []string
	for _, dep := range deps {
		if !depSet[dep] {
			depSet[dep] = true
			unique = append(unique, dep)
		}
	}

	sort.Strings(unique)
	return unique, nil
}

// DetectCycles returns a list of agents involved in circular dependencies.
func (dg *DependencyGraph) DetectCycles() []string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycleAgents []string

	for agentName := range dg.Agents {
		if !visited[agentName] {
			if dg.hasCycleDFS(agentName, visited, recStack) {
				cycleAgents = append(cycleAgents, agentName)
			}
		}
	}

	sort.Strings(cycleAgents)
	return cycleAgents
}

// GetAgent retrieves an agent by name from the graph.
func (dg *DependencyGraph) GetAgent(name string) *Agent {
	return dg.Agents[name]
}

// AgentCount returns the number of agents in the graph.
func (dg *DependencyGraph) AgentCount() int {
	return len(dg.Agents)
}

// EdgeCount returns the number of dependency edges in the graph.
func (dg *DependencyGraph) EdgeCount() int {
	count := 0
	for _, deps := range dg.Edges {
		count += len(deps)
	}
	return count
}

// GetAllAgents returns all agents in the graph.
func (dg *DependencyGraph) GetAllAgents() []*Agent {
	agents := make([]*Agent, 0, len(dg.Agents))
	for _, agent := range dg.Agents {
		agents = append(agents, agent)
	}

	// Sort by name for consistent output
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Name < agents[j].Name
	})

	return agents
}

// hasCycleDFS detects if there's a cycle in the graph using DFS.
func (dg *DependencyGraph) hasCycleDFS(agentName string, visited, recStack map[string]bool) bool {
	visited[agentName] = true
	recStack[agentName] = true

	// Check all dependencies of this agent
	for _, dep := range dg.Edges[agentName] {
		if !visited[dep] {
			if dg.hasCycleDFS(dep, visited, recStack) {
				return true
			}
		} else if recStack[dep] {
			// Back edge found - cycle detected
			return true
		}
	}

	recStack[agentName] = false
	return false
}

// topologicalSort performs DFS-based topological sort.
func (dg *DependencyGraph) topologicalSort(agentName string, visited map[string]bool, sorted *[]*Agent) {
	visited[agentName] = true

	// Visit all dependencies first
	for _, dep := range dg.Edges[agentName] {
		if !visited[dep] {
			dg.topologicalSort(dep, visited, sorted)
		}
	}

	// Add agent after visiting dependencies
	*sorted = append(*sorted, dg.Agents[agentName])
}

// getTransitiveDepsHelper recursively collects all transitive dependencies.
func (dg *DependencyGraph) getTransitiveDepsHelper(agentName string, visited map[string]bool, deps *[]string) {
	if visited[agentName] {
		return
	}
	visited[agentName] = true

	// Add all direct dependencies
	for _, dep := range dg.Edges[agentName] {
		*deps = append(*deps, dep)
		// Recursively add their dependencies
		dg.getTransitiveDepsHelper(dep, visited, deps)
	}
}

// ResolveConflicts checks if dependencies have conflicting versions.
// Returns error if conflicts detected.
func (dg *DependencyGraph) ResolveConflicts() error {
	// This is a placeholder for future version constraint checking
	// Will be implemented with version.go integration
	return nil
}

// String returns a string representation of the graph.
func (dg *DependencyGraph) String() string {
	var str string
	str += fmt.Sprintf("DependencyGraph: %d agents, %d edges\n", dg.AgentCount(), dg.EdgeCount())

	agents := dg.GetAllAgents()
	for _, agent := range agents {
		str += fmt.Sprintf("  %s: depends on [", agent.Name)
		deps := dg.Edges[agent.Name]
		for i, dep := range deps {
			if i > 0 {
				str += ", "
			}
			str += dep
		}
		str += "]\n"
	}

	return str
}
