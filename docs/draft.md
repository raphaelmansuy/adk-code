# Code Agent - Deep Analysis Notes

**Date**: November 12, 2025  
**Status**: In Progress - Core Analysis Phase  
**Scope**: Comprehensive codebase analysis with focus on architecture, patterns, and design decisions

---

## 1. PROJECT OVERVIEW

### What is Code Agent?
A CLI-based AI coding assistant powered by Google's ADK (Agent Development Kit) Go framework. It provides an interactive REPL for developers to ask coding questions, request code generation, and execute commands with AI assistance.

**Key Stat**: ~140 lines main.go + orchestrated component architecture (~12 internal packages, ~8 tool categories)

### Core Value Proposition
- **Multi-model support**: Gemini 2.5 Flash, OpenAI GPT-4o, Vertex AI
- **Rich terminal UI**: Glamour markdown rendering, typewriter effects, spinners
- **Agent-driven**: Uses Google ADK llmagent pattern for autonomous tool execution
- **Session persistence**: SQLite-backed session management
- **Extensive tooling**: 30+ tools across 8 categories (file ops, code editing, execution, workspace)

---

## 2. ARCHITECTURE PATTERNS

### 2.1 Builder Pattern with Orchestrator
The application uses a sophisticated builder pattern via `Orchestration.Orchestrator`:

```
Main.go
  └─> App.New(ctx, cfg)
        └─> Orchestrator.NewOrchestrator(ctx, cfg)
              ├─> WithDisplay()    [creates 4 display components]
              ├─> WithModel()      [initializes LLM + registry]
              ├─> WithAgent()      [creates ADK agent]
              ├─> WithSession()    [session + runner]
              └─> Build()          [returns composite Components]
```

