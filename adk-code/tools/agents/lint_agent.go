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

// LintAgentInput is the input for the lint agent tool
type LintAgentInput struct {
	// AgentName is the name of the agent to lint (required)
	AgentName string `json:"agent_name" jsonschema:"Name of the agent to lint (required)"`

	// FilePath is the path to the agent file to lint
	FilePath string `json:"file_path,omitempty" jsonschema:"Path to agent file to lint"`

	// IncludeWarnings includes warning-level issues in output
	IncludeWarnings bool `json:"include_warnings,omitempty" jsonschema:"Include warning-level lint issues (default: true)"`

	// IncludeInfo includes info-level issues in output
	IncludeInfo bool `json:"include_info,omitempty" jsonschema:"Include info-level lint issues (default: true)"`
}

// LintAgentOutput is the output of the lint agent tool
type LintAgentOutput struct {
	// Success indicates whether linting completed successfully
	Success bool `json:"success"`

	// AgentName is the name of the linted agent
	AgentName string `json:"agent_name"`

	// Passed indicates whether the agent passed all linting checks
	Passed bool `json:"passed"`

	// Summary is a short summary of the linting results
	Summary string `json:"summary"`

	// Errors is a list of error-level issues
	Errors []LintIssueOutput `json:"errors"`

	// Warnings is a list of warning-level issues
	Warnings []LintIssueOutput `json:"warnings"`

	// Info is a list of info-level issues
	Info []LintIssueOutput `json:"info"`

	// Total is the total number of issues found
	Total int `json:"total"`

	// Message is any additional message or error information
	Message string `json:"message"`
}

// LintIssueOutput is a linting issue in output format
type LintIssueOutput struct {
	Rule       string `json:"rule"`
	Message    string `json:"message"`
	Field      string `json:"field"`
	Suggestion string `json:"suggestion"`
}

// NewLintAgentTool creates a new lint agent tool
func NewLintAgentTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input LintAgentInput) LintAgentOutput {
		output := LintAgentOutput{
			Success:   false,
			Passed:    false,
			AgentName: input.AgentName,
			Errors:    []LintIssueOutput{},
			Warnings:  []LintIssueOutput{},
			Info:      []LintIssueOutput{},
		}

		// Validate input
		if input.AgentName == "" {
			output.Message = "agent_name is required"
			return output
		}

		// Set default values for flags
		includeWarnings := input.IncludeWarnings
		includeInfo := input.IncludeInfo
		if !input.IncludeWarnings && !input.IncludeInfo {
			// Default to including both
			includeWarnings = true
			includeInfo = true
		}

		// Load the agent
		var agent *agents.Agent
		if input.FilePath != "" {
			parsedAgent, err := agents.ParseAgentFile(input.FilePath)
			if err != nil {
				output.Message = fmt.Sprintf("Failed to parse agent file: %v", err)
				return output
			}
			agent = parsedAgent
		} else {
			// Try to discover the agent
			discoverer := agents.NewDiscoverer(".")
			result, err := discoverer.DiscoverAll()
			if err != nil {
				output.Message = fmt.Sprintf("Failed to discover agents: %v", err)
				return output
			}

			found := false
			for _, a := range result.Agents {
				if a.Name == input.AgentName {
					agent = a
					found = true
					break
				}
			}

			if !found {
				output.Message = fmt.Sprintf("Agent %q not found", input.AgentName)
				return output
			}
		}

		// Run linter
		linter := agents.NewLinter()
		result := linter.Lint(agent)

		output.Passed = result.Passed
		output.AgentName = result.AgentName

		// Collect issues by severity
		for _, issue := range result.Issues {
			issueOutput := LintIssueOutput{
				Rule:       issue.Rule,
				Message:    issue.Message,
				Field:      issue.Field,
				Suggestion: issue.Suggestion,
			}

			switch issue.Severity {
			case agents.SeverityError:
				output.Errors = append(output.Errors, issueOutput)
			case agents.SeverityWarning:
				if includeWarnings {
					output.Warnings = append(output.Warnings, issueOutput)
				}
			case agents.SeverityInfo:
				if includeInfo {
					output.Info = append(output.Info, issueOutput)
				}
			}
		}

		output.Total = result.ErrorCount + result.WarningCount + result.InfoCount
		output.Success = true

		// Build summary
		parts := []string{}
		if result.ErrorCount > 0 {
			parts = append(parts, fmt.Sprintf("%d error(s)", result.ErrorCount))
		}
		if result.WarningCount > 0 {
			parts = append(parts, fmt.Sprintf("%d warning(s)", result.WarningCount))
		}
		if result.InfoCount > 0 {
			parts = append(parts, fmt.Sprintf("%d info item(s)", result.InfoCount))
		}

		if len(parts) == 0 {
			output.Summary = "No linting issues found - agent passed all checks"
		} else {
			status := "Agent has issues:"
			if !result.Passed {
				status = "Agent failed validation:"
			}
			output.Summary = status + " " + strings.Join(parts, ", ")
		}

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "lint_agent",
		Description: "Lint an agent definition for best practices, naming conventions, and potential issues. Checks description quality, version format, author information, and dependency correctness.",
	}, handler)

	if err != nil {
		return nil, fmt.Errorf("failed to create lint_agent tool: %w", err)
	}

	// Register the tool
	common.Register(common.ToolMetadata{
		Tool:      t,
		Category:  common.CategorySearchDiscovery,
		Priority:  9,
		UsageHint: "Lint agent definitions for best practices and completeness",
	})

	return t, nil
}

// init registers the lint agent tool
func init() {
	// Tool is registered in NewLintAgentTool during creation
}
