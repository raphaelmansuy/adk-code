# adk-code Implementation Analysis & Approach

**Version**: 1.0  
**Date**: November 15, 2025  
**Status**: Analysis Phase  
**Focus**: How to implement Claude Code-like features leveraging existing adk-code architecture

---

## Executive Summary

adk-code already has **70% of the infrastructure needed** to implement Claude Code-like agents. The remaining work focuses on (1) building the subagent framework, (2) integrating MCP support, and (3) enhancing tool execution semantics. This document maps the gaps and proposes implementation approaches.

---

## 1. Existing Foundation Analysis

### 1.1 What adk-code Already Has ✓

**Agentic Loop** (fully implemented):
- Google ADK framework handles core agent orchestration
- Multi-turn tool execution with streaming
- Context management via Session subsystem
- Model abstraction layer supporting 3 backends

**Rich Tool Ecosystem** (~30 tools):
- File operations: Read, Write, Delete, Create, List
- Code search: Grep, Glob, Find
- Execution: Bash (with git integration)
- Analysis: LSP-based capabilities
- Display: Artifact creation and rendering

**Terminal-First UI** (Display subsystem):
- ANSI color support
- Markdown rendering
- Streaming output with spinners
- Event timeline visualization
- Proper error display

**Session Management**:
- Conversation history persistence
- Token tracking
- Session resumption capability
- Multiple LLM backend support

**REPL Interface**:
- Interactive mode with readline
- Built-in commands (/help, /models, /use)
- Command history
- Special command handling

### 1.2 What's Missing ✗

**Subagent Framework**:
- Partial support present: `pkg/agents` already implements file discovery, YAML frontmatter parsing, a generator template (`AgentGenerator`) and linting/validation. `tools/agents` implements `SubAgentManager` and converts agents into ADK tools (`tools/agents/subagent_tools.go`) so subagents can be invoked by the main agent today. See `adk-code/pkg/agents` for discovery and `adk-code/tools/agents` for create/edit/list tooling.
- Remaining work: `internal/agents` package (manager & router), Agent Router decision logic (explicit intent-scoring layer), full `/agents` REPL integration and explicit auto-delegation logic that centralizes delegation decisions instead of relying on LLM tool selection.

**MCP Support**:
- Partial support present: `internal/mcp/manager.go` can connect to external MCP servers (stdio, SSE, and streamable transports) and `internal/cli` provides `/mcp` commands to list/configure servers and tools.
- Remaining work: `adk-code mcp serve` (run adk-code as an MCP server), dynamic tool registration from adk-code to external MCP clients, and full resource provider feature parity.

**Tool Semantics**:
- No approval checkpoints (show diff before edit)
- No rollback capability
- Limited error recovery
- Minimal explicit action-taking philosophy

**Advanced Features**:
- No subagent chaining
- No resumable agents by ID
- No context boundaries between agents
- No dynamic tool discovery

---

## 2. Architecture Integration Points

### 2.1 Where Subagents Fit

**Current Flow**:
```
User Input → REPL → Agent → Tools → Display → Output
```

**Subagent-Enhanced Flow**:
```
User Input → REPL → Agent → [Decide: Main or Sub?]
                             ├→ Main Agent (coordinator)
                             ├→ Code Reviewer (subagent A, separate context)
                             ├→ Debugger (subagent B, separate context)
                             └→ Analyzer (subagent C, separate context)
                      → Synthesize Results → Display → Output
```

**Implementation Point**: Between Agent and Tools, insert "Agent Router" that decides which agent handles the request.

### 2.2 Where MCP Fits

**Current Architecture**:
```
adk-code (Agent) → Tools (30+) → Output
```

**MCP-Enhanced Architecture**:
```
┌─ adk-code as MCP Server ─────────────────┐
│  Exposes:                                 │
│  • Tools as MCP callables                │
│  • Resources (files, project info)       │
│  • Prompts (custom workflows)            │
│  Serves: Claude Desktop, other agents    │
└───────────────────────────────────────────┘
         ↕ (stdio/HTTP)
┌─ External MCP Servers ────────────────────┐
│  GitHub MCP → PR review, issue ops       │
│  Jira MCP → Ticket querying              │
│  Figma MCP → Design context              │
│  Slack MCP → Notifications               │
└───────────────────────────────────────────┘
```

