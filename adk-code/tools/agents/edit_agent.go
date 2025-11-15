package agents

import (
	"fmt"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/pkg/agents"
	common "adk-code/tools/base"
)

// EditAgentInput defines the input for the edit_agent tool
type EditAgentInput struct {
	// AgentName is the name of the agent to edit (required)
	AgentName string `json:"agent_name" jsonschema:"Name of the agent to edit"`

	// FilePath is the path to the agent file to edit
	FilePath string `json:"file_path,omitempty" jsonschema:"Path to agent file to edit"`

	// Field is the field to update (name, description, version, author, tags)
	Field string `json:"field" jsonschema:"Field to update: name, description, version, author, or tags"`

	// Value is the new value for the field
	Value string `json:"value" jsonschema:"New value for the field"`

	// CreateBackup indicates whether to create a backup before editing
	CreateBackup bool `json:"create_backup,omitempty" jsonschema:"Create backup before editing (default: true)"`
}

// EditAgentOutput defines the output for the edit_agent tool
type EditAgentOutput struct {
	// Success indicates whether the edit was successful
	Success bool `json:"success"`

	// AgentName is the name of the edited agent
	AgentName string `json:"agent_name"`

	// FilePath is the path to the agent file
	FilePath string `json:"file_path"`

	// Field is the field that was updated
	Field string `json:"field"`

	// OldValue is the previous value
	OldValue string `json:"old_value"`

	// NewValue is the new value
	NewValue string `json:"new_value"`

	// BackupPath is the path to the backup file (if created)
	BackupPath string `json:"backup_path,omitempty"`

	// Message is a human-readable message
	Message string `json:"message"`

	// Errors contains any validation errors
	Errors []string `json:"errors,omitempty"`

	// Warnings contains any validation warnings
	Warnings []string `json:"warnings,omitempty"`
}

// NewEditAgentTool creates a new edit_agent tool
func NewEditAgentTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input EditAgentInput) EditAgentOutput {
		output := EditAgentOutput{
			Success:   false,
			AgentName: input.AgentName,
			Field:     input.Field,
			Errors:    []string{},
			Warnings:  []string{},
		}

		// Validate input
		if input.AgentName == "" {
			output.Errors = append(output.Errors, "agent_name is required")
		}

		if input.Field == "" {
			output.Errors = append(output.Errors, "field is required")
		}

		validFields := map[string]bool{
			"name":        true,
			"description": true,
			"version":     true,
			"author":      true,
			"tags":        true,
		}

		if input.Field != "" && !validFields[input.Field] {
			output.Errors = append(output.Errors, fmt.Sprintf("invalid field: %s", input.Field))
		}

		if len(output.Errors) > 0 {
			output.Message = "Failed to edit agent: validation errors"
			return output
		}

		// Load the agent
		var agent *agents.Agent
		var filePath string

		if input.FilePath != "" {
			parsedAgent, err := agents.ParseAgentFile(input.FilePath)
			if err != nil {
				output.Errors = append(output.Errors, fmt.Sprintf("Failed to parse agent file: %v", err))
				output.Message = fmt.Sprintf("Failed to edit agent: %v", err)
				return output
			}
			agent = parsedAgent
			filePath = input.FilePath
		} else {
			// Try to discover the agent
			discoverer := agents.NewDiscoverer(".")
			result, err := discoverer.DiscoverAll()
			if err != nil {
				output.Errors = append(output.Errors, fmt.Sprintf("Failed to discover agents: %v", err))
				output.Message = fmt.Sprintf("Failed to discover agents: %v", err)
				return output
			}

			found := false
			for _, a := range result.Agents {
				if a.Name == input.AgentName {
					agent = a
					found = true
					filePath = a.Path
					break
				}
			}

			if !found {
				output.Errors = append(output.Errors, fmt.Sprintf("Agent %q not found", input.AgentName))
				output.Message = fmt.Sprintf("Agent %q not found", input.AgentName)
				return output
			}
		}

		// Store old value
		output.FilePath = filePath

		switch input.Field {
		case "name":
			output.OldValue = agent.Name
			agent.Name = input.Value
			output.NewValue = input.Value

		case "description":
			output.OldValue = agent.Description
			agent.Description = input.Value
			output.NewValue = input.Value

		case "version":
			output.OldValue = agent.Version
			agent.Version = input.Value
			output.NewValue = input.Value

		case "author":
			output.OldValue = agent.Author
			agent.Author = input.Value
			output.NewValue = input.Value

		case "tags":
			// Tags are comma-separated
			if agent.Tags == nil {
				output.OldValue = ""
			} else {
				output.OldValue = fmt.Sprintf("%v", agent.Tags)
			}
			agent.Tags = []string{input.Value}
			output.NewValue = input.Value
		}

		// Lint the edited agent to check for issues
		linter := agents.NewLinter()
		lintResult := linter.Lint(agent)

		// Collect warnings and errors
		for _, issue := range lintResult.Issues {
			if issue.Severity == agents.SeverityError {
				output.Errors = append(output.Errors, fmt.Sprintf("[%s] %s", issue.Rule, issue.Message))
			} else if issue.Severity == agents.SeverityWarning {
				output.Warnings = append(output.Warnings, fmt.Sprintf("[%s] %s", issue.Rule, issue.Message))
			}
		}

		output.Success = true
		output.Message = fmt.Sprintf("Agent '%s' field '%s' updated successfully from '%s' to '%s'",
			input.AgentName, input.Field, output.OldValue, output.NewValue)

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "agents-edit",
		Description: "Edit an existing agent definition with validation and linting",
	}, handler)

	if err != nil {
		return nil, fmt.Errorf("failed to create agents-edit tool: %w", err)
	}

	// Register the tool
	common.Register(common.ToolMetadata{
		Tool:      t,
		Category:  common.CategorySearchDiscovery,
		Priority:  7,
		UsageHint: "Edit existing agent definitions safely with validation",
	})

	return t, nil
}
