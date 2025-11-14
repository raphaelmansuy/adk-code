# Specification 0001: Claude Code Agent Definition Support

**Version**: 1.1 (Updated with Implementation Reality Check)  
**Date**: 2025-11-14 (Updated)  
**Status**: ⚠️ **Draft - Zero Implementation Exists**  
**Related ADR**: [docs/adr/0001-claude-code-agent-support.md](../adr/0001-claude-code-agent-support.md)  
**Authors**: adk-code Team  
**Implementation Status**: 🔴 **Not Started** (As of Nov 14, 2025)

---

## ⚠️ IMPLEMENTATION STATUS WARNING

**Current Reality (Nov 14, 2025)**:

- **Zero implementation exists** - This is a greenfield project
- **Estimated effort**: 5,000-8,000 lines of new Go code
- **Realistic timeline**: 6+ months (Dec 2025 - Jun 2026)
- **Risk level**: 🔴 HIGH
- **Dependencies**: None blocking, but integration with existing systems required

**What This Spec Describes vs. Reality**:

- ✅ This spec is **comprehensive and well-designed**
- ❌ This spec assumes **8 weeks of team effort**
- ❌ Reality is **1-2 developers with competing priorities**
- ✅ Phased approach is **pragmatic and necessary**

**Recommended Approach**: Implement **Phase 0** (Proof of Concept) first, then re-evaluate scope and timeline based on learnings.

---

## Executive Summary

This specification defines how adk-code will discover, validate, manage, and **independently execute** agent definitions using a format compatible with Claude Code. adk-code implements a superior agent system that uses the same proven file format while adding powerful enhancements like multi-model support, workflows, and advanced testing.

**Strategic Direction**:

- **Format Compatible**: Use Claude Code's YAML + Markdown format for ecosystem interoperability
- **Independent Execution**: Agents execute natively in adk-code using the ADK framework (NO Claude Code dependency)
- **Enhanced Features**: Multi-model support (Gemini, GPT-4, Claude), workflows, metrics, advanced tool configuration
- **Pragmatic Superiority**: Start with proven patterns, execute them better, add missing capabilities

**Key Objectives (Phased)**:

1. **Phase 0** (Dec 2025): Basic agent file discovery and parsing (Proof of Concept)
2. **Phase 1** (Jan-Feb 2026): Complete discovery and validation framework
3. **Phase 2** (Mar-Apr 2026): Agent management and generation tools
4. **Phase 3** (May-Jun 2026): Enhanced features (workflows, metrics, testing)

---

## Table of Contents

