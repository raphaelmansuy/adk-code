package agents

import (
	"fmt"
	"strings"
)

// AgentMetadataValidator validates agent metadata including version constraints
// and dependency requirements.
type AgentMetadataValidator struct {
	// Graph is the dependency graph
	Graph *DependencyGraph

	// AgentVersions maps agent name to its version
	AgentVersions map[string]*Version

	// Constraints maps agent name to version constraint
	Constraints map[string]*Constraint
}

// NewAgentMetadataValidator creates a new metadata validator.
func NewAgentMetadataValidator() *AgentMetadataValidator {
	return &AgentMetadataValidator{
		Graph:         NewDependencyGraph(),
		AgentVersions: make(map[string]*Version),
		Constraints:   make(map[string]*Constraint),
	}
}

// AddAgent adds an agent to the validator with optional version and constraint.
func (v *AgentMetadataValidator) AddAgent(agent *Agent, constraint string) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}

	// Add to graph
	if err := v.Graph.AddAgent(agent); err != nil {
		return err
	}

	// Parse and store version if present
	if agent.Version != "" {
		version, err := ParseVersion(agent.Version)
		if err != nil {
			return fmt.Errorf("invalid version %q for agent %q: %v", agent.Version, agent.Name, err)
		}
		v.AgentVersions[agent.Name] = version
	}

	// Parse and store constraint if provided
	if constraint != "" {
		c, err := ParseConstraint(constraint)
		if err != nil {
			return fmt.Errorf("invalid constraint %q for agent %q: %v", constraint, agent.Name, err)
		}
		v.Constraints[agent.Name] = c
	}

	return nil
}

// AddDependency adds a dependency relationship between agents.
func (v *AgentMetadataValidator) AddDependency(fromAgent, toAgent string) error {
	if err := v.Graph.AddEdge(fromAgent, toAgent); err != nil {
		return err
	}
	return nil
}

// ValidateDependencies checks that an agent's dependencies can be resolved
// and don't have circular references.
func (v *AgentMetadataValidator) ValidateDependencies(agentName string) error {
	if _, exists := v.Graph.Agents[agentName]; !exists {
		return fmt.Errorf("agent %q not found", agentName)
	}

	// Check for cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	if v.Graph.hasCycleDFS(agentName, visited, recStack) {
		return fmt.Errorf("circular dependency detected for agent %q", agentName)
	}

	// Verify all dependencies exist
	for _, depName := range v.Graph.Edges[agentName] {
		if _, exists := v.Graph.Agents[depName]; !exists {
			return fmt.Errorf("agent %q depends on non-existent agent %q", agentName, depName)
		}
	}

	return nil
}

// ValidateVersionConstraints checks that all agents satisfy their version constraints.
func (v *AgentMetadataValidator) ValidateVersionConstraints() error {
	for agentName, constraint := range v.Constraints {
		version, exists := v.AgentVersions[agentName]
		if !exists {
			// Agent has no version, can't validate constraint
			continue
		}

		if !constraint.Matches(version) {
			return fmt.Errorf("agent %q version %q does not satisfy constraint %q",
				agentName, version.String(), constraint.String())
		}
	}

	return nil
}

// ValidateDependencyVersions checks version constraints on all dependencies.
func (v *AgentMetadataValidator) ValidateDependencyVersions(agentName string) error {
	if _, exists := v.Graph.Agents[agentName]; !exists {
		return fmt.Errorf("agent %q not found", agentName)
	}

	// Get all transitive dependencies
	transitiveDeps, err := v.Graph.GetTransitiveDeps(agentName)
	if err != nil {
		return fmt.Errorf("failed to get transitive dependencies for %q: %v", agentName, err)
	}

	// Check version constraints on each dependency
	for _, depName := range transitiveDeps {
		depVersion, hasVersion := v.AgentVersions[depName]
		depConstraint, hasConstraint := v.Constraints[depName]

		if hasVersion && hasConstraint {
			if !depConstraint.Matches(depVersion) {
				return fmt.Errorf("dependency %q version %q does not satisfy constraint %q",
					depName, depVersion.String(), depConstraint.String())
			}
		}
	}

	return nil
}

