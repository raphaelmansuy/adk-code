# AI Coding Agent Instructions for `adk_training_go`

## Project Overview

This is a **code_agent CLI tool** - an AI-powered coding assistant built with Google ADK Go (Agent Development Kit). It demonstrates how to build sophisticated LLM agents with file operations, terminal execution, code search, and iterative problem solving.

**Key directories:**
- `code_agent/` - Main agent implementation using Google ADK (llmagent framework)
- `research/` - Reference implementations (adk-go SDK source, Cline repository)
- `doc/` - Documentation and comparisons

## Architecture & Core Concepts

### The Agent System (Google ADK Go)
The code_agent uses Google's ADK Go framework with the llmagent pattern:
1. **Agent** (agent/coding_agent.go) - Configures system prompt + tools
2. **Model** - Gemini 2.5 Flash (via genai client)
3. **Tools** - Autonomous functions agents can call (read/write/execute)
4. **Runner** - Manages sessions and agent execution loop

The agent operates **autonomously** - it receives a user request, generates tool calls, executes them, processes results, and iterates until completion.

### Workspace Management (Multi-Root Support)
The `workspace/` package abstracts file paths for multi-workspace support:
- **Backward compatible**: Single-directory mode by default
- **Smart initialization**: `SmartWorkspaceInitialization()` tries config → detection → fallback
- **Path resolution**: Supports workspace hints (`@frontend:src/index.ts`)
- **VCS detection**: Automatically finds Git/Mercurial repos, extracts commit hashes

For agents: Always use workspace hints for ambiguous paths in monorepo context. Paths are always relative to primary workspace.

### Display System (Rich Terminal Output)
The `display/` package renders agent output beautifully:
- **Renderer** - Handles output formats (rich/plain/json) and ANSI colors
- **Streaming** - Real-time output from agent tools and thinking
- **Markdown** - Converts markdown to formatted terminal output
- **Tool rendering** - Shows what tool is running + structured results

## Essential Development Workflows

### Building & Testing
```bash
make build           # Compile to ./code-agent binary
make test            # Run tests in all packages
make coverage        # Generate coverage report (opens HTML)
make check           # Run fmt, vet, lint, test (comprehensive)
```

### Running Locally
```bash
export GOOGLE_API_KEY="your-api-key"  # Required
make run                               # Build and run interactive CLI
./code-agent --output-format=rich      # With options
./code-agent --typewriter              # Enable typewriter effect
```

### Code Quality
```bash
make fmt             # Format code (go fmt)
make vet             # Run go vet
make lint            # Run golangci-lint
```

## Critical Patterns & Conventions

### Tool Definition Pattern (tools/ package)
Every tool follows this pattern:
```go
// 1. Define Input struct with JSON schema tags
type MyToolInput struct {
    Param1 string `json:"param1" jsonschema:"description"`
    Param2 *int   `json:"param2,omitempty" jsonschema:"optional param"`
}

// 2. Define Output struct
type MyToolOutput struct {
    Success bool   `json:"success"`
    Result  string `json:"result"`
    Error   string `json:"error,omitempty"`
}

// 3. Implement handler function
handler := func(ctx tool.Context, input MyToolInput) MyToolOutput {
    // Implementation
    return MyToolOutput{...}
}

// 4. Register with functiontool.New()
return functiontool.New(functiontool.Config{
    Name:        "tool_name",
    Description: "What this tool does",
}, handler)
```

See `tools/file_tools.go` for examples: `ReadFileTool`, `WriteFileTool`, `ReplaceInFileTool`.

### File Editing Best Practices
- **read_file** (line ranges) - Use `offset`/`limit` for large files
- **write_file** (atomic writes) - Default: `atomic=true`, creates directories, validates sizes
- **search_replace** (SEARCH/REPLACE blocks) - Most precise edits; whitespace-tolerant
- **edit_lines** (structural changes) - For fixing braces, adding imports
- **apply_patch** (complex edits) - Use `dry_run=true` first

**Critical**: For `write_file`, always provide **complete intended content** - never truncate.