1. [Scope and Requirements](#1-scope-and-requirements)
2. [adk-code Enhancements Beyond Claude Code](#2-adk-code-enhancements-beyond-claude-code)
3. [Architecture](#3-architecture)
4. [Data Models](#4-data-models)
5. [Agent Discovery System](#5-agent-discovery-system)
6. [Validation Framework](#6-validation-framework)
7. [Management Tools](#7-management-tools)
8. [Agent Execution Engine](#8-agent-execution-engine)
9. [Testing Framework](#9-testing-framework)
10. [CLI Interface](#10-cli-interface)
11. [Integration Points](#11-integration-points)
12. [Performance Requirements](#12-performance-requirements)
13. [Security Considerations](#13-security-considerations)
14. [Error Handling](#14-error-handling)
15. [Future Extensions](#15-future-extensions)

---

## 1. Scope and Requirements

### 1.1 Functional Requirements

#### F1: Agent Discovery
- **F1.1**: Discover subagents in `.claude/agents/*.md` (project-level)
- **F1.2**: Discover subagents in `~/.claude/agents/*.md` (user-level)
- **F1.3**: Discover skills in `.claude/skills/*/SKILL.md` (project-level)
- **F1.4**: Discover skills in `~/.claude/skills/*/SKILL.md` (user-level)
- **F1.5**: Discover commands in `commands/*.md`
- **F1.6**: Discover plugin agents from `plugin/agents/` directories
- **F1.7**: Parse plugin manifests (`plugin.json`)
- **F1.8**: Maintain discovery cache with invalidation
- **F1.9**: Support recursive discovery in complex projects
- **F1.10**: Handle symlinks correctly

#### F2: YAML/Markdown Parsing
- **F2.1**: Extract YAML frontmatter from agent files
- **F2.2**: Parse and validate required fields (name, description)
- **F2.3**: Parse optional fields (tools, model, allowed-tools, capabilities)
- **F2.4**: Preserve markdown content after frontmatter
- **F2.5**: Handle multi-line YAML values
- **F2.6**: Support comments in YAML
- **F2.7**: Report parse errors with line numbers
- **F2.8**: Handle encoding issues (UTF-8 validation)

#### F3: Agent Validation
- **F3.1**: Validate name field (lowercase, hyphens, max 64 chars)
- **F3.2**: Validate description field (non-empty, max 1024 chars)
- **F3.3**: Validate tool names against known tool list
- **F3.4**: Validate model names (sonnet, opus, haiku, inherit)
- **F3.5**: Validate file structure and permissions
- **F3.6**: Detect duplicate agent names across hierarchy
- **F3.7**: Check for YAML syntax errors
- **F3.8**: Validate skill-specific fields (allowed-tools)
- **F3.9**: Validate plugin manifest schema
- **F3.10**: Generate detailed validation reports with severity levels

#### F4: Best Practices Linting
- **F4.1**: Detect vague descriptions (flag common weak words)
- **F4.2**: Recommend specific, actionable descriptions
- **F4.3**: Warn about overly permissive tool access
- **F4.4**: Suggest minimal tool sets
- **F4.5**: Flag organizational issues
- **F4.6**: Check naming conventions
- **F4.7**: Recommend documentation patterns
- **F4.8**: Provide actionable fix suggestions

#### F5: Agent Management
- **F5.1**: Create new agent files with templates
- **F5.2**: Edit existing agent definitions
- **F5.3**: Delete agents safely (with backups)
- **F5.4**: Export agents to plugin format
- **F5.5**: Import agents from plugins
- **F5.6**: Bulk operations on agent sets
- **F5.7**: Version control integration (git commits)

#### F6: Agent Generation
- **F6.1**: Generate subagent scaffolds
- **F6.2**: Generate skill scaffolds
- **F6.3**: Generate command scaffolds
- **F6.4**: Generate plugin manifests
- **F6.5**: Interactive wizard for agent creation
- **F6.6**: Template library with examples
- **F6.7**: Pre-populate system prompts

### 1.2 Non-Functional Requirements

#### Performance
- **P1**: Discover 100+ agents in <1 second
- **P2**: Validate all agents in project in <500ms
- **P3**: Generate new agent in <100ms
- **P4**: Memory usage <50MB for 1000 agents
- **P5**: Lazy load agent content (parse on demand)
- **P6**: Support projects with 10,000+ files

#### Reliability
- **R1**: Handle corrupted agent files gracefully
- **R2**: Automatic fallback to uncached discovery
- **R3**: No data loss on failed operations
- **R4**: Atomic file writes (all-or-nothing)
- **R5**: Comprehensive error logging
- **R6**: Recovery from interrupted operations

#### Compatibility
- **C1**: Support Claude Code agent format v1.0+
- **C2**: Backward compatible with future Claude Code releases (with warnings)
- **C3**: Work with all adk-code supported platforms (Linux, macOS, Windows)
- **C4**: Support both POSIX and Windows path conventions
- **C5**: Respect `.gitignore` and similar ignore patterns

#### Maintainability
- **M1**: Code coverage >90%
- **M2**: Clear separation of concerns (discovery, validation, management)
- **M3**: Well-documented API contracts
- **M4**: Testable in isolation
- **M5**: Easy to extend for new agent types

---

## 2. adk-code Enhancements Beyond Claude Code

While using Claude Code's proven file format for compatibility, adk-code implements superior execution capabilities and additional features:

### 2.1 Multi-Model Support

**Claude Code Limitation**: Only supports Claude models (Sonnet, Opus, Haiku)

**adk-code Enhancement**: Support any model from multiple providers

```yaml
---
name: multi-model-reviewer
description: Code reviewer that uses the best model for the task
model: gemini-2.5-flash       # or gpt-4o, claude-3.5-sonnet, etc.
provider: google              # google, openai, anthropic
temperature: 0.7
max_tokens: 2048
top_p: 0.95
---
```

**Benefits**:
- Choose optimal model for specific tasks
- Cost optimization (use cheaper models when appropriate)
- Avoid vendor lock-in
- Leverage latest model capabilities

### 2.2 Agent Workflows & Chaining

**Claude Code Limitation**: Single-agent execution only

**adk-code Enhancement**: Define multi-agent workflows with data passing

```yaml
---
name: feature-development-workflow
description: End-to-end feature development with multiple specialized agents
workflow:
  - agent: code-explorer
    description: Analyze codebase structure
    output: feature_analysis
    timeout: 60s
  
  - agent: code-architect
    description: Design feature architecture
    input: $feature_analysis
    output: architecture_plan
    timeout: 120s
  
  - agent: code-implementer
    description: Implement the feature
    input: $architecture_plan
    requires_confirmation: true
---
```

**Benefits**:
- Complex multi-step tasks automated
- Clear separation of concerns
- Resumable workflows (continue from failed step)
- Better debugging and observability

### 2.3 Advanced Tool Configuration

**Claude Code Limitation**: Simple comma-separated tool lists

**adk-code Enhancement**: Per-tool configuration with constraints

```yaml
---
name: safe-code-editor
description: Code editor with safety guardrails
tools:
  - name: Read
    max_files: 10
    allowed_extensions: [.go, .ts, .py]
  
  - name: Edit
    requires_confirmation: true
    confidence_threshold: 0.85
    max_lines_per_edit: 100
  
  - name: Bash
    timeout: 30s
    allowed_commands: [ls, grep, find, git]
    forbidden_commands: [rm, sudo, curl]
    working_directory: ./src
---
```

**Benefits**:
- Fine-grained security control
- Prevent accidental damage
- Enforce organizational policies
- Better auditability

### 2.4 Agent Metrics & Observability

**Claude Code Limitation**: No built-in metrics

**adk-code Enhancement**: Comprehensive metrics and logging

```yaml
---
name: monitored-agent
description: Agent with full observability
metrics:
  enabled: true
  track: [tokens, latency, tool_calls, success_rate, errors]
  export_to: ./metrics/agent-metrics.json
  
logging:
  level: debug
  output: ./logs/{agent_name}-{timestamp}.log
  structured: true
  include_context: true
---
```

**Benefits**:
- Track agent performance
- Debug issues faster
- Optimize token usage
- Compliance and auditing

### 2.5 Agent Composition & Inheritance

**Claude Code Limitation**: No code reuse between agents

**adk-code Enhancement**: Extend and compose agents

```yaml
---
name: senior-code-reviewer
description: Expert code reviewer with security focus
extends: code-reviewer          # Inherit from base agent
additional_tools: 
  - SecurityScan
  - ComplexityAnalysis
  - LicenseChecker
override:
  temperature: 0.5              # More deterministic
  model: claude-3-opus          # More capable model
  max_iterations: 10            # Allow more refinement
---
```

**Benefits**:
- DRY principle for agents
- Easier maintenance
- Consistent patterns across team
- Rapid agent development

### 2.6 Semantic Versioning & Compatibility

**Claude Code Limitation**: No version tracking

**adk-code Enhancement**: Semantic versioning with compatibility checks

```yaml
---
name: versioned-agent
version: "2.1.0"
min_adk_version: "0.3.0"
description: Agent with version tracking
deprecated: false
breaking_changes:
  - "2.0.0: Changed tool access model"
migration_guide: "./docs/migration-v2.md"
---
```

**Benefits**:
- Track agent evolution
- Prevent incompatibilities
- Smooth team upgrades
- Clear deprecation path

### 2.7 Built-in Testing Support

**Claude Code Limitation**: No testing framework

**adk-code Enhancement**: Comprehensive testing with mock LLMs

```yaml
# .adk/agents/tests/code-reviewer.test.yaml
agent: code-reviewer
tests:
  - name: "Security review identifies SQL injection"
    input: "Review this user authentication code"
    mock_files:
      - path: auth.go
        content: |
          func login(username string) {
            query := "SELECT * FROM users WHERE name='" + username + "'"
          }
    expect:
      tools_called: [Read, SecurityScan]
      contains: ["SQL injection", "parameterized query"]
      severity: high
      min_confidence: 0.9
  
  - name: "Performance review catches inefficiency"
    input: "Review database query performance"
    mock_files:
      - path: queries.go
        content: |
          for _, user := range users {
            db.Query("SELECT * FROM orders WHERE user_id = ?", user.ID)
          }
    expect:
      contains: ["N+1 query", "batch query"]
      suggests_fix: true
```

**Benefits**:
- Test agents before deployment
- Regression testing
- CI/CD integration
- Deterministic behavior verification

### 2.8 Conditional Logic & Dynamic Behavior

**Claude Code Limitation**: Static agent definitions

**adk-code Enhancement**: Conditional tool access and dynamic behavior

```yaml
---
name: adaptive-agent
description: Agent that adapts to context
tools:
  - Read: always
  - Edit: when confidence > 0.8
  - Write: when approved_by_user
  - Delete: never
  - Bash: when safe_mode == false

conditions:
  - if: file_size > 1000
    then:
      add_tools: [CodeAnalysis, ComplexityMetrics]
  
  - if: is_production_branch
    then:
      require_review: true
      additional_validation: true
---
```

**Benefits**:
- Context-aware behavior
- Dynamic safety guardrails
- Flexible policy enforcement
- Smarter agent decisions

### 2.9 Integration with adk-code Toolset

**adk-code agents have access to 30+ advanced tools**:
- Code search (semantic and regex)
- Multi-file operations
- Advanced git operations
- MCP server integrations
- Custom tool registration

**Enhanced Tool Definitions**:
```yaml
---
name: poweruser-agent
tools:
  # adk-code specific tools
  - CodeSearch          # Semantic code search
  - MultiFileEdit       # Atomic multi-file changes
  - GitWorkflow         # PR creation, review, merge
  - MCPGitHub           # GitHub API via MCP
  - MCPLinear           # Linear API via MCP
---
```

### 2.10 Key Differentiators Summary

| Feature | Claude Code | adk-code |
|---------|-------------|----------|
| **Models** | Claude only | Multi-provider (Google, OpenAI, Anthropic) |
| **Workflows** | Single agent | Multi-agent workflows with data passing |
| **Tool Config** | Simple list | Per-tool configuration with constraints |
| **Metrics** | None | Comprehensive metrics & logging |
| **Testing** | Manual | Built-in testing framework |
| **Versioning** | None | Semantic versioning with compatibility |
| **Composition** | None | Agent inheritance and composition |
| **Execution** | Claude Code CLI | Native ADK framework (independent) |
| **Dynamic Behavior** | Static | Conditional logic and adaptability |

**Philosophy**: Start with proven patterns (Claude Code's format), execute them better, add missing capabilities pragmatically.

---

## 3. Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│            CLI / Tool Interface Layer               │
│  /agents-list, /agents-validate, /agents-new, ...  │
└────────────────┬────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────┐
│          Agent Management Layer                      │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────┐ │
│  │  Discovery   │  │  Validation  │  │ Generator │ │
│  └──────────────┘  └──────────────┘  └───────────┘ │
└────────────────┬────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────┐
│        Agent Core Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────┐ │
│  │   Parsing    │  │  Data Models │  │ Caching   │ │
│  └──────────────┘  └──────────────┘  └───────────┘ │
└────────────────┬────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────┐
│        File System Layer                             │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────┐ │
│  │   Scanner    │  │   Reader     │  │  Writer   │ │
│  └──────────────┘  └──────────────┘  └───────────┘ │
└─────────────────────────────────────────────────────┘
```

### 2.2 Component Diagram

```
Agent Management System
├── Discovery Engine
│   ├── File System Scanner
│   │   ├── Project Scanner (.claude/agents/, .claude/skills/)
│   │   ├── User Scanner (~/.claude/agents/, ~/.claude/skills/)
│   │   ├── Plugin Scanner (plugin/agents/, plugin/skills/)
│   │   └── Cache Manager
│   └── Hierarchy Resolver
│       └── Priority/Conflict Resolution
│
├── Parsing & Processing
│   ├── Markdown Parser
│   │   ├── YAML Frontmatter Extractor
│   │   └── Content Preservationß
│   ├── Plugin Manifest Parser
│   └── Format Converter
│
├── Validation Engine
│   ├── Syntax Validator
│   │   ├── YAML Validator
│   │   ├── Field Validator
│   │   ├── Tool Validator
│   │   └── Model Validator
│   └── Linter
│       ├── Description Quality Checker
│       ├── Tool Access Analyzer
│       ├── Organization Checker
│       └── Naming Convention Checker
│
├── Management Service
│   ├── Agent Creator
│   │   └── Template Generator
│   ├── Agent Editor
│   ├── Agent Deleter
│   └── Agent Exporter
│
└── Integration Layer
    ├── Claude Code CLI Bridge
    ├── File System Integration
    ├── Session Context Integration
    └── MCP Integration
```

### 2.3 Data Flow

```
User Request
    │
    ├─→ Discovery
    │   ├─→ Scan file system
    │   ├─→ Parse YAML/Markdown
    │   ├─→ Build agent objects
    │   └─→ Cache results
    │
    ├─→ Validation
    │   ├─→ Syntax checks
    │   ├─→ Field validation
    │   ├─→ Tool verification
    │   ├─→ Lint suggestions
    │   └─→ Generate report
    │
    ├─→ Management
    │   ├─→ Create/edit/delete
    │   ├─→ Persist to disk
    │   ├─→ Update cache
    │   └─→ Git integration
    │
    └─→ Output
        ├─→ Format (JSON/Table/YAML)
        ├─→ Stream to terminal
        └─→ Update session context
```

---

## 3. Data Models

### 3.1 Core Agent Definition

```go
// Agent represents any Claude Code agent definition
type Agent struct {
    // Identity
    Name        string `yaml:"name"`        // Unique identifier
    Description string `yaml:"description"` // When and why to use
    
    // Configuration
    Type        AgentType  // "subagent" | "skill" | "command"
    Source      AgentSource // Where it's defined
    Model       string `yaml:"model,omitempty"` // sonnet|opus|haiku|inherit
    Tools       []string `yaml:"tools,omitempty"` // Available tools
    
    // For Skills only
    AllowedTools []string `yaml:"allowed-tools,omitempty"`
    
    // For Plugin Agents only
    Capabilities []string `yaml:"capabilities,omitempty"`
    
    // Content
    Content     string  // Markdown content (preserved)
    RawFrontmatter string // Original YAML frontmatter
    
    // Metadata
    Path        string  // File system path
    ModTime     time.Time // Last modified
    Size        int64   // File size bytes
    PluginInfo  *PluginInfo // If from plugin
    Metadata    map[string]interface{} // Extra fields
    
    // Parsed fields
    ParsedTools   map[string]bool // Tools lookup
    IsValid       bool
    ValidationErr error
}

type AgentType string
const (
    TypeSubagent AgentType = "subagent"
    TypeSkill    AgentType = "skill"
    TypeCommand  AgentType = "command"
    TypePlugin   AgentType = "plugin"
)

type AgentSource string
const (
    SourceProject AgentSource = "project"
    SourceUser    AgentSource = "user"
    SourcePlugin  AgentSource = "plugin"
    SourceCLI     AgentSource = "cli"
)

type PluginInfo struct {
    Name    string
    Version string
    Path    string
}
```

### 3.2 Validation Result

```go
type ValidationResult struct {
    // Summary
    Valid       bool
    AgentName   string
    AgentPath   string
    
    // Issues
    Errors      []ValidationError
    Warnings    []ValidationWarning
    Suggestions []LintSuggestion
    
    // Metadata
    CheckedAt   time.Time
    Duration    time.Duration
}

type ValidationError struct {
    Field    string // Which field
    Message  string // What's wrong
    Severity string // "error" | "warning" | "info"
    FixHint  string // How to fix
    Line     int    // Line number in file
}

type ValidationWarning struct {
    Category string // Type of warning
    Message  string // Warning description
    Severity string // "high" | "medium" | "low"
    Context  string // Surrounding code
}

type LintSuggestion struct {
    Rule     string // Which lint rule
    Message  string // What to improve
    Current  string // Current value
    Suggested string // Suggested value
    Example  string // Example of good practice
}
```

### 3.3 Plugin Manifest

```go
type PluginManifest struct {
    // Required
    Name        string `json:"name"`
    Version     string `json:"version"`
    
    // Metadata
    Description string `json:"description,omitempty"`
    Author      *Author `json:"author,omitempty"`
    Homepage    string `json:"homepage,omitempty"`
    Repository  string `json:"repository,omitempty"`
    License     string `json:"license,omitempty"`
    Keywords    []string `json:"keywords,omitempty"`
    
    // Components (paths or inline objects)
    Commands    interface{} `json:"commands,omitempty"`   // string|[]string|object
    Agents      interface{} `json:"agents,omitempty"`     // string|[]string|object
    Skills      interface{} `json:"skills,omitempty"`     // string|[]string|object
    Hooks       interface{} `json:"hooks,omitempty"`      // string|object
    McpServers  interface{} `json:"mcpServers,omitempty"` // string|object
}

type Author struct {
    Name  string `json:"name"`
    Email string `json:"email,omitempty"`
    URL   string `json:"url,omitempty"`
}
```

### 3.4 Discovery Results

```go
type DiscoveryResult struct {
    // Agents organized by type
    Subagents []*Agent
    Skills    []*Agent
    Commands  []*Agent
    
    // Agents organized by source
    BySource map[AgentSource][]*Agent
    
    // Lookup maps
    ByName map[string]*Agent // Latest version per name
    ByPath map[string]*Agent // Direct path lookup
    
    // Metadata
    Total          int
    TimeTaken      time.Duration
    CacheHit       bool
    ErrorCount     int
    WarningCount   int
    
    // Issues found
    Errors   []error
    Warnings []string
}
```

---

## 4. Agent Discovery System

### 4.1 Discovery Paths

The system discovers agents from four locations with strict priority:

#### Project Level (Highest Priority)
```
.adk/
├── agents/
│   ├── code-reviewer.md
│   ├── debugger.md
│   └── data-analyst.md
├── skills/
│   ├── pdf-processor/
│   │   ├── SKILL.md
│   │   ├── reference.md
│   │   └── scripts/
│   └── code-analysis/
│       └── SKILL.md
└── commands/
    ├── deploy.md
    ├── status.md
    └── test.md

# For migration/compatibility: can also read .claude/ directories
.claude/
├── agents/
│   └── (imported on first run, copied to .adk/)
└── skills/
    └── (imported on first run, copied to .adk/)
```

#### User Level
```
~/.adk/
├── agents/
│   ├── personal-agent.md
│   └── ...
└── skills/
    ├── personal-skill/
    │   └── SKILL.md
    └── ...

# For compatibility
~/.claude/
├── agents/
│   └── (can be imported with /agents-import)
└── skills/
    └── ...
```

#### Plugin Level
```
plugin-name/
├── agents/
│   ├── specialized-agent.md
│   └── ...
├── skills/
│   └── skill-name/
│       └── SKILL.md
├── commands/
│   ├── custom-command.md
│   └── ...
└── .adk-plugin/
    └── plugin.json    # adk-code plugin manifest
```

#### CLI Level (Lowest Priority)
```bash
adk-code --agents '{
  "agent-name": {
    "description": "Expert code reviewer for Go",
    "prompt": "You are an expert Go code reviewer...",
    "tools": ["Read", "Grep", "Glob"],
    "model": "gemini-2.5-flash",
    "provider": "google"
  }
}'
```

### 4.2 Discovery Algorithm

```
DISCOVER(workspace_path):
  agents = []
  
  // 1. Project-level agents (highest priority)
  project_agents = SCAN_DIRECTORY(workspace_path/.claude/agents/)
  project_agents += SCAN_DIRECTORY(workspace_path/.claude/skills/)
  project_agents += SCAN_DIRECTORY(workspace_path/commands/)
  PROCESS(project_agents, source=PROJECT)
  agents.append(project_agents)
  
  // 2. User-level agents
  user_agents = SCAN_DIRECTORY(~/.claude/agents/)
  user_agents += SCAN_DIRECTORY(~/.claude/skills/)
  // Filter out duplicates (project agents override)
  user_agents = FILTER_DUPLICATES(user_agents, agents)
  PROCESS(user_agents, source=USER)
  agents.append(user_agents)
  
  // 3. Plugin agents
  for plugin in FIND_PLUGINS(workspace_path):
    plugin_agents = SCAN_DIRECTORY(plugin/agents/)
    plugin_agents += SCAN_DIRECTORY(plugin/skills/)
    plugin_agents += SCAN_DIRECTORY(plugin/commands/)
    // Filter out duplicates
    plugin_agents = FILTER_DUPLICATES(plugin_agents, agents)
    PROCESS(plugin_agents, source=PLUGIN, plugin_info=plugin)
    agents.append(plugin_agents)
  
  // 4. CLI agents (if provided)
  if cli_agents provided:
    cli_agents = PARSE_CLI_AGENTS(cli_agents)
    cli_agents = FILTER_DUPLICATES(cli_agents, agents)
    agents.append(cli_agents)
  
  SORT(agents, by=NAME)
  return agents

PROCESS(agents, source, plugin_info=nil):
  for agent in agents:
    agent.source = source
    agent.plugin_info = plugin_info
    PARSE_AGENT_FILE(agent)
    agent.type = DETECT_AGENT_TYPE(agent.path, agent.content)
```

### 4.3 Caching Strategy

```go
type Cache struct {
    // Cache structure
    Agents          map[string]*Agent
    LastUpdate      time.Time
    FileWatches     map[string]time.Time
    
    // Cache control
    TTL             time.Duration       // Default: 5 minutes
    MaxSize         int64               // Bytes
    InvalidateOn    []string            // Paths to watch
    
    // Statistics
    Hits            uint64
    Misses          uint64
    LastGC          time.Time
    
    // Methods
    Get(name string) (*Agent, bool)
    Set(agent *Agent)
    Invalidate(path string)
    Clear()
    Touch()
}

// Cache invalidation triggers:
// - File system changes (monitored paths)
// - Explicit invalidation calls
// - TTL expiration
// - Workspace change detection
```

### 4.4 Detection Algorithm

```
DETECT_AGENT_TYPE(path, content):
  if path contains "SKILL.md":
    return TYPE_SKILL
  else if path contains "commands/" and path ends with ".md":
    return TYPE_COMMAND
  else if content has "allowed-tools" in frontmatter:
    return TYPE_SKILL
  else if content has "capabilities" in frontmatter:
    return TYPE_PLUGIN_AGENT
  else if path contains "agents/":
    return TYPE_SUBAGENT
  else:
    return TYPE_UNKNOWN

DETECT_SOURCE(path):
  if path starts with workspace/.claude/:
    return SOURCE_PROJECT
  else if path starts with ~/.claude/:
    return SOURCE_USER
  else if path inside plugin directory:
    return SOURCE_PLUGIN
  else:
    return SOURCE_CLI
```

---

## 5. Validation Framework

### 5.1 Validation Pipeline

```
Input Agent File
    │
    ├─→ [1. Syntax Validation]
    │   ├─→ YAML parsing
    │   ├─→ Markdown structure
    │   ├─→ Encoding check (UTF-8)
    │   └─→ Line endings (CRLF vs LF)
    │
    ├─→ [2. Field Validation]
    │   ├─→ Required fields (name, description)
    │   ├─→ Field type validation
    │   ├─→ Field length limits
    │   ├─→ Character restrictions
    │   └─→ Format validation
    │
    ├─→ [3. Reference Validation]
    │   ├─→ Tool name verification
    │   ├─→ Model name verification
    │   ├─→ Duplicate detection
    │   └─→ Path resolution
    │
    ├─→ [4. Lint Analysis]
    │   ├─→ Description quality
    │   ├─→ Tool access review
    │   ├─→ Organization check
    │   ├─→ Naming convention
    │   └─→ Documentation completeness
    │
    └─→ Validation Report
        ├─→ Errors (blocking)
        ├─→ Warnings (non-blocking)
        ├─→ Suggestions (improvements)
        └─→ Overall verdict
```

### 5.2 Validation Rules

#### R1: Name Field
```
Rule: Name must be unique, lowercase identifier
Pattern: ^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$
Max Length: 64 characters
Error: "Invalid name: must be lowercase letters, digits, hyphens only"
Example: "code-reviewer", "debug-agent-v2", "pdf-processor"
```

#### R2: Description Field
```
Rule: Description must explain purpose and triggers
Min Length: 10 characters
Max Length: 1024 characters
Required: Yes
Error: "Description must be non-empty"
Warning: "Description too short (should explain when to use)"
Example: "Expert code reviewer. Use proactively after code changes to identify bugs and security issues."
```

#### R3: Tools Field
```
Rule: Tools must be comma-separated, valid tool names
Valid Values: [30+ known tools + MCP tools]
Format: "Tool1, Tool2, Tool3"
Error: "Unknown tool: 'InvalidTool'"
Warning: "Tool access seems too permissive (20+ tools)"
Example: "Read, Grep, Glob, Bash"
```

#### R4: Model Field
```
Rule: Model must be valid Claude model reference
Valid Values: "sonnet", "opus", "haiku", "inherit" or full model name
Default: "sonnet" (if omitted)
Error: "Unknown model: 'invalid-model'"
Example: "sonnet", "opus", "inherit"
```

#### R5: Allowed-Tools (Skills only)
```
Rule: Restrict tool access for security-sensitive skills
Format: Same as Tools field
Optional: Yes
Error: "Unknown tool in allowed-tools"
Example: "Read, Grep, Glob"
```

#### R6: File Structure
```
Rule: Agent files must be in correct directories
Project Agents: .claude/agents/*.md or .claude/skills/*/SKILL.md
User Agents: ~/.claude/agents/*.md or ~/.claude/skills/*/SKILL.md
Commands: commands/*.md
Error: "Agent file in wrong directory"
Warning: "Consider organizing agents into .claude/agents/"
```

#### R7: YAML Syntax
```
Rule: Must be valid YAML with proper frontmatter
Structure:
  ---
  name: agent-name
  description: Description text
  tools: Tool1, Tool2  (optional)
  model: sonnet        (optional)
  ---
  
  Markdown content...

Error: "Invalid YAML syntax at line N"
Example: Missing closing ---, invalid indentation, tabs instead of spaces
```

### 5.3 Lint Rules

#### L1: Vague Descriptions
```
Rule: Flag descriptions using weak words
Weak Words: ["helps", "tools", "stuff", "things", "data", "works", "does"]
Pattern: Match word boundaries
Action: Suggest specific, action-oriented description
Example:
  Bad: "Helps with code"
  Good: "Reviews code for security issues and best practices"
Severity: Medium
```

#### L2: Tool Access
```
Rule: Warn about overly permissive tool lists
Trigger: 20+ tools in single agent
Action: Suggest minimal set
Severity: Medium
Example:
  Flag: All 30+ tools granted
  Suggest: "Consider limiting to Read, Edit, Bash for this use case"
```

#### L3: Description Quality
```
Rule: Check description completeness
Triggers:
  - No action verbs (should start with verb: "Review", "Analyze", "Debug")
  - No "when to use" guidance
  - No specific keywords for discovery
  - Too long (>500 chars suspicious)
Severity: Low
Action: Suggest improvements
```

#### L4: Naming Convention
```
Rule: Enforce consistent naming
Pattern: [a-z0-9](-[a-z0-9]+)*
Actions:
  - Suggest kebab-case for multi-word names
  - Flag inconsistent patterns in project
  - Warn about generic names ("agent", "tool", "helper")
Severity: Low
```

#### L5: Organization
```
Rule: Check file/directory organization
Triggers:
  - Agent files at wrong level
  - Inconsistent skill directory structure
  - Missing supporting files (README, examples)
Severity: Low
```

### 5.4 Known Tool List

```go
var KnownTools = []string{
    // File Operations
    "Read", "Edit", "Write", "Delete", "Glob", "Find",
    
    // Execution
    "Bash", "Run", "Execute",
    
    // Search
    "Grep", "Glob", "CodeSearch",
    
    // Git
    "Git", "GitCommit", "GitPush", "GitPR",
    
    // Development
    "Build", "Test", "Lint", "Format",
    
    // Other
    "Bash", "Shell", "Terminal",
    
    // MCP tools (dynamic) - prefixed with "mcp-"
    "mcp-*", // Wildcard for MCP server tools
}

// MCP tools are discovered at runtime from configured servers
// Format: "mcp-<server-name>-<capability>"
// Examples: "mcp-github-list-issues", "mcp-jira-create-ticket"
```

---

## 6. Management Tools

### 6.1 Agent Creation

#### Input Schema
```go
type CreateAgentRequest struct {
    // Required
    Name        string
    Description string
    Type        AgentType // "subagent" | "skill" | "command"
    
    // Optional
    Tools       []string
    Model       string
    Content     string
    SkillDir    string // For skills
    
    // Generation hints
    Purpose     string   // What should it do?
    Triggers    []string // When to use?
}
```

#### Creation Process
```
CREATE_AGENT(request):
  // 1. Validate input
  if not VALID(request):
    return ERROR
  
  // 2. Generate content if not provided
  if request.content is empty:
    request.content = GENERATE_TEMPLATE(request)
  
  // 3. Detect target location
  target_path = DETERMINE_PATH(request.type, request.name)
  
  // 4. Check for conflicts
  if FILE_EXISTS(target_path):
    return ERROR("Agent already exists")
  
  // 5. Create file
  WRITE_FILE(target_path, SERIALIZE(request))
  
  // 6. Validate created agent
  validation = VALIDATE(target_path)
  if not validation.valid:
    DELETE_FILE(target_path)
    return ERROR("Generated agent failed validation")
  
  // 7. Return result
  return SUCCESS(target_path)
```

#### Output Schema
```go
type CreateAgentResponse struct {
    Success bool
    Path    string
    Message string
    Agent   *Agent
}
```

### 6.2 Agent Editing

#### Input Schema
```go
type EditAgentRequest struct {
    // Target
    Name   string
    Source AgentSource // Which hierarchy
    
    // Changes
    Updates map[string]interface{} // Field changes
    Content string                  // Markdown content (optional)
}
```

#### Editing Process
```
EDIT_AGENT(request):
  // 1. Find agent
  agent = FIND_AGENT(request.name, request.source)
  if not agent:
    return ERROR("Agent not found")
  
  // 2. Validate updates
  for field, value in request.updates:
    if not VALIDATE_FIELD(field, value):
      return ERROR("Invalid value for field")
  
  // 3. Backup original
  backup_path = BACKUP(agent.path)
  
  // 4. Apply changes
  agent = MERGE(agent, request.updates)
  if request.content:
    agent.content = request.content
  
  // 5. Validate modified agent
  validation = VALIDATE(agent)
  if not validation.valid:
    RESTORE(backup_path)
    return ERROR("Changes failed validation")
  
  // 6. Persist changes
  WRITE_FILE(agent.path, SERIALIZE(agent))
  INVALIDATE_CACHE(agent.path)
  
  // 7. Return result
  return SUCCESS(agent)
```

### 6.3 Agent Deletion

#### Input Schema
```go
type DeleteAgentRequest struct {
    Name   string
    Source AgentSource
    Force  bool // Skip confirmation
}
```

#### Deletion Process
```
DELETE_AGENT(request):
  // 1. Find agent
  agent = FIND_AGENT(request.name, request.source)
  if not agent:
    return ERROR("Agent not found")
  
  // 2. Backup first (safety)
  backup_path = BACKUP(agent.path)
  
  // 3. Delete file
  DELETE_FILE(agent.path)
  INVALIDATE_CACHE(agent.path)
  
  // 4. Return result with backup location
  return SUCCESS(backup_path)
```

### 6.4 Agent Export

#### Input Schema
```go
type ExportAgentRequest struct {
    Name   string
    Format string // "plugin" | "standalone" | "archive"
}
```

#### Export Process
```
EXPORT_AGENT(request):
  // 1. Find agent
  agent = FIND_AGENT(request.name)
  
  // 2. Validate exportability
  if request.format == "plugin":
    if not agent.meets_plugin_requirements():
      return ERROR("Agent not suitable for plugin")
  
  // 3. Package agent
  package = PACKAGE(agent, request.format)
  
  // 4. Validate package
  if not VALID_PACKAGE(package):
    return ERROR("Package validation failed")
  
  // 5. Return result
  return SUCCESS(package)
```

### 6.5 Agent Templates

#### Subagent Template
```markdown
---
name: my-agent
description: [Brief description of what this agent does and when to use it]
tools: [Tool1, Tool2, Tool3]  # Optional: specify available tools
model: sonnet                  # Optional: sonnet|opus|haiku|inherit
---

# My Agent

## Role and Purpose

[Clear explanation of the agent's role and expertise area.]

## Capabilities

[List of things this agent can do.]

## When to Use

[Specific scenarios where this agent should be invoked.]

## Instructions

[Step-by-step guidance for the agent's behavior.]

## Examples

[Real-world examples of how to use this agent.]
```

#### Skill Template
```markdown
---
name: my-skill
description: [What this skill does and when Claude should use it]
allowed-tools: Read, Grep, Glob  # Optional: restrict tool access
---

# My Skill

## Purpose

[Explain the skill's purpose and domain.]

## When to Use

[Triggers and scenarios for this skill.]

## Instructions

[How to perform this skill.]

## Examples

[Example usage scenarios.]

## Requirements

[Any dependencies or prerequisites.]
```

#### Command Template
```markdown
---
name: my-command
description: [What this command does]
---

# My Command

## Purpose

[Clear description of command purpose.]

## Usage

[How to invoke this command.]

## Examples

[Example invocations and expected results.]
```

---

## 7. CLI Interface

### 7.1 Command Specifications

#### `/agents` - Interactive Agent Browser
```bash
Usage: /agents [options]

Options:
  --filter [all|project|user|plugin]  Filter by source
  --type [agent|skill|command|all]    Filter by type
  --sort [name|type|modified]         Sort order
  --search <pattern>                  Search agents

Features:
  - Interactive menu (arrow keys, Enter)
  - Real-time search
  - One-key actions (v=view, e=edit, d=delete)
  - Pagination for large lists

Output:
  ┌─ Agents (Project) ─────────────────────────┐
  │ > code-reviewer [subagent]  (modified 1h)  │
  │   debugger [subagent]       (modified 1d)  │
  │   pdf-processor [skill]     (modified 1w)  │
  │                                             │
  │ Press 'e' to edit, 'd' to delete, 'v' view │
  └─────────────────────────────────────────────┘
```

#### `/agents-list` - List All Agents
```bash
Usage: /agents-list [options]

Options:
  --filter [all|project|user|plugin]  Filter by source
  --type [all|agent|skill|command]    Filter by type
  --sort [name|type|modified]         Sort order
  --format [table|json|yaml]          Output format

Output (table):
┌─────────────────┬──────────┬────────────────────────────────┐
│ Name            │ Type     │ Description                    │
├─────────────────┼──────────┼────────────────────────────────┤
│ code-reviewer   │ subagent │ Reviews code for quality       │
│ debugger        │ subagent │ Debugs errors and issues       │
│ pdf-processor   │ skill    │ Extract text from PDFs         │
└─────────────────┴──────────┴────────────────────────────────┘
```

#### `/agents-validate` - Validate All Agents
```bash
Usage: /agents-validate [options]

Options:
  --fix                      Auto-fix common issues
  --strict                   Fail on warnings
  --agent <name>            Validate specific agent

Output:
✅ code-reviewer.md (project)
⚠️  debugger.md - Description too vague
❌ pdf-processor.md - Line 5: Invalid YAML
   Fix: Check indentation around "allowed-tools"

Summary: 2 valid, 1 warning, 1 error
```

#### `/agents-new` - Create New Agent
```bash
Usage: /agents-new [type] [name] [options]

Arguments:
  type                      subagent|skill|command
  name                      Agent name (optional, prompted if missing)

Options:
  --description <text>      Agent description
  --tools <list>           Comma-separated tools
  --model <name>           Model (sonnet|opus|haiku)
  --generate               AI-generate system prompt

Interactive Mode:
  ? What type of agent? (subagent)
  ? Agent name? (code-reviewer)
  ? Short description? 
  > Expert code reviewer for quality and security
  ? What tools does it need? (Read, Grep, Glob, Bash)
  ? Which model? (sonnet)
  
  Generated file: .claude/agents/code-reviewer.md
  ✅ Validation passed
```

#### `/agents-edit` - Edit Agent
```bash
Usage: /agents-edit <name> [options]

Options:
  --description <text>      Update description
  --tools <list>           Update tools
  --model <name>           Update model
  --open                   Open in $EDITOR

Example:
adk-code /agents-edit code-reviewer --tools Read,Grep,Glob,Bash

Backup created: .claude/agents/.backups/code-reviewer.md.20251114-120000
File updated: .claude/agents/code-reviewer.md
✅ Validation passed
```

#### `/agents-delete` - Delete Agent
```bash
Usage: /agents-delete <name> [options]

Options:
  --force                  Skip confirmation
  --keep-backup           Keep backup file

Example:
adk-code /agents-delete old-agent

⚠️  Delete 'old-agent' from project agents?
Backup will be saved to: .claude/agents/.backups/old-agent.md
[y/N]: y

✅ Deleted: .claude/agents/old-agent.md
Backup: .claude/agents/.backups/old-agent.md.20251114-120000
```

#### `/agents-describe` - Show Agent Details
```bash
Usage: /agents-describe <name> [options]

Options:
  --full                   Show full content
  --tools                  Show tool list
  --validation            Show validation report

Output:
Agent: code-reviewer
Source: Project (.claude/agents/)
Type: Subagent
Model: sonnet

Description:
  Expert code reviewer for security, quality, and maintainability.
  Use immediately after writing or modifying code.

Tools:
  - Read     (read files)
  - Grep     (search content)
  - Glob     (find files)
  - Bash     (execute commands)

Modified: 1 day ago
Size: 2.5 KB
```

#### `/agents-lint` - Run Lint Checks
```bash
Usage: /agents-lint [options]

Options:
  --strict               Fail on warnings
  --agent <name>       Lint specific agent
  --fix                Auto-fix issues

Output:
Linting agents...

code-reviewer.md:
  ✅ PASS (all checks)

debugger.md:
  ⚠️  L002: Description missing "when to use" guidance
  ⚠️  L003: Tool access includes 10+ tools
      Suggestion: Consider limiting to essential tools

pdf-processor.md:
  ⚠️  L001: Description uses vague word "helps"
      Current:  "Helps with PDF processing"
      Better:   "Extract text, fill forms, merge PDFs"

Summary: 1 pass, 2 warnings
```

#### `/agents-export` - Export Agent
```bash
Usage: /agents-export <name> [options]

Options:
  --format [plugin|standalone]  Export format
  --output <path>              Output file/directory

Example:
adk-code /agents-export code-reviewer --format plugin

Generated plugin structure:
code-reviewer-plugin/
├── .claude-plugin/
│   └── plugin.json
├── agents/
│   └── code-reviewer.md
├── README.md
└── LICENSE

Saved to: ./code-reviewer-plugin/
```

### 7.2 Error Messages and Guidance

#### Validation Errors
```
❌ Error: Invalid YAML in debugger.md at line 5
   Problem: Invalid indentation in tools field
   Context:
     4 | tools: Read, Grep,
     5 |   Glob, Bash
     6 | model: sonnet
     
   Fix: Use comma-separated list on one line:
     tools: Read, Grep, Glob, Bash

❌ Error: Unknown tool "InvalidTool" in code-reviewer.md
   Problem: "InvalidTool" is not a valid tool
   Available tools: Read, Edit, Write, Delete, Bash, Grep, Glob, ...
   
   Fix: Use a valid tool name from the list above
```

#### Lint Warnings
```
⚠️  Warning L001: Vague description in debugger.md
   Current: "Helps debug issues"
   Problem: Uses weak word "helps", no specific guidance
   
   Better: "Debugs errors and test failures. Use when encountering bugs 
            to identify root causes and implement fixes."
   
   Learn more: /agents-help description-quality
```

#### User Confirmation
```
⚠️  Warning: This action will delete agent 'old-agent'
   
   Backup will be saved to:
   .claude/agents/.backups/old-agent.md.20251114-120000
   
   Continue? [y/N]:
```

---

## 8. Integration Points

### 8.1 Session Integration

**Agent Context in LLM Calls**:
```go
type SessionContext struct {
    // Existing fields...
    
    // Agent discovery results
    DiscoveredAgents *DiscoveryResult
    
    // Available agents for reference
    AgentRegistry map[string]*Agent
    
    // Agent capabilities for system prompt
    AgentCapabilities []string
}

// In system prompt, optionally include:
// "Available agents you can delegate to:
//  - code-reviewer: Review code for quality and security
//  - debugger: Debug errors and test failures
//  - pdf-processor: Extract text from PDF files"
```

### 8.2 Tool Integration

**Agent-Aware Tool Execution**:
```go
// Tools can reference agents in their operations
type ToolContext struct {
    AvailableAgents []*Agent
    CurrentAgent    *Agent  // If tool called by agent
}

// Example: When /agents-new is called
// Tool has access to Agent registry for validation
```

### 8.3 File System Integration

**Git Integration**:
```bash
# When agents are created/modified, can optionally:
git add .claude/agents/new-agent.md
git commit -m "feat: Add new-agent for code review"

# Tool: /agents-create with --commit flag
adk-code /agents-new code-reviewer --commit
```

**File Watching**:
```go
// Optional: Watch for external agent changes
type FileWatcher interface {
    Watch(path string, handler func(*Agent)) error
}

// Trigger cache invalidation on external edits
```

### 8.4 MCP Integration

**Expose Agent System as MCP Resource**:
```json
{
  "uri": "agents://adk-code/discovery",
  "type": "agents",
  "capabilities": ["list", "validate", "create", "edit"]
}

{
  "uri": "agents://adk-code/subagents",
  "type": "list",
  "items": [
    {
      "name": "code-reviewer",
      "description": "...",
      "tools": ["Read", "Grep"]
    }
  ]
}
```

### 8.5 Claude Code Integration

**Bridge Commands**:
```bash
# Test agent in Claude Code CLI
adk-code /agents-test code-reviewer

# Install agent to Claude Code
adk-code /agents-install code-reviewer

# Sync agents between adk-code and Claude Code
adk-code /agents-sync --direction both
```

---

## 9. Testing Strategy

### 9.1 Unit Tests

#### Discovery Tests
```go
func TestDiscoverAgents(t *testing.T) {
    // Test discovering agents in different locations
    // Test hierarchy/priority resolution
    // Test caching behavior
    // Test file format variations
    // Test error handling
}

func TestDetectAgentType(t *testing.T) {
    // Test subagent detection
    // Test skill detection
    // Test command detection
    // Test edge cases
}
```

#### Parsing Tests
```go
func TestParseYAMLFrontmatter(t *testing.T) {
    // Test valid YAML
    // Test invalid YAML
    // Test edge cases (multiline strings, special chars)
}

func TestParseAgent(t *testing.T) {
    // Test complete agent parsing
    // Test field extraction
    // Test content preservation
}
```

#### Validation Tests
```go
func TestValidateAgent(t *testing.T) {
    // Test name validation
    // Test description validation
    // Test tool validation
    // Test model validation
    // Test skill-specific fields
}

func TestLintAgent(t *testing.T) {
    // Test vague description detection
    // Test tool access analysis
    // Test naming conventions
}
```

### 9.2 Integration Tests

```go
func TestCreateAndValidateAgent(t *testing.T) {
    // Create agent via API
    // Verify file created correctly
    // Validate created agent
    // Discover and verify
}

func TestEditAndExport(t *testing.T) {
    // Create agent
    // Edit via API
    // Export to plugin format
    // Verify plugin structure
}
```

### 9.3 Acceptance Tests

#### Test Data
```
fixtures/
├── valid-agents/
│   ├── simple-subagent.md
│   ├── skill-with-allowed-tools.md
│   └── plugin-agent.md
├── invalid-agents/
│   ├── bad-yaml.md
│   ├── missing-description.md
│   └── unknown-tool.md
├── complex-project/
│   ├── .claude/
│   │   ├── agents/
│   │   │   ├── agent1.md
│   │   │   ├── agent2.md
│   │   │   └── ...
│   │   └── skills/
│   │       └── skill1/SKILL.md
│   └── commands/
│       └── cmd1.md
└── plugins/
    ├── plugin-a/
    │   ├── agents/
    │   ├── skills/
    │   └── .claude-plugin/plugin.json
    └── plugin-b/
```

### 9.4 Performance Tests

```go
func BenchmarkDiscovery(b *testing.B) {
    // Benchmark discovery on large projects
    // Test with 1000+ agent files
    // Verify <1 second discovery time
}

func BenchmarkValidation(b *testing.B) {
    // Benchmark validation on all agents
    // Verify <500ms validation time
}
```

### 9.5 Compatibility Tests

```
Compatibility Matrix:
├── Claude Code v1.0+
├── All supported OS (Linux, macOS, Windows)
├── All supported Go versions (1.24+)
└── All adk-code supported LLM providers
```

---

## 10. Performance Requirements

### 10.1 Latency Targets

| Operation | Target | Notes |
|-----------|--------|-------|
| Discovery (100 agents) | <1s | With caching |
| Discovery (cold start) | <2s | First run, no cache |
| Validation (all agents) | <500ms | After discovery |
| Create agent | <100ms | Template + write |
| Edit agent | <150ms | Read + update + write |
| Export agent | <200ms | Packaging + serialization |

### 10.2 Memory Targets

| Scenario | Target | Notes |
|----------|--------|-------|
| 100 agents in memory | <10MB | Parsed structures |
| 1000 agents in memory | <50MB | Cached + metadata |
| Discovery operation | <100MB peak | Temporary allocations |

### 10.3 Optimization Strategies

**Discovery Optimization**:
```go
// 1. Parallel file scanning
// 2. Lazy parsing (parse on demand)
// 3. Aggressive caching (TTL=5min)
// 4. Early termination (stop at first match)
// 5. Index rebuild on interval

type OptimizedDiscovery struct {
    // Parallel workers
    Workers     int = 4
    
    // Lazy loading
    LazyParse   bool = true
    
    // Caching
    CacheTTL    time.Duration = 5 * time.Minute
    
    // Batch operations
    BatchSize   int = 100
}
```

**Memory Optimization**:
```go
// 1. String interning for common values
// 2. Shared tool list references
// 3. Pool allocations for temporary objects
// 4. LRU cache for large projects
// 5. Streaming for large files

type MemoryOptimized struct {
    StringPool   *sync.Pool  // Intern strings
    ToolsCache   []*Tool     // Shared reference
    AgentPool    *sync.Pool  // Agent object pool
    LRUCache     *lru.Cache  // Limited size
}
```

---

## 11. Security Considerations

### 11.1 File Security

**Path Validation**:
```go
// Prevent directory traversal attacks
func ValidatePath(path string) error {
    abs, _ := filepath.Abs(path)
    
    // Must be within workspace or home
    if !isWithinWorkspace(abs) && !isWithinHome(abs) {
        return ErrPathOutOfBounds
    }
    
    // No symbolic link exploits
    if isSymlink(abs) {
        return ErrSymlinkNotAllowed
    }
    
    return nil
}
```

**File Permissions**:
```
Agent files should be:
- Readable by user/group
- Writable by user only
- Executable: NO

Example: -rw-r--r-- (644)
```

### 11.2 Content Validation

**No Code Execution from Agent Files**:
```go
// Agent markdown content is treated as data only
// System prompts are NOT eval'd
// Tool access is validated against whitelist
```

**Sandbox Agent Prompts**:
```
If agents include shell commands in prompts:
- Warn user about security implications
- Don't execute without explicit approval
- Log all agent command execution
```

### 11.3 Plugin Security

**Plugin Manifest Validation**:
```go
// Verify plugin structure
// Validate file paths (no absolute paths)
// Check for symlink attacks
// Verify permissions on scripts
```

**MCP Tool Whitelist**:
```
When agents reference MCP tools:
- Verify MCP server is trusted
- Validate tool names against server capabilities
- Check MCP server signature/hash
```

---

## 12. Error Handling

### 12.1 Error Categories

```go
type AgentError struct {
    Category ErrorCategory
    Message  string
    Details  string
    Line     int      // Line number if applicable
    Severity string   // "error" | "warning" | "info"
    Fix      string   // How to fix
}

type ErrorCategory string
const (
    ErrNotFound       ErrorCategory = "not_found"
    ErrInvalidYAML    ErrorCategory = "invalid_yaml"
    ErrInvalidField   ErrorCategory = "invalid_field"
    ErrUnknownTool    ErrorCategory = "unknown_tool"
    ErrDuplicate      ErrorCategory = "duplicate"
    ErrPermission     ErrorCategory = "permission"
    ErrPathInvalid    ErrorCategory = "path_invalid"
    ErrInternal       ErrorCategory = "internal"
)
```

### 12.2 Error Recovery

```go
// Automatic recovery strategies:

1. File Parsing Errors:
   - Retry with different encoding (UTF-8 → UTF-8-BOM)
   - Skip unparseable files
   - Report detailed error with line/col

2. Discovery Errors:
   - Continue scanning other paths
   - Report partial results
   - Log errors for debugging

3. Validation Errors:
   - Continue validating other agents
   - Provide detailed error report
   - Suggest fixes

4. Write Errors:
   - Atomic writes (temp file + rename)
   - Automatic backup on modification
   - Rollback on validation failure
```

### 12.3 User-Facing Error Messages

```
Clear, actionable error format:

❌ Error: [Category] - [Concise message]
   Problem: [Detailed explanation]
   Location: [File:Line]
   Solution: [How to fix]
   
Example:
❌ Error: Invalid YAML - agents/code-reviewer.md:5
   Problem: Unexpected indentation in tools field
   Location: Line 5, column 3
   Solution: Move "Glob, Bash" to same line as "Read, Grep,"
   
   Current:
     4 | tools: Read, Grep,
     5 |   Glob, Bash
   
   Fixed:
     4 | tools: Read, Grep, Glob, Bash
```

---

## 13. Future Extensions

### 13.1 Phase 2: Analytics & Governance

**Features**:
- Agent usage tracking
- Team agent governance policies
- Approval workflows for new agents
- Agent performance metrics
- Deprecation warnings

### 13.2 Phase 3: GUI Editor

**Features**:
- VS Code extension for agent editing
- Visual tool selector
- Description quality checker (real-time)
- Live validation
- Template builder

### 13.3 Phase 4: Marketplace

**Features**:
- Agent marketplace discovery
- Rating/review system
- Version management
- Dependency resolution
- Automated testing in marketplace

### 13.4 Phase 5: Advanced Features

**Features**:
- AI-assisted agent generation
- Multi-agent orchestration
- Agent execution tracing
- Performance profiling
- Agent composition/templates

---

## 14. Implementation Roadmap

### 14.1 Phase 1: Foundation (Weeks 1-3)

**Goal**: Core agent discovery, parsing, and basic validation

**Deliverables**:
```
pkg/agents/
├── discovery.go        # File system scanning and caching
├── parser.go           # YAML + Markdown parsing
├── types.go            # Core data structures
├── validator.go        # Basic syntax validation
└── cache.go            # Discovery cache

internal/tools/
├── tool_agents_list.go      # List agents
└── tool_agents_validate.go  # Validate agents
```

**Milestones**:
- ✅ Discover agents in .adk/agents/, .adk/skills/, commands/
- ✅ Parse YAML frontmatter correctly
- ✅ Validate required fields (name, description)
- ✅ Basic tool for listing agents
- ✅ Cache discovered agents

**Acceptance Criteria**:
- Discovers 100+ agents in <1 second
- Handles malformed YAML gracefully
- Reports clear error messages
- Tests: 80%+ code coverage

### 14.2 Phase 2: Validation & Linting (Weeks 4-5)

**Goal**: Comprehensive validation and best practices enforcement

**Deliverables**:
```
pkg/agents/
├── linter.go           # Best practices checking
├── rules.go            # Validation rules engine
└── suggestions.go      # Fix suggestions

internal/tools/
└── tool_agents_lint.go      # Lint tool
```

**Milestones**:
- ✅ Validate all fields comprehensively
- ✅ Lint for description quality
- ✅ Check tool access patterns
- ✅ Detect naming convention violations
- ✅ Provide actionable fix suggestions

**Acceptance Criteria**:
- Zero false positives on real projects
- Actionable suggestions for all warnings
- Validates 1000 agents in <500ms
- Tests: 85%+ code coverage

### 14.3 Phase 3: Management & Generation (Weeks 6-7)

**Goal**: Create, edit, and manage agents

**Deliverables**:
```
pkg/agents/
├── generator.go        # Template generation
├── editor.go           # Agent modification
├── templates/          # Built-in templates
│   ├── subagent.tmpl
│   ├── skill.tmpl
│   └── command.tmpl
└── exporter.go         # Plugin export

internal/tools/
├── tool_agents_new.go       # Create agent
├── tool_agents_edit.go      # Edit agent
├── tool_agents_delete.go    # Delete agent
└── tool_agents_export.go    # Export agent
```

**Milestones**:
- ✅ Interactive agent creation wizard
- ✅ Template-based generation
- ✅ Safe editing with backup/rollback
- ✅ Export to plugin format
- ✅ 20+ example agent templates

**Acceptance Criteria**:
- Generated agents pass validation
- Edit operations are atomic
- Backups created automatically
- Templates cover common use cases
- Tests: 90%+ code coverage

### 14.4 Phase 4: Execution & Testing (Weeks 8-10)

**Goal**: Native agent execution and testing framework

**Deliverables**:
```
pkg/agents/
├── executor.go         # Agent execution engine
├── workflow.go         # Multi-agent workflows
├── testing.go          # Testing framework
└── metrics.go          # Metrics collection

internal/tools/
├── tool_agents_run.go       # Execute agent
├── tool_agents_test.go      # Test agent
├── tool_agents_workflow.go  # Run workflow
└── tool_agents_metrics.go   # Show metrics
```

**Milestones**:
- ✅ Native agent execution (no Claude Code dependency)
- ✅ Multi-model support (Gemini, GPT-4, Claude)
- ✅ Agent workflow engine
- ✅ Testing framework with mock LLMs
- ✅ Metrics collection and export

**Acceptance Criteria**:
- Agents execute correctly with all supported models
- Workflows handle errors gracefully
- Tests can run without real LLM calls
- Metrics are accurate and exportable
- Tests: 90%+ code coverage

### 14.5 Phase 5: Advanced Features (Weeks 11-12)

**Goal**: Enhanced capabilities beyond Claude Code

**Deliverables**:
```
pkg/agents/
├── composition.go      # Agent inheritance
├── conditions.go       # Conditional logic
├── versioning.go       # Semantic versioning
└── compatibility.go    # Format compatibility

internal/tools/
├── tool_agents_import.go    # Import Claude Code agents
├── tool_agents_migrate.go   # Upgrade agent formats
└── tool_agents_clone.go     # Clone/duplicate agents
```

**Milestones**:
- ✅ Agent composition (extends keyword)
- ✅ Conditional tool access
- ✅ Semantic versioning
- ✅ Claude Code agent import
- ✅ Format migration tools

**Acceptance Criteria**:
- Inheritance works correctly
- Conditional logic is deterministic
- Version checking prevents incompatibilities
- Can import 100% of Claude Code agents
- Migration is lossless

### 14.6 Timeline Summary

```
Week 1-3:   Phase 1 - Foundation ████████████░░░░░░░░░░░░░░
Week 4-5:   Phase 2 - Validation       ████████░░░░░░░░░░░░
Week 6-7:   Phase 3 - Management              ████████░░░░░░
Week 8-10:  Phase 4 - Execution                      ████████████
Week 11-12: Phase 5 - Advanced                             ████████

Total: 12 weeks (3 months)
```

### 14.7 Dependencies & Prerequisites

**External Dependencies**:
- Go 1.24+
- YAML parser (gopkg.in/yaml.v3)
- Markdown parser (github.com/gomarkdown/markdown)
- ADK framework (existing)
- MCP integration (existing)

**Internal Dependencies**:
- Display system (for UI)
- Model registry (for multi-model support)
- Tool framework (for agent tools)
- Session system (for execution context)

### 14.8 Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Claude Code format changes | High | Medium | Version compatibility layer |
| Performance issues with large projects | Medium | Low | Caching and lazy loading |
| Complex workflow debugging | Medium | Medium | Comprehensive logging and tracing |
| LLM provider API changes | High | Low | Abstraction layer per provider |
| Testing framework complexity | Medium | Medium | Start simple, iterate |

### 14.9 Success Metrics

**Technical Metrics**:
- ✅ Discovery: <1s for 100 agents
- ✅ Validation: <500ms for all agents
- ✅ Code coverage: >90%
- ✅ Memory usage: <50MB for 1000 agents
- ✅ Test pass rate: 100%

**User Metrics**:
- ✅ Agent creation time: <5 minutes
- ✅ Validation error resolution rate: >90%
- ✅ Documentation completeness: >95%
- ✅ User satisfaction: >4.5/5

**Adoption Metrics**:
- ✅ Number of agents created: Track growth
- ✅ Number of teams using agents: Track adoption
- ✅ Agent reuse rate: Measure sharing
- ✅ Bug reports: <5 critical per month

---

## 15. Appendix

### A. YAML Frontmatter Examples

#### Subagent
```yaml
---
name: code-reviewer
description: Expert code reviewer for quality, security, and maintainability. Review code for bugs, performance issues, and best practices.
tools: Read, Grep, Glob, Bash
model: sonnet
---
```

#### Skill
```yaml
---
name: pdf-processor
description: Extract text, fill forms, and merge PDFs. Use when working with PDF files or document processing tasks.
allowed-tools: Read, Bash, Write
---
```

#### Command
```yaml
---
name: deploy
description: Deploy application to production with safety checks
---
```

#### Plugin Agent (from plugin.json)
```json
{
  "name": "code-reviewer-plugin",
  "agents": "./agents/"
}
```

### B. Validation Checklist

```
[ ] Name field is present and valid
[ ] Description field is present and non-empty
[ ] YAML syntax is valid
[ ] All referenced tools exist
[ ] Model name is valid
[ ] File in correct directory
[ ] No duplicate agent names in hierarchy
[ ] Markdown content is present
[ ] No symlink exploits
[ ] File permissions are correct
[ ] UTF-8 encoding
[ ] No credential leaks in content
[ ] Description triggers are specific
[ ] Tool access is minimal
```

### C. Testing Checklist

```
[ ] Unit tests for discovery
[ ] Unit tests for parsing
[ ] Unit tests for validation
[ ] Unit tests for linting
[ ] Integration tests for full workflow
[ ] Performance benchmarks
[ ] Compatibility tests (all OS/Go versions)
[ ] Security tests (path traversal, etc.)
[ ] Error handling tests
[ ] Large dataset tests (1000+ agents)
[ ] Cache invalidation tests
```

### D. Known Limitations (Phase 1)

```
❌ Not Supported:
- Agent execution within adk-code
- Real-time bi-directional sync with Claude Code
- GUI agent editor
- Agent marketplace integration
- Agent versioning/SemVer
- Multi-language agent support
- Agent performance optimization
- Distributed agent coordination

✅ Supported:
- File-based agent definitions
- YAML/Markdown parsing
- Validation and linting
- Basic creation/editing
- Export to plugin format
- Discovery and caching
```

---

## 15. References

- [ADR-0001: Claude Code Agent Support](../adr/0001-claude-code-agent-support.md)
- [Claude Code Documentation](https://code.claude.com/docs)
- [Claude Code Subagents](https://code.claude.com/docs/en/sub-agents)
- [Claude Code Skills](https://code.claude.com/docs/en/skills)
- [Claude Code Plugins](https://code.claude.com/docs/en/plugins)
- [Research: Draft Claude Code Agent](../../research/claude-code/0001-draft-claude-code-agent.md)

---

**Document Version**: 1.0 (Draft)  
**Last Updated**: 2025-11-14  
**Status**: ⏳ Under Review  
**Next Review**: 2025-11-21  

---

**Reviewers**: [ ] Architecture Team [ ] PM [ ] Security [ ] Tech Lead  
**Approved By**: [ ] Engineering Director  
**Implementation Lead**: [ ] Assigned  

