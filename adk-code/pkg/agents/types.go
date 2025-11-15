// Package agents provides agent definition discovery and management for adk-code.
// This package implements support for Claude Code agent definitions using a
// compatible YAML + Markdown file format.
package agents

import "time"

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
// Includes Phase 0 discovery fields and Phase 1 metadata enhancement.
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

	// Phase 1 Metadata Enhancement
	Version      string   // Semantic versioning (e.g., "1.0.0")
	Author       string   // Email or name of agent author
	Tags         []string // Categories/tags for the agent
	Dependencies []string // Names of agents this depends on

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
