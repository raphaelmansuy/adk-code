# Phase 0 Implementation Plan: Agent Definition Support

**Version**: 1.0  
**Date**: 2025-11-14  
**Status**: 📋 Ready for Implementation  
**Parent Spec**: [0001-agent-definition-support.md](./0001-agent-definition-support.md)  
**Related ADR**: [0001-claude-code-agent-support.md](../adr/0001-claude-code-agent-support.md)  
**Timeline**: December 2025 (4 weeks)  
**Goal**: Proof of Concept - Demonstrate basic agent file discovery and parsing

---

## Executive Summary

Phase 0 is the **foundation** for Claude Code agent support in adk-code. This phase proves the concept is viable by implementing minimal but functional agent discovery and parsing. Success here determines whether we proceed with Phases 1-3.

**Scope**: Build the smallest viable system that can:

1. Discover agent files in `.adk/agents/` directories
2. Parse YAML frontmatter and extract basic fields
3. Display discovered agents via one CLI command
4. Lay foundation for future phases

**Not in Scope for Phase 0**:

- ❌ Validation framework (Phase 1)
- ❌ Agent creation/editing (Phase 2)
- ❌ Skills or commands support (Phase 1)
- ❌ Plugin discovery (Phase 2)
- ❌ Caching or performance optimization (Phase 1)
- ❌ Comprehensive error handling (Phase 1)

**Success Criteria**:

- ✅ Can discover 10+ agent files in test directory
- ✅ Parses name and description from YAML frontmatter
- ✅ `/agents-list` command works and displays agents
- ✅ Code is tested (>80% coverage)
- ✅ Documentation for Phase 0 features exists
- ✅ Foundation is extensible for Phase 1

---

## Phase 0 Deliverables

### 1. Core Package: `pkg/agents` (~400 lines)

**File**: `pkg/agents/types.go` (~100 lines)

```go
package agents

import "time"

// AgentType represents the type of agent definition
type AgentType string

const (
    TypeSubagent AgentType = "subagent"
    TypeSkill    AgentType = "skill"
    TypeCommand  AgentType = "command"
)

// AgentSource indicates where the agent was discovered
type AgentSource string

const (
    SourceProject AgentSource = "project"
    SourceUser    AgentSource = "user"
)

// Agent represents a discovered agent definition (minimal Phase 0 version)
type Agent struct {
    // Identity
    Name        string
    Description string
    
    // Metadata
    Type        AgentType
    Source      AgentSource
    Path        string    // File path
    ModTime     time.Time // Last modified
    
    // Content (preserved for future phases)
    Content     string    // Markdown content
    RawYAML     string    // Original YAML frontmatter
}

// DiscoveryResult holds the results of agent discovery
type DiscoveryResult struct {
    Agents       []*Agent
    Total        int
    ErrorCount   int
    Errors       []error
    TimeTaken    time.Duration
}
```

**File**: `pkg/agents/parser.go` (~150 lines)

```go
package agents

import (
    "bufio"
    "bytes"
    "errors"
    "os"
    "strings"
    
    "gopkg.in/yaml.v3"
)

var (
    ErrNoFrontmatter = errors.New("no YAML frontmatter found")
    ErrInvalidYAML   = errors.New("invalid YAML syntax")
)

// ParseAgentFile reads and parses an agent definition file
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
    
    // Parse YAML into agent struct
    agent := &Agent{
        Path:    path,
        Content: string(markdownContent),
        RawYAML: string(yamlContent),
    }
    
    // Parse YAML fields
    var frontmatter struct {
        Name        string `yaml:"name"`
        Description string `yaml:"description"`
    }
    
    if err := yaml.Unmarshal(yamlContent, &frontmatter); err != nil {
        return nil, ErrInvalidYAML
    }
    
    agent.Name = frontmatter.Name
    agent.Description = frontmatter.Description
    
    // Get file info
    info, err := os.Stat(path)
    if err == nil {
        agent.ModTime = info.ModTime()
    }
    
    return agent, nil
}

// extractFrontmatter extracts YAML frontmatter from markdown content
// Expected format:
// ---
// name: agent-name
// description: Agent description
// ---
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
```

**File**: `pkg/agents/discovery.go` (~150 lines)

```go
package agents

import (
    "os"
    "path/filepath"
    "time"
)

// Discoverer finds agent definition files
type Discoverer struct {
    projectRoot string
}

// NewDiscoverer creates a new agent discoverer
func NewDiscoverer(projectRoot string) *Discoverer {
    return &Discoverer{
        projectRoot: projectRoot,
    }
}

// DiscoverAll finds all agent definitions (Phase 0: project-level only)
func (d *Discoverer) DiscoverAll() (*DiscoveryResult, error) {
    startTime := time.Now()
    
    result := &DiscoveryResult{
        Agents: make([]*Agent, 0),
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
```

