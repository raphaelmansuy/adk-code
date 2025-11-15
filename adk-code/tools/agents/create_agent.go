package agents

import (
	"fmt"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/pkg/agents"
	common "adk-code/tools/base"
)

// CreateAgentInput defines the input for the create_agent tool
type CreateAgentInput struct {
	// Name is the name of the agent (kebab-case)
	Name string `json:"name" jsonschema:"Name of the agent to create (kebab-case)"`

	// Description is a description of what the agent does
	Description string `json:"description" jsonschema:"Description of what the agent does (10-1024 characters)"`

	// TemplateType is the template type (subagent, skill, command)
	TemplateType string `json:"template_type" jsonschema:"Template type: subagent, skill, or command"`

	// Author is the author of the agent (optional)
	Author string `json:"author,omitempty" jsonschema:"Author name or email (optional)"`

	// Tags are tags for categorizing the agent (optional)
	Tags []string `json:"tags,omitempty" jsonschema:"Tags for categorizing the agent"`

	// Version is the initial version (optional, defaults to 1.0.0)
	Version string `json:"version,omitempty" jsonschema:"Initial version (defaults to 1.0.0)"`

	// TargetPath is where to write the agent file (optional)
	TargetPath string `json:"target_path,omitempty" jsonschema:"Where to write the agent file"`
}

// CreateAgentOutput defines the output for the create_agent tool
type CreateAgentOutput struct {
	// Success indicates whether the agent was created
	Success bool `json:"success"`

	// AgentName is the name of the created agent
	AgentName string `json:"agent_name"`

	// FilePath is the path where the agent was written
	FilePath string `json:"file_path"`

	// Content is the generated agent YAML content
	Content string `json:"content"`

	// Message is a human-readable message
	Message string `json:"message"`

	// Errors contains any validation errors
	Errors []string `json:"errors,omitempty"`

	// Warnings contains any validation warnings
	Warnings []string `json:"warnings,omitempty"`
}

// NewCreateAgentTool creates a new create_agent tool
func NewCreateAgentTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input CreateAgentInput) CreateAgentOutput {
		output := CreateAgentOutput{
			Success:   false,
			AgentName: input.Name,
			Errors:    []string{},
			Warnings:  []string{},
		}

		// Validate input
		if input.Name == "" {
			output.Errors = append(output.Errors, "agent name is required")
		}

		if input.Description == "" {
			output.Errors = append(output.Errors, "description is required")
		}

		if input.TemplateType == "" {
			output.Errors = append(output.Errors, "template_type is required")
		}

		if len(output.Errors) > 0 {
			output.Message = "Failed to create agent: validation errors"
			return output
		}

		// Create the agent generator
		generator := agents.NewAgentGenerator()

		// Prepare generator input
		genInput := agents.AgentGeneratorInput{
			Name:         input.Name,
			Description:  input.Description,
			TemplateType: agents.TemplateType(input.TemplateType),
			Author:       input.Author,
			Tags:         input.Tags,
			Version:      input.Version,
			TargetPath:   input.TargetPath,
		}

		// Generate the agent
		agent, err := generator.GenerateAgent(genInput)
		if err != nil {
			output.Errors = append(output.Errors, err.Error())
			output.Message = fmt.Sprintf("Failed to generate agent: %v", err)
			return output
		}

		// Lint the generated agent to check for issues
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

		// Write the agent to disk
		filePath, err := generator.WriteAgent(agent, input.TargetPath)
		if err != nil {
			output.Errors = append(output.Errors, err.Error())
			output.Message = fmt.Sprintf("Failed to write agent file: %v", err)
			return output
		}

		output.FilePath = filePath
		output.Content = agent.Description
		output.Success = true
		output.Message = fmt.Sprintf("Agent '%s' created successfully at %s", input.Name, filePath)

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "agents-create",
		Description: "Create a new agent from a template with validation and linting",
	}, handler)

	if err != nil {
		return nil, fmt.Errorf("failed to create agents-create tool: %w", err)
	}

	// Register the tool
	common.Register(common.ToolMetadata{
		Tool:      t,
		Category:  common.CategorySearchDiscovery,
		Priority:  8,
		UsageHint: "Create new agent definitions from templates",
	})

	return t, nil
}