**Pattern Benefits**:
- Clear separation of concerns
- Lazy initialization (only built what's needed)
- Error propagation at each step
- Dependency ordering enforced

### 2.2 Component Composition (NOT Inheritance)
Four major component groups orchestrated together:

| Component | Package | Key Types | Responsibility |
|-----------|---------|-----------|-----------------|
| **Display** | `internal/display/*` | Renderer, BannerRenderer, StreamingDisplay, TypewriterPrinter | Terminal UI, markdown rendering, rich output |
| **Model** | `pkg/models/*` | Registry, Config, Capabilities | LLM abstraction, provider handling, model selection |
| **Agent** | ADK framework | agent.Agent (from google.golang.org/adk) | Tool execution, agentic loop, context management |
| **Session** | `internal/session/*` | SessionManager, Runner, SessionTokens | Session persistence, token tracking, history |

### 2.3 Application Lifecycle

```
main()
  ├─> config.LoadFromEnv()           [env vars + CLI flags]
  ├─> clicommands.HandleSpecialCmds() [e.g., /new-session]
  ├─> app.New(ctx, cfg)              [orchestrate all components]
  ├─> application.Run()
  │    └─> repl.Run(ctx)
  │          ├─> readline loop (interactive)
  │          ├─> cli.HandleBuiltinCommand() [/help, /models, etc.]
  │          └─> processUserMessage()
  │               └─> agent.Run(ctx, userMsg)
  │                    └─> [agentic loop: think → tool call → result]
  └─> application.Close()            [cleanup resources]
```

---

## 3. TOOL ECOSYSTEM

### 3.1 Tool Registration Pattern

Every tool follows a 4-step pattern:

```go
// Step 1: Define Input/Output structs with JSON schema tags
type ReadFileInput struct {
    Path   string `json:"path" jsonschema:"..."`
    Offset *int   `json:"offset,omitempty" jsonschema:"..."`
    Limit  *int   `json:"limit,omitempty" jsonschema:"..."`
}

type ReadFileOutput struct {
    Content      string `json:"content"`
    Success      bool   `json:"success"`
    Error        string `json:"error,omitempty"`
    TotalLines   int    `json:"total_lines"`
    FilePath     string `json:"file_path"`
}

// Step 2: Create handler function
handler := func(ctx tool.Context, input ReadFileInput) ReadFileOutput {
    // Implementation with proper error handling
}

// Step 3: Wrap with functiontool.New()
t, err := functiontool.New(functiontool.Config{
    Name:        "read_file",
    Description: "Reads file content with optional offset/limit",
}, handler)

// Step 4: Register with tool registry
if err == nil {
    common.Register(common.ToolMetadata{
        Tool:      t,
        Category:  common.CategoryFileOperations,
        Priority:  1,
        UsageHint: "...",
    })
}
```

### 3.2 Tool Categories (8 Total)

| Category | Tools | Location | Key Functions |
|----------|-------|----------|----------------|
| **File Ops** | ReadFile, WriteFile, ReplaceInFile, ListDirectory, SearchFiles | `tools/file/` | Atomic writes, whitespace normalization |
| **Code Editing** | ApplyPatch, EditLines, SearchReplace | `tools/edit/` | Multi-format patch support (unified, v4a) |
| **Search/Discovery** | PreviewReplace, FileSearch | `tools/search/` | Dry-run preview, regex support |
| **Execution** | ExecuteCommand, ExecuteProgram, GrepSearch | `tools/exec/` | Terminal execution, output capture |
| **Workspace** | GetFileInfo, ListDirectory, ProjectAnalysis | `tools/workspace/` | VCS-aware path resolution |
| **Display** | DisplayMessage, UpdateTaskList | `tools/display/` | Agent-to-UI feedback channel |
| **V4A Patches** | ApplyV4APatch | `tools/v4a/` | Alternative patch format |
| **Base** | Registry, Error types | `tools/base/` | Tool discovery + error codes |

### 3.3 Tool Safety Features

Observed safeguards across tools:

```
✓ ReplaceInFile: Rejects empty replacements (would truncate)
✓ ReplaceInFile: Max replacement count validation (prevent accidents)
✓ ApplyPatch: Dry-run mode (--dry-run flag)
✓ ExecuteCommand: Working directory validation
✓ All tools: JSON schema validation + type safety
```

---

## 4. MODEL & LLM ABSTRACTION

### 4.1 Multi-Backend Support

**Three Backends**:
1. **Gemini**: google.golang.org/genai (Google's official SDK)
2. **Vertex AI**: Gemini models via GCP (requires GOOGLE_CLOUD_PROJECT)
3. **OpenAI**: OpenAI GPT-4o (requires OPENAI_API_KEY)

**Key File**: `pkg/models/registry.go` - Dynamic model resolution

### 4.2 Model Registry Design

```
Registry
  ├─ models: map[modelID] → Config       [canonical model definitions]
  ├─ aliases: map[shorthand] → modelID   [user-friendly shortcuts]
  └─ modelsByProvider: map[backend] → []modelID
```

**Resolution Priority**:
1. Explicit model ID from CLI (--model gemini-2.5-flash)
2. Backend from --backend flag or env var
3. Default model (gemini-2.5-flash)

**Factory Pattern**: `factories/` subdir registers models during init()

### 4.3 Config Structure

```go
type Config struct {
    ID             string         // canonical ID
    Name           string         // display name
    DisplayName    string         // UI-friendly name
    Backend        string         // "gemini" | "vertexai" | "openai"
    ContextWindow  int            // max tokens
    Capabilities   struct {       // vision, tools, long context, cost tier
        VisionSupport bool
        ToolUseSupport bool
        LongContextWindow bool
        CostTier string            // "economy" | "standard" | "premium"
    }
    RecommendedFor []string       // ["coding", "analysis", "creative"]
    IsDefault      bool
}
```

---

## 5. INTERNAL PACKAGES (Comprehensive Map)

### 5.1 `internal/app/`
**Responsibility**: Application lifecycle orchestration

| File | Purpose |
|------|---------|
| `app.go` | Main Application struct, Run(), Close(), REPL initialization |
| `components.go` | Type aliases for component re-export (backward compat) |
| `factories.go` | Component factory functions (deprecated, use Orchestrator) |
| `signals.go` | Unix signal handling (SIGINT, SIGTERM) |
| `session.go` | Session creation & management helpers |

**Key Methods**:
- `New(ctx, cfg)` - Creates app with orchestrated components
- `Run()` - Starts REPL loop
- `Close()` - Cleanup (session manager, signal handler, REPL)

### 5.2 `internal/orchestration/`
**Responsibility**: Component builder & initialization

| File | Purpose |
|------|---------|
| `builder.go` | Orchestrator fluent builder API |
| `components.go` | Component type definitions (Display, Model, Session) |
| `display.go` | Display component factory |
| `model.go` | Model registry & LLM initialization |
| `agent.go` | ADK agent creation + tool registration |
| `session.go` | Session manager + token tracking |

**Key Pattern**: All WithX() methods check prior errors, allowing:
```go
components, err := orchestrator.
    WithDisplay().
    WithModel().
    WithAgent().
    WithSession().
    Build()
```

### 5.3 `internal/repl/`
**Responsibility**: Read-Eval-Print Loop

| File | Purpose |
|------|---------|
| `repl.go` | Main REPL loop, readline integration, input processing |

**Features**:
- Readline instance with history file (~/.code_agent_history)
- Built-in command routing (/help, /models, /exit, etc.)
- Agent invocation with spinner feedback
- Event timeline collection for UI rendering
- Context cancellation awareness

### 5.4 `internal/display/`
**Responsibility**: Terminal UI rendering & output formatting

**Subpackages**:
- `banner/` - Welcome/start banners
- `renderer/` - Base Renderer (colors, styles, ANSI)
- `streaming/` - Real-time agent output (thinking, tool execution)
- `components/` - EventTimeline, event types
- `formatters/` - Output formatters (markdown, JSON, plain text)
- `styles/` - Color palettes, output format constants
- `terminal/` - Terminal width/height detection
- `tools/` - Tool execution display helpers
- `core/` - Core display logic

**Key Types**:
- `Renderer` - ANSI color/style application
- `StreamingDisplay` - Progressive output during agent execution
- `TypewriterPrinter` - Animated text output
- `EventTimeline` - Event collection for this request

### 5.5 `internal/session/`
**Responsibility**: Session persistence & token tracking

| File | Purpose |
|------|---------|
| `manager.go` | SessionManager CRUD operations |
| `models.go` | Session data models |
| `persistence/` | SQLite backend (gorm) |

**Key Abstraction**: Separates session service (interface) from implementation (SQLite)

### 5.6 `internal/config/`
**Responsibility**: Configuration from CLI flags & environment

**Key Method**: `LoadFromEnv()` - parses:
- CLI flags (--model, --backend, --output-format, etc.)
- Environment variables (GOOGLE_API_KEY, GOOGLE_CLOUD_PROJECT, etc.)
- Returns config + remaining args

### 5.7 `internal/cli/`
**Responsibility**: Built-in REPL commands

**Commands**:
- `/help` - Display help message
- `/models` - List available models
- `/use` - Switch models at runtime
- `/exit` / `/quit` - Exit REPL
- `/sessions` - Manage sessions

### 5.8 `internal/llm/`
**Responsibility**: LLM provider abstraction (Gemini, Vertex, OpenAI)

**Pattern**: Adapter pattern for each backend
- `gemini.go` - Google Gemini SDK integration
- `openai.go` - OpenAI SDK integration
- `vertex.go` - Vertex AI integration

### 5.9 Other Internal Packages

| Package | Purpose |
|---------|---------|
| `internal/tracking/` | Token usage tracking for sessions |
| `internal/runtime/` | Signal handling, context management |
| `internal/prompts/` | Agent system prompts |
| `internal/errors/` | Error handling utilities |

---

## 6. KEY DESIGN DECISIONS

### 6.1 Why ADK Framework?

**Chosen over alternatives** (Cline, Claude Code, direct API calls):
- ✓ Official Google framework for autonomous agents
- ✓ Tool abstraction layer (Tool interface, JSON schema support)
- ✓ Session management built-in
- ✓ Streaming & event-driven architecture
- ✓ Function calling with automatic type marshaling

### 6.2 Why Multiple LLM Backends?

**Strategic reasoning**:
- 🎯 Vendor lock-in avoidance (Gemini → OpenAI → Vertex)
- 🎯 Cost optimization (pick cheapest model for task)
- 🎯 Feature coverage (some models have vision, others have function calling)
- 🎯 Fallback strategy (if one API rate-limits, switch backends)

### 6.3 Tool Registration at Package Init

**Pattern**: Each tool's `NewXxxTool()` calls `common.Register()` during `init()`

**Pro**: Automatic discovery, no manual registry maintenance  
**Con**: Makes mocking harder in tests (global state)

### 6.4 Component-based Display

Rather than monolithic renderer, **8 specialized packages** under `display/`:
- Concern separation (banner ≠ streaming ≠ formatting)
- Testability (mock individual components)
- Extensibility (add new output formats without touching core)

### 6.5 Orchestrator Pattern > Factory Pattern

**Old approach**: Separate factory functions (still exist in `factories.go`)  
**New approach**: Orchestrator with WithX() methods

**Why**:
- Single place to understand component dependencies
- Fluent API is more readable
- Error collection at each step (fail-fast, not panic)

---

## 7. DATA FLOWS

### 7.1 User Interaction → Agent Execution

```
REPL.readline()
  └─> processUserMessage(ctx, input)
       ├─> Create genai.Content with user text
       ├─> agent.Run(ctx, content)
       │    └─> [ADK agentic loop]
       │         ├─ Call LLM with context
       │         ├─ Parse tool calls from response
       │         ├─ Execute tools (file read, execute command, etc.)
       │         ├─ Collect results in timeline
       │         └─ Repeat until stop
       ├─> Render timeline (thinking, tool execution, results)
       └─> Collect token usage & session state
```

### 7.2 Model Selection Flow

```
CLI: --model gemini-2.5-flash --backend gemini
  └─> config.LoadFromEnv()
       └─> orchestration.InitializeModelComponents(ctx, cfg)
            ├─> registry.ResolveModel(cfg.Model, cfg.Backend)
            │    └─ Returns Config (with ID, Backend, Capabilities)
            └─> llm.NewLLM(provider, modelID, apiKey)
                 └─ Creates google.golang.org/adk/model.LLM instance
```

### 7.3 Session & Token Tracking

```
SessionManager (SQLite backend)
  ├─ CreateSession(ctx, userID, sessionName)
  ├─ GetSession(ctx, userID, sessionID)
  ├─ ListSessions(ctx, userID)
  └─ DeleteSession(ctx, userID, sessionID)

SessionTokens (in-memory tracking)
  ├─ Track input tokens per request
  ├─ Track output tokens per request
  ├─ Accumulate totals in session
  └─ Display in UI
```

---

## 8. KEY FILES TO UNDERSTAND

### Essential Reading Order (Recommended)
1. **`main.go`** (140 lines) - Entry point, initialization
2. **`internal/orchestration/builder.go`** (140 lines) - Component orchestration
3. **`internal/app/app.go`** (140 lines) - Lifecycle management
4. **`internal/repl/repl.go`** (245 lines) - Interactive loop
5. **`tools/file/file_tools.go`** (150 lines) - Tool pattern example
6. **`internal/display/renderer.go`** - Terminal UI facade
7. **`pkg/models/registry.go`** (218 lines) - Model selection logic

### Total Critical Code: ~1000 lines (easily digestible)

---

## 9. CONVENTIONS & PATTERNS

### 9.1 Error Handling
- Uses `pkg/errors/` package with error codes
- Output structs always have Success + Error fields
- Tools never panic (return error in output struct)

### 9.2 Testing
- Unit tests in `*_test.go` files
- Test factories in `pkg/testutil/`
- Makefile has `make test`, `make coverage`, `make check`

### 9.3 Configuration
- Env vars for secrets (GOOGLE_API_KEY, etc.)
- CLI flags for behavior (--model, --output-format, etc.)
- Precedence: CLI flags > env vars > defaults

### 9.4 Code Organization
- `cmd/` patterns (none currently, uses main.go directly)
- `pkg/` for public reusable code (models, errors, workspace)
- `internal/` for app-specific code (repl, display, session)
- `tools/` for tool ecosystem (flat structure, ~30 tools)

---

## 10. EXTERNAL DEPENDENCIES (High-Level)

| Dependency | Purpose | Version |
|------------|---------|---------|
| `google.golang.org/adk` (local fork) | Agent framework, tool abstraction | v0.0.0 (replaced) |
| `google.golang.org/genai` | Gemini API SDK | v1.20.0 |
| `github.com/openai/openai-go` | OpenAI API SDK | v3.8.1 |
| `github.com/charmbracelet/glamour` | Markdown rendering | v0.10.0 |
| `github.com/charmbracelet/lipgloss` | Terminal styling | v1.1.1 |
| `github.com/chzyer/readline` | Interactive CLI | v1.5.1 |
| `gorm.io/gorm` + `sqlite` | Session persistence | v1.31.0 |

---

## 11. STRENGTHS & OBSERVATIONS

### ✅ What's Well Done
1. **Clean separation of concerns** - Each package has clear responsibility
2. **Tool abstraction** - Extensible, type-safe tool system
3. **Multi-backend support** - Strategic abstraction for LLM providers
4. **Component composition** - Orchestrator pattern scales well
5. **Rich terminal UI** - Professional markdown rendering + streaming output
6. **Error handling** - Consistent error struct across all tools
7. **Testing infrastructure** - Makefile targets, test utilities

### ⚠️ Areas for Refinement
1. **Tool registration** - Global state (init functions) makes mocking hard
2. **Orchestrator size** - `orchestration/` package gets large with all factories
3. **REPL complexity** - `repl.go` (245 lines) handles too many concerns
4. **Documentation** - Limited inline comments on complex flows

---

## 12. TODO FOR DOCUMENTATION

- [ ] **ARCHITECTURE.md** - System design diagram, component interaction
- [ ] **TOOL_DEVELOPMENT.md** - Step-by-step guide for adding new tools
- [ ] **QUICK_REFERENCE.md** - Common commands, environment variables
- [ ] **API_INTEGRATION.md** - Backend setup (Gemini, Vertex AI, OpenAI)

---

## Summary Statistics

| Metric | Count |
|--------|-------|
| Go packages | 20+ |
| Tool categories | 8 |
| Total tools | ~30 |
| Main code files | ~15 |
| Internal packages | 11 |
| Lines of critical code | ~1000 |
| Supported LLM backends | 3 |
| CLI commands | 6+ |

