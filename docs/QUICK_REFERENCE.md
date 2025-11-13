# Quick Reference Guide

One-page cheat sheet for common Code Agent tasks.

---

## Building & Running

```bash
# Build the application
make build              # Binary goes to ../bin/adk-code

# Run with default settings
../bin/adk-code

# Run with specific model
../bin/adk-code --model gemini-2.5-flash

# Run with custom session
../bin/adk-code --session my-project

# Build + run in one command
make run
```

---

## Environment Setup

### Gemini (Google AI) - Default

```bash
export GOOGLE_API_KEY=your-api-key-here
../bin/adk-code
```

### Vertex AI (GCP)

```bash
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_CLOUD_LOCATION=us-central1
export GOOGLE_GENAI_USE_VERTEXAI=true
../bin/adk-code
```

### OpenAI

```bash
export OPENAI_API_KEY=sk-...
../bin/adk-code --model gpt-4o
```

---

## CLI Flags Reference

```bash
# Model Selection
--model gemini-2.5-flash          # Specific model (ID or shorthand)
--backend gemini                  # Backend provider only

# Session Management
--session my-session              # Named session (default: "default")
--db /path/to/sessions.db        # Custom database location

# Output & Display
--output-format plain             # Output format: plain, rich (default), json
--typewriter                      # Enable typewriter effect (slower, animated)

# Working Directory
--working-directory /path/to/src # Agent's base directory (default: cwd)

# Model Reasoning/Thinking
--enable-thinking                 # Enable long reasoning (default: true)
--thinking-budget 5000           # Max thinking tokens (default: 1024)
```

---

## In-REPL Commands

```bash
❯ /help              # Show available commands
❯ /models            # List all available models
❯ /use gemini-2.5-flash    # Switch to different model
❯ /sessions          # List all sessions
❯ /clear             # Clear screen
❯ /exit              # Exit (same as Ctrl+C)
❯ /quit              # Exit
```

---

## Common Prompts & Examples

### Code Generation

```
❯ Create a Python FastAPI server with authentication
❯ Write a Go CLI tool that reads CSV files
❯ Generate a React component for a user profile form
```

### File Analysis

```
❯ What's in the main.go file?
❯ Show me the structure of the tools/ directory
❯ Find all TODO comments in the codebase
```

### Code Editing

```
❯ Add error handling to the ReadFile function
❯ Refactor the REPL loop to be more modular
❯ Update all imports to use the new package structure
```

### Testing

```
❯ Write unit tests for the tool registry
❯ How do I run the test suite?
❯ Can you add integration tests for the agent?
```

---

## Development Commands

```bash
# Code quality checks (run before committing)
make check              # fmt + vet + lint + test

# Individual checks
make fmt               # Format code
make vet               # Run go vet
make lint              # Run linters (requires golangci-lint)
make test              # Run tests
make test-short        # Run short tests only

# Coverage
make coverage          # Generate HTML coverage report

# Cleanup
make clean             # Remove build artifacts

# Dependencies
make deps              # Download dependencies
make deps-tidy         # Tidy go.mod and go.sum
make deps-update       # Update all dependencies
make deps-verify       # Verify dependencies
```

---

## Project Structure At A Glance

```
adk-code/
├── main.go                      # Entry point (140 lines)
├── go.mod, go.sum              # Dependencies
├── Makefile                    # Build targets
├── bin/                        # Compiled binaries
├── internal/                   # App-specific code
│   ├── app/                    # Application lifecycle
│   ├── orchestration/          # Component builder
│   ├── repl/                   # Interactive loop
│   ├── display/                # Terminal UI (8 subpackages)
│   ├── session/                # Persistence + token tracking
│   ├── config/                 # Configuration loading
│   ├── cli/                    # Built-in commands
│   ├── llm/                    # LLM provider abstraction
│   ├── runtime/                # Signal handling
│   ├── tracking/               # Token tracking
│   └── prompts/                # System prompts
├── pkg/                        # Public/reusable code
│   ├── models/                 # Model registry (3 backends)
│   ├── errors/                 # Error handling
│   ├── workspace/              # Path resolution, VCS awareness
│   └── testutil/               # Test utilities
├── tools/                      # Tool ecosystem (~30 tools)
│   ├── file/                   # File operations
│   ├── edit/                   # Code editing
│   ├── exec/                   # Command execution
│   ├── search/                 # Search & discovery
│   ├── display/                # Agent→UI messaging
│   ├── workspace/              # Workspace analysis
│   ├── v4a/                    # V4A patch format
│   └── base/                   # Tool registry
├── docs/                       # Documentation
└── examples/                   # Example code
```

---

## File Locations

```bash
# Executable
../bin/adk-code

# Session database
~/.adk-code/sessions.db

# History file
~/.code_agent_history

# Config file (future)
~/.adk-code/config.json
```

---

## Key Files to Know