**Implementation Point**: New package `internal/mcp/` for server, client, and tool adaptation.

### 2.3 Existing Components to Leverage

**Session System** (`internal/session/*`):
- Already manages agent lifecycle
- Stores conversation history
- Tracks tokens per session
- **Reuse for**: Subagent session storage and resumption

**Model System** (`pkg/models/*`):
- Registry-based model selection
- Multi-backend support (Gemini, OpenAI, VertexAI)
- **Reuse for**: Model selection per subagent

**Display System** (`internal/display/*`):
- Excellent streaming output
- Markdown rendering
- **Reuse for**: Subagent result visualization

**REPL System** (`internal/cli/*`):
- Command handling
- Interactive mode
- **Reuse for**: New `/agents` and `/mcp` commands

---

## 3. Implementation Approach by Component

### 3.1 Subagent Framework

#### Design: File-Based Storage

**Rationale**:
- Versioning: Store in `.adk/agents/` alongside code
- Discoverability: Easy to list and search
- Editing: Can edit with any text editor
- Format: YAML frontmatter + Markdown (simple, human-readable)

**File Structure**:
```
.adk/agents/
  ├── code-reviewer.md        # Project-level
  ├── debugger.md
  ├── test-runner.md
  └── custom-agent.md

~/.adk/agents/
  ├── analyzer.md             # User-level (fallback)
  ├── optimizer.md
  └── researcher.md
```

**File Format**:
```yaml
---
name: code-reviewer
description: Expert code review specialist. Proactively invoked after code changes.
tools: Read, Grep, Glob, Bash
model: sonnet
---

You are a senior code reviewer ensuring high standards of code quality...

Focus on:
- Code clarity and maintainability
- Security vulnerabilities
- Performance issues
- Test coverage
```

#### Implementation Steps

1. **Create SubAgentManager** (`internal/agents/manager.go`):
   ```go
   type SubAgentManager struct {
       projectAgentsDir string
       userAgentsDir   string
       agentCache      map[string]*SubAgent
   }
   
   func (m *SubAgentManager) List() []*SubAgent
   func (m *SubAgentManager) Load(name string) (*SubAgent, error)
   func (m *SubAgentManager) Save(agent *SubAgent) error
   func (m *SubAgentManager) Delete(name string) error
   ```

2. **Define SubAgent Type** (`internal/agents/types.go`):
   ```go
   type SubAgent struct {
       Name        string
       Description string
       Prompt      string
       Tools       []string
       Model       string
       Scope       string // "project" or "user"
   }
   ```

3. **YAML Parsing** (`internal/agents/parser.go`):
   - Parse YAML frontmatter
   - Extract markdown body as prompt
   - Validate required fields
   - Handle defaults

4. **Agent Router** (`internal/agents/router.go`):
   ```go
   type AgentRouter struct {
       manager *SubAgentManager
   }
   
   func (r *AgentRouter) Route(ctx context.Context, request string) (*Agent, error)
       // Decides: main agent vs which subagent to use
   ```

5. **REPL Command** (`internal/cli/commands/agents.go`):
   ```go
   /agents              # List all subagents
   /agents create       # Interactive creation
   /agents edit <name>  # Edit existing
   /agents delete <name> # Remove
   ```

#### Integration with Existing Code

**Hook into Agent execution** (in `app.Run()`):
```go
// Before: directly invoke agent
// response := agent.Run(ctx, input)

// After: check if we should delegate to subagent
router := agents.NewRouter(subAgentManager)
selectedAgent, err := router.Route(ctx, input)
if selectedAgent != nil {
    // Run subagent in isolated context
    subAgentCtx := createSubAgentContext(ctx, selectedAgent)
    response := selectedAgent.Run(subAgentCtx, input)
    // Collect results and synthesize in main agent
} else {
    // Run main agent
    response := agent.Run(ctx, input)
}
```

---

### 3.2 MCP Integration

#### Design: Modular Server