### System Prompt Evolution
- **Legacy prompt** - Basic tool definitions (see `coding_agent.go`)
- **Enhanced prompt** (enhanced_prompt.go) - Better tool selection guide, safety practices
- Workspace context is **dynamically injected** at agent initialization (BuildEnvironmentContext)

**Key injections:**
- Workspace summary and paths
- Environment metadata (Git remotes, commit hashes)
- Path usage conventions (relative to primary workspace)

## Integration Points & Data Flows

### Agent → Tools → Workspace → File I/O
```
User Input → Agent.Run()
  ↓
  generates tool calls based on system prompt
  ↓
Tool execution (read_file, write_file, etc.)
  ↓
Workspace resolver normalizes paths
  ↓
Actual file I/O operations
```

### Session Management
- **Session** (ADK concept) - Tracks conversation history + agent state
- **Runner** - Orchestrates sessions and streams events
- Main.go creates in-memory session service; productionizable to database

### Event Streaming
Tools and model output stream as events (text, function calls, responses). Main.go prints events with enhanced rendering:
- Spinner for "thinking"
- Tool execution banners
- Result parsing
- Error handling

## Common Implementation Patterns

### Adding a New Tool
1. Define Input/Output structs in `tools/your_tool.go`
2. Implement handler function with validation
3. Call `functiontool.New()` to register
4. Add to agent's tool list in `coding_agent.go` (NewCodingAgent function)
5. Document in system prompt (enhanced_prompt.go)
6. Add tests (see `file_tools_test.go`)

### Detecting Project Context
Use `GetProjectRoot()` (agent/coding_agent.go) - traverses upward to find `go.mod`.

### Multi-Workspace Support (Feature Flag)
```go
config := codingagent.Config{
    Model:                    model,
    WorkingDirectory:         workingDir,
    EnableMultiWorkspace:     true,  // Feature flag
}
```

When enabled, uses `.workspace.json` config or auto-detects workspace structure.

## Testing & Debugging

### Running Tests
```bash
go test ./...           # All tests
go test ./tools/...     # Tools package only
go test -v -run TestReadFile ./tools  # Specific test
```

### Common Debug Patterns
- **Path issues**: Use `list_directory` to verify structure
- **Command execution**: Check working_dir and argument quoting
- **File encoding**: Tools assume UTF-8; binary files may fail gracefully
- **Workspace resolution**: Check `.workspace.json` or use `GetPrimaryRoot()`

### Test Structure (file_tools_test.go example)
- Create temp files
- Call tools directly (bypass agent)
- Assert Success/Error fields
- Clean up with defer

## Performance & Constraints

- **Large files**: Always use `offset`/`limit` in read_file
- **Search**: `grep_search` scans entire files; optimize patterns
- **Atomic writes**: Small overhead but prevents data corruption
- **Workspace detection**: Max depth/count configurable (defaults: depth=3, count=10)

## Git & VCS Awareness

The workspace package automatically detects:
- Git: `.git` directories, commit hash, remote URLs
- Mercurial: `.hg` directories

Useful for agents that need VCS context. VCS metadata stored in WorkspaceRoot struct (see workspace/types.go).

## Key Files to Reference

| File | Purpose |
|------|---------|
| `code_agent/main.go` | CLI entry point, event printing |
| `code_agent/agent/coding_agent.go` | Agent factory, tool registration |
| `code_agent/agent/enhanced_prompt.go` | System prompt template |
| `code_agent/tools/file_tools.go` | Core read/write/edit tools |
| `code_agent/tools/terminal_tools.go` | Command execution |
| `code_agent/workspace/manager.go` | Workspace orchestration |
| `code_agent/workspace/resolver.go` | Path resolution logic |
| `code_agent/display/renderer.go` | Output formatting |

## Dependencies & External Systems

- **google.golang.org/adk** - Agent Development Kit (llmagent, runner, session)
- **google.golang.org/genai** - Gemini API client
- **Built on**: Go 1.24+ (check go.mod for exact version)

