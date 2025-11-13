# Code Agent: System Architecture & Design

## Executive Summary

**Code Agent** is an intelligent CLI coding assistant built on Google's ADK framework. It orchestrates three major subsystems—**Display** (terminal UI), **Model** (LLM provider abstraction), and **Agent** (agentic reasoning loop)—through a clean component composition pattern.

### Key Metrics
- **~1000 lines** of critical code (highly scalable for learning)
- **~30 tools** across 8 categories
- **3 LLM backends** (Gemini, Vertex AI, OpenAI)
- **Zero external tool frameworks** (no Cline, Claude Code, LangChain dependency)

---

## 1. System Architecture Overview

### 1.1 High-Level Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                          USER                                   │
│                     (Terminal / REPL)                          │
└──────────────────────────────┬──────────────────────────────────┘
                               │ "How do I write a Rust server?"
                               ▼
                    ┌──────────────────────┐
                    │    REPL Loop         │
                    │  (readline + CLI)    │
                    │                      │
                    │ • Prompt> input      │
                    │ • Parse built-in cmds│
                    │ • Invoke agent       │
                    └──────┬───────────────┘
                           │ genai.Content
                           ▼
         ┌─────────────────────────────────────┐
         │      Agent (ADK Framework)          │
         │  ┌──────────────────────────────────┤
         │  │ Agentic Loop:                    │
         │  │ 1. Call LLM with context         │
         │  │ 2. Parse tool calls from response│
         │  │ 3. Execute tools (read file,     │
         │  │    execute command, etc.)        │
         │  │ 4. Stream results to Display     │
         │  │ 5. Repeat until stop_reason      │
         │  └──────────────────────────────────┤
         └────────────────┬──────────────────────┘
                          │
            ┌─────────────┼─────────────┐
            │             │             │
            ▼             ▼             ▼
      ┌─────────┐  ┌─────────┐  ┌──────────┐
      │  Tools  │  │   LLM   │  │ Display  │
      │         │  │ Backend │  │  Output  │
      │ • File  │  │         │  │ Rendering│
      │ • Edit  │  │Gemini   │  │          │
      │ • Exec  │  │Vertex AI│  │ • Colors │
      │ • Search│  │OpenAI   │  │ • Markdown
      │ (30+)   │  │         │  │ • Spinner│
      └─────────┘  └─────────┘  └──────────┘
