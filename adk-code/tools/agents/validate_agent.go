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

// ValidateAgentInput defines input for the validate_agent tool.
type ValidateAgentInput struct {
	// AgentName is the agent to validate
	AgentName string `json:"agent_name" jsonschema:"Name of the agent to validate (required)"`

	// CheckDependencies validates dependency relationships
	CheckDependencies bool `json:"check_dependencies,omitempty" jsonschema:"Validate dependency relationships"`

	// CheckVersions validates version constraints
	CheckVersions bool `json:"check_versions,omitempty" jsonschema:"Validate version constraints"`

	// CheckExecutionRequirements validates execution requirements
	CheckExecutionRequirements bool `json:"check_execution_requirements,omitempty" jsonschema:"Validate execution requirements"`

	// Detailed includes detailed validation information
	Detailed bool `json:"detailed,omitempty" jsonschema:"Include detailed validation information"`
}

// ValidateAgentOutput defines output for the validate_agent tool.
type ValidateAgentOutput struct {
	// AgentName is the validated agent
	AgentName string `json:"agent_name"`

	// Valid indicates if agent passes all checks
	Valid bool `json:"valid"`

	// DependencyValidation results
	DependencyValidation *ValidationResult `json:"dependency_validation,omitempty"`

	// VersionValidation results
	VersionValidation *ValidationResult `json:"version_validation,omitempty"`

	// ExecutionValidation results
	ExecutionValidation *CompatibilityResult `json:"execution_validation,omitempty"`

	// Issues contains all validation issues found
	Issues []ValidationIssue `json:"issues"`

	// Warnings contains advisories
	Warnings []string `json:"warnings,omitempty"`

	// Summary is a human-readable validation summary
	Summary string `json:"summary"`

	// Error is any error that occurred
	Error string `json:"error,omitempty"`
}

// ValidationResult contains results of a specific validation check.
type ValidationResult struct {
	Name    string   `json:"name"`
	Valid   bool     `json:"valid"`
	Details []string `json:"details,omitempty"`
}

