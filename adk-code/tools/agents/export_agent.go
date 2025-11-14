package agents

import (
	"fmt"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/pkg/agents"
	common "adk-code/tools/base"
)

// ExportAgentInput defines the input for the export_agent tool
type ExportAgentInput struct {
	// AgentName is the name of the agent to export (required)
	AgentName string `json:"agent_name" jsonschema:"Name of the agent to export"`

	// FilePath is the path to the agent file to export
	FilePath string `json:"file_path,omitempty" jsonschema:"Path to agent file to export"`

	// Format is the export format (markdown, json, yaml, plugin)
	Format string `json:"format,omitempty" jsonschema:"Export format: markdown, json, yaml, or plugin (default: markdown)"`

	// OutputPath is where to write the exported file
	OutputPath string `json:"output_path,omitempty" jsonschema:"Where to write the exported file"`

	// IncludeMetadata indicates whether to include metadata in export
	IncludeMetadata bool `json:"include_metadata,omitempty" jsonschema:"Include metadata in export (default: true)"`
}

// ExportAgentOutput defines the output for the export_agent tool
type ExportAgentOutput struct {
	// Success indicates whether the export was successful
	Success bool `json:"success"`

	// AgentName is the name of the exported agent
	AgentName string `json:"agent_name"`

	// FilePath is the path to the original agent file
	FilePath string `json:"file_path"`

	// ExportPath is the path where the agent was exported
	ExportPath string `json:"export_path"`

	// Format is the export format used
	Format string `json:"format"`

	// Content is the exported content
	Content string `json:"content"`

	// Size is the size of the exported file in bytes
	Size int64 `json:"size"`

	// Message is a human-readable message
	Message string `json:"message"`

	// Errors contains any validation errors
	Errors []string `json:"errors,omitempty"`

	// Warnings contains any validation warnings
	Warnings []string `json:"warnings,omitempty"`
}

// NewExportAgentTool creates a new export_agent tool
func NewExportAgentTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ExportAgentInput) ExportAgentOutput {
		output := ExportAgentOutput{
			Success:   false,
			AgentName: input.AgentName,
			Errors:    []string{},
			Warnings:  []string{},
		}

		// Validate input
		if input.AgentName == "" {
			output.Errors = append(output.Errors, "agent_name is required")
		}

		// Default format to markdown if not specified
		if input.Format == "" {
			input.Format = "markdown"
		}

		validFormats := map[string]bool{
			"markdown": true,
			"json":     true,
			"yaml":     true,
			"plugin":   true,
		}

		if !validFormats[input.Format] {
			output.Errors = append(output.Errors, fmt.Sprintf("invalid format: %s", input.Format))
		}

		if len(output.Errors) > 0 {
			output.Message = "Failed to export agent: validation errors"
			return output
		}

		// Load the agent
		var agent *agents.Agent
		var filePath string

		if input.FilePath != "" {
			parsedAgent, err := agents.ParseAgentFile(input.FilePath)
			if err != nil {
				output.Errors = append(output.Errors, fmt.Sprintf("Failed to parse agent file: %v", err))
				output.Message = fmt.Sprintf("Failed to export agent: %v", err)
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

		output.FilePath = filePath
		output.Format = input.Format

		// Lint the agent to check for issues
		linter := agents.NewLinter()
		lintResult := linter.Lint(agent)

		// Collect warnings
		for _, issue := range lintResult.Issues {
			if issue.Severity == agents.SeverityWarning {
				output.Warnings = append(output.Warnings, fmt.Sprintf("[%s] %s", issue.Rule, issue.Message))
			}
		}

		// Format the export based on the requested format
		switch input.Format {
		case "markdown":
			output.Content = formatAsMarkdown(agent)
		case "json":
			output.Content = formatAsJSON(agent)
		case "yaml":
			output.Content = formatAsYAML(agent)
		case "plugin":
			output.Content = formatAsPlugin(agent)
		}

		// Set the export path
		if input.OutputPath != "" {
			output.ExportPath = input.OutputPath
		} else {
			output.ExportPath = fmt.Sprintf("%s-%s.%s", agent.Name, input.Format, getFileExtension(input.Format))
		}

		output.Size = int64(len(output.Content))
		output.Success = true
		output.Message = fmt.Sprintf("Agent '%s' exported successfully to %s in %s format",
			input.AgentName, output.ExportPath, input.Format)

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "agents-export",
		Description: "Export an agent definition in various formats (markdown, json, yaml, plugin)",
	}, handler)

	if err != nil {
		return nil, fmt.Errorf("failed to create agents-export tool: %w", err)
	}

	// Register the tool
	common.Register(common.ToolMetadata{
		Tool:      t,
		Category:  common.CategorySearchDiscovery,
		Priority:  6,
		UsageHint: "Export agent definitions in various formats",
	})

	return t, nil
}

