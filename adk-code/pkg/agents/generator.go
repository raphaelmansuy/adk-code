package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TemplateType represents the type of agent template
type TemplateType string

const (
	TemplateSubagent TemplateType = "subagent"
	TemplateSkill    TemplateType = "skill"
	TemplateCommand  TemplateType = "command"
)

// AgentGeneratorInput defines parameters for agent generation
type AgentGeneratorInput struct {
	Name         string
	Description  string
	TemplateType TemplateType
	Author       string
	Version      string
	Tags         []string
	TargetPath   string // Directory to create agent in
}

// AgentGenerator generates agent definition files
type AgentGenerator struct {
	templates map[TemplateType]string
}

// NewAgentGenerator creates a new agent generator
func NewAgentGenerator() *AgentGenerator {
	return &AgentGenerator{
		templates: initializeTemplates(),
	}
}

// GenerateAgent creates a new agent file from a template
func (g *AgentGenerator) GenerateAgent(input AgentGeneratorInput) (*Agent, error) {
	// Validate input
	if err := validateGeneratorInput(input); err != nil {
		return nil, err
	}

	// Set default template type
	if input.TemplateType == "" {
		input.TemplateType = TemplateSubagent
	}

	// Get template
	template, exists := g.templates[input.TemplateType]
	if !exists {
		return nil, fmt.Errorf("unknown template type: %s", input.TemplateType)
	}

	// Set defaults
	if input.Version == "" {
		input.Version = "1.0.0"
	}
	if input.Author == "" {
		input.Author = "Unknown"
	}

	// Create YAML frontmatter
	frontmatter := generateFrontmatter(input)

	// Create agent object
	agent := &Agent{
		Name:         input.Name,
		Description:  input.Description,
		Type:         TypeSubagent,
		Source:       SourceProject,
		Version:      input.Version,
		Author:       input.Author,
		Tags:         input.Tags,
		Dependencies: []string{},
		Content:      template,
		RawYAML:      frontmatter,
		ModTime:      time.Now(),
	}

	return agent, nil
}

// WriteAgent writes an agent to disk
func (g *AgentGenerator) WriteAgent(agent *Agent, basePath string) (string, error) {
	// Create .adk/agents directory if needed
	agentDir := filepath.Join(basePath, ".adk", "agents")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Create filename
	filename := agent.Name + ".md"
	filePath := filepath.Join(agentDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); err == nil {
		return "", fmt.Errorf("agent file already exists: %s", filePath)
	}

	// Write file
	content := agent.RawYAML + agent.Content
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write agent file: %w", err)
	}

	return filePath, nil
}

// generateFrontmatter creates YAML frontmatter for an agent
func generateFrontmatter(input AgentGeneratorInput) string {
	frontmatter := "---\n"
	frontmatter += fmt.Sprintf("name: %s\n", input.Name)
	frontmatter += fmt.Sprintf("description: %s\n", input.Description)

	if input.Version != "" {
		frontmatter += fmt.Sprintf("version: %s\n", input.Version)
	}

	if input.Author != "" {
		frontmatter += fmt.Sprintf("author: %s\n", input.Author)
	}

	if len(input.Tags) > 0 {
		frontmatter += "tags:\n"
		for _, tag := range input.Tags {
			frontmatter += fmt.Sprintf("  - %s\n", tag)
		}
	}

	frontmatter += "---\n"
	return frontmatter
}

// validateGeneratorInput validates agent generator input
func validateGeneratorInput(input AgentGeneratorInput) error {
	if input.Name == "" {
		return fmt.Errorf("agent name is required")
	}

	if !isKebabCase(input.Name) {
		return fmt.Errorf("agent name must be kebab-case, got: %s", input.Name)
	}

	if input.Description == "" {
		return fmt.Errorf("agent description is required")
	}

	if len(input.Description) < 10 {
		return fmt.Errorf("agent description must be at least 10 characters")
	}

	if len(input.Description) > 1024 {
		return fmt.Errorf("agent description must be at most 1024 characters")
	}

	return nil
}

// initializeTemplates creates built-in agent templates
func initializeTemplates() map[TemplateType]string {
	return map[TemplateType]string{
		TemplateSubagent: `## Overview

This agent specializes in [specific domain/task]. It is designed to help with [main purpose].

## Capabilities

- [Capability 1]
- [Capability 2]
- [Capability 3]

## Usage

To use this agent, provide:
1. [Input type 1]
2. [Input type 2]

## Example

[Provide an example of typical usage]

## Notes

- This agent requires [any dependencies or permissions]
- Performance considerations: [if any]
`,
		TemplateSkill: `## Skill Description

This skill provides specialized functionality for [domain].

## Methods

### Method Name

**Purpose**: [Description of what this method does]

**Parameters**:
- param1: [description]
- param2: [description]

**Returns**: [Description of return value]

## Implementation Notes

[Any relevant implementation details or considerations]
`,
		TemplateCommand: `## Command Description

This command-line utility provides [functionality].

## Syntax

` + "`" + `command-name [options] [arguments]` + "`" + `

## Options

- ` + "`" + `--option1` + "`" + `: [Description]
- ` + "`" + `--option2` + "`" + `: [Description]
- ` + "`" + `--help` + "`" + `: Show help message

## Examples

` + "```bash" + `
# Example 1: [Description]
command-name arg1 arg2

# Example 2: [Description]
command-name --option1 value
` + "```" + `

## Exit Codes

- 0: Success
- 1: General error
- 2: Misuse of command
`,
	}
}

// GetAvailableTemplates returns list of available template types
func GetAvailableTemplates() []TemplateType {
	return []TemplateType{
		TemplateSubagent,
		TemplateSkill,
		TemplateCommand,
	}
}

// Template returns the template content for a given type
func (g *AgentGenerator) Template(t TemplateType) (string, error) {
	template, exists := g.templates[t]
	if !exists {
		return "", fmt.Errorf("unknown template type: %s", t)
	}
	return template, nil
}

// CustomizeTemplate allows users to customize a template before generation
func (g *AgentGenerator) CustomizeTemplate(t TemplateType, customContent string) error {
	if _, exists := g.templates[t]; !exists {
		return fmt.Errorf("unknown template type: %s", t)
	}
	g.templates[t] = customContent
	return nil
}

// isKebabCase is a helper function (imported from linter)
// In actual usage, this would be shared or imported