// CompatibilityResult contains compatibility check results.
type CompatibilityResult struct {
	Compatible bool     `json:"compatible"`
	Issues     []string `json:"issues,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	Details    []string `json:"details,omitempty"`
}

// ValidationIssue represents a single validation issue.
type ValidationIssue struct {
	Category string `json:"category"` // dependency, version, execution, metadata
	Message  string `json:"message"`
	Severity string `json:"severity"` // error, warning
}

// NewValidateAgentTool creates a new validate_agent tool.
func NewValidateAgentTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ValidateAgentInput) ValidateAgentOutput {
		output := ValidateAgentOutput{
			AgentName: input.AgentName,
			Valid:     true,
			Issues:    []ValidationIssue{},
			Warnings:  []string{},
		}

		// Validate input
		if input.AgentName == "" {
			output.Error = "agent_name is required"
			output.Valid = false
			return output
		}

		// Default to checking all if none specified
		if !input.CheckDependencies && !input.CheckVersions && !input.CheckExecutionRequirements {
			input.CheckDependencies = true
			input.CheckVersions = true
			input.CheckExecutionRequirements = true
		}

		// Discover all agents
		discoverer := agents.NewDiscoverer(".")
		result, err := discoverer.DiscoverAll()
		if err != nil {
			output.Error = fmt.Sprintf("discovery failed: %v", err)
			output.Valid = false
			return output
		}

		// Find the agent to validate
		var targetAgent *agents.Agent
		for _, agent := range result.Agents {
			if agent.Name == input.AgentName {
				targetAgent = agent
				break
			}
		}

		if targetAgent == nil {
			output.Error = fmt.Sprintf("agent %q not found", input.AgentName)
			output.Valid = false
			return output
		}

		// Build dependency graph
		graph, err := agents.BuildGraphFromDiscovery(result)
		if err != nil {
			output.Error = fmt.Sprintf("failed to build dependency graph: %v", err)
			output.Valid = false
			return output
		}

		// Validate dependencies
		if input.CheckDependencies {
			output.DependencyValidation = validateDependencies(graph, targetAgent, input.Detailed)
			if !output.DependencyValidation.Valid {
				output.Valid = false
				for _, detail := range output.DependencyValidation.Details {
					output.Issues = append(output.Issues, ValidationIssue{
						Category: "dependency",
						Message:  detail,
						Severity: "error",
					})
				}
			}
		}

		// Validate versions
		if input.CheckVersions {
			output.VersionValidation = validateVersions(graph, targetAgent, input.Detailed)
			if !output.VersionValidation.Valid {
				output.Valid = false
				for _, detail := range output.VersionValidation.Details {
					output.Issues = append(output.Issues, ValidationIssue{
						Category: "version",
						Message:  detail,
						Severity: "error",
					})
				}
			}
		}

		// Validate execution requirements
		if input.CheckExecutionRequirements {
			output.ExecutionValidation = validateExecution(targetAgent, input.Detailed)
			if !output.ExecutionValidation.Compatible {
				output.Valid = false
				for _, issue := range output.ExecutionValidation.Issues {
					output.Issues = append(output.Issues, ValidationIssue{
						Category: "execution",
						Message:  issue,
						Severity: "error",
					})
				}
			}
			for _, warning := range output.ExecutionValidation.Warnings {
				output.Warnings = append(output.Warnings, warning)
			}
		}

		// Generate summary
		output.Summary = generateValidationSummary(input.AgentName, output.Valid, len(output.Issues), len(output.Warnings))

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "validate_agent",
		Description: "Validate an agent definition comprehensively. Checks dependencies, versions, execution requirements, and metadata for consistency and correctness.",
	}, handler)

	if err != nil {
		return nil, fmt.Errorf("failed to create validate_agent tool: %w", err)
	}

	// Register the tool
	common.Register(common.ToolMetadata{
		Tool:      t,
		Category:  common.CategorySearchDiscovery,
		Priority:  8,
		UsageHint: "Validate agent definitions for correctness and completeness",
	})

	return t, nil
}

// validateDependencies checks dependency relationships.
func validateDependencies(graph *agents.DependencyGraph, agent *agents.Agent, detailed bool) *ValidationResult {
	result := &ValidationResult{
		Name:    "Dependencies",
		Valid:   true,
		Details: []string{},
	}

	// Check for cycles
	if err := validateAgentDeps(graph, agent.Name); err != nil {
		result.Valid = false
		result.Details = append(result.Details, fmt.Sprintf("Dependency validation failed: %v", err))
		return result
	}

	// Check if all dependencies exist
	for _, depName := range agent.Dependencies {
		if _, exists := graph.Agents[depName]; !exists {
			result.Valid = false
			result.Details = append(result.Details, fmt.Sprintf("Dependency %q not found in discovered agents", depName))
		}
	}

	if detailed && result.Valid {
		resolved, _ := graph.ResolveDependencies(agent.Name)
		result.Details = append(result.Details, fmt.Sprintf("Dependency chain: %d agents in execution order", len(resolved)))
	}

	return result
}

// validateAgentDeps checks if agent dependencies are valid.
func validateAgentDeps(graph *agents.DependencyGraph, agentName string) error {
	if _, exists := graph.Agents[agentName]; !exists {
		return fmt.Errorf("agent %q not found", agentName)
	}

	// Use ResolveDependencies to check for cycles
	// If there's a cycle, it will error
	_, err := graph.ResolveDependencies(agentName)
	return err
}

// validateVersions checks version constraints.
func validateVersions(graph *agents.DependencyGraph, agent *agents.Agent, detailed bool) *ValidationResult {
	result := &ValidationResult{
		Name:    "Versions",
		Valid:   true,
		Details: []string{},
	}

	// Check if agent has version
	if agent.Version == "" {
		result.Details = append(result.Details, "Agent has no version specified")
		// This is not an error, just informational
		if detailed {
			result.Details = append(result.Details, "Consider adding version field to agent metadata")
		}
		return result
	}

	// Parse version
	version, err := agents.ParseVersion(agent.Version)
	if err != nil {
		result.Valid = false
		result.Details = append(result.Details, fmt.Sprintf("Invalid version format: %v", err))
		return result
	}

	if detailed {
		result.Details = append(result.Details, fmt.Sprintf("Version: %s", version.String()))
	}

	return result
}

// validateExecution checks execution compatibility.
func validateExecution(agent *agents.Agent, detailed bool) *CompatibilityResult {
	result := &CompatibilityResult{
		Compatible: true,
		Issues:     []string{},
		Warnings:   []string{},
	}

	// Check for required fields
	if agent.Name == "" {
		result.Compatible = false
		result.Issues = append(result.Issues, "Agent name is required")
	}

	if agent.Description == "" {
		result.Compatible = false
		result.Issues = append(result.Issues, "Agent description is required")
	}

	// Check if agent can be discovered
	if agent.Path == "" {
		result.Warnings = append(result.Warnings, "Agent has no file path")
	}

	// Check metadata completeness
	if agent.Author == "" {
		result.Warnings = append(result.Warnings, "Agent author is not specified")
	}

	if len(agent.Tags) == 0 {
		result.Warnings = append(result.Warnings, "Agent has no tags for categorization")
	}

	if detailed {
		if agent.Type != "" {
			result.Details = append(result.Details, fmt.Sprintf("Type: %s", agent.Type))
		}
		if agent.Source != "" {
			result.Details = append(result.Details, fmt.Sprintf("Source: %s", agent.Source))
		}
		if agent.Version != "" {
			result.Details = append(result.Details, fmt.Sprintf("Version: %s", agent.Version))
		}
	}

	return result
}

// generateValidationSummary creates a human-readable validation summary.
func generateValidationSummary(agentName string, valid bool, issueCount, warningCount int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Validation Report for %q\n", agentName))
	sb.WriteString("=" + strings.Repeat("=", len(agentName)+20) + "\n\n")

	if valid {
		sb.WriteString("✓ VALID: Agent passes all validation checks\n")
	} else {
		sb.WriteString("✗ INVALID: Agent has validation errors\n")
	}

	if issueCount > 0 {
		sb.WriteString(fmt.Sprintf("\nIssues: %d\n", issueCount))
	}

	if warningCount > 0 {
		sb.WriteString(fmt.Sprintf("Warnings: %d\n", warningCount))
	}

	return sb.String()
}

func init() {
	// Register the tool
	if _, err := NewValidateAgentTool(); err != nil {
		// Log error if needed
		_ = err
	}
}
