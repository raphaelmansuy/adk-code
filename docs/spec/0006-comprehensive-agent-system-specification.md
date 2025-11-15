# Comprehensive ADK-Code Agent System Specification

**Version**: 2.0  
**Status**: 📋 **CURRENT STATE - Strategic & Implementation Guide**  
**Date**: November 15, 2025  
**Author**: adk-code Team  
**Branch**: `feat/agent-definition-support-phase2`

---

## Executive Summary

This specification documents the **complete vision and current implementation state** of the adk-code Agent System—a sophisticated multi-agent framework that extends Google's ADK with Claude Code-compatible agent definitions, autonomous execution, and advanced orchestration capabilities.

### What You Need to Know

**🎯 Why Deep Understanding Matters**:
1. **Foundation**: The agent system is the core value proposition of adk-code
2. **Complexity**: Multi-agent systems require careful architecture (state, coordination, dependencies)
3. **Strategy**: adk-code is positioned as the "glue" between Anthropic and Google ecosystems
4. **Execution**: Phase 2 is underway; understanding all phases helps execution efficiency
5. **Innovation**: This is first-mover advantage in cross-platform agent orchestration

**📊 Implementation Status**:
- ✅ Phase 0: **COMPLETE** (Agent discovery & parsing)
- ✅ Phase 1: **COMPLETE** (Configuration, multi-path, validation, linting)
- 🔄 Phase 2: **IN PROGRESS** (Generation, editing, execution)
- 📅 Phase 3: **PLANNED** (Advanced features: workflows, metrics, testing)

---

## Part 1: Strategic Vision & Context

### 1.1 Why Claude Code Agent Support?

**The Problem We're Solving**:
```
Today's developer workflows are fragmented:
- Claude Code users locked into Claude models (Sonnet, Opus, Haiku)
- No way to define reusable agents across tools
- Each tool reinvents agent concepts independently
- No ecosystem interoperability
```

**Our Solution**:
```
adk-code becomes the "orchestration layer":
- Uses Claude Code's proven agent format (YAML + Markdown)
- Executes agents with ANY LLM (Gemini, GPT-4, Claude)
- Adds features Claude Code doesn't have (workflows, multi-agent, metrics)
- Independent execution (no Claude Code dependency needed)
```

**Strategic Advantages**:
1. **Format Compatibility**: Users can share agents across Anthropic/Google ecosystems
2. **Model Flexibility**: Use the right model for each task (cost + capability optimization)
3. **Independence**: adk-code agents execute natively in the ADK framework
4. **Extensibility**: Agents as tools enable infinite composition
5. **Market Position**: First platform to bridge Anthropic and Google ecosystems

### 1.2 Core Concepts: Agents as Tools

**Key Insight**: In adk-code, agents themselves become tools that the main agent can delegate to.

```
┌─────────────────────────────────────────────────────┐
│          Main User Request                          │
│     "Build a feature end-to-end"                   │
└──────────────────┬──────────────────────────────────┘
                   │
        ┌──────────▼──────────────┐
        │   Orchestrator Agent    │
        │  (Analyzes request)     │
        └──┬──────────────────┬───┘
           │                  │
    ┌──────▼────────┐  ┌──────▼────────┐
    │  Code-Reviewer│  │  Code-Architect
    │    Subagent   │  │    Subagent    
    │  (Tool Call)  │  │  (Tool Call)   
    └───────────────┘  └────────────────┘
           │                  │
        Results merge and synthesize
           │
        ┌──▼────────────────────────────┐
        │  Implementer Agent (Tool Call) │
        │  Creates the feature           │
        └───────────────────────────────┘
           │
        Final deliverable
```

**How It Works**:
1. User asks the main agent for a task
2. Main agent analyzes and decides which subagents to involve
3. Each subagent executes as an autonomous tool invocation
4. Results are collected and synthesized
5. Final response delivered to user

**Why This Matters**:
- Scalable: Add agents without modifying core
- Composable: Any agent can delegate to any other
- Observable: Each step tracked separately
- Testable: Agents can be tested in isolation
- Cost-effective: Right tool for right job

