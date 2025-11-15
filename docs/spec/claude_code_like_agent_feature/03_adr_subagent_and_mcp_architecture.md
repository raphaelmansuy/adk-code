# ADR-0005: Claude Code-Like Agent Architecture with Subagents and MCP

**Status**: Proposed  
**Date**: November 15, 2025  
**Author**: Engineering Team  
**Scope**: adk-code core architecture evolution

---

## Context

adk-code is built on Google's ADK framework and provides an agentic CLI coding assistant. Currently, it operates as a single agent with ~30 tools. The goal is to evolve it toward a Claude Code-like system where:

1. **Specialized subagents** handle specific tasks (code review, debugging, testing)
2. **Model Context Protocol (MCP)** allows integration with external tools
3. **Enhanced tool semantics** provide transparent, action-oriented workflows

This ADR defines the architectural decisions to achieve these goals while maintaining backward compatibility and leveraging existing infrastructure.

---

## Problem Statement

**Current Limitations**:
1. Single-agent architecture cannot specialize (reviewer + debugger competing for context)
2. No mechanism to restrict tools per task (safety/focus)
3. Token usage inefficient for large conversations
4. No integration path for external tools (GitHub, Jira, Figma, etc.)
5. No mechanism for composing complex workflows

Note: Since this ADR was drafted, we have partial Phase-0/Phase-1 implementations:
- `pkg/agents` implements discovery, YAML frontmatter parsing, a generator template and a linter for agent definitions.
- `adk-code/tools/agents` exposes create/edit/list tools for agents.
- `internal/mcp/manager.go` provides an MCP client manager that connects to external MCP servers (stdio/SSE/streamable).
These components advance Phase 0 and indicate the system is partly implemented; remaining work includes the router, enforcement/approval flows, and a serving mode for adk-code as an MCP server.

**Desired Outcome**:
- Autonomous agents that delegate specialized tasks
- Clean integration with external tools/data sources
- Transparent, audit-able decision-making
- Efficient token usage and fast iteration

---

## Decision

We adopt a **hierarchical multi-agent architecture** with three key components:

### 1. Subagent Framework

Create a file-based subagent system that enables task-specific agent delegation:

**Design**:
- Store subagent definitions in `.adk/agents/` (project) and `~/.adk/agents/` (user)
- Use YAML frontmatter + Markdown format for human-readability and version control
- Each subagent has isolated context, restricted tool access, and custom system prompt
- Main agent decides when/how to delegate using auto-detection + explicit invocation

**Rationale**:
- ✓ File-based = easy versioning, editing, sharing
- ✓ Markdown = human-readable, low maintenance
- ✓ Isolated context = token efficiency + specialization
- ✓ Tool restrictions = safety + focus

**Alternative Considered**: Database-driven subagents
- ✗ Adds complexity, harder to version control
- ✗ Requires migration path from file-based if we change later

### 2. MCP Integration Layer

Implement adk-code as both an MCP server and client:

**Design**:
- Expose adk-code tools as MCP callables: `adk-code mcp serve`
- Allow external MCP servers to extend capabilities (GitHub, Jira, etc.)
- Dynamic tool discovery and permission enforcement
- Streaming results and error handling

**Rationale**:
- ✓ Standard protocol (MCP) not proprietary
- ✓ Composable (adk-code + other agents + external tools)
- ✓ Extensible (new MCP servers added without code changes)
- ✓ Community tools available (50+ MCP servers already exist)

**Alternative Considered**: Custom plugin system
- ✗ Proprietary, maintenance burden
- ✗ Reinvents the wheel (MCP already standardized)

### 3. Agent Router

Insert decision logic between user request and tool execution:

**Design**:
- Router examines request intent and available subagents
- Matches request to best-fit agent (main or subagent)
- Creates appropriate context for selected agent
- Synthesizes results back for presentation

**Rationale**:
- ✓ Extensible (new subagents detected automatically)
- ✓ Clean separation (router ≠ agent logic)
- ✓ Transparent (router decisions can be logged)
- ✓ Backward compatible (main agent is default)

**Alternative Considered**: Hardcoded delegation
- ✗ Not scalable, requires code changes for new subagents
- ✗ Brittle (tightly coupled)