**Rationale**:
- Extensible: Easy to add new MCP servers later
- Testable: Can mock MCP in tests
- Maintainable: Clear separation of concerns

**Architecture**:
```
internal/mcp/
  ├── server.go          # MCP server implementation
  ├── resources.go       # Resource providers (files, project info)
  ├── tools.go           # Tool exposure & adapters
  ├── prompts.go         # Custom MCP prompts
  └── client.go          # Client for connecting to external MCPs
```

#### Implementation Steps

1. **MCP Server** (`internal/mcp/server.go`):
   ```go
   type MCPServer struct {
       tools     []MCPTool
       resources []MCPResource
       prompts   []MCPPrompt
   }
   
   func (s *MCPServer) Start() error // stdio transport
   func (s *MCPServer) RegisterTool(tool MCPTool)
   func (s *MCPServer) RegisterResource(res MCPResource)
   ```

2. **Tool Adapter**:
   - Map adk-code tools to MCP tool definitions
   - Handle permission checks
   - Stream results
   - Error handling

3. **Resource Types**:
   - `file://path` - File content
   - `project://info` - Project structure
   - `git://status` - Git state
   - `workspace://structure` - Workspace layout

4. **Command**: `adk-code mcp serve`
   ```go
   // In main.go, add handler for "mcp serve" subcommand
   if args[0] == "mcp" && args[1] == "serve" {
       server := mcp.NewServer(components)
       server.Start() // Listen on stdio
   }
   ```

5. **MCP Client** (`internal/mcp/client.go`):
   - Connect to external MCP servers
   - Auto-discover tools
   - Handle authentication
   - Integrate results into agent context

#### Integration with Existing Code

**Add to components** (in `internal/app/app.go`):
```go
type Components struct {
    Display *DisplayComponents
    Model   *ModelComponents
    Agent   agent.Agent
    Session *SessionComponents
    MCP     *MCPComponents  // NEW
}

// In New():
mcp, err := InitMCPComponents(ctx, cfg)
```

---

### 3.3 Tool Execution Enhancement

#### Current State

Tools execute commands but don't have approval flows or diff previews.

#### Proposed Enhancement

1. **Edit Tool Approval**:
   ```go
   // Before applying edit, show diff
   diff := computeDiff(original, modified)
   
   if needsApproval && !autoApprove {
       display.ShowDiff(diff)
       approved := promptUserApproval()
       if !approved {
           return ErrEditRejected
       }
   }
   
   applyEdit(path, oldString, newString)
   ```

2. **Rollback Support**:
   ```go
   // Track changes for undo
   changes := []Change{
       {Tool: "Edit", File: "main.go", Before: old, After: new},
   }
   
   // Later: user can undo
   /undo  // Reverts last change
   ```

3. **Error Recovery**:
   ```go
   // On tool failure, agent can:
   // 1. Analyze error message
   // 2. Propose fix
   // 3. Retry with adjusted parameters
   // 4. Fall back to subagent (debugger)
   ```

---

## 4. Data Flow Diagram: Subagent Execution

```
┌──────────────────┐
│  User Request    │
│  "Fix the bug"   │
└────────┬─────────┘
         │
         ▼
┌──────────────────────────────┐
│  Main Agent                  │
│  • Analyze request           │
│  • Search for error context  │
│  • Decide next step          │
└────────┬─────────────────────┘
         │
         ├─→ Is this a bug/error?
         │   YES ↓
         │
    ┌────▼──────────────────────────────────┐
    │  Subagent Router                      │
    │  Match: Debugger subagent available?  │
    │  YES → Create subagent context        │
    └────┬─────────────────────────────────┘
         │
    ┌────▼──────────────────────────────┐
    │  Debugger Subagent (separate ctx) │
    │  • Read error logs                │
    │  • Identify root cause            │
    │  • Propose and implement fix      │
    │  • Verify fix works               │
    │  → Return findings + recommendation
    └────┬───────────────────────────────┘
         │
    ┌────▼─────────────────────────────────┐
    │  Main Agent (reconvenes)             │
    │  • Receives subagent findings        │
    │  • Synthesizes results              │
    │  • Formats for user                 │
    └────┬────────────────────────────────┘
         │
         ▼
    ┌──────────────────────┐
    │  Display/Output      │
    │  Show results        │
    │  to user             │
    └──────────────────────┘
```