```

### 1.2 Component Architecture (4-Part System)

The application uses **composition over inheritance**. Four components work together:

| Component | Package | Role | Key Type(s) |
|-----------|---------|------|------------|
| **Display** | `internal/display/*` | Terminal UI rendering, markdown formatting, streaming output | `Renderer`, `StreamingDisplay` |
| **Model** | `pkg/models/*` | LLM provider abstraction, model selection, capability tracking | `Registry`, `Config`, `LLM` |
| **Agent** | ADK framework (local fork) | Agentic loop, tool execution, context management | `agent.Agent` |
| **Session** | `internal/session/*` | Persistence, token tracking, history management | `SessionManager`, `Runner` |

All orchestrated via:
```go
type Components struct {
    Display *DisplayComponents
    Model   *ModelComponents
    Agent   agent.Agent
    Session *SessionComponents
}
```

---

## 2. Detailed Component Analysis

### 2.1 Display Subsystem (`internal/display/*`)

**Purpose**: All terminal UI rendering and output formatting

**Architecture**:
```
display/
├── renderer/          # Core Renderer (colors, styles)
├── streaming/         # Real-time output during agent execution
├── banner/           # Welcome messages
├── components/       # EventTimeline, event types
├── formatters/       # Output format strategies (markdown, JSON)
├── styles/          # Color palettes, output constants
├── terminal/        # Terminal capabilities detection
├── tooling/         # Tool execution rendering
└── tools/           # Display tools (DisplayMessage)
```

**Key Classes**:

```go
// Core renderer for ANSI colors and formatting
type Renderer struct {
    boldFn, colorFn func(string) string
    // Methods: Bold(), Cyan(), Red(), Green(), Yellow()
}

// Real-time output during agent thinking/execution
type StreamingDisplay struct {
    // Handles thinking tokens, tool calls, results
}

// Timeline for collecting events in a request
type EventTimeline struct {
    events []TimelineEvent
}

// Event types: Thinking, Executing, Result, Success, Error, Progress
enum EventType {
    EventTypeThinking
    EventTypeExecuting
    EventTypeResult
    EventTypeSuccess
    EventTypeError
    EventTypeProgress
}
```

**Output Format Support**:
- **Rich** (default): Colors, markdown rendering, spinners, progress
- **Plain**: No ANSI codes
- **JSON**: Structured output for machines

**Usage Example**:
```go
spinner := display.NewSpinner(renderer, "Processing...")
spinner.Start()
// ... do work ...
spinner.Stop()

timeline := display.NewEventTimeline()
timeline.AddEvent(display.EventTypeThinking, "Agent is thinking...")
renderer.RenderTimeline(timeline)
```

### 2.2 Model Subsystem (`pkg/models/*`)

**Purpose**: LLM provider abstraction and model selection

**Architecture**:
```
models/
├── registry.go         # Model catalog and resolution
├── types.go           # Config, Capabilities structs
├── adapter.go         # LLM interface adapter
├── factories/         # Model registration (init functions)
├── gemini.go          # Gemini SDK integration
├── openai.go          # OpenAI SDK integration
└── openai_adapter.go  # OpenAI → ADK adapter
```

**Model Registry Pattern**:

```go
type Registry struct {
    models           map[string]Config    // canonical definitions
    aliases          map[string]string    // user shortcuts
    modelsByProvider map[string][]string  // provider grouping
}

// Resolution: explicit model ID > explicit backend > default
func (r *Registry) ResolveModel(modelID, backend string) Config
```

**Supported Models**:

| Backend | Models |
|---------|--------|
| **Gemini** | gemini-2.5-flash (default), gemini-1.5-pro, gemini-1.5-flash |
| **Vertex AI** | gemini-2.5-flash-vertex, gemini-1.5-pro-vertex |
| **OpenAI** | gpt-4o, gpt-4-turbo |

**Model Config Example**:
```go
type Config struct {
    ID             string
    Name           string
    DisplayName    string
    Backend        string  // "gemini" | "vertexai" | "openai"
    ContextWindow  int     // e.g., 1,000,000
    Capabilities   struct {
        VisionSupport      bool
        ToolUseSupport     bool
        LongContextWindow  bool
        CostTier           string  // "economy", "standard", "premium"
    }
    RecommendedFor []string  // ["coding", "analysis"]
    IsDefault      bool
}
```

### 2.3 Agent Subsystem (ADK Framework)

**Purpose**: Autonomous reasoning loop with tool execution

**Integration Point**: `google.golang.org/adk/agent`

**Agentic Loop**:
```
1. Agent receives user message + context
2. Calls LLM backend with:
   - System prompt (from internal/prompts/)
   - Conversation history
   - Available tools (with JSON schema)
3. LLM responds with:
   - Thinking (optional, if enabled)
   - Tool calls or final response
4. Agent executes tools:
   - Invokes each tool with validated input
   - Collects outputs
5. Agent appends tool results to context
6. Loop back to step 2 (until stop_reason = END_TURN or ERROR)
```

**Tool Registration**:
```go
// Tools are discovered dynamically
agent.RegisterTool(tool.Tool) // Each tool has:
// - Name: "read_file"
// - Description: "Reads file content..."
// - InputType: ReadFileInput (with JSON schema)
// - OutputType: ReadFileOutput
// - Handler: func(ctx, input) output
```

**Agent Uses** (orchestration/agent.go):
```go
func InitializeAgentComponent(ctx context.Context, cfg *config.Config, llm model.LLM) (agent.Agent, error) {
    // 1. Create LLM backend client
    // 2. Register all tools
    // 3. Create agent with system prompt
    // 4. Return agent.Agent
}
```

### 2.4 Session Subsystem (`internal/session/*`)

**Purpose**: Persistence, state management, token tracking

**Architecture**:
```
session/
├── manager.go              # Session CRUD
├── models.go              # Session data structures
└── persistence/
    ├── sqlite_service.go   # GORM SQLite backend
    └── migrations/         # DB schema
```

**SessionManager**:
```go
type SessionManager struct {
    sessionService session.Service  // Interface: Create, Get, List, Delete
    dbPath         string
    appName        string
}

// Methods:
// • CreateSession(ctx, userID, sessionName) Session
// • GetSession(ctx, userID, sessionID) Session
// • ListSessions(ctx, userID) []Session
// • DeleteSession(ctx, userID, sessionID) error
```

**Token Tracking**:
```go
type SessionTokens struct {
    summary struct {
        totalInputTokens  int
        totalOutputTokens int
        requestCount      int
    }
}

// Methods: TrackRequest(input, output), GetSummary()
```

**Persistence**: SQLite via GORM
```
~/.adk-code/sessions.db
  ├── sessions table
  │   ├── app_name
  │   ├── user_id
  │   ├── session_id
  │   ├── state (JSON)
  │   └── created_at, updated_at
  └── messages table (conversation history)
```

---

## 3. Application Lifecycle

### 3.1 Startup Sequence

```go
// main.go
func main() {
    ctx := context.Background()
    
    // 1. Load configuration from CLI + env
    cfg, args := config.LoadFromEnv()
    
    // 2. Handle special commands (/new-session, /list-sessions)
    if clicommands.HandleSpecialCommands(ctx, args, &cfg) {
        os.Exit(0)
    }
    
    // 3. Create and run application
    application, err := app.New(ctx, &cfg)
    application.Run()
}
```

### 3.2 Application Initialization (app.New)

```go
// internal/app/app.go
func New(ctx context.Context, cfg *config.Config) (*Application, error) {
    // 1. Setup signal handling (Ctrl+C awareness)
    signalHandler := runtime.NewSignalHandler(ctx)
    ctx = signalHandler.Context()
    
    // 2. Orchestrate all components
    components, err := orchestration.NewOrchestrator(ctx, cfg).
        WithDisplay().     // Terminal UI, renderers, streaming
        WithModel().       // LLM registry, model selection
        WithAgent().       // ADK agent with tools
        WithSession().     // Session persistence, token tracking
        Build()
    
    // 3. Print welcome banner
    fmt.Print(components.Display.BannerRenderer.RenderStartBanner(...))
    
    // 4. Initialize REPL
    repl, err := repl.New(repl.Config{...})
    
    return &Application{
        ctx: ctx,
        agent: components.Agent,
        repl: repl,
        ...
    }, nil
}
```

### 3.3 REPL Loop (repl.Run)

```go
// internal/repl/repl.go
func (r *REPL) Run(ctx context.Context) {
    for {
        // 1. Read user input
        input, err := r.readline.Readline()
        
        // 2. Handle built-in commands (/help, /models, /use)
        if cli.HandleBuiltinCommand(input, ...) {
            continue
        }
        
        // 3. Create user message
        userMsg := &genai.Content{
            Role: genai.RoleUser,
            Parts: []*genai.Part{{Text: input}},
        }
        
        // 4. Run agent
        result, err := r.runner.Run(ctx, userMsg)
        
        // 5. Collect timeline and render
        timeline := display.NewEventTimeline()
        // ... populate timeline from agent results ...
        r.renderer.RenderTimeline(timeline)
        
        // 6. Track tokens in session
        r.sessionTokens.TrackRequest(inputTokens, outputTokens)
    }
}
```

---

## 4. Tool Ecosystem

### 4.1 Tool Registration Pattern (4 Steps)

Every tool follows this pattern:

```go
// Step 1: Define Input/Output with JSON schema
type MyToolInput struct {
    Param string `json:"param" jsonschema:"Description of param"`
}

type MyToolOutput struct {
    Success bool   `json:"success"`
    Result  string `json:"result,omitempty"`
    Error   string `json:"error,omitempty"`
}

// Step 2: Create handler
func NewMyTool() tool.Tool {
    handler := func(ctx tool.Context, input MyToolInput) MyToolOutput {
        // Implementation
        if err != nil {
            return MyToolOutput{Success: false, Error: err.Error()}
        }
        return MyToolOutput{Success: true, Result: "..."}
    }
    
    // Step 3: Wrap with functiontool
    t, _ := functiontool.New(
        functiontool.Config{
            Name:        "my_tool",
            Description: "Does something useful",
        },
        handler,
    )
    
    // Step 4: Register globally
    common.Register(common.ToolMetadata{
        Tool:      t,
        Category:  common.CategoryFileOperations,
        Priority:  1,
        UsageHint: "...",
    })
    
    return t
}
```

### 4.2 Tool Categories (8 Total)

| Category | Purpose | Tools |
|----------|---------|-------|
| **File Operations** | Read, write, list files | ReadFile, WriteFile, ReplaceInFile, ListDirectory, SearchFiles |
| **Code Editing** | Apply patches, edit lines | ApplyPatch, EditLines, SearchReplace |
| **Search & Discovery** | Find code, preview changes | PreviewReplace, FileSearch |
| **Execution** | Run commands, shell scripts | ExecuteCommand, ExecuteProgram, GrepSearch |
| **Workspace** | Project analysis | FileInfo, WorkspaceAnalysis |
| **Display** | Agent→UI feedback | DisplayMessage, UpdateTaskList |
| **V4A Patches** | Alternative patch format | ApplyV4APatch |
| **Base** | Tool discovery | Registry, ErrorCodes |

### 4.3 Key Tools by Frequency of Use

```
🔴 Daily Use:
  • ReadFile: Core for understanding code
  • WriteFile: Create new files
  • ExecuteCommand: Run code, tests, builds

🟡 Common Use:
  • ReplaceInFile: Make targeted edits
  • EditLines: Line-by-line changes
  • GrepSearch: Find occurrences

🟢 Specialized Use:
  • ApplyPatch: Complex multi-file changes
  • SearchFiles: Bulk discovery
  • V4A patches: Diff-based edits
```

---

## 5. MCP (Model Context Protocol) Support

### 5.1 What is MCP?

**Model Context Protocol (MCP)** enables the agent to dynamically connect to external tool servers at runtime, extending capabilities beyond built-in tools. Instead of coding tools directly into Code Agent, you can spin up MCP servers and configure them in a JSON file.

**Key Benefits**:

- 🔌 **Unlimited Tools**: Add new tools without modifying Code Agent
- 🌐 **External Servers**: Connect to community-provided or custom MCP servers
- ⚡ **Easy Integration**: Simple JSON configuration
- 🔄 **Hot Reload**: Reload servers without restarting agent
- 🛡️ **Isolation**: Tools run in separate processes (safer, easier to manage)

### 5.2 MCP Architecture

**Components**:

| Component | Package | Role |
|-----------|---------|------|
| **MCP Config** | `internal/config/mcp.go` | Loads MCP server definitions from JSON |
| **MCP Manager** | `pkg/mcp/manager.go` | Multi-server orchestration |
| **Transport Factory** | `pkg/mcp/transport.go` | Creates MCP client transports (stdio, SSE, HTTP) |
| **MCP Commands** | `internal/cli/commands/mcp.go` | User-facing `/mcp` commands in REPL |
| **ADK mcptoolset** | `google.golang.org/adk/tool/mcptoolset` | MCP protocol implementation (Google ADK) |

**Data Flow**:

```
┌─────────────────┐
│  MCP Config     │
│  (JSON file)    │
└────────┬────────┘
         │ {servers: {...}}
         ▼
┌─────────────────────────┐
│   MCP Manager           │
│  (Multi-server coord)   │
└──────┬──────┬──────┬────┘
       │      │      │
       ▼      ▼      ▼
  ┌────────┐ ┌────────┐ ┌────────┐
  │Server 1│ │Server 2│ │Server 3│
  │(stdio) │ │(SSE)   │ │(HTTP)  │
  └────┬───┘ └───┬────┘ └───┬────┘
       │        │           │
       ▼        ▼           ▼
  ┌─────────────────────────────┐
  │   Agent (via mcptoolset)    │
  │  Tools available to LLM     │
  └─────────────────────────────┘
```

### 5.3 Configuration Format

**File Location**: `~/.adk-code/config.json` (default) or `AK_CONFIG_PATH` env var

**Example Configuration**:

```json
{
  "mcp": {
    "servers": {
      "filesystem": {
        "type": "stdio",
        "command": "mcp-server-filesystem",
        "args": ["--root", "/home/user/projects"]
      },
      "github": {
        "type": "stdio",
        "command": "mcp-server-github",
        "env": {
          "GITHUB_TOKEN": "${GITHUB_TOKEN}"
        }
      },
      "web_scraper": {
        "type": "sse",
        "url": "http://localhost:8080/sse"
      },
      "database": {
        "type": "http",
        "url": "http://localhost:3000",
        "headers": {
          "Authorization": "Bearer ${DB_TOKEN}"
        }
      }
    }
  }
}
```

**Server Configuration Options**:

```go
type ServerConfig struct {
    // Transport type
    Type     string                 // "stdio" | "sse" | "http"
    
    // For stdio transport
    Command  string                 // e.g., "mcp-server-filesystem"
    Args     []string               // Command arguments
    
    // For SSE/HTTP transport
    URL      string                 // e.g., "http://localhost:8080/sse"
    
    // Shared options
    Env      map[string]string      // Environment variables (supports ${VAR} substitution)
    Headers  map[string]string      // HTTP headers for SSE/HTTP (supports ${VAR})
    Timeout  int                    // Connection timeout (seconds)
    Tools    []string               // Optional: whitelist specific tools (if empty, all tools exposed)
}
```

### 5.4 MCP Manager

**Responsibilities**:

- Load server configurations from JSON
- Create MCP transports (stdio, SSE, HTTP)
- Initialize mcptoolset instances for each server
- Aggregate tools from all servers
- Handle server lifecycle (start, stop, reload)
- Track server connection status

**Usage in Agent**:

```go
// In orchestration/agent.go
func InitializeAgentComponent(ctx context.Context, cfg *config.Config, llm model.LLM) (agent.Agent, error) {
    // 1. Initialize built-in tools
    builtinTools := registerBuiltinTools()
    
    // 2. Initialize MCP servers and aggregate tools
    mcpManager, err := mcp.NewManager(cfg.MCP)
    mcpTools, err := mcpManager.LoadAllTools(ctx)
    
    // 3. Combine built-in + MCP tools
    allTools := append(builtinTools, mcpTools...)
    
    // 4. Create agent with combined tools
    return agent.New(llm, allTools, systemPrompt)
}
```

### 5.5 CLI Commands for MCP

**In-REPL Commands**:

```bash
/mcp list
  → Displays all connected MCP servers and status
  
/mcp tools <server>
  → Lists all tools available from a specific server
  
/mcp reload
  → Hot-reload MCP servers without restarting agent
  
/mcp status
  → Shows connection details and health of each server
```

**Example Session**:

```bash
> /mcp list
Connected MCP Servers:
  ✓ filesystem  (stdio) - 15 tools
  ✓ github      (stdio) - 8 tools
  ○ web_scraper (sse)   - Connection pending...
  ✗ database    (http)  - Failed: Connection refused

> /mcp tools filesystem
Tools from filesystem:
  • read_file(path: string) → string
  • write_file(path: string, content: string) → bool
  • list_directory(path: string) → string[]
  ... (12 more)

> /mcp reload
Reloading MCP servers...
✓ filesystem reloaded
✓ github reloaded
✗ web_scraper failed: timeout
```

### 5.6 How MCP Tools Appear to the Agent

Once configured, MCP tools are indistinguishable from built-in tools. The agent can:

```bash
User: "Clone the repo and read the README"
       ↓
Agent: [Uses github MCP tool to clone]
       [Uses filesystem MCP tool to read]
       ↓
User: [Gets result]
```

**Tool Naming**:

- Built-in: `read_file`, `execute_command`, etc.
- MCP: `mcp_<server>_<tool_name>` (e.g., `mcp_github_clone_repo`)
- Optional: Configure tool prefix in config

### 5.7 Transport Types

| Transport | Best For | Setup |
|-----------|----------|-------|
| **stdio** | Local CLI tools, simple deployment | Run command, inherit stdin/stdout |
| **SSE** | Web services, long-running servers | HTTP GET with event stream |
| **HTTP** | REST APIs, microservices | HTTP POST with request body |

---

## 6. Configuration & Environment

### 6.1 Configuration Loading (config.LoadFromEnv)

```text
Priority (highest to lowest):
1. CLI flags           (e.g., --model gemini-2.5-flash)
2. Environment vars    (e.g., GOOGLE_API_KEY)
3. Config file         (e.g., ~/.adk-code/config.json) [future]
4. Defaults            (e.g., gemini-2.5-flash, ~/.adk-code/sessions.db)
```

### 6.2 CLI Flags

```bash
# LLM Configuration
code-agent --model gemini/2.5-flash          # Explicit model
code-agent --backend gemini                  # Explicit backend
code-agent --model gpt-4o                    # OpenAI model

# Session Management
code-agent --session my-session              # Named session
code-agent --db /path/to/sessions.db        # Custom DB location

# Output & UI
code-agent --output-format plain             # plain | rich | json
code-agent --typewriter                      # Enable typewriter effect

# Working Directory
code-agent --working-directory /path/to/src # Agent's working dir

# Thinking/Reasoning
code-agent --enable-thinking                 # Enable long thinking
code-agent --thinking-budget 5000            # Max thinking tokens
```

### 6.3 Environment Variables

```bash
# Gemini (Google AI)
export GOOGLE_API_KEY=...                    # API key for Gemini

# Vertex AI (GCP)
export GOOGLE_CLOUD_PROJECT=my-project       # GCP project ID
export GOOGLE_CLOUD_LOCATION=us-central1     # GCP region
export GOOGLE_GENAI_USE_VERTEXAI=true       # Enable Vertex AI

# OpenAI
export OPENAI_API_KEY=...                    # OpenAI API key
```

---

## 6. Error Handling & Safety

### 6.1 Error Types (`pkg/errors/`)

```go
type ErrorCode string

const (
    CodeFileNotFound   ErrorCode = "file_not_found"
    CodeFileNotReadable           = "file_not_readable"
    CodeToolNotFound              = "tool_not_found"
    CodeInternal                  = "internal_error"
    CodeExecution                 = "execution_failed"
    CodeValidation                = "validation_failed"
)

type ToolError struct {
    Code    ErrorCode
    Message string
    Details string
}
```

### 6.2 Tool Safeguards

```go
// ReplaceInFile: Reject empty replacements (would delete content)
if input.NewText == "" {
    return error("Use edit_lines with mode='delete' for deletions")
}

// ReplaceInFile: Max replacement count to prevent accidents
if count > maxAllowed {
    return error("Too many replacements. Use preview_replace_in_file first")
}

// ApplyPatch: Dry-run mode to preview changes
applyPatch(..., dryRun=true)  // Shows changes without applying

// All tools: Type-safe JSON schema validation
// (enforced by ADK framework at runtime)
```

---

## 7. Key Design Patterns

### 7.1 Builder Pattern with Orchestrator

**Problem**: Multiple components with dependencies (Display → Agent → Session)

**Solution**: Fluent orchestrator
```go
components, err := NewOrchestrator(ctx, cfg).
    WithDisplay().
    WithModel().
    WithAgent().
    WithSession().
    Build()
```

**Benefits**:
- Single place to see component ordering
- Error propagation at each step
- Dependencies automatically checked

### 7.2 Tool Factory Pattern

**Problem**: 30+ tools need to be created and registered

**Solution**: Each tool's `NewXxxTool()` calls `common.Register()` during `init()`

**Files**: `tools/*/xxx_tool.go` (e.g., `tools/file/read_tool.go`)

**Benefit**: Automatic discovery, no manual registry

### 7.3 Adapter Pattern for LLM Backends

**Problem**: Gemini, OpenAI, Vertex AI have different APIs

**Solution**: Adapter layer in `pkg/models/`
```go
type LLMAdapter interface {
    GenerateContent(ctx, prompt, tools) Response
}