| File | Purpose | Size |
|------|---------|------|
| `main.go` | Entry point, initialization | 140 lines |
| `internal/app/app.go` | Application lifecycle | 140 lines |
| `internal/orchestration/builder.go` | Component wiring | 140 lines |
| `internal/repl/repl.go` | Interactive loop | 245 lines |
| `tools/file/read_tool.go` | Tool pattern example | 130 lines |
| `pkg/models/registry.go` | Model selection | 218 lines |

---

## Tool Categories & Examples

```
📁 File Operations      read_file, write_file, replace_in_file, list_directory
✏️ Code Editing         apply_patch, edit_lines, search_replace
🔍 Search/Discovery    preview_replace, search_files, grep_search
⚙️ Execution           execute_command, execute_program
🏢 Workspace           get_file_info, project_analysis
💬 Display             display_message, update_task_list
📦 V4A Patches         apply_v4a_patch
🎛️ Base/Registry       tool_registry, error_types
```

---

## Testing a Single Tool

```bash
# Test a specific package
cd adk-code/tools/file
go test -v ./...

# Test one file
go test -v -run TestReadFile

# Test with coverage
go test -v -cover ./...

# Run from code_agent root
make test
```

---

## Debugging Tips

### Enable Debug Output

Most tools support quiet flags or environment variables:

```bash
# Set verbosity
export DEBUG=1
../bin/adk-code

# Or via RUST_LOG-like pattern (if supported)
export RUST_LOG=debug
../bin/adk-code
```

### Check Tool Registration

In REPL:
```
❯ /help
[Shows all registered tools]
```

### Verify Environment

```bash
# Check if API key is set
echo $GOOGLE_API_KEY
echo $OPENAI_API_KEY

# Check model resolution
../bin/adk-code --model gpt-4o  # Should work if OpenAI key is set
```

### View Session History

Sessions are stored in SQLite:
```bash
# Install sqlite3 CLI if needed
brew install sqlite3

# View sessions
sqlite3 ~/.adk-code/sessions.db "SELECT * FROM sessions;"
sqlite3 ~/.adk-code/sessions.db "SELECT * FROM messages;"
```

---

## Common Issues & Solutions

### "API key not found"

**Solution**: Set environment variable
```bash
export GOOGLE_API_KEY=your-key
```

### "Model not found"

**Solution**: List available models
```
❯ /models
```

Then use one of the listed models:
```bash
../bin/adk-code --model gemini-1.5-pro
```

### "Permission denied" on binary

**Solution**: Make executable
```bash
chmod +x ../bin/adk-code
```

### Tests failing

**Solution**: Run quality checks
```bash
make check    # Formats, vets, lints, tests
```

### Tool not showing up

**Solution**: Verify registration
1. Check tool has `init()` function
2. Verify `common.Register()` is called
3. Check tool is exported in `tools/tools.go`
4. Restart REPL

---

## Architecture in 30 Seconds

```
User Input
   ↓
REPL (readline loop)
   ↓
Agent (ADK agentic loop)
   ↓
Call LLM (Gemini/OpenAI/Vertex)
   ↓
Execute Tools (file ops, commands, etc.)
   ↓
Stream Results to Display
   ↓
Render in Terminal (colors, markdown, spinner)
```

---

## Learning Path

1. **Day 1**: Read `main.go` (140 lines) → understand entry point
2. **Day 1**: Read `internal/orchestration/builder.go` (140 lines) → understand components
3. **Day 2**: Build & run: `make build && make run`
4. **Day 2**: Try some built-in commands: `/help`, `/models`
5. **Day 3**: Read `tools/file/read_tool.go` (130 lines) → understand tool pattern
6. **Day 3**: Create a simple tool (following TOOL_DEVELOPMENT.md)
7. **Day 4**: Read `internal/display/renderer.go` → understand UI
8. **Day 5**: Study agent integration in `internal/repl/repl.go`

---

## Useful Makefile Targets

```bash
make build          # Compile binary
make run            # Build + execute
make test           # Run all tests
make check          # fmt + vet + lint + test (pre-commit)
make coverage       # Generate coverage report
make clean          # Remove artifacts
make help           # Show all targets
make deps-update    # Update dependencies
make lint           # Run linters
make fmt            # Format code
make vet            # Run go vet
```

---

## Resources

- **Architecture Details**: `docs/ARCHITECTURE.md`
- **Tool Development**: `docs/TOOL_DEVELOPMENT.md`
- **Deep Analysis**: `docs/draft.md`
- **Codebase**: `adk-code/`

---

## Key Concepts

| Concept | Definition |
|---------|-----------|
| **Orchestrator** | Builder that wires Display, Model, Agent, Session together |
| **Component** | Independent subsystem (Display, Model, Agent, or Session) |
| **Tool** | Callable function (read file, execute command, etc.) available to agent |
| **Agent** | ADK framework's agentic loop (think → call tools → collect results) |
| **REPL** | Read-Eval-Print Loop (interactive CLI) |
| **Backend** | LLM provider (Gemini, Vertex AI, OpenAI) |
| **Registry** | Catalog of available models or tools |
| **Session** | Named conversation thread with history + tokens |