---

## 5. Storage & Persistence Model

### Session Storage

**Current Implementation:**
```
~/.adk/
  └── sessions.db (SQLite database)
```

adk-code uses SQLite for session persistence (implemented in `internal/session/persistence/sqlite.go`).
This provides:
- Efficient querying and indexing
- ACID transactions for reliability
- Built-in support for complex queries
- No need for file format parsing

The SQLite database stores:
- Full conversation transcripts
- Tool calls and results
- Session metadata (model, tokens, duration)
- Resumable state for long-running tasks

### Subagent Definitions

```
Project:
.adk/agents/
  ├── code-reviewer.md
  ├── test-engineer.md
  ├── debugger.md
  ├── architect.md
  └── documentation-writer.md

User:
~/.adk/agents/
  ├── general-debugger.md
  └── performance-analyzer.md
```

---

## 6. Testing Strategy

### Unit Tests

- SubAgent parsing and validation
- MCP server tool registration
- Tool execution and approval flow
- Router decision logic

### Integration Tests

- End-to-end subagent invocation
- MCP server start/stop
- Tool chain execution
- Session resumption

### E2E Tests

- Full workflow: request → subagent → result
- MCP connection and tool exposure
- Complex subagent chaining
- Error recovery

---

## 7. Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Subagent context explosion | Token waste | Separate context windows, summaries |
| MCP server instability | Service degradation | Graceful fallback, error handling |
| Tool permission bypass | Security issue | Clear approval semantics, logging |
| Large project slowdown | UX regression | Lazy loading, caching, summarization |

---

## 8. Performance Considerations

### Optimization Points

1. **Subagent Initialization**:
   - Cache parsed subagent definitions
   - Lazy load only used agents
   - Reuse model instances

2. **Tool Execution**:
   - Parallel tool calls where safe
   - Result streaming (don't buffer)
   - Incremental output

3. **MCP Integration**:
   - Lazy connect to remote servers
   - Cache tool definitions
   - Timeout on slow servers

### Target Metrics

- Subagent invocation: <500ms overhead
- MCP server start: <1s
- Tool execution: <2s per call
- Overall session: <5 min for typical workflow

---

## 9. Security Considerations

### Tool Execution

- ✓ Show diff before file edits
- ✓ Require approval for destructive ops
- ✓ Log all tool executions
- ✓ Sandbox bash execution (consider)

### MCP Integration

- ✓ Validate external MCP servers
- ✓ Permission checks propagated
- ✓ Authentication tokens secured
- ✓ Error messages don't leak secrets

### Subagent System

- ✓ Subagent prompts are user-controlled
- ✓ Tool restrictions enforced
- ✓ No self-modifying agents
- ✓ Audit trail for delegations

---

## 10. Implementation Roadmap

### Phase 1: Subagent MVP (Weeks 1-3)
1. SubAgentManager implementation
2. File-based storage (YAML parsing)
3. `/agents` REPL command
4. Agent router (basic delegation)
5. 5 default subagents
6. Tests & documentation

### Phase 2: MCP Integration (Weeks 4-6)
1. MCP server scaffolding
2. Tool exposure
3. `mcp serve` command
4. Resource registration
5. External MCP client
6. Tests & documentation

### Phase 3: Enhancement (Weeks 7+)
1. Subagent chaining
2. Resumable agents
3. Advanced approval flows
4. Performance optimization
5. Production hardening

---

## 11. Success Metrics

- [x] All features from Specification document implementable
- [ ] Code complexity remains manageable (<200 lines per package)
- [ ] Test coverage >80%
- [ ] Performance within targets (see 8.)
- [ ] User can create custom subagent in <5 minutes
- [ ] MCP server exposes all tools correctly

---

## References

- adk-code Architecture: `/docs/ARCHITECTURE.md`
- Specification: `01_claude_code_agent_specification.md`
- ADK Go: https://github.com/google/adk-go
- MCP Spec: https://modelcontextprotocol.io