// Implementations:
// - GeminiAdapter (wraps google.golang.org/genai)
// - OpenAIAdapter (wraps github.com/openai/openai-go)
// - VertexAIAdapter (wraps google.golang.org/genai with Vertex endpoint)
```

### 7.4 Component Composition (over Inheritance)

**Problem**: Need modular, testable subsystems

**Solution**: Four independent components composed in orchestration
```
Display ────┐
Model ──────├─→ Application.Run()
Agent ──────┤
Session ────┘
```

**Benefit**: Can test each component independently

---

## 8. Deployment & Execution Modes

### 8.1 Interactive Mode (Default)

```bash
$ code-agent
❯ How do I write a Python FastAPI server?
[Agent thinks, executes tools, displays results]
❯ Can you add authentication?
...
```

### 8.2 Batch Mode (Via Stdin)

```bash
$ echo "Create a Rust CLI app" | code-agent --session batch-1
[Executes once, exits]
```

### 8.3 Session Persistence

```bash
# Create session
$ code-agent --session project-alpha

# Session state saved to ~/.adk-code/sessions.db
# Later:
$ code-agent --session project-alpha  # Resumes with history
```

---

## 9. Testing Strategy

### 9.1 Test Organization

```
adk-code/
├── internal/app/
│   ├── app.go
│   ├── app_init_test.go          # Initialization tests
│   ├── repl_test.go              # REPL behavior
│   └── ...
├── tools/
│   ├── file/
│   │   ├── read_tool.go
│   │   └── file_tools_test.go    # Tool unit tests
│   └── ...
└── pkg/models/
    ├── models_test.go            # Registry tests
    └── ...
