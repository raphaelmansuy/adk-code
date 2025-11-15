# Claude Code Like Agent Feature - Research & Implementation Scratchpad

**Date Started**: November 15, 2025  
**Objective**: Implement a Claude Code-like agent in adk-code leveraging Google ADK Go infrastructure

---

## 📚 RESEARCH FINDINGS

### Claude Code Agent Architecture (High-Level)

**Claude Code Key Characteristics**:
1. **Terminal-First Design**: Lives in terminal, integrates with developer workflow (not separate chat window)
2. **Agentic Loop Core**: Runs standard agentic loop: call LLM → parse tools → execute → stream results → iterate
3. **Rich Tool Set**: 30+ tools for file operations, code editing, execution, search, git, etc.
4. **MCP Integration**: Model Context Protocol for extending capabilities with external tools/data (GitHub, Figma, Slack, Jira, Databases)
5. **Subagents/Delegation**: Can delegate to specialized subagents (code-reviewer, debugger, analyzer)
6. **Context Management**: Separate context window per subagent to prevent context pollution
7. **Tool Execution**: Direct file editing, command execution, Git operations - takes action, not just suggests
8. **Unix Philosophy**: Composable, scriptable, works in pipes (e.g., `cat logs | claude -p "analyze"`)

**Claude Code MCP Features**:
- HTTP, SSE (deprecated), Stdio transports for tool connections
- Resource references (@resource notation)
- MCP prompts as slash commands
- Enterprise-managed MCP configurations
- Plugin-provided MCP servers

**Claude Code Subagent System**:
- Pre-configured AI personalities for specific tasks
- Separate context windows from main conversation
- Tool restrictions per subagent
- Custom system prompts
- Built-in subagents: Plan, Code-reviewer, Debugger, Data scientist
- Can be chained (subagent → subagent workflows)
- Resumable subagents with agent IDs for long-running tasks
- Stored in `.claude/agents/` (project) or `~/.claude/agents/` (user) as Markdown YAML files

**CLI Capabilities**:
- `-p` (print mode): Non-interactive query, exit after response
- `-c` (continue): Resume most recent conversation
- `-r` (resume): Resume specific session by ID
- `--agents`: JSON for defining custom subagents
- `--system-prompt`: Complete control over system instructions
- `--append-system-prompt`: Add to default instructions
- `--output-format`: text, json, stream-json
- `--max-turns`: Limit agentic iterations

---

### Google ADK Go Framework Analysis

**ADK Purpose**: Flexible, modular framework for building sophisticated AI agents

**Key Components**:
1. **Agent Loop**: Core agentic iteration (LLM call → tool parsing → execution → repeat)
2. **Tool System**: Pre-built tools + custom functions integration
3. **Model Abstraction**: Support for multiple LLM backends (Gemini, others)
4. **Session Management**: Conversation persistence and history
5. **Multi-Agent Orchestration**: Compose specialized agents
6. **Deployment Ready**: Cloud-native (Google Cloud Run), containerizable

**ADK Go Strengths**:
- Idiomatic Go (leverages concurrency, performance)
- Code-first development
- Modular multi-agent composition
- Model-agnostic and deployment-agnostic

**ADK Go Packages** (from research/adk-go):
- `agent/`: Core agent loop and orchestration
- `model/`: LLM provider abstraction
- `tool/`: Tool definition and execution
- `runner/`: Agent execution runner
- `session/`: Conversation persistence
- `server/`: Server/RPC capabilities
- `memory/`: Context management
- `artifact/`: Output artifacts handling

---

### Current adk-code Architecture

**4-Part Component System**:
1. **Display** (`internal/display/*`): Terminal UI, colors, markdown rendering, streaming output
2. **Model** (`pkg/models/*`): LLM provider abstraction (Gemini, OpenAI, VertexAI)
3. **Agent** (ADK Framework fork): Agentic loop, tool execution
4. **Session** (`internal/session/*`): Persistence, token tracking, history

**adk-code Features**:
- ~1000 lines of critical code (highly scalable)
- ~30 tools across 8 categories
- 3 LLM backends
- REPL with built-in commands (/help, /models, /use, etc.)
- Workspace with multi-root path resolution
- Session persistence
- Token tracking
- Supports custom system prompts

**Current Tool Categories** (inferred from architecture):
1. File operations (Read, Write, etc.)
2. Code editing (Edit, etc.)
3. Search (Grep, Glob, etc.)
4. Execution (Bash, etc.)
5. Git operations
6. Display/Output
7. Codebase navigation
8. Artifact management

