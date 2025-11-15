# Claude Code-Like Agent Feature Specification

**Version**: 1.0  
**Date**: November 15, 2025  
**Status**: Specification Phase  
**Objective**: Define what a "Claude Code-like agent" means in the context of adk-code

---

## Executive Summary

A **Claude Code-like agent** is an agentic system that autonomously understands developer intent and takes direct, measurable action on code—rather than merely suggesting changes. This specification defines the architectural and behavioral characteristics that enable this capability.

---

## 1. Core Principles

### 1.1 Takes Action, Not Suggestions
- **Modifies files** in place using Edit tool
- **Executes commands** without user approval for read-only operations
- **Creates commits** with proper git integration
- **Requires explicit approval** only for destructive operations (delete, force push)

### 1.2 Developer-Centric Design
- **Terminal-first**: Integrates into existing shell workflows
- **Composable**: Works with pipes, redirects, and automation
- **Scriptable**: CLI-first architecture allows automation via CI/CD
- **Transparent**: Shows diffs, command outputs, git changes clearly

### 1.3 Agentic Reasoning Loop
- **Multi-turn iteration**: Agent reasons across multiple tool calls
- **Context-aware**: Understands project structure, git state, recent changes
- **Tool-grounded**: Uses actual tools to gather information, not hallucination
- **Self-correcting**: Detects failures and adapts approach

### 1.4 Specialized Capabilities Through Subagents
- **Code Reviewer**: Quality, security, maintainability checks
- **Debugger**: Root cause analysis and fixes
- **Test Runner**: Test execution and failure analysis
- **Analyzer**: Performance, complexity, and architecture insights
- **Main Agent**: Orchestrator and general-purpose assistant

---

## 2. Essential Features

### 2.1 Rich Tool Set

**File Operations**:
- `Read(path)` - Read file contents with line ranges
- `Edit(path, oldString, newString)` - Precise in-place edits
- `Create(path, content)` - Create new files
- `Delete(path)` - Remove files (requires approval)
- `List(path)` - Directory listing with filtering

**Code Search & Navigation**:
- `Grep(pattern, paths)` - Text search with regex
- `Glob(pattern)` - Find files by pattern
- `Find(pattern, paths)` - Extended find with filters

**Execution**:
- `Bash(command)` - Execute shell commands
- `Bash(git ...)` - Git operations (status, diff, commit, branch)
- `RunTests(...)` - Execute test suite with output capture

**Analysis**:
- `LanguageServer(action, path)` - LSP-based analysis (type checking, diagnostics)
- `CreateArtifact(name, content, type)` - Create shareable artifacts

### 2.2 Subagent System

**Subagent Definition** (Markdown YAML format):
```yaml
---
name: code-reviewer
description: Expert code review specialist. Proactively invoked after code changes.
tools: Read, Grep, Glob, Bash
model: sonnet
---

You are a senior code reviewer ensuring high standards...
```

**Subagent Manager**:
- `.adk/agents/` (project-level, version controlled)
- `~/.adk/agents/` (user-level, personal)
- CLI command `/agents` for management (list, create, edit, delete)

Notes from current implementation:
- Discovery, generator, and linting are provided in `pkg/agents` and the tools `agents-create` and `agents-edit` exist under `adk-code/tools/agents`.
- Full REPL `/agents` integration and agent router are planned but not yet complete.
- Runtime subagent delegation currently uses ADK's agent-as-tool pattern: subagents are wrapped as tools via `agenttool.New()` (see `tools/agents/subagent_tools.go`). This allows the LLM to select a subagent tool at runtime; an explicit central Agent Router (intent scoring and pre-routing checks) is planned for Phase 2 if additional auditability or fine-grained routing is required.

**Subagent Behavior**:
- Separate context window (prevents main conversation pollution)
- Tool restrictions (reduce scope of concern)
- Auto-delegation (agent decides when to use subagent)
- Explicit invocation (user can request specific subagent)

### 2.3 Multi-Agent Orchestration

**Delegation Pattern**:
1. User provides high-level request
2. Main agent analyzes intent
3. Delegates specialized tasks to relevant subagents
4. Subagents execute and report findings
5. Main agent synthesizes results and presents to user

**Workflow Examples**:
```
User: "Improve this code"
  → Analyzer finds performance issues
  → Reviewer suggests best practices
  → Optimizer implements changes
  → Tester runs verification
  → Report: 3 changes, all tests pass
```