```

### 9.2 Test Utilities

```go
// pkg/testutil/ provides:
// • Mock factories
// • Temporary file/directory helpers
// • Mock LLM for testing agent behavior
```

### 9.3 Makefile Targets

```bash
make test           # Run all tests
make coverage       # Generate coverage report
make check          # fmt + vet + lint + test (pre-commit check)
```

---

## 10. Extensibility Guide

### 10.1 Adding a New Tool

1. Create `tools/CATEGORY/new_tool.go`
2. Define Input/Output structs
3. Implement NewXxxTool() with handler
4. Call common.Register() in init()
5. Export in `tools/tools.go`
6. Run `make test` to verify

### 10.2 Adding a New LLM Backend

1. Create `pkg/models/backends/new_backend.go`
2. Implement LLMAdapter interface
3. Register in model factory
4. Add to pkg/models/factories/
5. Test with CLI: `--backend new-backend`

### 10.3 Adding Display Format

1. Create `internal/display/formatters/new_format.go`
2. Implement Formatter interface
3. Register in display factory
4. Test with CLI: `--output-format new-format`

---

## 11. Performance Considerations

### 11.1 Memory Management

- **Event timeline**: Bounded to current request (auto-cleared)
- **Session history**: Persisted to SQLite (not in-memory)
- **Tool outputs**: Streamed when possible (not buffered)

### 11.2 Concurrency

- **Agent loop**: Synchronous (waits for tool results)
- **Display rendering**: Concurrent (streaming output via goroutines)
- **Context cancellation**: Honored throughout (Ctrl+C safe)

### 11.3 File Operations

- **Large files**: Use `offset`/`limit` in ReadFile tool
- **Atomic writes**: Always write complete content (no truncation)
- **Workspace detection**: VCS-aware (Git, Mercurial)

---

## 12. Summary: Key Takeaways

| Aspect | What Makes It Special |
|--------|----------------------|
| **Architecture** | Clean 4-part composition (Display, Model, Agent, Session) |
| **Scalability** | ~1000 lines core → easy to understand, extend |
| **Tool System** | Type-safe, JSON schema validated, auto-registered |
| **LLM Support** | 3 backends (Gemini, Vertex, OpenAI), swappable at runtime |
| **UX** | Rich terminal rendering, streaming output, spinner feedback |
| **Persistence** | SQLite sessions, conversation history, token tracking |
| **Testing** | Comprehensive test coverage, Makefile targets |

---

## Next Steps: Further Learning

1. **Run the app**: `make build && make run`
2. **Read main.go**: Entry point (140 lines)
3. **Study orchestration/builder.go**: Component wiring (140 lines)
4. **Explore tools/file/read_tool.go**: Tool pattern (100 lines)
5. **Modify repl.go**: Add a new built-in command
6. **Create a new tool**: Follow the 4-step pattern