---

## Architecture

### Component Diagram

```
┌──────────────────────────────────────────────────────────────┐
│ USER / REPL                                                  │
│ Input: "Fix the bug"                                         │
└──────────────────────┬───────────────────────────────────────┘
                       │
          ┌────────────▼──────────────────┐
          │ Agent Router                  │
          │ ├─ Analysis of intent         │
          │ ├─ Subagent matcher          │
          │ └─ Context creator           │
          └────────┬───────────┬──────────┘
                   │           │
        ┌──────────┴─┐    ┌────┴──────────┐
        │            │    │               │
   ┌────▼──────┐  ┌─▼────▼─────┐  ┌──────▼────┐
   │ Main Agent│  │Subagent A   │  │Subagent B │
   │(General)  │  │(Debugger)   │  │(Reviewer) │
   └─────┬─────┘  └──────┬──────┘  └───┬───────┘
         │               │             │
         └───────┬───────┴─────┬───────┘
                 │             │
           ┌─────▼──────────────▼──────┐
           │ Tool Execution System     │
           │ ├─ Read, Edit, Bash       │
           │ ├─ Grep, Glob, Find       │
           │ ├─ Git operations         │
           │ ├─ Approval checks        │
           │ └─ Error recovery         │
           └─────┬───────────┬─────────┘
                 │           │
         ┌───────┴─┐   ┌─────▼────────┐
         │         │   │              │
      ┌──▼─────────▼─┐ │ MCP Servers  │
      │ File System  │ │ (External)   │
      │ & Workspace  │ │              │
      └──────────────┘ └──────────────┘
```

### Data Flow: Subagent Execution

```
Input: "fix the build error"
  ↓
Router: Analyze intent
  → Contains "error", "fix", "build"
  → Match: Debugger subagent (high confidence)
  ↓
Create subagent context:
  → Model: Debugger system prompt
  → Tools: [Read, Edit, Bash, Grep] (restricted)
  → Context: Last error message, build logs
  ↓
Debugger runs:
  1. Bash: get build error output
  2. Read: examine failing test/code
  3. Edit: apply fix
  4. Bash: verify fix works
  ↓
Return to main agent:
  → Findings: "Root cause was missing import X"
  → Action: "Applied fix to file.rs"
  → Status: "Verified by running tests"
  ↓
Main agent synthesizes and presents to user
```

---

## Decision Details

### Subagent Definition Format

```yaml
---
name: code-reviewer
description: Expert code review specialist. Use proactively after code changes.
tools: Read, Grep, Glob, Bash
model: sonnet
---

You are a senior code reviewer ensuring high standards of code quality and security.

When invoked:
1. Run git diff to see recent changes
2. Review for best practices, security issues, performance
3. Provide actionable feedback
```

**Why YAML + Markdown**:
- Simple text format (no database)
- Version-controllable in git
- Human-readable and editable
- Extensible (add fields as needed)
- No parsing complexity

### MCP Server Implementation

```go
// Start MCP server exposing adk-code tools
adk-code mcp serve

// External agent/tool can then:
// - Call adk-code's Read tool via MCP
// - Access project files as resources
// - Execute bash commands
// - See results in real-time
```

**Example Use Case**:
```json
{
  "mcpServers": {
    "adk-code": {
      "command": "adk-code",
      "args": ["mcp", "serve"],
      "type": "stdio"
    }
  }
}
```

Claude Desktop (or another agent) can now use adk-code's tools as if they were native.

---

## Implementation Strategy

### Phase 1: Subagent Framework (Weeks 1-3)
- [ ] Create `internal/agents/` package with manager, types, parser
- [ ] Implement `/agents` REPL command
- [ ] Build agent router
- [ ] Create 5 default subagents (code-reviewer, debugger, test-engineer, architect, documentation-writer)
- [ ] Add tests and documentation

**Deliverable**: Subagent MVP - users can see subagents, invocation works

### Phase 2: MCP Integration (Weeks 4-6)
- [ ] Create `internal/mcp/` package
- [ ] Implement MCP server with stdio transport
- [ ] Expose tools and resources
- [ ] Add `mcp serve` command
- [ ] Implement external MCP client
- [ ] Tests and documentation

