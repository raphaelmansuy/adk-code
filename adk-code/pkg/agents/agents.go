// Package agents provides agent definition discovery and management for adk-code.
// This package implements support for Claude Code agent definitions using a
// compatible YAML + Markdown file format.
//
// Phase 0 Implementation:
// - Basic agent file discovery in .adk/agents/ directory
// - YAML frontmatter parsing (name and description)
// - Simple agent listing capability
package agents

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	// ErrNoFrontmatter is returned when a file has no YAML frontmatter
	ErrNoFrontmatter = errors.New("no YAML frontmatter found")

	// ErrInvalidYAML is returned when YAML syntax is invalid
	ErrInvalidYAML = errors.New("invalid YAML syntax")

	// ErrMissingName is returned when name field is missing
	ErrMissingName = errors.New("missing required field: name")

	// ErrMissingDescription is returned when description field is missing
	ErrMissingDescription = errors.New("missing required field: description")
)

// AgentType represents the type of agent definition
type AgentType string

const (
	TypeSubagent AgentType = "subagent"
	TypeSkill    AgentType = "skill"
	TypeCommand  AgentType = "command"
	TypePlugin   AgentType = "plugin"
)

// String returns the string representation of AgentType
func (t AgentType) String() string {
	return string(t)
}

// AgentSource indicates where the agent was discovered
type AgentSource string

const (
	SourceProject AgentSource = "project"
	SourceUser    AgentSource = "user"
	SourcePlugin  AgentSource = "plugin"
	SourceCLI     AgentSource = "cli"
)

// String returns the string representation of AgentSource
func (s AgentSource) String() string {
	return string(s)
}

// Agent represents a discovered agent definition.
// Phase 0 version includes minimal fields needed for discovery and listing.
type Agent struct {
	// Identity
	Name        string
	Description string

	// Type and Source
	Type   AgentType
	Source AgentSource

	// File Information
	Path    string    // File path relative to project root
	ModTime time.Time // Last modified time

	// Content (preserved for future phases)
	Content string // Markdown content after frontmatter
	RawYAML string // Original YAML frontmatter for round-tripping
}

// DiscoveryResult holds the results of agent discovery operations
type DiscoveryResult struct {
	// Discovered agents
	Agents []*Agent

	// Summary statistics
	Total      int           // Total agents found
	ErrorCount int           // Number of errors encountered
	TimeTaken  time.Duration // Time spent discovering

	// Error tracking
	Errors []error
}

// IsEmpty returns true if no agents were discovered
func (r *DiscoveryResult) IsEmpty() bool {
	return r.Total == 0
}

// HasErrors returns true if any errors occurred during discovery
func (r *DiscoveryResult) HasErrors() bool {
	return r.ErrorCount > 0 || len(r.Errors) > 0
}

// Discoverer finds agent definition files in a project.
// Phase 0 implementation scans .adk/agents/ directory only.
type Discoverer struct {
	projectRoot string
}

// NewDiscoverer creates a new agent discoverer for the given project root
func NewDiscoverer(projectRoot string) *Discoverer {
	return &Discoverer{
		projectRoot: projectRoot,
	}
}

// DiscoverAll finds all agent definitions in the project.
// Phase 0: Only scans project-level .adk/agents/ directory.
// Returns a DiscoveryResult with discovered agents or errors.
func (d *Discoverer) DiscoverAll() (*DiscoveryResult, error) {
	startTime := time.Now()

	result := &DiscoveryResult{
		Agents: make([]*Agent, 0),
		Errors: make([]error, 0),
	}

	// Phase 0: Only scan project-level .adk/agents/
	agentsDir := filepath.Join(d.projectRoot, ".adk", "agents")

	// Check if directory exists
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		// Not an error - just no agents yet
		result.TimeTaken = time.Since(startTime)
		return result, nil
	}

	// Walk the agents directory
	err := filepath.Walk(agentsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.ErrorCount++
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .md files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Parse the agent file
		agent, parseErr := ParseAgentFile(path)
		if parseErr != nil {
			result.Errors = append(result.Errors, parseErr)
			result.ErrorCount++
			return nil // Continue walking
		}

		// Set source and type
		agent.Source = SourceProject
		agent.Type = TypeSubagent // Phase 0: assume all are subagents

		result.Agents = append(result.Agents, agent)
		result.Total++

		return nil
	})

	if err != nil {
		result.Errors = append(result.Errors, err)
		result.ErrorCount++
	}

	result.TimeTaken = time.Since(startTime)
	return result, nil
}

// DiscoverProjectAgents finds agents only in project-level directory
func (d *Discoverer) DiscoverProjectAgents() (*DiscoveryResult, error) {
	return d.DiscoverAll() // Phase 0: only project agents anyway
}

// ParseAgentFile reads and parses an agent definition file.
// The file format is YAML frontmatter followed by markdown content.
func ParseAgentFile(path string) (*Agent, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Extract YAML frontmatter and markdown content
	yamlContent, markdownContent, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Parse YAML fields
	var frontmatter struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}

	if err := yaml.Unmarshal(yamlContent, &frontmatter); err != nil {
		return nil, ErrInvalidYAML
	}

	// Validate required fields
	if frontmatter.Name == "" {
		return nil, ErrMissingName
	}
	if frontmatter.Description == "" {
		return nil, ErrMissingDescription
	}

	// Create agent
	agent := &Agent{
		Name:        frontmatter.Name,
		Description: frontmatter.Description,
		Content:     string(markdownContent),
		RawYAML:     string(yamlContent),
		Path:        path,
	}

	// Get file modification time
	info, err := os.Stat(path)
	if err == nil {
		agent.ModTime = info.ModTime()
	}

	return agent, nil
}

// extractFrontmatter extracts YAML frontmatter from markdown content.
// Expected format:
//
//	---
//	name: agent-name
//	description: Agent description
//	---
//
// Markdown content...
func extractFrontmatter(content []byte) (yaml []byte, markdown []byte, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))

	// First line must be "---"
	if !scanner.Scan() || scanner.Text() != "---" {
		return nil, nil, ErrNoFrontmatter
	}

	// Read YAML until closing "---"
	var yamlLines []string
	foundClosing := false

	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			foundClosing = true
			break
		}
		yamlLines = append(yamlLines, line)
	}

	if !foundClosing {
		return nil, nil, ErrNoFrontmatter
	}

	// Remaining content is markdown
	var markdownLines []string
	for scanner.Scan() {
		markdownLines = append(markdownLines, scanner.Text())
	}

	yaml = []byte(strings.Join(yamlLines, "\n"))
	markdown = []byte(strings.Join(markdownLines, "\n"))

	return yaml, markdown, nil
}
