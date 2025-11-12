# Codex Features Analysis for Code Agent

**Date**: November 12, 2025  
**Purpose**: Document high-value features from OpenAI's Codex CLI that could enhance code_agent  
**Status**: Active Investigation  

## Executive Summary

This document catalogues features discovered in Codex CLI (OpenAI's production coding agent written in Rust) that could provide significant value to code_agent. Codex is a mature, feature-rich autonomous coding agent with several patterns and capabilities that are directly transferable to Go.

**Key Finding**: Codex has solved several hard problems that code_agent currently lacks:
- OS-level sandboxing with granular approval policies
- Session persistence and resumable conversations
- MCP (Model Context Protocol) integration for extensible tools
- Sophisticated context compression (plan tool)
- File change previews before execution
- Zero data retention (ZDR) compliance option

---

## 1. OS-Level Sandboxing & Approval Policies

### Discovery
Codex implements sophisticated OS-level sandboxing mechanisms to constrain what the agent can do without user approval. This is critical for safety and user trust.

### Key Features

**Platform-Specific Sandboxing**:
- **macOS 12+**: Uses Apple Seatbelt via `sandbox-exec` to enforce sandbox profiles
- **Linux**: Combines Landlock + seccomp APIs to restrict filesystem and network access
- **Windows**: Uses restricted tokens derived from AppContainer profiles (experimental)

**Approval Policies** (granular control):
1. `"untrusted"` - Prompt before every risky operation (read-only by default)
2. `"on-failure"` - Ask to retry outside sandbox only if sandboxed command fails
3. `"on-request"` - Let model decide when to escalate and ask user
4. `"never"` - Model runs without approval (fully autonomous)

**Sandbox Modes**:
- `read-only` - Can read files, no writes/network (safe default)
- `workspace-write` - Can write to workspace + /tmp, no network (unless enabled)
- `danger-full-access` - No sandboxing (for container environments)

### Value for Code Agent

**Current Limitation**: code_agent has no execution constraints—tools run with full permissions.

**Solution**: Implement granular approval system allowing users to:
- ✅ Run agent in read-only mode (safe exploration)
- ✅ Require approval for writes/commands (controlled automation)
- ✅ Enable workspace-write mode for trusted projects (faster iteration)
- ✅ Configure approval policies globally or per-project

**Implementation Path**:
1. Add approval policy config to settings
2. Implement tool-level approval checking before execution
3. Optional: Add OS sandboxing (macOS/Linux first)
4. Create approval UI/flow in REPL

**Effort**: Medium (80-120 hours for config + approval UI; OS sandboxing adds 40-60)  
**ROI**: Very High (builds user trust, enables safe autonomous operation)

---

## 2. Session Persistence & Resumable Conversations

### Discovery
Codex maintains persistent session state allowing users to resume conversations and tasks. This is critical for long-running operations.

### Key Features

**Session Management**:
- Sessions stored in `~/.codex/sessions/` as SQLite databases
- Each session has a unique UUID identifier
- Full conversation history is preserved
- Can resume with `codex resume` or `codex resume <SESSION_ID>`
- Can resume to a specific point in conversation with `Esc-Esc` (backtrack mode)

**Resume Workflow**:
```bash
codex "Start a task"              # Session 1 starts
codex resume --last               # Resume most recent session
codex resume <SESSION_ID>          # Resume specific session
codex resume --last "New task"     # Resume and continue with new task
```

**Session Picker UI**:
- Interactive TUI shows recent sessions
- Shows conversation preview
- Can select and resume from any session

### Value for Code Agent

**Current Limitation**: Each run is independent; complex tasks that exceed context window are lost.

**Solution**: Add session persistence to code_agent enabling:
- ✅ Resume interrupted tasks (`/resume <id>`)
- ✅ Continue long operations without restarting
- ✅ Maintain task state across sessions
- ✅ Keep conversation history for reference

**Implementation Path**:
1. Add session storage (use SQLite or file-based JSON)
2. Serialize conversation state and results
3. Add `resume` command to REPL
4. Implement session picker UI
5. Extend `/status` to show session ID and timestamp

**Effort**: High (120-150 hours for storage layer + resumption logic)  
**ROI**: Very High (essential for real work, enables recovery from crashes)

**Go Adaptation**: 
- Use `sqlite3` Go package (or file-based JSON for simplicity)
- Store sessions in `$CODE_AGENT_HOME/sessions/`
- Persist after each turn automatically

---

## 3. Non-Interactive Mode (`codex exec`)

### Discovery
Codex provides a CLI mode for automation and scripting without interactive TUI.

### Key Features

**Basic Usage**:
```bash
codex exec "count the total number of lines of code"
codex exec --full-auto "Fix all lint errors"
codex exec --json "Explain this error"
```

**Output Modes**:
1. **Default**: Streams activity to stderr, final message to stdout (pipeline-friendly)
2. **JSON Mode** (`--json`): Streams events as JSONL with structured item types
3. **Structured Output** (`--output-schema`): Constrains output to JSON schema

**Event Types** (in JSON mode):
- `thread.started` / `turn.started` / `turn.completed`
- `item.started` / `item.updated` / `item.completed`
- `command_execution`, `file_change`, `mcp_tool_call`, `web_search`, `todo_list`
- `reasoning`, `agent_message`, `error`

**Resume in Exec**:
```bash
codex exec "First task" --json > first.jsonl
codex exec resume --last "Continue with second task" --json >> first.jsonl
```

### Value for Code Agent

**Current Limitation**: code_agent is interactive REPL only; no CI/automation mode.

**Solution**: Add exec mode enabling:
- ✅ Automation in CI/CD pipelines
- ✅ Structured output for tool integration
- ✅ Headless operation on servers
- ✅ JSON event streaming for monitoring

**Implementation Path**:
1. Add `--exec` flag to CLI
2. Implement non-interactive event loop
3. Add JSON output formatter
4. Support output schema validation
5. Implement session resumption in exec

**Effort**: High (140-180 hours including JSON schema support)  
**ROI**: High (enables enterprise integrations, CI/CD usage)

**Go Adaptation**:
- Use `json.Encoder` for JSONL streaming
- Create event types mirroring Codex structure
- Reuse REPL event system

---

## 4. Slash Commands (`/model`, `/approvals`, `/review`, `/new`, etc.)

### Discovery
Codex provides a set of built-in slash commands for quick operations and configuration changes mid-session.

### Key Features

**Built-in Commands**:
| Command | Purpose |
|---------|---------|
| `/model` | Choose model and reasoning effort |
| `/approvals` | Configure approval policy and sandbox |
| `/review` | Review current changes, find issues |
| `/new` | Start new chat during conversation |
| `/init` | Create AGENTS.md instructions file |
| `/compact` | Summarize conversation (context limit management) |
| `/undo` | Ask agent to undo a turn |
| `/diff` | Show git diff (incl. untracked files) |
| `/mention` | Mention a file (inject content) |
| `/status` | Show session config and token usage |
| `/mcp` | List configured MCP tools |
| `/logout` | Log out |
| `/quit` / `/exit` | Exit session |
| `/feedback` | Send logs to maintainers |

### Value for Code Agent

**Current Limitation**: Limited built-in commands; no model switching, approval management, or review tools.

**Solution**: Implement slash commands for common operations:
- ✅ `/model` - Switch models/reasoning (already exists)
- ✅ `/approvals` - Configure approval policy
- ✅ `/review` - Run code review on changes
- ✅ `/new` - Start new conversation
- ✅ `/diff` - Show git diff
- ✅ `/status` - Show session details
- ✅ `/undo` - Revert last operation
- ✅ `/mention` - Inject file content

**Implementation Path**:
1. Add command parsing to REPL
2. Implement each command handler
3. Add autocomplete for command names
4. Create help system (`/help`)

**Effort**: Medium (60-80 hours for full set)  
**ROI**: Medium-High (improves UX, discoverability)

---

## 5. MCP (Model Context Protocol) Integration

### Discovery
Codex supports MCP servers as a way to extend its capabilities with custom tools. This is critical for extensibility.

### Key Features

**Server Types**:
1. **STDIO Servers** - Launch via command, communicate over stdin/stdout
2. **Streamable HTTP** - Talk to HTTP servers (localhost or remote)

**Configuration**:
```toml
[mcp_servers.example]
command = "npx"
args = ["-y", "mcp-server"]
env = { "API_KEY" = "value" }
cwd = "/path/to/server"
startup_timeout_sec = 20
tool_timeout_sec = 30
enabled_tools = ["search"]
disabled_tools = ["dangerous_tool"]
```

**CLI Commands**:
```bash
codex mcp add docs -- docs-server --port 4000
codex mcp list
codex mcp list --json
codex mcp get docs
codex mcp remove docs
codex mcp login SERVER_NAME  # OAuth support
```

**MCP as Server**:
```bash
codex mcp-server  # Run Codex as MCP server for other agents
```

### Value for Code Agent

**Current Limitation**: Tool ecosystem is static; no extensibility without code changes.

**Solution**: Implement MCP client enabling:
- ✅ User-provided tool servers
- ✅ Custom tool discovery
- ✅ Extensible agent capabilities
- ✅ Integration with MCP ecosystem

**Implementation Path**:
1. Add MCP client library (use existing Go MCP libraries)
2. Implement STDIO and HTTP server transports
3. Add MCP tool registration to tool registry
4. Implement config in settings
5. Add CLI commands for MCP management

**Effort**: Very High (200-240 hours including MCP spec implementation)  
**ROI**: Very High (future-proofs architecture, enables community extensions)

---

## 6. Smart Context Management & Token Tracking

### Discovery
Codex tracks token usage and can proactively summarize conversations to stay within context limits.

### Key Features

**Token Tracking**:
- Reports input/output/cached tokens per turn
- Tracks context window utilization
- Shows in `/status` command
- Warns when approaching limit

**Auto-Compaction** (`/compact`):
- Summarizes conversation when reaching 75% of context window
- Preserves critical information
- User can manually trigger at any time
- Creates "focus chain" of checkpoints

**Focus Chain**:
- Markdown-based task progression tracking
- Captures completed steps
- Maintains context for long tasks
- File-editable (user can review/edit)

### Value for Code Agent

**Current Limitation**: No token tracking; tasks fail when context exhausted.

**Solution**: Add context management features:
- ✅ Token counting per turn (integrate with model provider)
- ✅ Context utilization warning
- ✅ Auto-summary at threshold
- ✅ Manual `/compact` command
- ✅ Task progress file tracking

**Implementation Path**:
1. Add token counting to display layer
2. Implement auto-summary detection
3. Create task tracking format (markdown-based)
4. Add `/compact` command
5. Persist progress file in workspace

**Effort**: Medium (100-120 hours for full feature)  
**ROI**: Very High (critical for long-running tasks)

---

## 7. AGENTS.md (Project-Level Instructions)

### Discovery
Codex discovers and uses `AGENTS.md` files throughout the project hierarchy to provide context-specific instructions.

### Key Features

**Discovery Hierarchy**:
1. `~/.codex/AGENTS.md` - Global personal guidance
2. Repository root down to current directory - Project instructions
3. `AGENTS.md.override` - Directory-specific override

**Usage**:
- Automatically loaded and merged top-down
- Provides instructions without CLI prompts
- Monorepo-aware (per-directory overrides)
- Can be created with `/init` command

**Example**:
```markdown
# Project Guidelines

This is a Go project. Use standard Go conventions.

## Code Style
- Run `gofmt` before committing
- Use `golangci-lint` for linting

## Testing
- Write tests for all public functions
- Use table-driven tests pattern
```

### Value for Code Agent

**Current Limitation**: No project-level instruction discovery.

**Solution**: Implement AGENTS.md discovery:
- ✅ Search for AGENTS.md in hierarchy
- ✅ Load and merge instructions
- ✅ Pass to enhanced prompt
- ✅ Support directory overrides
- ✅ Add `/init` to create template

**Implementation Path**:
1. Add instruction discovery to workspace loading
2. Merge instructions into system prompt
3. Add `/init` command to create template
4. Document format and best practices

**Effort**: Low-Medium (40-60 hours)  
**ROI**: High (improves code quality and consistency)

---

## 8. Change Preview & Diff Display

### Discovery
Codex shows file changes before executing, allowing users to review and approve changes.

### Key Features

**Multi-File Diff Preview**:
- Shows changes side-by-side before applying
- Highlights added/modified/deleted lines
- Allows selective application (not all-or-nothing)
- Interactive approval flow

**Git Integration**:
- `/diff` shows current git diff
- Includes untracked files
- Integrates with approval system

### Value for Code Agent

**Current Limitation**: Tools execute immediately without showing impact.

**Solution**: Add change preview feature:
- ✅ Show file changes before applying
- ✅ Allow user review of diffs
- ✅ Selective file application
- ✅ Git-aware diffs

**Implementation Path**:
1. Implement diff calculation before writes
2. Add diff display in REPL
3. Add approval confirmation
4. Wire into file write tools

**Effort**: Medium (80-100 hours)  
**ROI**: High (builds user confidence, enables review)

---

## 9. Plan Tool & Task Tracking

### Discovery
Codex has a plan tool that the agent can use to structure complex tasks.

### Key Features

**Plan Tool**:
- Agent can create and update task plans
- Tracks completed vs. pending steps
- Markdown-based format
- Displayed in real-time

**Task Tracking**:
- Agent references plan when making decisions
- Updates plan as work progresses
- Helps with context management
- Visible in event stream

### Value for Code Agent

**Current Limitation**: No built-in planning/tracking tool for agent.

**Solution**: Add plan tool enabling:
- ✅ Agent-driven task planning
- ✅ Progress visibility
- ✅ Checkpoint management
- ✅ Recovery from interruptions

**Implementation Path**:
1. Add plan tool to tool registry
2. Implement plan storage (file-based)
3. Add plan display/updates
4. Wire into enhanced prompt

**Effort**: Medium (90-110 hours)  
**ROI**: High (critical for complex tasks)

---

## 10. Configuration Profiles

### Discovery
Codex supports named profiles for different configurations, allowing users to switch presets.

### Key Features

**Profile System**:
```toml
[profiles.gpt4]
model = "gpt-4"
approval_policy = "on-request"
sandbox_mode = "workspace-write"

[profiles.readonly]
model = "gpt-3.5-turbo"
approval_policy = "untrusted"
sandbox_mode = "read-only"
```

**Usage**:
```bash
codex --profile gpt4      # Use GPT-4 profile
codex --profile readonly  # Use read-only profile
```

### Value for Code Agent

**Current Limitation**: Must re-specify settings each session.

**Solution**: Add profile system:
- ✅ Named configuration presets
- ✅ Quick switching via CLI flag
- ✅ Per-project profiles
- ✅ Default profile setting

**Implementation Path**:
1. Add profiles table to config
2. Implement profile loading
3. Add `--profile` CLI flag
4. Handle merging with CLI overrides

**Effort**: Low (30-40 hours)  
**ROI**: Medium (improves UX for power users)

---

## 11. Zero Data Retention (ZDR)

### Discovery
Codex supports ZDR compliance for sensitive environments.

### Key Features

**ZDR Mode**:
- Option to run without any data persistence
- No session history saved
- No logs written
- Useful for security-sensitive work

### Value for Code Agent

**Current Limitation**: All sessions are persisted.

**Solution**: Add ZDR option:
- ✅ Optional no-persistence mode
- ✅ Disable all logging in ZDR mode
- ✅ No session history
- ✅ Configurable per session

**Implementation Path**:
1. Add `--zdr` flag to CLI
2. Disable session persistence when enabled
3. Suppress all logging
4. Document compliance guarantees

**Effort**: Low (20-30 hours)  
**ROI**: Medium (enables enterprise/sensitive use)

---

## 12. Structured Output & JSON Schema

### Discovery
Codex can constrain output to a JSON schema for programmatic consumption.

### Key Features

**Output Schema**:
```bash
codex exec "Extract project info" --output-schema ~/schema.json
```

**Schema Validation**:
- Must follow OpenAI strict schema rules
- Agent constrained to match schema
- JSON output guaranteed

### Value for Code Agent

**Current Limitation**: No way to get structured output from agent.

**Solution**: Add output schema support:
- ✅ Accept JSON schema as input
- ✅ Constrain agent output
- ✅ Return validated JSON

**Implementation Path**:
1. Accept schema in CLI or request
2. Pass schema to enhanced prompt
3. Validate output against schema
4. Return JSON or error

**Effort**: Medium (60-80 hours with schema validation library)  
**ROI**: Medium-High (enables integrations)

---

## 13. Web Search Tool

### Discovery
Codex can perform web searches as part of task execution.

### Key Features

**Web Search**:
- Optional tool the agent can call
- Requires `web_search_request` feature flag
- Can be disabled in config
- Respects approval policies

### Value for Code Agent

**Current Limitation**: Agent cannot search the web.

**Solution**: Add web search tool:
- ✅ Optional web search capability
- ✅ Approval gating
- ✅ Feature flag control
- ✅ Integration with enhanced prompt

**Implementation Path**:
1. Implement web search tool (use Search API)
2. Add feature flag
3. Wire into tool registry
4. Add approval gate

**Effort**: High (100-120 hours including API integration)  
**ROI**: Medium (nice-to-have, not critical)

---

## 14. Image Viewing & Attachment

### Discovery
Codex can view and reason about images.

### Key Features

**Image Input**:
- Paste images via Ctrl+V / Cmd+V
- Attach via `-i` CLI flag
- Multiple images supported
- Workspace-scoped access

### Value for Code Agent

**Current Limitation**: No image support.

**Solution**: Add image support:
- ✅ Accept image files as input
- ✅ Pass to vision models
- ✅ Constrain to workspace
- ✅ Show in conversation

**Implementation Path**:
1. Add image file inputs to prompts
2. Detect vision model capability
3. Encode images for API
4. Add display support

**Effort**: Low-Medium (50-70 hours)  
**ROI**: Low-Medium (useful but not essential)

---

## 15. OpenTelemetry Integration

### Discovery
Codex can emit structured telemetry events via OpenTelemetry.

### Key Features

**OTEL Support**:
- Optional export to OTLP/HTTP or OTLP/gRPC collectors
- Tracks all agent operations
- Event types for monitoring
- Disabled by default (privacy-first)

### Value for Code Agent

**Current Limitation**: No structured observability.

**Solution**: Add OTEL support:
- ✅ Optional telemetry export
- ✅ Structured event types
- ✅ Monitoring dashboard integration
- ✅ Privacy-first (disabled by default)

**Implementation Path**:
1. Add OTEL SDK setup
2. Define event types
3. Emit events from agent loop
4. Add config for exporters

**Effort**: Medium (70-100 hours)  
**ROI**: Medium (useful for operators, not users)

---

## Feature Priority Matrix

| Feature | Value | Effort | Impact | ROI | Priority |
|---------|-------|--------|--------|-----|----------|
| Approval & Sandboxing | Very High | Medium | High | 2.0x | 🔴 P0 |
| Session Persistence | Very High | High | Very High | 1.8x | 🔴 P0 |
| Slash Commands | High | Medium | Medium | 1.5x | 🟠 P1 |
| MCP Integration | Very High | Very High | Very High | 1.4x | 🟠 P1 |
| Context Management | Very High | Medium | Very High | 2.0x | 🔴 P0 |
| AGENTS.md Discovery | High | Low | Medium | 2.5x | 🟢 P2 |
| Change Preview | High | Medium | Medium | 1.5x | 🟠 P1 |
| Plan Tool | High | Medium | High | 1.8x | 🟠 P1 |
| Config Profiles | Medium | Low | Low | 1.3x | 🟢 P2 |
| Non-Interactive Mode | High | High | Medium | 1.2x | 🟠 P1 |
| Web Search | Medium | High | Low | 0.8x | 🟢 P3 |
| Image Support | Medium | Low-Medium | Low | 0.9x | 🟢 P3 |
| JSON Schema Output | Medium | Medium | Medium | 1.2x | 🟢 P2 |
| ZDR Mode | Low | Low | Low | 0.8x | 🟢 P3 |
| OTEL Integration | Medium | Medium | Low | 0.7x | 🟢 P3 |

---

## Implementation Roadmap

### Phase 1: Foundation (Weeks 1-2)
**Essential for safe operation** - 80-100 hours

- [ ] Approval & Sandboxing (config + approval UI)
- [ ] Context Management (token tracking, `/compact`)
- [ ] AGENTS.md Discovery (instruction loading)

**Target**: Safe, context-aware autonomous operation

### Phase 2: Usability (Weeks 3-4)
**Improve user experience** - 100-120 hours

- [ ] Session Persistence (resume capability)
- [ ] Slash Commands (common operations)
- [ ] Plan Tool (task tracking)
- [ ] Change Preview (review before commit)

**Target**: Production-ready for real work

### Phase 3: Extensibility (Weeks 5-6)
**Future-proof architecture** - 200-240 hours

- [ ] MCP Integration (custom tool support)
- [ ] Non-Interactive Mode (automation/CI)
- [ ] Config Profiles (preset management)

**Target**: Enterprise-ready, integrable

### Phase 4: Polish (Weeks 7+)
**Nice-to-haves and optimization** - 100-150 hours

- [ ] Web Search Tool
- [ ] Image Support
- [ ] JSON Schema Output
- [ ] OTEL Integration
- [ ] ZDR Mode

---

## Go Implementation Considerations

### Architecture Patterns

**From Codex Rust Codebase**:

1. **Tool Registry Pattern**
   ```rust
   // Codex pattern
   pub struct ToolRegistry {
       handlers: HashMap<String, Box<dyn ToolHandler>>,
   }
   
   // Go equivalent
   type ToolRegistry struct {
       handlers map[string]ToolHandler
   }
   ```

2. **Event Streaming Pattern**
   ```rust
   // Codex emits structured events
   enum ToolEvent {
       Started,
       Updated,
       Completed,
       Failed,
   }
   
   // Go equivalent
   type ToolEvent struct {
       Type      string
       ItemID    string
       Timestamp time.Time
       Data      interface{}
   }
   ```

3. **Config Merging Pattern**
   ```rust
   // Config hierarchy with defaults
   // CLI overrides > Profile > config.toml > defaults
   
   // Go equivalent
   // Use viper for config management with override precedence
   ```

### Libraries to Consider

- **Config**: `github.com/spf13/viper` (Codex uses TOML files)
- **Validation**: `github.com/go-playground/validator`
- **MCP**: Pending Go MCP library (check modelcontextprotocol.io)
- **OTEL**: `go.opentelemetry.io/otel`
- **SQLite**: `github.com/mattn/go-sqlite3` for session storage

---

## Comparison: Codex vs Code Agent

| Aspect | Codex | Code Agent | Gap |
|--------|-------|-----------|-----|
| **Model Support** | Multiple (GPT-4, o3, etc) | Gemini, OpenAI, Vertex | ✅ More broad |
| **Sandboxing** | OS-level (macOS, Linux, Windows) | None | ❌ Critical gap |
| **Session Persistence** | Yes (SQLite) | No | ❌ Critical gap |
| **Approval Control** | Granular policies | All-or-nothing | ❌ Significant gap |
| **Context Management** | Auto-summary | Manual | ❌ Important gap |
| **MCP Support** | Full client + server | None | ❌ Major gap |
| **Tool Ecosystem** | Extensible | Fixed set | ❌ Significant gap |
| **Configuration** | Rich (TOML, profiles) | Basic | ❌ Minor gap |
| **Automation Mode** | Yes (exec) | No | ❌ Important gap |
| **Terminal Support** | TUI + non-interactive | REPL only | ✅ Simpler is fine |

---

## Key Learnings

### 1. Safety is Non-Negotiable
Codex's OS-level sandboxing + granular approvals are table-stakes for autonomous operation. Users need to trust the agent.

### 2. Long Tasks Need Context Management
Auto-summary at 75% capacity is elegant. Prevents context window crashes.

### 3. Extensibility is Critical
MCP support makes Codex future-proof. Custom tools via user-provided servers is powerful.

### 4. Session Resumption is Essential
Crashing or long tasks are inevitable. Being able to resume is critical for production use.

### 5. Configuration Matters
Profiles, environment-specific settings, and per-project instructions make the tool work across diverse use cases.

### 6. Events Over Logs
Structured event streams (JSON Lines) enable better monitoring, automation, and integration.

---

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| OS sandboxing complexity | Start with approval policies; add OS sandboxing later |
| Session storage overhead | Use simple file-based format; migrate to SQLite if needed |
| MCP implementation effort | Consider using existing Go MCP libraries; may not need full spec |
| Context management tuning | Start with simple summarization; tune thresholds based on usage |
| Approval UX complexity | Keep initial implementation simple; iterate based on feedback |

---

## Next Steps

### Week 1-2
1. Implement approval policy config
2. Add approval UI to REPL
3. Add context tracking to display

### Week 3-4
1. Implement session storage (file-based JSON)
2. Add resume command
3. Test with multi-turn scenarios

### Week 5+
1. Evaluate Go MCP libraries
2. Plan MCP client implementation
3. Design extensibility points

---

## References

- **Codex Docs**: `/research/codex/docs/`
- **Codex Config**: `/research/codex/docs/config.md`
- **Codex Sandbox**: `/research/codex/docs/sandbox.md`
- **Codex Exec**: `/research/codex/docs/exec.md`
- **Codex Source**: `/research/codex/codex-rs/`
- **MCP Spec**: https://modelcontextprotocol.io/

---

## Document History

| Date | Change | Author |
|------|--------|--------|
| 2025-11-12 | Initial analysis | Copilot |

---

**Status**: ✅ Complete for initial investigation  
**Next Review**: After Phase 1 implementation begins