**Deliverable**: adk-code works as MCP server, can connect to external servers

### Phase 3: Enhancement (Weeks 7+)
- [ ] Subagent chaining
- [ ] Resumable subagent by ID
- [ ] Advanced tool approval flows
- [ ] Performance optimization
- [ ] Production hardening

**Deliverable**: Production-ready Claude Code-like system

---

## Backward Compatibility

**✓ Full compatibility** with existing behavior:
- Existing `/help`, `/models`, `/use` commands unchanged
- Existing tool set unchanged
- Existing session format compatible
- Main agent behavior as default (no change)

**New features are opt-in**:
- Subagents only used if created
- MCP only enabled if explicitly started
- Tool approvals only if requested

---

## Testing Plan

### Unit Tests
- SubAgent parsing and validation
- Router decision logic
- MCP tool registration
- Tool approval flows

### Integration Tests
- Subagent invocation end-to-end
- MCP server startup and tool exposure
- Tool chain execution with subagents
- Session persistence

### E2E Tests
- Full workflow: user request → subagent → result
- Complex workflows (subagent chaining)
- Error scenarios and recovery

**Target**: >80% code coverage for all new packages

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Subagent context explosion | Medium | Token waste, cost | Separate contexts, summarization |
| MCP server instability | Low | Service disruption | Graceful fallback, error handling |
| Tool permission bypass | Low | Security issue | Clear semantics, logging, testing |
| Performance regression | Medium | UX degradation | Profiling, caching, lazy loading |

---

## Costs

### Development
- ~3 weeks for Phase 1 (subagents)
- ~2 weeks for Phase 2 (MCP)
- ~2 weeks for Phase 3+ (enhancements)
- ~0.5 weeks for testing & documentation per phase

### Maintenance
- New subagent examples/docs
- MCP server monitoring
- Subagent best practices guide
- Periodic security reviews

### Operational
- No new infrastructure required
- No additional costs (uses existing models)
- Slightly higher token usage (beneficial - more efficient)

---

## Alternatives Considered

### 1. Single-Agent with Tool Specialization
- ✗ Single context window still limited
- ✗ Can't prevent token bloat
- ✗ No clear delegation mechanism

### 2. External Agent Orchestration
- ✗ Dependency on external system
- ✗ Less control over behavior
- ✗ Integration complexity

### 3. Plugin System Instead of MCP
- ✗ Proprietary, duplicates work
- ✗ Maintenance burden
- ✗ Fragmented ecosystem

### 4. Database-Driven Subagents
- ✗ Adds infrastructure
- ✗ Harder to version control
- ✗ Migration complexity

---

## Decision Criteria Met

- ✓ **Feasible**: Leverages existing ADK infrastructure
- ✓ **Backward Compatible**: No breaking changes to existing API
- ✓ **Extensible**: Easy to add new subagents/MCP servers
- ✓ **Standards-Based**: Uses MCP (not proprietary)
- ✓ **User-Friendly**: Simple YAML format for definitions
- ✓ **Scalable**: Can handle many subagents without perf impact
- ✓ **Testable**: Clear boundaries, mockable components
- ✓ **Maintainable**: Clean separation of concerns

---

## Follow-Up Decisions

This ADR enables future decisions on:
1. Plugin system (Phase 4)
2. Distributed agent execution (Phase 5)
3. Agent marketplace (Phase 6)
4. Model-specific agents (as new models available)
5. Advanced scheduling/orchestration

---

## References

- ADK Go Documentation: https://github.com/google/adk-go
- Model Context Protocol: https://modelcontextprotocol.io
- Claude Code Subagents: https://code.claude.com/docs/en/sub-agents
- adk-code Architecture: `/docs/ARCHITECTURE.md`
- Specification: `01_claude_code_agent_specification.md`
- Implementation Approach: `02_adk_code_implementation_approach.md`

---

## Sign-Off

- [ ] Product Lead
- [ ] Engineering Lead
- [ ] Security Review
- [ ] Architecture Review

**Date Approved**: _________________
