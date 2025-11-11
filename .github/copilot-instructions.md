# AI Coding Agent Instructions for `adk_training_go`

## Quick Context

**What**: AI-powered CLI agent (Google ADK Go) with file I/O, terminal execution, and code search.
**Structure**: `code_agent/` (main), `research/` (reference impls), `doc/` (design docs)
**Tech Stack**: Go 1.24+, Gemini 2.5 Flash API, ADK llmagent framework

## Architecture Essentials

**Agent Lifecycle**: User request → Agent.Run() → Tool calls → Process results → Iterate until complete

**Key Components**:
- **Agent** (`coding_agent.go`) - System prompt + tool registry
- **Model** - Gemini 2.5 Flash (via genai)
- **Tools** - Autonomous callables (file ops, exec, search)
- **Workspace** - Multi-root path resolution with VCS awareness (Git/Mercurial)
- **Display** - Rich terminal rendering (ANSI, markdown, streaming)

## Essential Commands

```bash
make check              # Quality gate: fmt, vet, lint, test (RUN BEFORE COMMITTING)
make test               # Run all tests
make build              # Compile to ./code-agent
make run                # Build and run (requires GOOGLE_API_KEY)
```

## Critical Patterns

**Tool Definition** - All tools follow this template:
```go
// 1. Input/Output structs with JSON schema tags
type MyInput struct {
    Param string `json:"param" jsonschema:"description"`
}
type MyOutput struct {
    Success bool   `json:"success"`
    Error   string `json:"error,omitempty"`
}

// 2. Handler function → functiontool.New()
handler := func(ctx tool.Context, input MyInput) MyOutput { /* impl */ }
return functiontool.New(functiontool.Config{Name: "tool_name"}, handler)
```
Reference: `tools/file_tools.go` (ReadFileTool, WriteFileTool, ReplaceInFileTool)

**File Editing Best Practices**:
- Large files: Use `offset`/`limit` in read_file
- Atomic writes: Always provide **complete content** (never truncate)
- Search/replace: Use whitespace-tolerant SEARCH/REPLACE blocks
- Complex edits: Use `dry_run=true` first

**Path Resolution**:
- Always relative to primary workspace
- Use workspace hints for ambiguous paths: `@frontend:src/index.ts`
- Never use absolute paths from working_dir

<integration_data_flows>
## Implementation Workflow

**Step 1: Research** → Find similar code in `tools/` → Read reference files

**Step 2: Define** → Input/Output structs with JSON schema tags

**Step 3: Implement** → Handler function → `functiontool.New()` → Unit tests

**Step 4: Wire** → Add to `coding_agent.go` tool list → Document in `enhanced_prompt.go`

**Step 5: Verify** → Run `make check` → Verify all tests pass

<testing_debugging>
## Debugging Guide

- **Path issues**: Use workspace hints → check `.workspace.json` → verify with `list_directory`
- **Tool behavior**: Add dry_run mode → test handler directly → bypass agent for isolation
- **Execution**: Check working_dir → verify argument quoting → inspect environment
- **File encoding**: Tools assume UTF-8 → binary files fail gracefully</testing_debugging>

<performance_constraints>
## Key Constraints

- **Large files**: Use `offset`/`limit` in read_file to avoid memory bloat
- **Atomic writes**: Always include **complete content** (never truncate)
- **Path resolution**: Always relative to primary workspace, never absolute working_dir paths
- **Tool registration**: Add to both `coding_agent.go` tool list AND document in `enhanced_prompt.go`</performance_constraints>

<vcs_awareness>
## VCS Awareness

Workspace package auto-detects Git (`.git`, commit hash, remote URLs) and Mercurial (`.hg`). Use WorkspaceRoot struct for VCS metadata (workspace/types.go).</vcs_awareness>

<quick_wins>
## Quick Wins for Agents

✅ **To implement a new file operation**: Copy `WriteFileTool` pattern from `tools/file_tools.go`
✅ **To add shell execution**: See `ExecuteCommandTool` in `tools/terminal_tools.go`
✅ **To handle paths in monorepo**: Use workspace.FormatPathWithHint() + resolver.ResolvePath()
✅ **To test changes**: Run `make test` or specific test file
✅ **To debug tool behavior**: Implement dry_run mode (see `ApplyPatchTool`)
</quick_wins>

<boundaries_gotchas>
## Important Project Boundaries & Gotchas

**What This Codebase Is NOT**:
- Not a full LLM framework (demonstrates one pattern; research/ has alternatives)
- Not production-ready as-is (in-memory session; use for demos/prototypes)
- Not a replacement for Cline/Claude Code (learning reference implementation)

**Critical Implementation Details**:
- **Path resolution**: Always relative to primary workspace, never absolute paths from working_dir
- **Tool registration**: Must add to NewCodingAgent tool list AND document in enhanced_prompt.go
- **Error handling**: Use Success/Error fields in Output structs (not panics or thrown exceptions)
- **File safety**: Atomic writes prevent corruption; size validation prevents accidental truncation

**Common Mistakes to Avoid**:
1. ❌ Forgetting to add new tool to agent's tool list in `coding_agent.go`
2. ❌ Truncating files in write_file - ALWAYS include complete content
3. ❌ Using absolute paths instead of workspace-relative paths
4. ❌ Not testing tool handler functions directly before wiring up
5. ❌ Ignoring workspace hints in monorepo paths</boundaries_gotchas>

<development_workflow>
## Development Workflow

**Before starting**: Understand intent → Search codebase for similar patterns → Reference existing implementations → Plan approach

**During implementation**:
- Follow the tool pattern: Input struct → Output struct → handler → functiontool.New()
- Test incrementally: Verify each piece before wiring
- Run `make check` before committing
- Update `enhanced_prompt.go` if adding/changing tools

**When complete**: Summarize your work in a new log file under `logs/YYYY-MM-DD-hh-mm_task_name.md`. Briefly describe:
- What was implemented or changed
- What worked well
- Any challenges or blockers encountered
- Key learnings or follow-up actions

This helps track progress and share insights for future contributors.
</development_workflow>

<reference_implementations>
## Reference Implementations in ./research

Use these as learning references:
- **adk-go/**: Official Google ADK Go framework - inspect llmagent pattern, session management
- **cline/**: Alternative agent design (TypeScript/MCP protocol) - different tool abstraction patterns

When stuck on how to implement something, search the research folder for similar patterns.
</reference_implementations>