### 2. Tool: `/agents-list` (~150 lines)

**File**: `tools/agents/list_agents_tool.go`

```go
package agents

import (
    "fmt"
    "strings"
    
    "google.golang.org/adk/tool"
    "google.golang.org/adk/tool/functiontool"
    
    agentpkg "adk-code/pkg/agents"
    common "adk-code/tools/base"
)

// ListAgentsInput defines the input for listing agents
type ListAgentsInput struct {
    Format string `json:"format" desc:"Output format: 'table' or 'json' (default: table)"`
}

// ListAgentsOutput defines the output structure
type ListAgentsOutput struct {
    Success bool              `json:"success"`
    Message string            `json:"message"`
    Agents  []*AgentSummary   `json:"agents,omitempty"`
    Total   int               `json:"total"`
}

// AgentSummary is a simplified agent for display
type AgentSummary struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Type        string `json:"type"`
    Source      string `json:"source"`
}

// NewListAgentsTool creates the agents-list tool
func NewListAgentsTool(projectRoot string) (tool.Tool, error) {
    handler := func(ctx tool.Context, input ListAgentsInput) ListAgentsOutput {
        // Discover agents
        discoverer := agentpkg.NewDiscoverer(projectRoot)
        result, err := discoverer.DiscoverAll()
        
        if err != nil {
            return ListAgentsOutput{
                Success: false,
                Message: fmt.Sprintf("Discovery failed: %v", err),
            }
        }
        
        // Convert to summaries
        summaries := make([]*AgentSummary, 0, result.Total)
        for _, agent := range result.Agents {
            summaries = append(summaries, &AgentSummary{
                Name:        agent.Name,
                Description: agent.Description,
                Type:        string(agent.Type),
                Source:      string(agent.Source),
            })
        }
        
        // Format output based on request
        format := input.Format
        if format == "" {
            format = "table"
        }
        
        var message string
        if format == "table" {
            message = formatAsTable(summaries)
        } else {
            message = fmt.Sprintf("Found %d agents", result.Total)
        }
        
        return ListAgentsOutput{
            Success: true,
            Message: message,
            Agents:  summaries,
            Total:   result.Total,
        }
    }
    
    t, err := functiontool.New(functiontool.Config{
        Name:        "agents_list",
        Description: "List all discovered agent definitions in the project",
        Handler:     handler,
    })
    
    if err != nil {
        return nil, err
    }
    
    // Register with global registry
    metadata := common.ToolMetadata{
        Tool:        t,
        Category:    common.CategoryAgents,
        Description: "List agent definitions",
        Examples: []string{
            "List all agents",
            "Show me available agents",
        },
    }
    
    if err := common.Register(metadata); err != nil {
        return nil, err
    }
    
    return t, nil
}

func formatAsTable(agents []*AgentSummary) string {
    if len(agents) == 0 {
        return "No agents found"
    }
    
    var b strings.Builder
    b.WriteString("\n")
    b.WriteString("┌────────────────────┬──────────┬────────────────────────────────────────┐\n")
    b.WriteString("│ Name               │ Type     │ Description                            │\n")
    b.WriteString("├────────────────────┼──────────┼────────────────────────────────────────┤\n")
    
    for _, agent := range agents {
        name := truncate(agent.Name, 18)
        agentType := truncate(agent.Type, 8)
        desc := truncate(agent.Description, 38)
        b.WriteString(fmt.Sprintf("│ %-18s │ %-8s │ %-38s │\n", name, agentType, desc))
    }
    
    b.WriteString("└────────────────────┴──────────┴────────────────────────────────────────┘\n")
    return b.String()
}

func truncate(s string, max int) string {
    if len(s) <= max {
        return s
    }
    return s[:max-3] + "..."
}
```

### 3. Tests (~250 lines)

**File**: `pkg/agents/parser_test.go`

```go
package agents

import (
    "os"
    "path/filepath"
    "testing"
)

func TestParseAgentFile(t *testing.T) {
    // Create temp directory
    tmpDir := t.TempDir()
    
    // Test case 1: Valid agent file
    validAgent := `---
name: test-agent
description: A test agent for unit tests
---

# Test Agent