### 2.4 Context Management

**Per-Agent Context**:
- Separate conversation history per subagent
- Independent token counters
- Non-overlapping file reads (efficient)

**Context Boundaries**:
- Main agent: Full codebase context
- Subagents: Focused context (e.g., only test files for test-engineer)
- Result sharing: Subagent findings fed back to main

### 2.5 MCP (Model Context Protocol) Support

- **MCP Server Exposure**:
- adk-code has a built-in MCP manager (`internal/mcp/manager.go`) that can connect to external MCP servers (stdio, HTTP/SSE, streamable transports). The REPL exposes `/mcp` commands for listing and inspecting configured servers and tools. Note: `adk-code mcp serve` (adk-code acting as an MCP server) is a planned feature.
- External tools (Claude Desktop, other agents) can call adk-code's tools
- Enables agent composition: adk-code + Figma + GitHub + Slack together

**MCP Resource Types**:
- Files: `@adk:file://path/to/file`
- Project info: `@adk:project://structure`
- Git state: `@adk:git://status`

**MCP Tool Registration**:
- Dynamic tool discovery
- Permission-aware (respect tool restrictions)
- Streaming output support

---

## 3. Behavioral Specifications

### 3.1 Approval Checkpoints

**Read-Only Operations** (no approval needed):
- Read files
- Search/grep
- Examine git history
- List directories
- Run tests (if no modifications)

**Approval Required**:
- Edit files (show diff first)
- Delete files/directories
- Force git operations
- Reset branches
- Clear caches

**Always Transparent**:
- Show command execution and output
- Display diffs before applying
- Provide rollback information

### 3.2 Error Handling

**Automatic Retry**:
- Tool execution timeout → retry once
- Network failure → retry with backoff
- File not found → search similar names, ask user

**Graceful Degradation**:
- MCP server unavailable → continue without it
- Model timeout → switch to faster model
- Token limit → summarize context

**Error Recovery**:
- On edit failure: show error, suggest alternatives
- On execution failure: capture stderr, analyze, propose fix
- Undo/rollback: `git checkout` for failed edits

### 3.3 Tool Execution Philosophy

**"Act, Don't Suggest"**:
- ✓ Edit files in place
- ✓ Create commits with meaningful messages
- ✓ Execute commands and capture output
- ✗ Suggest edits and wait for approval
- ✗ Tell user to "run this command"

**Transparency**:
- Always show what will change before making changes
- Provide full command output
- Log all tool executions to session

---

## 4. REPL & CLI Interface

### 4.1 Interactive Mode (`adk-code` or `adk-code "query"`)

**Slash Commands**:
- `/agents` - List/create/edit subagents
- `/help` - Show available commands
- `/mcp` - Manage MCP servers and authentication
- `/models` - List/switch models
- `/clear` - Clear conversation history
- `/exit` or `Ctrl+C` - Exit session

**Subagent Invocation**:
```
> Use the code-reviewer subagent to check my changes
> Ask the debugger to investigate this error
> Have the test-engineer verify the fix
```

### 4.2 Non-Interactive Mode (`adk-code -p "query"`)

**Flags**:
- `-p, --print` - Query and exit (return response)
- `-c, --continue` - Resume last session
- `-r, --resume <id>` - Resume specific agent session
- `--agents` - Define subagents via JSON
- `--system-prompt` - Custom system prompt
- `--append-system-prompt` - Add to default prompt
- `--output-format` - json, text, stream-json
- `--max-turns` - Limit agentic iterations

**Example Scripts**:
```bash
# Review changes before commit
adk-code -p "Review my changes using the code-reviewer subagent" --max-turns 3

# One-off analysis
echo "main.rs" | adk-code -p "Explain this file" --max-turns 1

# CI/CD integration
adk-code -p "Run tests and fix any failures" --dangerously-skip-permissions
```

### 4.3 MCP Server Mode (`adk-code mcp serve`)

**Interface**:
- Exposes tools as MCP callables
- Streams results in real-time
- Respects authentication/permissions
- Handles tool errors gracefully

---

## 5. Session & Persistence

### 5.1 Session Storage

**Current Implementation**:
```
~/.adk/
  └── sessions.db (SQLite database)
```