// formatAsMarkdown formats an agent as markdown
func formatAsMarkdown(agent *agents.Agent) string {
	content := fmt.Sprintf("# %s\n\n", agent.Name)
	content += fmt.Sprintf("**Version:** %s\n\n", agent.Version)
	content += fmt.Sprintf("**Author:** %s\n\n", agent.Author)
	content += fmt.Sprintf("**Description:** %s\n\n", agent.Description)

	if len(agent.Tags) > 0 {
		content += "**Tags:**\n"
		for _, tag := range agent.Tags {
			content += fmt.Sprintf("- %s\n", tag)
		}
		content += "\n"
	}

	if len(agent.Dependencies) > 0 {
		content += "**Dependencies:**\n"
		for _, dep := range agent.Dependencies {
			content += fmt.Sprintf("- %s\n", dep)
		}
		content += "\n"
	}

	if agent.Content != "" {
		content += "## Content\n\n"
		content += agent.Content
	}

	return content
}

// formatAsJSON formats an agent as JSON
func formatAsJSON(agent *agents.Agent) string {
	return fmt.Sprintf(`{
  "name": "%s",
  "version": "%s",
  "author": "%s",
  "description": "%s",
  "type": "%s",
  "source": "%s",
  "tags": %v,
  "dependencies": %v,
  "path": "%s"
}
`, agent.Name, agent.Version, agent.Author, agent.Description,
		agent.Type, agent.Source, agent.Tags, agent.Dependencies, agent.Path)
}

// formatAsYAML formats an agent as YAML
func formatAsYAML(agent *agents.Agent) string {
	yaml := "---\n"
	yaml += fmt.Sprintf("name: %s\n", agent.Name)
	yaml += fmt.Sprintf("version: %s\n", agent.Version)
	yaml += fmt.Sprintf("author: %s\n", agent.Author)
	yaml += fmt.Sprintf("description: |\n")
	yaml += fmt.Sprintf("  %s\n", agent.Description)
	yaml += fmt.Sprintf("type: %s\n", agent.Type)
	yaml += fmt.Sprintf("source: %s\n", agent.Source)

	if len(agent.Tags) > 0 {
		yaml += "tags:\n"
		for _, tag := range agent.Tags {
			yaml += fmt.Sprintf("  - %s\n", tag)
		}
	}

	if len(agent.Dependencies) > 0 {
		yaml += "dependencies:\n"
		for _, dep := range agent.Dependencies {
			yaml += fmt.Sprintf("  - %s\n", dep)
		}
	}

	yaml += "---\n"

	if agent.Content != "" {
		yaml += agent.Content
	}

	return yaml
}

// formatAsPlugin formats an agent as a plugin descriptor
func formatAsPlugin(agent *agents.Agent) string {
	plugin := fmt.Sprintf(`# Plugin: %s

## Metadata
- **Name:** %s
- **Version:** %s
- **Author:** %s
- **Type:** %s

## Description
%s

## Features
`, agent.Name, agent.Name, agent.Version, agent.Author, agent.Type, agent.Description)

	if len(agent.Tags) > 0 {
		for _, tag := range agent.Tags {
			plugin += fmt.Sprintf("- %s\n", tag)
		}
	} else {
		plugin += "- No features specified\n"
	}

	if len(agent.Dependencies) > 0 {
		plugin += "\n## Dependencies\n"
		for _, dep := range agent.Dependencies {
			plugin += fmt.Sprintf("- %s\n", dep)
		}
	}

	return plugin
}

// getFileExtension returns the file extension for a format
func getFileExtension(format string) string {
	switch format {
	case "json":
		return "json"
	case "yaml":
		return "yaml"
	case "plugin":
		return "plugin.md"
	default:
		return "md"
	}
}