This is the markdown content.
`
    
    validPath := filepath.Join(tmpDir, "valid.md")
    if err := os.WriteFile(validPath, []byte(validAgent), 0644); err != nil {
        t.Fatal(err)
    }
    
    agent, err := ParseAgentFile(validPath)
    if err != nil {
        t.Errorf("ParseAgentFile() error = %v, want nil", err)
    }
    
    if agent.Name != "test-agent" {
        t.Errorf("agent.Name = %q, want %q", agent.Name, "test-agent")
    }
    
    if agent.Description != "A test agent for unit tests" {
        t.Errorf("agent.Description = %q, want %q", agent.Description, "A test agent for unit tests")
    }
    
    // Test case 2: Missing frontmatter
    noFrontmatter := `# Agent without frontmatter`
    
    invalidPath := filepath.Join(tmpDir, "invalid.md")
    if err := os.WriteFile(invalidPath, []byte(noFrontmatter), 0644); err != nil {
        t.Fatal(err)
    }
    
    _, err = ParseAgentFile(invalidPath)
    if err != ErrNoFrontmatter {
        t.Errorf("ParseAgentFile() error = %v, want %v", err, ErrNoFrontmatter)
    }
}
```

**File**: `pkg/agents/discovery_test.go`

```go
package agents

import (
    "os"
    "path/filepath"
    "testing"
)

func TestDiscovererDiscoverAll(t *testing.T) {
    // Create temp project structure
    tmpDir := t.TempDir()
    agentsDir := filepath.Join(tmpDir, ".adk", "agents")
    
    if err := os.MkdirAll(agentsDir, 0755); err != nil {
        t.Fatal(err)
    }
    
    // Create test agent files
    agents := []struct {
        name    string
        content string
    }{
        {
            name: "agent1.md",
            content: `---
name: agent1
description: First test agent
---
Content 1`,
        },
        {
            name: "agent2.md",
            content: `---
name: agent2
description: Second test agent
---
Content 2`,
        },
    }
    
    for _, agent := range agents {
        path := filepath.Join(agentsDir, agent.name)
        if err := os.WriteFile(path, []byte(agent.content), 0644); err != nil {
            t.Fatal(err)
        }
    }
    
    // Discover agents
    discoverer := NewDiscoverer(tmpDir)
    result, err := discoverer.DiscoverAll()
    
    if err != nil {
        t.Errorf("DiscoverAll() error = %v, want nil", err)
    }
    
    if result.Total != 2 {
        t.Errorf("result.Total = %d, want 2", result.Total)
    }
    
    if len(result.Agents) != 2 {
        t.Errorf("len(result.Agents) = %d, want 2", len(result.Agents))
    }
}
```

### 4. Integration with Tool Registry

**File**: `tools/agents/init.go`

```go
package agents

import (
    "os"
    common "adk-code/tools/base"
)

func init() {
    // Auto-register agents-list tool during package init
    projectRoot, _ := os.Getwd()
    tool, err := NewListAgentsTool(projectRoot)
    if err != nil {
        // Log error but don't panic - tool registration failures are non-fatal
        return
    }
    
    _ = tool // Tool is already registered in NewListAgentsTool
}
```

### 5. Update Tool Registry Category

**File**: `tools/base/registry.go` (add new category)

```go
const (
    // ... existing categories ...
    CategoryAgents    ToolCategory = "agents"  // NEW
)
```

### 6. Documentation

**File**: `docs/tools/agents-list.md`

```markdown
# agents-list Tool

## Description

Lists all discovered agent definitions in the current project.

## Phase 0 Scope

- Discovers agents in `.adk/agents/` directory only
- Parses name and description from YAML frontmatter
- Displays results in table or JSON format

## Usage

```bash
# List all agents in table format
adk-code /agents-list

# List agents in JSON format
adk-code /agents-list --format json
```

## Output Format

### Table (default)

```
┌────────────────────┬──────────┬────────────────────────────────────────┐
│ Name               │ Type     │ Description                            │
├────────────────────┼──────────┼────────────────────────────────────────┤
│ code-reviewer      │ subagent │ Reviews code for quality and security  │
│ debugger           │ subagent │ Debugs errors and test failures        │
└────────────────────┴──────────┴────────────────────────────────────────┘
```

### JSON

```json
{
  "success": true,
  "message": "Found 2 agents",
  "agents": [
    {
      "name": "code-reviewer",
      "description": "Reviews code for quality and security",
      "type": "subagent",
      "source": "project"
    }
  ],
  "total": 2
}
```

## Limitations (Phase 0)

