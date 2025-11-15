# ADR-0001: Claude Code Agent Definition Support

**Status**: Proposed  
**Date**: 2025-11-14  
**Authors**: adk-code Team  
**Related Spec**: [spec/0001-agent-definition-support.md](../../spec/0001-agent-definition-support.md)  
**References**: [research/claude-code/0001-draft-claude-code-agent.md](../../research/claude-code/0001-draft-claude-code-agent.md)

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Context](#context)
3. [Decision](#decision)
4. [Rationale](#rationale)
5. [Implementation Strategy](#implementation-strategy)
6. [API Design](#api-design)
7. [Integration Points](#integration-points)
8. [Consequences](#consequences)
9. [Alternatives Considered](#alternatives-considered)
10. [Future Considerations](#future-considerations)

---

## Problem Statement

**The Challenge**:
- Claude Code provides a sophisticated agent ecosystem (subagents, skills, commands, plugins)
- adk-code currently has no native support for Claude Code agent definitions
- Users cannot discover, validate, or manage Claude Code agents within adk-code
- No tooling exists to help developers author, test, or share Claude Code agents
- Integration between adk-code's agent system and Claude Code's ecosystem is undefined

**Impact**:
- Lost opportunity for ecosystem interoperability
- Developers must switch contexts to manage Claude Code agents
- No validation of agent configurations before deployment
- Missed chance to leverage Claude Code's extensibility model
- Reputation risk if ecosystem expectations aren't met

**Goal**: Enable adk-code to become a first-class agent development and management platform for Claude Code while maintaining architectural independence.

---

## Context

### Historical Context

**Claude Code Architecture**:
- Launched with subagents (specialized AI personalities)
- Extended with Skills (model-invoked capabilities)
- Plugin system for distribution (agents, commands, skills, hooks, MCP servers)
- Markdown-based configuration with YAML frontmatter
- File-based persistence in `.claude/` directories

**adk-code's Current State**:
- Multi-model AI coding assistant (Google ADK framework)
- 30+ built-in tools for file operations, code search, execution
- Support for Gemini, OpenAI, Vertex AI
- Session persistence and streaming responses
- MCP integration for extensibility

### Technical Landscape

**Claude Code Agent Types** (4 primary):
1. **Subagents** - Pre-configured AI personalities with specialized expertise
2. **Skills** - Model-invoked capabilities for specific domains
3. **Commands** - User-invoked slash commands
4. **Plugin Agents** - Agents bundled and distributed via plugins

**Key Characteristics**:
- Hierarchical discovery (project > user > plugin > CLI)
- YAML-frontmatter + Markdown content format
- Granular tool access control
- Model selection per agent (sonnet, opus, haiku, inherit)
- Description-driven discovery mechanism
- Resumable agent instances with isolated contexts

### Market Drivers

**Why Now?**:
1. **Growing AI Agent Ecosystem**: Multiple vendors (Anthropic, OpenAI, Google) investing in agent frameworks
2. **Integration Expectations**: Users expect tools to interoperate
3. **Developer Experience**: Developers want unified tooling across ecosystems
4. **Enterprise Demand**: Organizations need consistent agent governance
5. **Market Differentiation**: First-mover advantage in cross-platform agent support

---

## Decision

### Core Decision

**We will implement full-fidelity Claude Code agent definition support in adk-code as an independent, superior alternative that uses the same file format but with enhanced capabilities**:

1. **Agent Discovery System**: Automatically find and catalog agent definitions across all source hierarchies
2. **Validation Framework**: Comprehensive linting and verification with graceful degradation
3. **Management Tools**: Create, edit, test, and manage agent definitions natively in adk-code
4. **Execution Engine**: Native agent execution within adk-code (no Claude Code dependency)
5. **Enhanced Features**: Multi-model support, workflows, metrics, and advanced capabilities beyond Claude Code

### Strategic Clarifications

**Independence**: adk-code operates independently - it does NOT depend on Claude Code being installed. Agents are executed natively within adk-code using the ADK framework.

**Format Compatibility**: We use the same file format (YAML + Markdown) to enable ecosystem interoperability, but execution is entirely within adk-code.

**Enhancement Philosophy**: Start with Claude Code's proven patterns, then pragmatically exceed them with multi-model support, advanced workflows, and superior tooling.

### Specific Commitments

✅ **In Scope (Phase 1-3)**:
- Support all 4 agent types: subagents, skills, commands, and plugin agents
- YAML/Markdown parsing and validation with graceful degradation
- File discovery across all source hierarchies (.adk/ directories)
- Native agent execution within adk-code (NO Claude Code dependency)
- Multi-model support (Gemini, OpenAI, Anthropic)
- Agent templating and code generation
- Comprehensive validation with helpful error messages
- Tool access control validation
- Agent testing framework (mock and integration testing)
- Agent documentation generation
- CLI-based agent management commands
- Version management with semantic versioning
- Agent metrics and observability

✅ **Out of Scope (Future Phases)**:
- Graphical agent editor (future VS Code extension)
- Agent marketplace integration
- Agent training/fine-tuning capabilities
- Distributed agent coordination
- Cross-tool agent invocation (adk-code agents don't call Claude Code agents)

---

## Rationale

### Why Support Claude Code Agents?

**1. Strategic Alignment**
- Both adk-code and Claude Code are AI-powered coding assistants
- Shared developer audience (coding productivity focus)
- Complementary strengths: adk-code (multi-model, terminal-native) vs Claude Code (specialized, plugin-rich)
- Ecosystem play: Position adk-code as agent orchestration platform

**2. Technical Synergy**
- ADK framework already handles agentic loops (similar patterns to Claude Code)
- adk-code's MCP integration aligns with Claude Code's MCP ecosystem
- Both use markdown-based configuration (natural fit)
- Tool abstraction is compatible (30+ tools → skill-like capabilities)

**3. Developer Experience**
- Unified agent management across tools
- Automatic validation catches mistakes early
- Template generation reduces boilerplate
- Documentation synthesis from definitions
- Single source of truth for team agents

**4. Competitive Advantage**
- First agent platform to offer cross-ecosystem support
- Positions adk-code as "glue" between Anthropic and Google ecosystems
- Addresses real pain point: today's agents are fragmented

**5. Independence & Superiority**
- adk-code agents execute independently (no Claude Code dependency)
- Format compatibility enables sharing, but execution is superior
- Multi-model support (Claude, Gemini, GPT-4) vs Claude-only
- Enhanced features: workflows, metrics, advanced tool configuration
- Natural evolution of MCP support with adk-code's toolset

### Why This Design?

**Layered Approach**:
- **Layer 1 (Discovery)**: Know what agents exist
- **Layer 2 (Validation)**: Ensure agents are correct
- **Layer 3 (Management)**: Create and maintain agents
- **Layer 4 (Integration)**: Connect with Claude Code ecosystem

Benefits:
- Incremental implementation (Phase 1 → 2 → 3 → 4)
- Clear separation of concerns
- Testable at each layer
- Extensible foundation

**Hierarchical File Support**:
- `.adk/agents/` (project - highest priority)
- `~/.adk/agents/` (user)
- Plugin `agents/` directories
- CLI-defined agents (lowest priority)
- **Compatibility**: Can also read `.claude/agents/` for migration

Benefits:
- Uses adk-code's native directory structure
- Compatible with Claude Code format (can import)
- Respects project boundaries
- Enables team sharing
- Provides escape hatches
- Clear separation of adk-code vs Claude Code configs

**YAML + Markdown Format**:
- Matches Claude Code exactly (no format translation)
- Human-readable and editable
- Version control friendly
- Simple to parse and validate

Benefits:
- Zero friction with Claude Code
- Natural for developers
- Git-friendly diffs
- No proprietary formats

---

## Implementation Strategy

### Phase 1: Foundation (Weeks 1-2)

**Goal**: Core agent discovery and parsing

**Deliverables**:
1. `pkg/agents/discovery.go` - File system scanner
   - Recursive search for `.claude/agents/*.md`
   - Search in `~/.claude/agents/`
   - Plugin agent discovery
   - Caching layer

2. `pkg/agents/parser.go` - Markdown + YAML parser
   - YAML frontmatter extraction
   - Field validation
   - Content preservation
   - Error reporting

3. `pkg/agents/types.go` - Data structures
   - `AgentDefinition` struct
   - `PluginManifest` struct
   - Validation interfaces

4. Tool: `tool-agents-list` - List all discovered agents
5. Tool: `tool-agents-validate` - Basic syntax checking

**Acceptance Criteria**:
- ✅ Discover 50+ agents in complex project
- ✅ Parse all agent types correctly
- ✅ Report validation errors clearly
- ✅ Handle malformed files gracefully

### Phase 2: Validation & Linting (Weeks 3-4)

**Goal**: Comprehensive agent validation and best practices enforcement

**Deliverables**:
1. `pkg/agents/validator.go` - Multi-pass validation
   - YAML syntax checking
   - Field requirement validation
   - Tool name verification (against known 30+ tools + MCP tools)
   - Model name validation
   - Description quality analysis
   - Tool conflict detection

2. `pkg/agents/linter.go` - Best practices checking
   - Vague description detection
   - Overly permissive tool access
   - Missing documentation
   - Naming convention violations
   - Path organization issues
   - Conflicting agent definitions

3. Tool: `tool-agents-lint` - Run full validation suite
4. Tool: `tool-agents-describe` - Generate documentation

**Validation Rules**:
- Name: lowercase, digits, hyphens; max 64 chars; unique per hierarchy
- Description: non-empty, max 1024 chars, contains action verbs and triggers
- Tools: comma-separated, all must exist in known tool list
- Model: must be sonnet|opus|haiku|inherit or valid model name
- YAML: valid syntax, proper indentation, no tabs

**Lint Rules**:
1. **Description Quality**
   - Flag: Vague words ("helps", "tools", "data")
   - Flag: Missing "when to use" guidance
   - Flag: No specific keywords for discovery
   - Warn: Too short (<50 chars)
   - Warn: Too long (>500 chars is suspicious)

2. **Tool Access**
   - Warn: Tools field has 20+ tools (overly permissive)
   - Warn: Potentially conflicting tool combinations
   - Suggest: Minimal tool set recommendations

3. **Organization**
   - Warn: Agent files in wrong location
   - Warn: Inconsistent naming conventions
   - Warn: Agent duplicates across hierarchies

**Acceptance Criteria**:
- ✅ Validate 100% of Claude Code agent spec
- ✅ Provide actionable lint suggestions
- ✅ No false positives in real codebase
- ✅ Lint runs in <500ms on large projects

### Phase 3: Management & Generation (Weeks 5-6)

**Goal**: Create and manage agent definitions

**Deliverables**:
1. `pkg/agents/generator.go` - Agent template generation
   - Subagent scaffold template
   - Skill scaffold template
   - Command scaffold template
   - Plugin manifest template
   - Custom prompt builders

2. `pkg/agents/editor.go` - Agent modification
   - Update frontmatter fields
   - Modify tool access lists
   - Change model selection
   - Preserve markdown content

3. Tool: `tool-agents-new` - Create new agent with guided wizard
4. Tool: `tool-agents-edit` - Modify existing agent
5. Tool: `tool-agents-export` - Export to plugin format

**Generation Features**:
- Interactive prompts for agent purpose
- Auto-suggest tool requirements
- Generate system prompt skeleton
- Add markdown structure
- Validate generated content

**Example Usage**:
```bash
# Interactive agent creation
adk-code /agents-new
→ Agent type? (subagent|skill|command)
→ Agent name? (code-reviewer)
→ What should it do?
→ When should it be used?
→ What tools does it need?
[Generates .claude/agents/code-reviewer.md]

# Validate the new agent
adk-code /agents-validate
→ ✅ code-reviewer.md: All validation passed

# Modify an agent
adk-code /agents-edit --agent code-reviewer --tools Read,Grep,Glob
```

**Acceptance Criteria**:
- ✅ Generate valid agents that pass validation
- ✅ Support all 4 agent types
- ✅ Preserve formatting on edits
- ✅ Handle complex scenarios (plugin agents, MCP tools)

### Phase 4: Enhanced Features & Documentation (Weeks 7-8)

**Goal**: Implement adk-code-specific enhancements that surpass Claude Code

**Deliverables**:
1. Multi-model support
   - Support Gemini, OpenAI, Anthropic models
   - Per-agent model configuration
   - Model-specific parameter tuning (temperature, max_tokens)
   - Automatic fallback and retry logic

2. Agent workflows & chaining
   - Sequential agent execution
   - Conditional branching
   - Data passing between agents
   - Workflow visualization

3. Advanced tool configuration
   - Per-tool timeout and resource limits
   - Conditional tool access (confidence thresholds)
   - Tool parameter customization
   - Dynamic tool discovery

4. Agent metrics & observability
   - Token usage tracking
   - Latency monitoring
   - Success/failure rates
   - Tool usage statistics
   - Export metrics to structured logs

5. Testing framework
   - Mock LLM for deterministic tests
   - Test file format (.adk/agents/tests/)
   - Integration testing support
   - CI/CD integration

6. Documentation
   - Comprehensive agent authoring guide
   - Best practices documentation
   - API reference
   - 20+ example agent templates
   - Migration guide from Claude Code
   - Video tutorials

7. Example agents
   - `adk-code-agents` collection showing best practices
   - Multi-model agent examples
   - Workflow examples
   - Testing examples

**Acceptance Criteria**:
- ✅ Agents work with Gemini, GPT-4, and Claude models
- ✅ Workflow chaining works correctly
- ✅ Metrics are accurate and exportable
- ✅ Testing framework covers 90%+ of use cases
- ✅ Documentation is comprehensive
- ✅ Agents created in adk-code can be imported to Claude Code (format compatible)

---

## API Design

### Discovery API

```go
// Agent discovery and enumeration
type AgentDiscovery interface {
    // Find all agents across hierarchy
    FindAllAgents(ctx context.Context) ([]*AgentDefinition, error)
    
    // Find agents by source
    FindProjectAgents(ctx context.Context) ([]*AgentDefinition, error)
    FindUserAgents(ctx context.Context) ([]*AgentDefinition, error)
    FindPluginAgents(ctx context.Context) ([]*AgentDefinition, error)
    
    // Find specific agent
    FindAgent(ctx context.Context, name string) (*AgentDefinition, error)
    
    // Watch for changes (optional)
    Watch(ctx context.Context, handler func(*AgentDefinition)) error
}

type AgentDefinition struct {
    Name        string            // Unique identifier
    Description string            // When and why to use
    Type        AgentType         // "subagent" | "skill" | "command"
    Tools       []string          // Available tools
    Model       string            // "sonnet" | "opus" | "haiku" | "inherit"
    Content     string            // Markdown content
    Path        string            // File path
    Source      AgentSource       // "project" | "user" | "plugin" | "cli"
    PluginInfo  *PluginMetadata   // If from plugin
    Metadata    map[string]string // Extra fields
}

type AgentType string
const (
    TypeSubagent AgentType = "subagent"
    TypeSkill    AgentType = "skill"
    TypeCommand  AgentType = "command"
)

type AgentSource string
const (
    SourceProject AgentSource = "project"
    SourceUser    AgentSource = "user"
    SourcePlugin  AgentSource = "plugin"
    SourceCLI     AgentSource = "cli"
)
```

### Validation API

```go
// Agent validation
type AgentValidator interface {
    // Validate single agent
    Validate(ctx context.Context, agent *AgentDefinition) (*ValidationResult, error)
    
    // Validate all agents
    ValidateAll(ctx context.Context) ([]*ValidationResult, error)
    
    // Get validation rules
    GetRules() *ValidationRules
}

type ValidationResult struct {
    Agent    *AgentDefinition
    Valid    bool
    Errors   []ValidationError
    Warnings []ValidationWarning
    Lint     []LintSuggestion
}

type ValidationError struct {
    Field   string // Which field has error
    Message string // Error description
    Severity string // "error" | "warning"
}

type LintSuggestion struct {
    Rule    string // Which rule triggered
    Message string // What to improve
    Example string // How to fix
}
```

### Management API

```go
// Agent creation and modification
type AgentManager interface {
    // Create new agent
    Create(ctx context.Context, def *AgentDefinition) error
    
    // Update existing agent
    Update(ctx context.Context, def *AgentDefinition) error
    
    // Delete agent
    Delete(ctx context.Context, name string, source AgentSource) error
    
    // Export agent (to plugin or standalone)
    Export(ctx context.Context, name string, format string) ([]byte, error)
    
    // Import agent
    Import(ctx context.Context, data []byte, source AgentSource) error
}

// Agent generation
type AgentGenerator interface {
    // Generate from template
    Generate(ctx context.Context, req *GenerateRequest) (*AgentDefinition, error)
    
    // Get available templates
    ListTemplates() []string
}

type GenerateRequest struct {
    Type        AgentType
    Name        string
    Description string
    Purpose     string
    Tools       []string
    Model       string
}
```

### Tool Implementation Example

```go
// Tool: agents-list
type AgentsListInput struct {
    Filter string `json:"filter"`       // "all" | "project" | "user" | "plugin"
    Sort   string `json:"sort"`         // "name" | "type" | "modified"
    Format string `json:"format"`       // "json" | "table" | "yaml"
}

type AgentsListOutput struct {
    Success bool                 `json:"success"`
    Agents  []*AgentDefinition  `json:"agents"`
    Total   int                 `json:"total"`
    Message string              `json:"message"`
}

func handleAgentsList(ctx Context, input AgentsListInput) AgentsListOutput {
    // Discover agents
    discovery := NewAgentDiscovery(ctx.WorkspacePaths)
    agents, err := discovery.FindAllAgents(ctx)
    if err != nil {
        return AgentsListOutput{Success: false, Message: err.Error()}
    }
    
    // Filter and sort
    agents = filterAgents(agents, input.Filter)
    sortAgents(agents, input.Sort)
    
    // Format output
    return AgentsListOutput{
        Success: true,
        Agents:  agents,
        Total:   len(agents),
        Message: fmt.Sprintf("Found %d agents", len(agents)),
    }
}
```

---

## adk-code Enhancements Beyond Claude Code

To pragmatically surpass Claude Code while maintaining format compatibility, adk-code will add:

### 1. Multi-Model Support
Claude Code is Claude-only. adk-code supports multiple providers:

```yaml
---
name: code-reviewer
model: gemini-2.5-flash      # or gpt-4o, claude-3.5-sonnet
provider: google              # google, openai, anthropic
temperature: 0.7
max_tokens: 2048
---
```

### 2. Agent Workflows & Chaining
```yaml
---
name: feature-builder
description: Builds complete features end-to-end
workflow:
  - agent: code-explorer
    output: feature_analysis
  - agent: code-architect
    input: $feature_analysis
    output: architecture
  - agent: code-implementer
    input: $architecture
    timeout: 300s
---
```

### 3. Advanced Tool Configuration
```yaml
---
name: safe-editor
tools:
  - name: Read
    max_files: 10
  - name: Edit
    requires_confirmation: true
    confidence_threshold: 0.8
  - name: Bash
    timeout: 30s
    allowed_commands: [ls, grep, find, git]
---
```

### 4. Agent Metrics & Observability
```yaml
---
name: monitored-agent
metrics:
  enabled: true
  track: [tokens, latency, tool_calls, success_rate]
  export_to: ./logs/agent-metrics.json
logging:
  level: debug
  output: ./logs/{agent_name}-{timestamp}.log
---
```

### 5. Agent Composition (Inheritance)
```yaml
---
name: senior-reviewer
extends: code-reviewer        # Inherit from base agent
additional_tools: [SecurityScan, ComplexityAnalysis]
override:
  temperature: 0.5
  model: claude-3-opus
---
```

### 6. Semantic Versioning & Compatibility
```yaml
---
name: my-agent
version: "2.1.0"
min_adk_version: "0.3.0"
description: Advanced code reviewer
deprecated: false
---
```

### 7. Testing Support (Built-in)
```yaml
# .adk/agents/tests/code-reviewer.test.yaml
agent: code-reviewer
tests:
  - name: "Reviews code correctly"
    input: "Review the authenticate function"
    expect:
      tools_called: [Read, Grep]
      contains: ["security", "validation"]
      min_confidence: 0.7
```

**Key Difference**: adk-code agents are **more powerful** while remaining **format compatible** for basic use cases.

---

## Integration Points

### 1. CLI Integration

**New Commands**:
```bash
# Discovery
adk-code /agents                    # Interactive agent browser
adk-code /agents-list               # List agents with filtering
adk-code /agents-find <name>        # Find specific agent
adk-code /agents-import <path>      # Import Claude Code agent

# Validation
adk-code /agents-validate            # Validate all agents
adk-code /agents-lint                # Run best practices checks
adk-code /agents-migrate             # Upgrade old agent formats

# Management
adk-code /agents-new                 # Create new agent (wizard)
adk-code /agents-edit <name>        # Edit existing agent
adk-code /agents-delete <name>      # Delete agent
adk-code /agents-clone <name>       # Duplicate agent

# Execution
adk-code /agents-run <name>         # Execute agent
adk-code /agents-test <name>        # Run agent tests
adk-code /agents-workflow <name>    # Execute workflow

# Documentation
adk-code /agents-describe <name>    # Show agent documentation
adk-code /agents-docs               # Generate markdown docs
adk-code /agents-metrics <name>     # Show agent usage metrics

# Export/Share
adk-code /agents-export <name>      # Export as plugin
adk-code /agents-share <name>       # Package for distribution
```

### 2. Session Integration

**Session-scoped Agent Access**:
- Auto-detect agents in project
- Make agent metadata available to LLM context
- Allow LLM to reference agent definitions
- Support agent-aware prompts

### 3. Tool Integration

**Related Tools**:
- `file-read` → Read agent markdown files
- `file-write` → Create/update agent files
- `bash` → Call Claude Code CLI
- `code-search` → Find agents in codebase
- `model-selection` → Set model for agents

### 4. MCP Integration

**MCP Capabilities**:
- Expose agent discovery as MCP resource
- Stream agent changes via MCP events
- Allow external tools to query agents
- Enable bidirectional sync with Claude Code

### 5. Plugin Integration

**Potential adk-code Plugin**:
```json
{
  "name": "adk-code-agents",
  "version": "1.0.0",
  "description": "Claude Code agent management for adk-code",
  "authors": "adk-code Team",
  "agents": "./agents/",
  "skills": "./skills/",
  "commands": "./commands/",
  "hooks": "./hooks.json"
}
```

---

## Consequences

### Positive Consequences

✅ **Enhanced Developer Experience**
- Unified agent management across tools
- Automatic validation prevents errors
- Reduced context switching
- Better discoverability of team agents

✅ **Ecosystem Expansion**
- First tool to bridge Anthropic and Google ecosystems
- Positions adk-code as agent orchestration platform
- Attracts Claude Code users to adk-code
- Opens new market segments

✅ **Technical Benefits**
- Robust validation framework
- Reusable agent discovery code
- Foundation for future agent features
- Extensible architecture

✅ **Strategic Advantages**
- First-mover advantage in cross-platform support
- Differentiates from pure Claude Code tooling
- Strengthens positioning vs. Cline
- Increases switching costs for users

### Negative Consequences

⚠️ **Implementation Complexity**
- Comprehensive agent system requires significant engineering effort
- Multi-model support adds complexity
- Workflow engine needs careful design
- Testing framework is non-trivial

⚠️ **Performance Considerations**
- Agent discovery could be slow in large projects
- File system scanning adds latency
- Caching requirements increase complexity
- Metrics collection adds overhead

⚠️ **Maintenance Considerations**
- Need comprehensive test coverage
- Ongoing documentation updates
- Community support expectations
- Must maintain format compatibility for ecosystem benefit

⚠️ **Format Evolution**
- Basic format follows Claude Code patterns
- adk-code enhancements might not work in Claude Code
- Need clear documentation on compatibility boundaries
- Users may be confused about which features are portable

### Mitigation Strategies

**For Implementation Complexity**:
- Strict phase boundaries with clear milestones
- Incremental implementation and testing
- Reuse existing ADK framework patterns
- Clear separation of concerns (discovery, validation, execution)

**For Performance**:
- Efficient caching strategies (in-memory + disk)
- Background discovery with async loading
- Lazy loading of agents
- Benchmarking for each release
- Profiling and optimization sprints

**For Maintenance**:
- Comprehensive automated tests (unit, integration, e2e)
- CI/CD pipeline with test coverage targets
- Documentation as code (keep in sync)
- Clear versioning and deprecation policy

**For Format Evolution**:
- Clear documentation on compatibility boundaries
- Validation warnings for adk-code-specific features
- Export option for "Claude Code compatible" format
- Version field in agent definitions
- Graceful degradation for unknown fields

---

## Alternatives Considered

### Alternative 1: Ignore Claude Code Integration

**Approach**: Focus solely on adk-code's native agent system, ignore Claude Code.

**Pros**:
- No maintenance burden
- No dependency on Claude Code
- Simpler codebase

**Cons**:
- Lost ecosystem opportunity
- Fragmented tooling for users
- Competitive disadvantage vs. Cline
- Ignores real user need for interoperability

**Decision**: ❌ Rejected - Ignores market realities

### Alternative 2: Wrapper/Proxy Approach

**Approach**: Make adk-code a thin wrapper around Claude Code CLI.

**Pros**:
- Minimal new code
- Always compatible
- Leverage Claude Code's stability

**Cons**:
- adk-code becomes secondary tool
- Can't improve tooling
- No validation or management
- Poor user experience
- Not differentiating

**Decision**: ❌ Rejected - Insufficient value

### Alternative 3: Proprietary Format Conversion

**Approach**: Convert Claude Code agents to adk-code format.

**Pros**:
- Full control over format
- Can optimize for our needs
- No dependency on Claude Code format changes

**Cons**:
- Breaking change if Claude Code updates
- Complex bidirectional sync
- Confusing for users (two formats)
- Maintenance nightmare
- Not source-of-truth compatible

**Decision**: ❌ Rejected - Format fragmentation risk

### Alternative 4: Format-Compatible, Feature-Superior Approach (Selected)

**Approach**: Use Claude Code's proven file format (YAML + Markdown) but implement a superior execution engine with enhanced features in adk-code.

**Pros**:
- Format compatibility enables ecosystem interoperability
- Users can share basic agents between tools
- Builds on proven patterns (Claude Code's design is good)
- adk-code agents are more powerful (multi-model, workflows, metrics)
- No Claude Code dependency - fully independent execution
- Can import Claude Code agents and enhance them
- Git-friendly diffs
- Natural for developers

**Cons**:
- Basic format follows Claude Code patterns (but we add enhancements)
- Must document which features are adk-code-specific
- Users may try to use adk-code features in Claude Code (will be ignored)

**Decision**: ✅ **Selected** - Best of both worlds: compatibility where it matters, superiority where it counts

**Key Principle**: "Steal the good ideas, execute them better, add missing capabilities"

---

## Future Considerations

### Phase 2+ Possibilities

**Short-term (Months 3-6)**:
- 📋 Agent marketplace integration
- 🔄 Bidirectional sync with Claude Code
- 📊 Agent usage analytics
- 🤝 Team agent governance policies
- 🧪 Agent testing framework

**Medium-term (Months 6-12)**:
- 🎨 GUI agent editor (VS Code extension)
- 🚀 Agent performance optimization
- 🔐 Agent security scanning
- 📱 Mobile agent management
- 🌍 Multi-language agent support

**Long-term (12+ Months)**:
- 🤖 AI-assisted agent generation
- 🔀 Cross-platform agent orchestration
- 📦 Agent runtime/container support
- 🧬 Agent evolution and versioning
- 🌐 Federated agent ecosystem

### Compatibility Considerations

**Claude Code Evolution Planning**:
- Monitor Claude Code releases for breaking changes
- Maintain compatibility matrix
- Automated compatibility testing
- Clear upgrade guidance
- Graceful deprecation strategy

**Version Support**:
```
adk-code v1.0.0 supports Claude Code agents v1.0-v1.5
adk-code v1.1.0 supports Claude Code agents v1.0-v2.0
adk-code v2.0.0 supports Claude Code agents v2.0+ (v1.x with warnings)
```

### Ecosystem Expansion

**Potential Integrations**:
- 🔗 OpenAI GPT assistants API
- 🔗 Google Vertex AI agents
- 🔗 Custom agent frameworks (LangChain, etc.)
- 🔗 Agent training/fine-tuning services
- 🔗 Agent deployment platforms

---

## Decision Log

| Date | Decision | Rationale | Outcome |
|------|----------|-----------|---------|
| 2025-11-14 | Implement Phase 1-2 | Market demand + feasibility | Approved |
| 2025-11-14 | Use native format | Zero friction with Claude Code | Approved |
| 2025-11-14 | Four-layer architecture | Incremental, testable, extensible | Approved |
| 2025-11-14 | Hierarchical discovery | Matches Claude Code model | Approved |

---

## References

- [Claude Code Subagents Documentation](https://code.claude.com/docs/en/sub-agents)
- [Claude Code Agent Skills](https://code.claude.com/docs/en/skills)
- [Claude Code Plugins Reference](https://code.claude.com/docs/en/plugins-reference)
- [Draft Research Notes](../../research/claude-code/0001-draft-claude-code-agent.md)
- [Specification Document](../../spec/0001-agent-definition-support.md)

---

## Resolved Strategic Questions

### Q1: How deeply should adk-code support Claude Code agents?
**Answer**: Full-fidelity implementation with enhancements. Use same file format, but with superior execution and advanced features.

### Q2: Should adk-code agents call Claude Code agents or vice versa?
**Answer**: No. adk-code agents call adk-code agents only. Independent execution environment.

### Q3: How to handle agent versioning and backwards compatibility?
**Answer**: Semantic versioning with graceful degradation. Support `version` and `min_adk_version` fields. Maintain backwards compatibility for 2 major versions.

### Q4: What level of validation is appropriate?
**Answer**: Strict validation with helpful errors for blocking issues. Warnings for best practices. Suggestions for improvements. Not blocking for unknown fields (graceful degradation).

### Q5: Should adk-code be a Claude Code plugin itself?
**Answer**: No. adk-code is a standalone alternative, not a Claude Code plugin.

### Q6: How to test agents without running Claude Code?
**Answer**: Comprehensive adk-code-native testing framework with mock LLMs, test file format, integration tests, and CI/CD support.

---

**Status**: ⚠️  **APPROVED WITH REALITY CHECK** - Zero Implementation Exists  
**Implementation Start**: December 2025 (Earliest Realistic Date)  
**Target Completion**: Phase 1 by February 2026, Phase 2 by April 2026, Phase 3 by June 2026  
**DRI**: adk-code Architecture Team  
**Risk Level**: 🔴 HIGH - Large scope, no existing code, ambitious timeline

---

## ⚠️ IMPLEMENTATION REALITY CHECK (Added Nov 14, 2025)

### Current State Assessment

**What Exists Today (Nov 14, 2025)**:

- ✅ Robust tool infrastructure (~27K lines of Go code, 30+ tools)
- ✅ Dynamic tool registry with metadata system
- ✅ MCP server integration
- ✅ Multi-model support (Gemini, OpenAI, Vertex AI, Ollama)
- ✅ Workspace management with VCS awareness
- ✅ Session persistence and tracking
- ✅ Strong error handling and testing framework

**What Does NOT Exist**:

- ❌ NO agent definition discovery system
- ❌ NO YAML/Markdown parser for agent files
- ❌ NO validation framework for agent definitions
- ❌ NO agent management tools (create/edit/delete)
- ❌ NO `.adk/agents/` or `.claude/agents/` directory scanning
- ❌ NO skill or command file support
- ❌ NO plugin manifest parsing
- ❌ NO agent execution within adk-code
- ❌ NO CLI commands for agent operations

**Gap Analysis**:
This ADR describes a comprehensive system (4 phases, 8 weeks) but **ZERO implementation exists**. This is essentially a **greenfield project** requiring:

- Estimated **5,000-8,000 lines of new Go code**
- New `pkg/agents` package (core logic)
- New `tools/agents` package (CLI tools)
- 10+ new tools for agent operations
- Integration across display, REPL, session layers
- Comprehensive test suite (20+ test files)

### Revised Implementation Strategy

**Reality**: Original timeline (8 weeks) is **HIGHLY OPTIMISTIC** given zero existing code and single-contributor constraints.

**Pragmatic Approach**:

1. **Phase 0 (Dec 2025)**: Foundation & Proof of Concept
   - Basic agent file discovery
   - Simple YAML parser
   - Minimal validation
   - 1-2 CLI commands
   - Target: 1000 lines of code

2. **Phase 1 (Jan-Feb 2026)**: Core Discovery & Validation
   - Complete discovery system
   - Robust YAML/Markdown parsing
   - Validation framework
   - 5+ CLI commands
   - Target: +2000 lines of code

3. **Phase 2 (Mar-Apr 2026)**: Management & Generation
   - Agent creation templates
   - Agent editing tools
   - Documentation generation
   - Testing framework basics
   - Target: +2000 lines of code

4. **Phase 3 (May-Jun 2026)**: Enhanced Features
   - Multi-model agent execution
   - Basic workflows
   - Metrics collection
   - Advanced validation
   - Target: +1500 lines of code

**Key Dependencies & Risks**:

- **Single Contributor**: Original plan assumes team, reality is 1-2 developers
- **Competing Priorities**: Other adk-code features and maintenance
- **Testing Requirements**: Each phase needs 90%+ test coverage
- **Integration Complexity**: Touching 5+ major subsystems
- **Documentation**: Each phase requires comprehensive docs

**Success Criteria (Revised)**:

- Phase 0: Basic discovery working by end of Dec 2025
- Phase 1: Agents can be listed and validated by end of Feb 2026
- Phase 2: Agents can be created and managed by end of Apr 2026
- Phase 3: Full feature parity with spec by end of Jun 2026

**Failure Modes to Watch**:

1. Scope creep (adding features not in ADR)
2. Insufficient testing (rushing to "done")
3. Poor error handling (not production-ready)
4. Documentation debt (skipping docs to ship faster)
5. Integration bugs (breaking existing tools)

**Recommendation**:

- ✅ Proceed with **Phase 0 only** for December 2025
- 🔍 Re-evaluate after Phase 0 completion
- 📊 Track velocity and adjust timeline
- ⚠️ Be prepared to de-scope or delay Phases 2-3

---