### 1.3 The Multi-Agent Paradigm

**Single Agent Limitations**:
- Can't parallelize tasks (sequential only)
- Limited context for complex problems
- 4-8 turns before context fills up
- Can't specialize for specific domains

**Multi-Agent Advantages** (Proven by Anthropic's Research system):
- Parallel exploration (5-10 agents working simultaneously)
- Domain specialization (architect vs. coder vs. reviewer)
- Information compression (each agent focuses on slice)
- 90%+ performance improvement for breadth-first tasks
- Token efficiency through context windows

**adk-code Multi-Agent Strategy**:
```
Token Budget Allocation (1M token budget):
┌─────────────────────────────────────────────┐
│ Lead Agent (Orchestrator)     - 100K tokens │
├─────────────────────────────────────────────┤
│ Subagent 1 (Parallel work)    - 200K tokens │
│ Subagent 2 (Parallel work)    - 200K tokens │
│ Subagent 3 (Parallel work)    - 200K tokens │
│ Subagent 4 (Parallel work)    - 200K tokens │
├─────────────────────────────────────────────┤
│ Reserve (synthesis, retries)  - 100K tokens │
└─────────────────────────────────────────────┘

Result: 5x parallelization = higher quality output
```

---

## Part 2: Current Implementation Status

### 2.1 What Exists TODAY (as of Nov 15, 2025)

#### Phase 0: Agent Discovery ✅ COMPLETE

**What Works**:
```go
// 1. Agent file parsing
agent := agents.ParseAgentFile(".adk/agents/code-reviewer.md")
// Returns: Agent with name, description, tools, model

// 2. Project-level discovery
discoverer := agents.NewDiscoverer(".")
result, _ := discoverer.DiscoverAll()
// Returns: []*Agent with all discovered agents + errors

// 3. CLI integration
/agents                           // Lists all agents
/agents --type subagent          // Filter by type
/agents --detailed               // Show metadata
```

**Files Implemented**:
- `pkg/agents/agents.go` (500+ lines)
  - `Agent` struct with name, description, type, source
  - `Discoverer` for scanning `.adk/agents/` 
  - `ParseAgentFile()` for YAML+Markdown parsing
  - Full error handling

- `tools/agents/agents_tool.go` (140+ lines)
  - `list_agents` tool for CLI integration
  - Filtering by type/source
  - Detailed metadata display

**Test Coverage**: 89% (22/22 tests passing)

---

#### Phase 1: Enhanced Discovery & Configuration ✅ COMPLETE

**What Works**:
```go
// 1. Multi-path discovery (beyond just .adk/agents/)
config := agents.LoadConfig(".")
// Searches: .adk/agents/, ~/.adk/agents/, plugins/

// 2. Extended metadata
agent.Version = "1.0.0"
agent.Author = "raphael@example.com"
agent.Tags = []string{"coding", "review"}
agent.Dependencies = []string{"code-formatter"}

// 3. Agent validation
validator := agents.NewValidator()
result := validator.Validate(agent)
// Returns: ValidationResult with errors/warnings

// 4. Linting for best practices
linter := agents.NewLinter()
lintResult := linter.Lint(agent)
// Checks: vague descriptions, naming conventions, etc.
```

**Files Implemented**:
- `pkg/agents/config.go` (200+ lines)
  - Multi-path configuration
  - Source hierarchies (project > user > plugin)
  - Config file support

- `pkg/agents/linter.go` (500+ lines)
  - 11 built-in linting rules
  - Description quality checks
  - Naming convention validation
  - Dependency validation

- `tools/agents/validate_agent.go` (150+ lines)
- `tools/agents/lint_agent.go` (150+ lines)

**Test Coverage**: 85%+ (comprehensive integration tests)

---

#### Phase 2: Generation & Execution 🔄 IN PROGRESS

**What Works Now**:
```go
// 1. Agent generation from templates
input := agents.AgentGeneratorInput{
    Name: "my-agent",
    Description: "Does something useful",
    TemplateType: agents.TemplateSubagent,
    Author: "me",
}
generator := agents.NewAgentGenerator()
agent, _ := generator.GenerateAgent(input)
agent, _ := generator.WriteAgent(agent, ".")  // Saves to disk

// 2. Agent editing
// Files: edit_agent.go, create_agent.go (120+ lines each)

// 3. Agent execution (PREVIEW)
runner := agents.NewAgentRunner(discoverer)
result, _ := runner.Execute(agents.ExecutionContext{
    Agent: agent,
    Params: map[string]interface{}{"topic": "Go"},
    Timeout: 30 * time.Second,
})
// Returns: ExecutionResult with output, exit code, duration
```

**Files Implemented**:
- `pkg/agents/generator.go` (275+ lines)
  - Template-based agent generation
  - YAML frontmatter creation
  - Input validation

- `pkg/agents/execution.go` (385+ lines)
  - `ExecutionContext` for parameterized execution
  - `ExecutionResult` with timing + exit codes
  - `AgentRunner` for autonomous execution
  - Context-aware execution with timeouts

- `tools/agents/create_agent.go` (120+ lines)
- `tools/agents/edit_agent.go` (120+ lines)
- `tools/agents/run_agent.go` (150+ lines)

**Current Status**: 
- ✅ Generation framework works
- ✅ Execution framework works
- 🔄 CLI integration in progress (`/run-agent` command partial)

---

#### Phase 3: Advanced Features 📅 PLANNED

**Dependency Management** (files exist):
- `pkg/agents/dependencies.go` (150+ lines)
- Dependency graph resolution
- Circular dependency detection
- Transitive dependency tracking

**Version Management** (files exist):
- `pkg/agents/version.go`
- Semantic version parsing
- Compatibility checking

**Execution Strategies** (files exist):
- `pkg/agents/execution_strategies.go` (200+ lines)
- Sequential vs. parallel execution
- Workflow orchestration patterns
- Error recovery strategies

**Integration Systems** (partial):
- `pkg/agents/metadata_integration.go`
- ADK framework integration hooks
- Tool availability tracking

---

### 2.2 Actual vs. Planned Timeline

| Phase | Status | Planned | Actual | Variance |
|-------|--------|---------|--------|----------|
| 0 | ✅ DONE | 2 weeks | 1.5 weeks | **-25% (ahead)** |
| 1 | ✅ DONE | 4 weeks | 3 weeks | **-25% (ahead)** |
| 2 | 🔄 IN PROGRESS | 4 weeks | +2 weeks | **+50% (behind)** |
| 3 | 📅 PLANNED | 2 weeks | TBD | TBD |

**Key Insight**: Early phases finished faster than planned; Phase 2 more complex than initially estimated.

---

### 2.3 Code Organization

```
adk-code/
├── pkg/agents/                          # Core agent system
│   ├── agents.go                        # Discovery + parsing (Phase 0)
│   ├── config.go                        # Configuration (Phase 1)
│   ├── linter.go                        # Best practices (Phase 1)
│   ├── generator.go                     # Template generation (Phase 2)
│   ├── execution.go                     # Agent execution (Phase 2)
│   ├── dependencies.go                  # Dependency resolution (Phase 3)
│   ├── version.go                       # Semantic versioning (Phase 3)
│   ├── execution_strategies.go          # Workflow patterns (Phase 3)
│   ├── metadata_integration.go          # ADK hooks (Phase 3)
│   ├── types.go                         # Core data structures
│   └── *_test.go                        # 85%+ code coverage
│
├── tools/agents/                        # CLI tools
│   ├── agents_tool.go                   # /agents command
│   ├── validate_agent.go                # /agents-validate
│   ├── lint_agent.go                    # /agents-lint
│   ├── create_agent.go                  # /agents-new
│   ├── edit_agent.go                    # /agents-edit
│   ├── run_agent.go                     # /run-agent (preview)
│   ├── export_agent.go                  # /agents-export
│   ├── dependency_graph.go              # /agents-graph
│   └── *_test.go
│
├── .adk/agents/                         # Example agents
│   ├── architect.md
│   ├── code-reviewer.md
│   ├── debugger.md
│   ├── documentation-writer.md
│   └── test-engineer.md
│
└── docs/spec/
    ├── 0001-agent-definition-support.md (original spec)
    ├── 0004-phase0-completion.md        (Phase 0 report)
    ├── THIS FILE                        (strategic overview)
    └── 0006-implementation-phases/
        ├── phase-2-detailed-plan.md
        ├── phase-3-roadmap.md
        └── execution-guide.md
```

---

## Part 3: Deep Technical Understanding

### 3.1 Agent Definition Format (YAML + Markdown)

**Format**:
```markdown
---
name: agent-name                          # Required: kebab-case, unique
description: What does this agent do?     # Required: 50-500 chars
version: 1.0.0                            # Optional: semantic version
author: author@example.com                # Optional: email/name
tags: [coding, review, security]          # Optional: categories
dependencies: [formatter, linter]         # Optional: agent names
tools: Read, Grep, Glob, Bash            # Optional: tools this uses
model: sonnet                             # Optional: claude-3.5-sonnet
---

# Agent Title

## Role and Purpose
Detailed explanation of what the agent does and when to use it.

## Capabilities
- Bullet list of specific capabilities
- How it adds value

## When to Use
Specific scenarios where this agent is most effective

## Instructions
Step-by-step process the agent follows

## Example
Show how to use the agent effectively
```

**Why This Format?**:
1. **Human-Readable**: Works in text editors, GitHub, wikis
2. **Claude Code Compatible**: Same format they use
3. **Self-Documenting**: Markdown content serves as instructions
4. **Extensible**: YAML allows future fields easily
5. **Versionable**: Works perfectly with Git

**Agent Types**:
```yaml
name: code-reviewer                   # Type: "subagent" (default)
name: pdf-processor                   # Type: "skill" (model-invoked)
name: /deploy                         # Type: "command" (user-invoked)
# Type: "plugin" (comes from plugin manifest)
```

### 3.2 Agent Discovery Hierarchy

**Search Order** (highest to lowest priority):
```
1. Project Level:        .adk/agents/
2. User Level:           ~/.adk/agents/
3. Plugin Agents:        ./plugins/*/agents/
4. CLI Definition:       From --agents JSON flag
```

**First Match Wins**:
- Same agent name in multiple locations? Use from highest priority
- Allows overrides (user agent shadows project agent)

**Configuration** (`.adk/config.yaml`):
```yaml
agents:
  skip_missing: false          # Fail if a path doesn't exist?
  project_path: .adk/agents
  user_paths:
    - ~/.adk/agents
    - ~/.local/adk/agents
  plugin_paths:
    - ./plugins/*/agents
  source_priorities:
    - project
    - user
    - plugin
```

### 3.3 Validation & Linting System

**Validation** (Structural - must pass):
```
✅ Name field exists and is kebab-case
✅ Description field exists (50-500 chars)
✅ All referenced tools exist
✅ Model name is valid
✅ Dependencies form no circular references
✅ YAML syntax is valid
```

**Linting** (Best Practices - should pass):
```
⚠️  Description is too vague (uses words like "helps", "data", "tools")
⚠️  Missing author field (Phase 1 enhancement)
⚠️  Too many tools (>15 = overly permissive)
⚠️  Missing documentation section
⚠️  Unusual characters in agent name
⚠️  Unversioned agent (recommended for Phase 1+)
```

**Example Output**:
```
❯ /agents-validate code-reviewer

✓ code-reviewer.md passes structural validation

⚠️  WARNINGS (3):
  - Missing author field (recommended for team sharing)
  - Description could be more specific (currently 120 chars)
  - No version specified (recommending semantic versioning)

Agent is ready to use but could be improved.
```

### 3.4 Agent Execution Model

**Execution Context**:
```go
ExecutionContext{
    Agent:         *Agent,
    Params:        map[string]interface{},
    Timeout:       time.Duration,
    WorkDir:       string,
    Env:           map[string]string,
    CaptureOutput: bool,
    Context:       context.Context,
}
```

**Execution Flow**:
```
1. Validate execution requirements
2. Set up execution environment (working dir, env vars)
3. Create context with timeout
4. Invoke agent with parameterized prompt
5. Capture output (stdout, stderr)
6. Track execution time
7. Return ExecutionResult
```

**ExecutionResult**:
```go
ExecutionResult{
    Output:    string,              // Captured output
    Error:     string,              // Error message (if any)
    ExitCode:  int,                 // Process exit code
    Duration:  time.Duration,       // Execution time
    Success:   bool,                // Success flag
    Stderr:    string,              // Stderr if captured separately
    StartTime: time.Time,
    EndTime:   time.Time,
}
```

### 3.5 Agents as Tools

**How Subagents Become Tools**:

```go
// When agent discovers subagents, each becomes a potential tool

discoverer.DiscoverAll()  // Returns []*Agent
// For each agent:
//   1. Convert to tool.Tool interface
//   2. Add to available tools for main agent
//   3. Tool name: agent_<agent_name>
//   4. Tool input: { "request": "what to ask agent" }
```

**Example Tool Invocation**:
```
User: "Build a feature with reviews"
    ↓
Main Agent thinks: "This needs code review. Let me delegate."
    ↓
Tool Call: agent_code_reviewer({
    "request": "Review this implementation for bugs",
    "code": "..."
})
    ↓
Subagent (code-reviewer) executes
    ↓
Result: "Found 3 issues: ..."
    ↓
Main agent integrates result
    ↓
User: Refined implementation
```

**Tool Integration Points**:
- Agent metadata available to system prompt
- Agent capabilities communicated to LLM
- Tool descriptions generated from agent descriptions
- Execution captured and returned to caller

---

## Part 4: Understanding the Phases

### Phase 0: Discovery (✅ COMPLETE)

**Goal**: Prove the system works at all

**Scope**:
- Scan `.adk/agents/` directory
- Parse YAML frontmatter + Markdown
- List agents with basic filtering
- Handle errors gracefully
- Write comprehensive tests

**Outcome**: ✅ Delivered
- 22/22 tests passing
- 89% code coverage
- Full error handling
- Extensible architecture

**Why Phases Matter**:
- Phase 0 proves concept
- Don't overengineer early
- Validate assumptions with users
- Early feedback loops

---

### Phase 1: Configuration & Validation (✅ COMPLETE)

**Goal**: Make it production-ready

**Scope**:
- Multi-path discovery (project, user, plugins)
- Configuration file support
- Extended metadata (version, author, tags, dependencies)
- Validation framework (structural checks)
- Linting framework (best practices)
- Improved CLI commands

**Outcome**: ✅ Delivered
- Config system in place
- 11 linting rules
- Full validation framework
- Better error messages

**Why This Phase**:
- Users need customization
- Organizations have standards
- Early validation catches bugs
- Metadata enables advanced features

---

### Phase 2: Generation & Execution (🔄 IN PROGRESS)

**Goal**: Make agents usable

**Scope**:
- Agent template generation
- Agent editing (modify existing agents)
- Agent execution/invocation
- /agents-new, /agents-edit, /run-agent commands
- CLI integration
- Preview of agent system capabilities

**Challenges**:
- Execution context complexity
- Agent parameterization
- Error recovery in long-running processes
- CLI UX for complex operations

**Timeline**: +2 weeks from original estimate
- Complexity higher than Phase 0/1
- More edge cases to handle
- Integration with ADK framework more involved

**What's Hard Here**:
- Long-running agent execution
- State management across turns
- Handling agent dependencies
- Integration with LLM context

---

### Phase 3: Advanced Features (📅 PLANNED)

**Goal**: Unlock full power

**Scope**:
- Agent workflows (chaining agents together)
- Parallel agent execution
- Dependency resolution
- Agent composition/inheritance
- Metrics & observability
- Testing framework for agents
- Advanced CLI commands

**Key Features**:
```yaml
---
name: feature-builder
workflow:
  - agent: architect
    output: design
  - agent: implementer
    input: $design
    output: code
  - agent: reviewer
    input: $code
    timeout: 120s
```

**Why Later**:
- Depends on Phase 2 working well
- Requires extensive testing
- Architectural decisions needed
- Multi-agent coordination complex

---

## Part 5: The "Why" Behind Key Decisions

### 5.1 YAML + Markdown Format

**Alternative Considered**: JSON-only, TOML, custom DSL

**Chosen**: YAML frontmatter + Markdown content

**Rationale**:
1. **Human-Readable**: Non-technical people can understand
2. **Ecosystem**: Already used by Claude Code, Jekyll, Hugo
3. **Self-Documenting**: The instructions are the content
4. **Version Control**: Works well with Git diffs
5. **Tooling**: YAML parsers, Markdown renderers everywhere

### 5.2 Agent Discovery via File System

**Alternative Considered**: Database, API registry, configuration files only

**Chosen**: File system scan with optional config

**Rationale**:
1. **Simplicity**: No external dependencies
2. **Composability**: Works with Git naturally
3. **Discoverability**: Browse with `ls`, `find`, GitHub web
4. **Forking**: Users can copy agents easily
5. **Diff-Friendly**: See what changed in agent definitions

### 5.3 Validation vs. Linting Separation

**Validation** (structural, must pass):
- Name format correct
- Required fields present
- Syntax valid

**Linting** (best practices, should pass):
- Description quality
- Tool access minimal
- Documentation complete

**Why Separate?**:
- Users can work with invalid agents (with warnings)
- Clear distinction: errors vs. warnings
- Linting rules can be disabled/customized
- Validation gates releases, linting guides development

### 5.4 Multi-Phase Approach

**Alternative**: Ship everything at once

**Chosen**: 4 phases, 8-12 weeks

**Rationale**:
1. **Risk Mitigation**: Fail fast, early feedback
2. **User Feedback**: Each phase informs next
3. **Learning**: Team learns from each phase
4. **Scope Control**: Prevents gold-plating
5. **Momentum**: Visible progress maintains team morale

---

## Part 6: Implementing Phases in Practice

### 6.1 How to Execute Phase 2

**Current State** (as of Nov 15):
- ✅ Scaffolding exists
- ✅ Data structures in place  
- ✅ Tests written (TDD pattern)
- 🔄 CLI integration partial
- 🔄 Execution flow working but needs refinement

**Next Steps** (Action Items):

1. **Finish CLI Integration** (2-3 days):
   ```
   ❯ /agents-new code-formatter       # Create agent
   ❯ /agents-edit code-formatter      # Edit agent
   ❯ /run-agent code-formatter "format this" # Execute
   ```

2. **Complete Execution Flow** (3-4 days):
   - Hook agents into tool system
   - Test with actual agent runs
   - Add error recovery
   - Validate timeout handling

3. **Testing** (2-3 days):
   - Integration tests for full workflow
   - Edge case handling
   - Performance baseline

4. **Documentation** (1-2 days):
   - API docs
   - Usage examples
   - Troubleshooting guide

### 6.2 Phase 2 Risk Areas

**High Risk**:
- ❌ Execution context management (complex state)
- ❌ Agent parameterization (unclear API)
- ❌ LLM integration points (where to hook?)

**Medium Risk**:
- ⚠️ Error recovery (long-running processes)
- ⚠️ Dependency resolution (circular deps?)
- ⚠️ Tool availability (which tools in which context?)

**Low Risk**:
- ✅ Generation framework (straightforward templating)
- ✅ CLI commands (proven pattern from Phase 0/1)
- ✅ Testing (TDD already done)

### 6.3 How to Plan Phase 3

**Before Starting Phase 3**:
1. ✅ Complete Phase 2 fully
2. ✅ Get user feedback on Phase 2
3. ✅ Measure performance/token usage
4. ✅ Document lessons learned
5. ✅ Re-evaluate scope based on learnings

**Phase 3 Decision Points**:
- Q: Do users actually want agent workflows?
- Q: Is multi-agent execution valuable enough?
- Q: What's the MVP for Phase 3?
- Q: Can Phase 3 be split further?

**Likely Phase 3 Approach**:
- Start with sequential workflows (simpler)
- Add parallelization if needed
- Build metrics incrementally
- User-driven feature requests

---

## Part 7: Key Insights for Success

### 7.1 What Makes Multi-Agent Systems Hard

**1. State Management**:
```
Problem: 5 agents running in parallel, each modifying shared state
Solution: Event-driven architecture, immutable messages between agents
Cost: ~20% more code, but much more reliable
```

**2. Coordination Overhead**:
```
Problem: Agents need to talk to each other
Solution: Shared context window (expensive!) or message passing
Cost: Token usage explodes (15x vs. single agent)
```

**3. Observability**:
```
Problem: Hard to debug multi-agent systems
Solution: Full tracing, detailed logging, replay capability
Cost: Infrastructure complexity increases significantly
```

**4. Error Recovery**:
```
Problem: One agent fails; whole system breaks
Solution: Checkpoints, idempotency, graceful degradation
Cost: ~25% more code, special patterns needed
```

### 7.2 Why Phases Are Critical

**Without Phases**: 
- Ship broken features
- Can't recover from mistakes
- Users hit bugs in production
- Team burns out

**With Phases**:
- Each phase is a checkpoint
- Gather feedback early
- Adjust based on learnings
- Build momentum and confidence

### 7.3 The "Agents as Tools" Pattern

**Why This Matters**:
- Subagents don't need special execution mode
- Reuse all existing tool infrastructure
- Composable: any agent can delegate to any other
- Scalable: add agents without code changes

**How It Works**:
```
Agent 1 requests tool "agent_code_reviewer"
    → Tool invocation goes through normal ADK flow
    → Subagent (code-reviewer) runs with its own context
    → Result returned as tool output
    → Agent 1 integrates result
```

**Implementation Complexity**:
- Low-level: Just another tool
- High-level: Enables powerful orchestration

---

## Part 8: Success Criteria & Measurement

### 8.1 Phase 2 Success Criteria

**Technical**:
- ✅ All tests passing (22+new ones)
- ✅ 85%+ code coverage
- ✅ Zero compilation errors
- ✅ Agents can be created/edited/executed from CLI

**User Experience**:
- ✅ Clear error messages
- ✅ Help documentation complete
- ✅ Works with 5 example agents
- ✅ Beginner can create agent in <5 minutes

**Performance**:
- ✅ Agent discovery <100ms
- ✅ Agent execution <timeout
- ✅ No memory leaks
- ✅ CLI responsive (<1s feedback)

### 8.2 Phase 3 Success Criteria

**Workflows**:
- ✅ Sequential agent chains work
- ✅ Data passing between agents
- ✅ Dependency resolution correct
- ✅ Circular dependency detection

**Metrics**:
- ✅ Token usage tracked
- ✅ Execution time measured
- ✅ Success/failure rates calculated
- ✅ Exportable metrics

**Testing**:
- ✅ Mock LLM for deterministic tests
- ✅ Test coverage >85%
- ✅ Performance benchmarks
- ✅ Integration tests

---

## Part 9: Conclusion & Roadmap

### Why This Understanding Matters

```
A developer understanding this specification can:

1. Implement features in alignment with architecture
2. Anticipate challenges in later phases
3. Make good trade-off decisions
4. Debug complex multi-agent scenarios
5. Extend system intelligently
6. Mentor other developers on patterns
```

### The Path Forward

```
Nov 2025:  Phase 2 completion (Generation, Execution, CLI)
Dec 2025:  Phase 2 polish + Phase 3 planning
Jan 2026:  Phase 3 implementation (Workflows, Metrics)
Feb 2026:  Full agent system release (v1.0)

Parallel: User feedback, documentation, community building
```

### Why adk-code Matters

**In the Landscape**:
- Claude Code (Anthropic): Agent-first but single vendor
- Cline (VS Code): Powerful but not designed for agents
- LangGraph (LangChain): Complex, requires learning new framework
- adk-code: **Simple, multi-vendor, agent-native**

**Our Unique Position**:
```
adk-code = Simple Multi-Agent Orchestration + Any LLM + Agent-First

Not:
- LangChain replacement (too heavy)
- Claude Code replacement (it's complementary)
- Another framework (we're a tool + framework hybrid)

But:
- The glue between Anthropic and Google ecosystems
- The agent orchestration layer for CLI/terminal
- The place to compose agents for complex tasks
```

---

## Appendix A: File Structure Reference

```
pkg/agents/
├── types.go (100 lines)
│   └── Core structs: Agent, DiscoveryResult, ExecutionContext, etc.
│
├── agents.go (500 lines)
│   ├── Discoverer struct and methods
│   ├── ParseAgentFile function
│   ├── Frontmatter extraction
│   └── Error types (ErrNoFrontmatter, etc.)
│
├── config.go (200 lines)
│   ├── Config struct
│   ├── Multi-path configuration
│   ├── Source hierarchies
│   └── LoadConfig function
│
├── linter.go (500 lines)
│   ├── Linter struct
│   ├── 11 built-in rules
│   ├── LintRule interface
│   └── LintResult struct
│
├── generator.go (275 lines)
│   ├── AgentGenerator struct
│   ├── Template system
│   ├── GenerateAgent method
│   └── WriteAgent method
│
├── execution.go (385 lines)
│   ├── ExecutionContext struct
│   ├── ExecutionResult struct
│   ├── Executor interface
│   ├── AgentRunner implementation
│   └── Execute method with timeouts
│
├── dependencies.go (150 lines)
│   ├── Dependency graph resolution
│   ├── Circular dependency detection
│   └── Transitive closure calculation
│
├── version.go (100 lines)
│   ├── Semantic version parsing
│   ├── Compatibility checking
│   └── Version constraints
│
├── execution_strategies.go (200 lines)
│   ├── Sequential execution
│   ├── Parallel execution
│   ├── Workflow patterns
│   └── Error recovery strategies
│
├── metadata_integration.go (150 lines)
│   ├── ADK framework hooks
│   ├── Tool availability tracking
│   ├── Context building
│   └── Capability inference
│
└── *_test.go (85%+ coverage)
    └── Comprehensive test suite

tools/agents/
├── agents_tool.go (200 lines)
│   └── list_agents tool and filtering
│
├── validate_agent.go (150 lines)
│   └── /agents-validate command
│
├── lint_agent.go (150 lines)
│   └── /agents-lint command
│
├── create_agent.go (120 lines)
│   └── /agents-new command
│
├── edit_agent.go (120 lines)
│   └── /agents-edit command
│
├── run_agent.go (150 lines)
│   └── /run-agent command
│
├── export_agent.go (100 lines)
│   └── /agents-export command
│
├── dependency_graph.go (120 lines)
│   └── /agents-graph command (Phase 3)
│
└── *_test.go (85%+ coverage)
    └── CLI tool tests
```

---

## Appendix B: Example Agent Definitions

See `.adk/agents/` directory:
- `architect.md` - System design specialist
- `code-reviewer.md` - Quality & security reviewer
- `debugger.md` - Bug finder & fixer
- `documentation-writer.md` - Technical writer
- `test-engineer.md` - Testing specialist

---

## Appendix C: Glossary

| Term | Definition |
|------|-----------|
| **Agent** | An autonomous AI entity with defined capabilities, tools, and instructions |
| **Subagent** | An agent that can be delegated to by another agent |
| **Skill** | An agent invoked by the model itself (not by user) |
| **Command** | An agent invoked via CLI command (e.g., /deploy) |
| **Workflow** | Sequential or parallel execution of multiple agents |
| **Tool** | Callable function that agents can invoke (read file, execute command, etc.) |
| **Agent Tool** | An agent exposed as a tool to other agents |
| **Orchestrator** | The main agent that delegates to subagents |
| **ExecutionContext** | Parameters and configuration for running an agent |
| **ExecutionResult** | Output and metadata from running an agent |
| **Discovery** | Process of finding and cataloging agent definitions |
| **Validation** | Structural checks (must pass) |
| **Linting** | Best practices checks (should pass) |
| **Multi-Agent System** | Multiple agents working together (in parallel or sequence) |
| **Token Budget** | Total number of tokens available for multi-agent execution |

---

**End of Specification**

**Document Version**: 2.0  
**Last Updated**: November 15, 2025  
**Status**: Current Implementation Guide  
**Maintainers**: adk-code Team