---

## 🎯 CLAUDE CODE-LIKE AGENT VISION FOR adk-code

### What Makes Claude Code "Like Claude Code"?

**Non-Negotiable Features**:
1. **Takes Direct Action**: Modifies files, executes commands, creates commits - doesn't just suggest
2. **Terminal Integration**: Seamlessly embedded in developer workflow
3. **Rich Context Awareness**: Understands entire codebase, workspace structure, Git state
4. **Specialized Tool Set**: Purpose-built for code-centric tasks
5. **Composable/Scriptable**: Works in pipes, CLI-driven, automation-friendly
6. **Agentic Reasoning**: Multi-turn reasoning with tool use, not just one-shot completion
7. **Subagent Delegation**: Can delegate specialized tasks to focused agents
8. **Context Windows**: Separate contexts per agent to manage token usage

---

## 💡 IMPLEMENTATION IDEAS & INSIGHTS

### Quick Wins (Phase 1: MVP)

1. **Subagent Framework** (Already partially in ADK)
  - Leverage existing agent loop to create specialized subagents
  - Partial support present in `pkg/agents`: discovery, generator and linting
  - Agent discovery and basic tools available: `adk-code/tools/agents` (create/edit/list)
   - Store in `.adk/agents/` (project) and `~/.adk/agents/` (user)
  - Implement `/agents` REPL command for subagent management (planned; tools exist but full REPL integration is in-progress)

2. **Enhanced Tool Execution**
   - Implement "approval checkpoint" system (before file modifications)
   - Show diffs before applying changes
   - Support rollback/undo
   - Track which tools executed and their side effects

3. **MCP Integration** (Phase 1.5)
   - Expose adk-code as MCP server (`adk-code mcp serve`)
   - Connect to external MCP servers (GitHub, Jira, etc.)
   - Reference resources with @notation
   - Register MCP tools dynamically

4. **Advanced Delegation**
   - Auto-delegate to relevant subagent (code-reviewer after edits)
   - Subagent chaining (analyzer → reviewer → optimizer)
   - Context sharing between subagents

### Architecture Alignment

**Where adk-code Fits**:
```
ADK Framework provides:
  - Core agentic loop ✓ (already have)
  - Multi-agent composition ✓ (leverage)
  - Tool framework ✓ (leverage)
  - Session management ✓ (already have)

adk-code extends with:
  - Terminal-first UX (Display subsystem) ✓
  - Code-specific tools (add/enhance)
  - Subagent orchestration (build)
  - MCP integration (build)
  - REPL/CLI surface (build)
```

### Sub-Agent Design Pattern

```go
// Pattern for subagent definition
type SubAgent struct {
    Name        string              // "code-reviewer"
    Description string              // When to use this agent
    Prompt      string              // System prompt
    Tools       []string            // Tool permissions
    Model       ModelAlias          // Which model to use
}

// Example definitions:
SubAgents = [
  {
    Name: "code-reviewer",
    Description: "Expert code reviewer for quality and security checks",
    Prompt: "You are a senior code reviewer...",
    Tools: ["Read", "Grep", "Glob", "Bash"],
    Model: "sonnet"
  },
  {
    Name: "debugger",
    Description: "Debugging specialist for errors and test failures",
    Prompt: "You are an expert debugger...",
    Tools: ["Read", "Edit", "Bash", "Grep", "Glob"]
  },
  {
    Name: "test-runner",
    Description: "Test automation and CI/CD specialist",
    Prompt: "You are a test engineer...",
    Tools: ["Read", "Edit", "Bash", "Glob"]
  }
]
```

### MCP Server Implementation Approach

```go
// adk-code as MCP server:
// 1. Expose standard ADK tools as MCP resources/tools
// 2. Allow other apps to invoke adk-code's tools
// 3. Example: Claude Desktop config
//    "adk-code": {
//      "command": "adk-code mcp serve"
//    }

// MCP server provides:
// - Resources: Files, project info
// - Tools: Read, Edit, Bash, Grep, Glob, etc.
// - Prompts: Custom workflows
```

---

## 🔮 VISION: MULTI-AGENT ORCHESTRATION

### Future Architecture

```
┌─────────────────────────────────────────────┐
│          Main Agent (REPL)                   │
│  "What should I do next?"                    │
└─────────────┬───────────────────────────────┘
              │
              ├─→ [Code Analyzer]  → Finds issues
              ├─→ [Code Reviewer]  → Reviews changes
              ├─→ [Test Runner]    → Runs tests
              ├─→ [Debugger]       → Fixes errors
              └─→ [Optimizer]      → Improves performance
              
Main agent delegates and coordinates responses
```

