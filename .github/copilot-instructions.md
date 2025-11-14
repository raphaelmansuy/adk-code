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

**Testing REPL Commands:**
- Test with `echo "/<command>" | timeout 10 ./adk-code` to see full output
- Use `grep` to filter specific sections: `... | grep "Ollama"`
- Count output lines: `... | grep "model" | wc -l`
- This helps verify dynamic content is working correctly

**Dynamic Content (like Ollama models):**
- Always implement graceful fallback when external services unavailable
- Log errors without crashing the REPL
- Show user-friendly message if discovery fails
- Example: "Using cached models" vs "Fetching from server"
</repl_best_practices>