adk-code uses SQLite for session persistence via `internal/session/persistence/sqlite.go`.
This provides efficient storage, querying, and ACID transactions.

**Session Data Stored**:
- Full conversation transcript
- Tool calls and results
- Metadata (model, tokens, duration)
- Resumable state (for long-running tasks)
- Session relationships (main agent ↔ subagents)

### 5.2 Resume Capability

**Resume by ID**:
```
adk-code -r "abc123" "Continue analyzing the performance issue"
```

**Resume Most Recent**:
```
adk-code -c "Keep investigating"
```

**Use Case**: Long-running analysis that spans multiple sessions without losing context.

---

## 6. Token & Cost Management

### 6.1 Context Optimization

**Main Agent**:
- Full codebase context (summarized for large projects)
- Conversation history (last N turns)
- Tool results (streamed, not buffered)

**Subagents**:
- Minimal context (only relevant files)
- Short system prompt
- No conversation history (fresh start)

**Result**: ~30-40% token reduction vs. single large agent

### 6.2 Cost Tracking

**Per-Session Metrics**:
- Input tokens consumed
- Output tokens generated
- Tool execution count
- Estimated cost (model-dependent)

**Display in UI**:
```
Session complete
━━━━━━━━━━━━━━━━
Tokens: 2,345 in / 1,234 out (est. $0.045)
Tools: 8 executions
Duration: 45 seconds
```

---

## 7. Success Criteria

### Functional Requirements

- [ ] Subagent framework fully implemented and operational
- [ ] At least 5 default subagents (code-reviewer, debugger, test-engineer, architect, documentation-writer)
- [ ] MCP server interface working (adk-code mcp serve)
- [ ] All core tools executable (Read, Edit, Bash, Grep, Glob)
- [ ] Approval checkpoint system for destructive operations
- [ ] Session persistence and resume capability
- [ ] REPL commands for subagent management

### Quality Requirements

- [ ] >95% tool execution success rate
- [ ] <2s latency for tool calls (excluding network)
- [ ] Graceful error handling for all failure modes
- [ ] Clear, actionable error messages
- [ ] Proper token tracking and cost visibility

### User Experience

- [ ] Intuitive subagent delegation (auto + manual)
- [ ] Transparent diff display before edits
- [ ] Clear command output and tool results
- [ ] Fast feedback loops for iteration
- [ ] Proper context window management

---

## 8. Non-Goals

- **GUI/Web Interface**: Terminal-first only (web access via MCP)
- **Plugin System**: Out of scope for MVP (Phase 3)
- **Advanced Reasoning**: Stick to standard agentic loop
- **Model Training**: Use existing models only
- **Self-Modification**: Security-first, no self-updating agents

---

## 9. Comparison Matrix: Claude Code vs adk-code Target

| Feature | Claude Code | adk-code Target | Priority |
|---------|------------|-----------------|----------|
| Agentic loop | Native | ADK-based | High |
| Tool execution | 30+ tools | ~30 tools | High |
| Subagents | Built-in | To build | High |
| MCP support | Native | To build | Medium |
| Terminal UX | Excellent | Good (Display) | Medium |
| Direct action | Yes | Yes | High |
| Context mgmt | Per-agent | Per-agent | High |
| Plugin system | Yes | Not planned | Low |

---

## 10. Phase Gate Checklist

### Before Phase 1 Starts
- [x] Specification written and reviewed
- [ ] ADR created
- [ ] Roadmap detailed
- [ ] Team alignment confirmed

### Phase 1 MVP Complete
- [ ] Subagent framework (file-based storage)
- [ ] `/agents` REPL command
- [ ] 5 default subagents
- [ ] Auto-delegation logic
- [ ] Documentation

### Phase 2 Complete
- [ ] MCP server implementation
- [ ] Remote server integration
- [ ] Resource exposure
- [ ] Authentication flow

### Phase 3+ Complete
- [ ] Advanced features (chaining, resume)
- [ ] Performance optimization
- [ ] Security hardening
- [ ] Production-ready

---

## References

- Claude Code Documentation: https://code.claude.com/docs
- Claude Code Subagents: https://code.claude.com/docs/en/sub-agents
- Claude Code MCP: https://code.claude.com/docs/en/mcp
- ADK Go: https://github.com/google/adk-go
- adk-code Architecture: /docs/ARCHITECTURE.md