### Tool Access Control

```
Main Agent:       All tools (with approval prompts)
Code Reviewer:    Read, Grep, Glob, Bash (no edit)
Debugger:         Read, Edit, Bash, Grep (full execution)
Test Runner:      Read, Bash, Edit (test files only)
Code Analyzer:    Read, Grep, Glob (read-only)
```

---

## 📊 QUICK REFERENCE: CLAUDE CODE VS adk-code

| Aspect | Claude Code | adk-code (Target) | Status |
|--------|------------|----------|--------|
| **Core Loop** | Agentic | Agentic | ✓ (Have ADK) |
| **Tools** | 30+ code-specific | ~30 tools | ✓ (Have base) |
| **Subagents** | Yes (built-in) | Need to build | ⚠️ (In progress) |
| **MCP Support** | Yes (native) | Need to build | ⚠️ (Planned) |
| **Terminal UX** | REPL + colors | REPL + colors | ✓ (Have Display) |
| **Direct Action** | Yes (Edit, Bash) | Yes (Edit, Bash) | ✓ (Have tools) |
| **Context Mgmt** | Per-subagent | Per-subagent | ⚠️ (Partial) |
| **Plugin System** | Yes | Not yet | ❌ (Future) |

---

## 🏗️ IMPLEMENTATION PHASES

### Phase 0: Discovery & Planning (Current)
- [x] Research Claude Code architecture
- [x] Understand Google ADK Go
- [x] Analyze current adk-code
- [ ] Write specification documents
- [ ] Create ADR
- [ ] Plan implementation steps

-### Phase 1: Subagent Framework (MVP)
- [ ] Subagent manager implementation
-- [x] `.adk/agents/` file-based discovery implemented (`pkg/agents`)
-- [x] `/agents` REPL tooling scaffolding available as tools (`agents-create`, `agents-edit`, `list_agents`) - full REPL UI work remains
- [ ] Auto-delegation logic
- [ ] Default subagents (reviewer, debugger, analyzer)

### Phase 2: MCP Integration
- [ ] MCP server interface
- [ ] Resource exposure system
- [ ] Tool registration for MCP
- [ ] `mcp serve` command
- [ ] Remote server integration

### Phase 3: Advanced Features
- [ ] Subagent chaining
- [ ] Resume/resumable agents
- [ ] Permission system refinement
- [ ] Plugin system foundation

### Phase 4: Polish & Production
- [ ] Performance optimization
- [ ] Security hardening
- [ ] Documentation
- [ ] Example agents & workflows
- [ ] Integration testing

---

## 🎓 LEARNINGS & KEY INSIGHTS

1. **ADK Already Provides 70% of Needs**: The existing ADK framework in adk-code handles the core agentic loop. We're primarily building on top (subagents, MCP).

2. **Display Subsystem is Excellent**: adk-code's terminal UI/Display layer is more sophisticated than basic Claude Code CLI - good foundation.

3. **MCP is Industry Standard**: We should treat MCP as a first-class feature from the start, not an afterthought.

4. **Subagents are Game-Changer**: Separate context windows prevent token bloat and allow specialization. This should be a core differentiator.

5. **Tool Execution Philosophy Matters**: "Takes action" vs "suggests actions" is fundamental to Claude Code's appeal. Ensure tools actually modify state.

6. **Composability is Key**: Make adk-code easy to script and pipeline. Unix philosophy applies to AI agents too.

---

## 📝 NEXT STEPS

1. **Write Detailed Specification** (this scratchpad → `01_claude_code_agent_specification.md`)
2. **Write Implementation Analysis** (→ `02_adk_code_implementation_approach.md`)
3. **Create ADR** (→ `03_adr_subagent_and_mcp_architecture.md`)
4. **Create Roadmap** (→ `04_implementation_roadmap.md`)

Each document should be:
- **Concise**: No padding, dense with information
- **Actionable**: Clear next steps, specific tasks
- **Precise**: Technical details, architecture patterns
- **High-Value**: Every sentence earns its place

---

## 🔗 REFERENCES

- Claude Code Docs: https://code.claude.com/docs
- Claude Code MCP: https://code.claude.com/docs/en/mcp
- Claude Code Subagents: https://code.claude.com/docs/en/sub-agents
- ADK Go Repo: https://github.com/google/adk-go
- adk-code Architecture: /docs/ARCHITECTURE.md