API Key required: `GOOGLE_API_KEY` environment variable (Gemini API).

---

## Quick Wins for Agents

✅ **To implement a new file operation**: Copy `WriteFileTool` pattern from `tools/file_tools.go`
✅ **To add shell execution**: See `ExecuteCommandTool` in `tools/terminal_tools.go`
✅ **To handle paths in monorepo**: Use workspace.FormatPathWithHint() + resolver.ResolvePath()
✅ **To test changes**: Run `make test` or specific test file
✅ **To debug tool behavior**: Implement dry_run mode (see `ApplyPatchTool`)

## Important Project Boundaries & Gotchas

### What This Codebase Is NOT
- **Not a full LLM framework**: It demonstrates one pattern (llmagent) but research/ has alternatives
- **Not production-ready as-is**: Session service is in-memory; use for demos/prototypes
- **Not a replacement for Cline/Claude Code**: It's a learning reference implementation

### Critical Implementation Details
- **Path resolution**: Always relative to primary workspace, never absolute paths from working_dir
- **Tool registration**: Must add to NewCodingAgent tool list AND document in enhanced_prompt.go
- **Error handling**: Use Success/Error fields in Output structs (not panics or thrown exceptions)
- **File safety**: Atomic writes prevent corruption; size validation prevents accidental truncation

### Common Mistakes to Avoid
1. ❌ Forgetting to add new tool to agent's tool list in `coding_agent.go`
2. ❌ Truncating files in write_file - ALWAYS include complete content
3. ❌ Using absolute paths instead of workspace-relative paths
4. ❌ Not testing tool handler functions directly before wiring up
5. ❌ Ignoring workspace hints in monorepo paths

## Development Workflow

### Before Starting Work
1. **Understand the intent clearly** - What problem are you solving? Read related issues/PRs in research/
2. **Search the codebase first** - Is there existing similar code? Check `tools/` for patterns
3. **Reference existing implementations** - Find a similar tool and use it as a template
4. **Plan your approach** - What files need changes? In what order?

### While Working on a Task
- **Create a brainstorm file** for notes: `brainstorm/YYYY-MM-DD-HH-MM-task_name.md`
  - This is a **draft document** - don't worry about formatting
  - Jot down ideas, decisions, dead ends, learnings
  - Use for reflection and debugging if things go wrong

### During Implementation
- **Follow the tool pattern**: Input struct → Output struct → handler → functiontool.New()
- **Test incrementally**: Verify each piece before wiring it together
- **Run `make check`** before committing (fmt, vet, lint, test all pass)
- **Update enhanced_prompt.go** if adding/changing tools

### When Task Is Complete
- **Write a summary log** in `logs/YYYY-MM-DD-HH-MM_task_name.md`
  - Concise summary of what was done (not a play-by-play)
  - What worked, what didn't, key learnings
  - Not formatted - just notes for future reference
- **Clean up brainstorm files** optionally (they're working docs)

### Example Workflow
```
1. brainstorm/2025-11-10-14-30-add_lint_tool.md
   - Notes about what a lint tool would do
   - Design questions, research findings
   
2. Implement tool in tools/lint_tools.go following WriteFileTool pattern
   - Define LintInput, LintOutput structs
   - Implement handler
   - Register with functiontool.New()
   - Add tests to lint_tools_test.go
   
3. Update agent/coding_agent.go
   - Import and create lintTool
   - Add to agent's tool list
   
4. Update agent/enhanced_prompt.go
   - Document the new tool in system prompt
   
5. Run make check - verify all tests pass
   
6. logs/2025-11-10-14-30_add_lint_tool.md
   - Implemented new lint_tool that wraps golangci-lint
   - Output structured as JSON for agent consumption
   - Key learning: tools should always return Success/Error in output struct
```

## Reference Implementations in ./research

Use these as learning references:
- **adk-go/**: Official Google ADK Go framework - inspect llmagent pattern, session management
- **cline/**: Alternative agent design (TypeScript/MCP protocol) - different tool abstraction patterns

When stuck on how to implement something, search the research folder for similar patterns.