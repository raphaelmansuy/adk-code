# AI Coding Agent Instructions for `adk_training_go`

## Quick Context

**What**: AI-powered CLI agent (Google ADK Go) with file I/O, terminal execution, and code search.
**Structure**: `adk-code/` (main), `research/` (reference impls), `doc/` (design docs)
**Tech Stack**: Go 1.24+, Gemini 2.5 Flash API, ADK llmagent framework

## Architecture Essentials

**Agent Lifecycle**: User request → REPL.Run() → Agent.Run() → Tool calls → Process results → Stream to Display → Iterate until complete

**4-Part Component Architecture** (details in `docs/ARCHITECTURE.md`):
- **Display** (`internal/display/*`) - Terminal UI, ANSI colors, markdown rendering, event streaming
- **Model** (`pkg/models/*`) - LLM provider abstraction (Gemini, Vertex AI, OpenAI), model registry, capability tracking
- **Agent** (ADK Framework) - Agentic loop, tool execution, context management
- **Session** (`internal/session/*`) - Persistence, token tracking, conversation history

**Key Systems**:
- **Workspace** - Multi-root path resolution with VCS awareness (Git/Mercurial)
- **Tools** - 30+ autonomous callables across 8 categories (file ops, code editing, execution, search, etc.)
- **REPL** - Interactive read-eval-print loop with built-in commands (`/help`, `/models`, `/use`)
- **Configuration** - CLI flags + environment variables with precedence resolution

**References**:
- See `docs/ARCHITECTURE.md` for system design & data flows
- See `docs/TOOL_DEVELOPMENT.md` for tool creation patterns
- See `docs/QUICK_REFERENCE.md` for CLI flags and configuration

## Essential Commands

```bash
make check              # Quality gate: fmt, vet, lint, test (RUN BEFORE COMMITTING)
make test               # Run all tests
make build              # Compile to ./adk-code
make run                # Build and run (requires GOOGLE_API_KEY)
```

## Important: Terminal Safety

**Avoid Terminal Crashes with Long Output:**
- When creating git commits or PRs with `gh` CLI, use **short, concise messages**
- Long multi-line bodies in terminal commands can crash the shell
- For detailed information, use GitHub web UI after PR creation
- Prefer: `gh pr create --title "Short title" --body "One line description"`
- Example commands that work:
  - `gh pr create --title "feat: Feature name" --body "Brief description"`
  - Use `--body-file <file>` for longer content in a file
- When in doubt, use the GitHub web interface to add extended descriptions

**Pipe Commands to Avoid Pagination Issues:**
- Commands that produce lots of output (git log, gh pr list, etc.) may trigger pager and hang
- **ALWAYS pipe to prevent interactive pager from blocking:**
  - ❌ `gh pr list` (may hang waiting for user input)
  - ✅ `gh pr list | cat` (disables pager, outputs all at once)
  - ✅ `git log --oneline | head -20` (filter output)
  - ✅ `git log --no-pager` (explicitly disable pager)
- Use pipes with common utilities to control output:
  - `| head -N` - Show first N lines
  - `| tail -N` - Show last N lines
  - `| grep "pattern"` - Filter by pattern
  - `| wc -l` - Count lines
  - `| cat` - Disable pager and output directly
- For long-running commands, use `timeout`:
  - `timeout 5 ./bin/adk-code` - Kill after 5 seconds
  - `timeout 10 gh pr view <number>` - Prevent hanging on API calls

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

<repl_best_practices>
## REPL Command Best Practices

**Large Output Handling:**
- REPL commands like `/providers`, `/models`, `/tools` may output 50+ lines
- The pager system handles this automatically - use SPACE to continue, Q to quit
- **Never force-print large outputs** to terminal without pagination
- If implementing new REPL commands that produce >30 lines of output:
  - Use the pagination system: `Display.DisplayPaged(lines)`
  - Split output into multiple pages when needed
  - Provide navigation hints to users

**Testing REPL Commands with Proper Piping:**
- Always pipe test commands to avoid pagination hangs
- Test with proper timeout and piping:
  - ✅ `echo "/providers" | timeout 10 ./adk-code 2>&1 | head -50` - Show first 50 lines
  - ✅ `echo "/models" | timeout 10 ./adk-code 2>&1 | grep "gemini"` - Filter by provider
  - ✅ `echo "/tools" | timeout 10 ./adk-code 2>&1 | tail -20` - Show last 20 lines
  - ✅ `echo "/providers" | timeout 10 ./adk-code 2>&1 | wc -l` - Count total output lines
- Key patterns:
  - `timeout N` prevents hanging on long-running commands
  - `2>&1` captures both stdout and stderr
  - `| head/tail` controls output volume
  - `| grep` filters for specific patterns
  - `| wc -l` counts lines to verify expected output
- Never use interactive commands without timeout/redirect:
  - ❌ `./adk-code` alone (interactive mode)
  - ✅ `echo "/command" | timeout 5 ./adk-code` (non-interactive)

**Dynamic Content (like Ollama models):**
- Always implement graceful fallback when external services unavailable
- Log errors without crashing the REPL
- Show user-friendly message if discovery fails
- Example: "Using cached models" vs "Fetching from server"
</repl_best_practices>