- Only scans `.adk/agents/` directory (not user-level or plugins)
- No validation of agent definitions
- No caching (re-scans on each call)
- All agents assumed to be type "subagent"
```

---

## Implementation Checklist

### Week 1: Foundation (Dec 1-7, 2025)

- [ ] Create `pkg/agents` package structure
- [ ] Implement `types.go` with basic data models
- [ ] Implement `parser.go` with YAML frontmatter extraction
- [ ] Write unit tests for parser (>80% coverage)
- [ ] Document parser API

### Week 2: Discovery (Dec 8-14, 2025)

- [ ] Implement `discovery.go` with file system scanning
- [ ] Write unit tests for discovery (>80% coverage)
- [ ] Test with sample agent files (create 5+ test agents)
- [ ] Document discovery API

### Week 3: CLI Tool (Dec 15-21, 2025)

- [ ] Create `tools/agents` package
- [ ] Implement `agents-list` tool
- [ ] Add to tool registry
- [ ] Write integration tests
- [ ] Create tool documentation

### Week 4: Polish & Testing (Dec 22-31, 2025)

- [ ] End-to-end testing with real agent files
- [ ] Fix bugs discovered during testing
- [ ] Code review and refactoring
- [ ] Update main README with Phase 0 status
- [ ] Demo video or screenshot for documentation

---

## Test Plan

### Unit Tests (Target: >80% Coverage)

1. **Parser Tests** (`pkg/agents/parser_test.go`)
   - Valid YAML frontmatter parsing
   - Missing frontmatter handling
   - Invalid YAML syntax
   - Multi-line values
   - Edge cases (empty files, no content)

2. **Discovery Tests** (`pkg/agents/discovery_test.go`)
   - Empty directory
   - Single agent
   - Multiple agents
   - Non-.md files (should skip)
   - Nested directories (should walk)
   - Permission errors

### Integration Tests

1. **Tool Integration** (`tools/agents/list_agents_tool_test.go`)
   - Tool registration works
   - Tool can be invoked
   - Output formats (table, JSON)
   - Error handling

2. **End-to-End Tests**
   - Create test project with agents
   - Run `adk-code /agents-list`
   - Verify output correctness

---

## Success Metrics

**Quantitative**:

- ✅ >80% test coverage for all new code
- ✅ Parse 10+ sample agent files without errors
- ✅ Discovery completes in <100ms for 10 agents
- ✅ Zero crashes or panics in testing

**Qualitative**:

- ✅ Code is readable and well-documented
- ✅ Patterns are consistent with existing adk-code style
- ✅ Foundation is extensible for Phase 1
- ✅ Team confidence in continuing to Phase 1

---

## Risk Mitigation

### Risk: YAML Parsing Complexity

**Mitigation**: Use `gopkg.in/yaml.v3` library (mature, well-tested)

### Risk: File System Edge Cases

**Mitigation**: Comprehensive test suite with edge cases, use `filepath` package

### Risk: Integration with Existing Tools

**Mitigation**: Follow existing tool patterns (file/, exec/, etc.)

### Risk: Time Overrun

**Mitigation**: Phase 0 scope is minimal - can ship partial if needed

---

## Dependencies

**External Libraries**:

- `gopkg.in/yaml.v3` - YAML parsing (already used in adk-code)
- No new dependencies required

**Internal Dependencies**:

- `tools/base` - Tool registry
- `pkg/workspace` - Project root detection (existing)

---

## Phase 0 → Phase 1 Transition

At the end of Phase 0, we will have:

✅ **Foundation in place**:

- Basic data models
- Parser infrastructure
- Discovery infrastructure
- One working CLI command

📋 **Phase 1 will add**:

- Validation framework
- Skills and commands support
- User-level agent discovery (~/.adk/agents/)
- Caching for performance
- More CLI commands (validate, describe)

🔍 **Decision Point**:

After Phase 0, assess:

1. Did we achieve success criteria?
2. Were estimates accurate?
3. Is the architecture sound?
4. Should we continue to Phase 1?

---

## References

- [Parent Spec: 0001-agent-definition-support.md](./0001-agent-definition-support.md)
- [ADR: 0001-claude-code-agent-support.md](../adr/0001-claude-code-agent-support.md)
- [Research: Claude Code Agent Patterns](../../research/claude-code/0001-draft-claude-code-agent.md)
- [Tool Development Guide](../TOOL_DEVELOPMENT.md)

---

**Document Status**: ✅ Ready for Implementation  
**Next Review**: After Week 2 (Mid-December 2025)  
**Success Checkpoint**: End of December 2025