// ValidateAgent performs comprehensive validation of an agent.
// This includes dependency resolution, version constraints, and circular references.
func (v *AgentMetadataValidator) ValidateAgent(agentName string) (*ValidationReport, error) {
	report := &ValidationReport{
		AgentName: agentName,
		Valid:     true,
		Issues:    []string{},
	}

	if _, exists := v.Graph.Agents[agentName]; !exists {
		return nil, fmt.Errorf("agent %q not found", agentName)
	}

	// Check dependencies exist and are not circular
	if err := v.ValidateDependencies(agentName); err != nil {
		report.Valid = false
		report.Issues = append(report.Issues, fmt.Sprintf("Dependency validation failed: %v", err))
	}

	// Check version constraints
	if constraint, exists := v.Constraints[agentName]; exists {
		if version, hasVersion := v.AgentVersions[agentName]; hasVersion {
			if !constraint.Matches(version) {
				report.Valid = false
				report.Issues = append(report.Issues, fmt.Sprintf("Version constraint violation: %q does not satisfy %q",
					version.String(), constraint.String()))
			}
		}
	}

	// Check dependency version constraints
	if err := v.ValidateDependencyVersions(agentName); err != nil {
		report.Valid = false
		report.Issues = append(report.Issues, fmt.Sprintf("Dependency version validation failed: %v", err))
	}

	// Get resolved dependencies
	resolved, err := v.Graph.ResolveDependencies(agentName)
	if err == nil {
		report.ResolvedDependencies = make([]string, 0, len(resolved))
		for _, dep := range resolved {
			report.ResolvedDependencies = append(report.ResolvedDependencies, dep.Name)
		}
	}

	return report, nil
}

// ValidationReport contains results of agent validation.
type ValidationReport struct {
	// AgentName is the name of the validated agent
	AgentName string

	// Valid indicates if validation passed
	Valid bool

	// Issues contains validation issues found
	Issues []string

	// ResolvedDependencies contains the ordered dependency chain
	ResolvedDependencies []string
}

// String returns a string representation of the validation report.
func (r *ValidationReport) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Agent: %s\n", r.AgentName))
	sb.WriteString(fmt.Sprintf("Valid: %v\n", r.Valid))

	if len(r.Issues) > 0 {
		sb.WriteString("Issues:\n")
		for _, issue := range r.Issues {
			sb.WriteString(fmt.Sprintf("  - %s\n", issue))
		}
	}

	if len(r.ResolvedDependencies) > 0 {
		sb.WriteString("Dependencies (in execution order):\n")
		for i, dep := range r.ResolvedDependencies {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, dep))
		}
	}

	return sb.String()
}

// BuildGraphFromDiscovery builds a dependency graph from discovered agents.
func BuildGraphFromDiscovery(result *DiscoveryResult) (*DependencyGraph, error) {
	if result == nil {
		return nil, fmt.Errorf("discovery result is nil")
	}

	graph := NewDependencyGraph()

	// Add all agents to the graph
	for _, agent := range result.Agents {
		if err := graph.AddAgent(agent); err != nil {
			return nil, fmt.Errorf("failed to add agent %q: %v", agent.Name, err)
		}
	}

	// Add dependency edges
	for _, agent := range result.Agents {
		for _, depName := range agent.Dependencies {
			// Silently skip dependencies to non-existent agents (may be from plugins)
			if _, exists := graph.Agents[depName]; exists {
				if err := graph.AddEdge(agent.Name, depName); err != nil {
					// Log but continue - edge might already exist
					continue
				}
			}
		}
	}

	return graph, nil
}

// ResolveAgentDependencies resolves the dependency chain for an agent.
// Returns agents in execution order (dependencies first).
func ResolveAgentDependencies(discoverer *Discoverer, agentName string) ([]*Agent, error) {
	if discoverer == nil {
		return nil, fmt.Errorf("discoverer is nil")
	}

	// Discover all agents
	result, err := discoverer.DiscoverAll()
	if err != nil {
		return nil, fmt.Errorf("failed to discover agents: %v", err)
	}

	// Build dependency graph
	graph, err := BuildGraphFromDiscovery(result)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %v", err)
	}

	// Resolve dependencies for the target agent
	resolved, err := graph.ResolveDependencies(agentName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies for %q: %v", agentName, err)
	}

	return resolved, nil
}

// Package-level constants and helpers

var (
	// ErrNilAgent is returned when agent is nil
	ErrNilAgent = fmt.Errorf("agent is nil")

	// ErrNilGraph is returned when graph is nil
	ErrNilGraph = fmt.Errorf("graph is nil")

	// ErrInvalidConstraint is returned when version constraint is invalid
	ErrInvalidConstraint = fmt.Errorf("invalid version constraint")
)

// GetAgentMetadata extracts metadata from an agent for validation.
func GetAgentMetadata(agent *Agent) map[string]interface{} {
	if agent == nil {
		return nil
	}

	return map[string]interface{}{
		"name":         agent.Name,
		"version":      agent.Version,
		"dependencies": agent.Dependencies,
		"author":       agent.Author,
		"tags":         agent.Tags,
		"type":         agent.Type,
		"source":       agent.Source,
	}
}
