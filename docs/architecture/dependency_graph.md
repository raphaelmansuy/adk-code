# Package Dependency Graph

**Generated**: November 12, 2025  
**Project**: code_agent (Go 1.24.4)  
**Total Packages**: ~100 Go files across 20+ packages

## Internal Package Dependency Map

```
┌─────────────────────────────────────────────────────┐
│                  Main Application                    │
│                    main.go                           │
│         (CLI entry point - minimal logic)            │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│              internal/app Package                    │
│     (Application orchestration & lifecycle)          │
│  ├─ Components factory                              │
│  ├─ REPL initialization                             │
│  ├─ Session initialization                          │
│  ├─ Model initialization                            │
│  ├─ Signal handling                                 │
│  └─ Application runner                              │
└────┬────────────┬───────────┬───────────┬───────────┘
     │            │           │           │
     ▼            ▼           ▼           ▼
   REPL        Session     Model      Display
   Setup       Manager     Provider   Components
     │            │           │           │
     ▼            ▼           ▼           ▼
┌────────────┐┌───────────┐┌──────────┐┌─────────┐
│  pkg/cli   ││  session/ ││pkg/llm   ││ display │
│(Commands)  ││ (Manager) ││(Models)  ││(Render) │
└────────────┘└─────────┬─┘└────┬─────┘└────┬────┘
                        │       │           │
                        ▼       ▼           ▼
            ┌───────────────┐  ┌──────────┐ ┌──────────┐
            │internal/data/ │  │pkg/models│ │ styles/  │
            │(Repository)   │  │(Config)  │ │(ANSI)    │
            └───────┬───────┘  └──────────┘ └──────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼
    ┌────────────┐         ┌─────────────┐
    │ SQLite     │         │ In-Memory   │
    │ Storage    │         │ Storage     │
    └────────────┘         └─────────────┘


Agent System
┌──────────────────────────────────────────────┐
│            agent Package                      │
│      (ADK LLMAgent wrapper)                   │
│  ├─ Tool registry                            │
│  ├─ Agent lifecycle                          │
│  ├─ Prompt engineering                       │
│  └─ Tool execution loop                      │
└──────┬─────────────────┬──────────────────────┘
       │                 │
       ▼                 ▼
    Tools          Display Events
   Registry        ├─ Agent thinking
   ├─ file/*       ├─ Tool calls
   ├─ edit/*       ├─ Tool results
   ├─ exec/*       └─ Responses
   ├─ search/*
   └─ workspace/*


Display System (Deep)
┌──────────────────────────────────────────────┐
│         display Package (Main)                │
│  ├─ Facade (backward compatibility)          │
│  ├─ Factory (component creation)             │
│  ├─ Event handling                           │
│  └─ Rendering coordination                   │
└──────┬──────────────────────────────────────┘
       │
   ┌───┴──────────┬────────────┬──────────┬────────────┐
   │              │            │          │            │
   ▼              ▼            ▼          ▼            ▼
┌────────┐  ┌───────┐   ┌────────┐  ┌───────┐  ┌──────────┐
│styles/ │  │terminal│   │formatters│ │renderers│ │components│
│(ANSI)  │  │(Util)  │   │(Custom)  │ │(Content)│ │(Spinner) │
└────────┘  └───────┘   └────────┘  └───────┘  └──────────┘
   │
   └─► ANSI escape codes & color codes
```

## Inter-Package Dependencies

### High-Level Flows

#### 1. User Input Flow
```
CLI Flags/Args
    ↓
pkg/cli (Command parsing)
    ↓
cmd/commands (Command handlers)
    ↓
internal/app (Orchestration)
    ↓
agent.Agent (Core logic)
    ↓
tools/* (Execution)
```

#### 2. Model Resolution Flow
```
CLI --model flag
    ↓
pkg/cli (Parse)
    ↓
pkg/models (Resolve)
    ↓
internal/llm (Backend init)
    ↓
agent.Config (Set model)
```

#### 3. Session Management Flow
```
CLI --session flag
    ↓
internal/app (Initialize)
    ↓
session.Manager (Load/Create)
    ↓
internal/data/* (Persistence)
```

#### 4. Display Flow
```
agent.Agent (Tool events)
    ↓
display.Facade (Route events)
    ↓
display/renderers (Format)
    ↓
display/styles (ANSI codes)
    ↓
Terminal (Output)
```

## Package Inventory with Responsibilities

### Core Packages (Used by Everything)

| Package | Purpose | Dependencies |
|---------|---------|--------------|
| `agent` | LLM agent orchestration via ADK | `display`, `tools`, `pkg/errors` |
| `internal/app` | Application lifecycle & setup | All other internal packages |
| `pkg/errors` | Error types & utilities | None (foundational) |
| `pkg/cli` | CLI argument parsing & model resolution | `pkg/models`, `pkg/errors` |

### Data & Storage Packages

| Package | Purpose | Dependencies |
|---------|---------|--------------|
| `session` | Session management (high-level API) | `internal/data`, `pkg/errors` |
| `internal/data` | Repository interfaces (abstraction) | None (defines contracts) |
| `internal/data/sqlite` | SQLite session persistence | SQLite3 driver |
| `internal/data/memory` | In-memory session storage | None |

### LLM & Model Packages

| Package | Purpose | Dependencies |
|---------|---------|--------------|
| `pkg/models` | Model registry & resolution | `pkg/errors`, model configs |
| `pkg/models/factories` | Model factory implementations | Provider SDKs |
| `internal/llm` | LLM provider abstraction | Provider SDKs |
| `internal/llm/backends` | Gemini, OpenAI, VertexAI adapters | Provider SDKs |

### Display & Output Packages

| Package | Purpose | Dependencies |
|---------|---------|--------------|
| `display` | Main display facade & factory | All display/* subpackages |
| `display/styles` | ANSI colors & text formatting | None |
| `display/terminal` | Terminal utilities | OS system calls |
| `display/components` | UI components (spinner, banner) | `display/styles` |
| `display/renderers` | Content renderers (markdown, etc) | Rendering libraries |
| `display/formatters` | Custom output formatters | None |

### Tool Packages (Agent Tools)

| Package | Purpose | Dependencies |
|---------|---------|--------------|
| `tools/common` | Tool registry & base types | `pkg/errors` |
| `tools/file` | File I/O operations | `pkg/errors`, `workspace` |
| `tools/edit` | File editing (patches, replacement) | `tools/file` |
| `tools/exec` | Command execution | `pkg/errors` |
| `tools/search` | Workspace search operations | `workspace` |
| `tools/workspace` | Workspace analysis | None |
| `tools/display` | Display message tool | `display` |
| `tools/v4a` | V4A patch format parsing | None |

### Infrastructure Packages

| Package | Purpose | Dependencies |
|---------|---------|--------------|
| `workspace` | Workspace detection & management | `pkg/errors` |
| `tracking` | Token usage tracking | `pkg/models` |
| `internal/config` | Configuration types | `pkg/errors` |
| `cmd/commands` | CLI command implementations | All others |

## Dependency Coupling Analysis

### Tightest Coupling (May need refactoring)
1. **display → styles** - 100% dependency (necessary)
2. **agent → tools/common** - 100% dependency (by design)
3. **internal/app → all packages** - Many dependencies (orchestrator role)

### Moderate Coupling (OK for now)
1. **tools/file → workspace** - Needed for path resolution
2. **session → internal/data** - Repository pattern (good)
3. **pkg/cli → pkg/models** - Model resolution (logical)

### Healthy Isolation (Well-separated)
1. **display/styles** - No dependencies on app logic ✅
2. **pkg/errors** - No dependencies at all ✅
3. **tools/v4a** - Standalone patch parsing ✅
4. **workspace** - Minimal dependencies ✅

## External Dependency Map

### Critical External Dependencies
- `google.golang.org/genai` - Gemini API client
- `github.com/openai/openai-go` - OpenAI SDK
- `cloud.google.com/go/vertexai` - VertexAI SDK
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/charmbracelet/glamour` - Markdown rendering
- `gorm.io/gorm` - Database ORM (for data package)

### Optional External Dependencies
- `github.com/chzyer/readline` - Interactive input
- Various charmbracelet libraries - Terminal styling

## Import Path Standardization

### Current Patterns
```go
// Internal packages
import "code_agent/internal/app"
import "code_agent/internal/data"
import "code_agent/internal/llm"

// Public packages
import "code_agent/pkg/cli"
import "code_agent/pkg/errors"
import "code_agent/pkg/models"

// Tools
import "code_agent/tools/file"
import "code_agent/tools/edit"

// Other
import "code_agent/agent"
import "code_agent/display"
```

## Unused/Low-Use Dependencies
(To investigate in Phase 5+)

- `cmd/` - Commands package is large, may need refactoring
- `examples/` - Example code (OK to leave)

## Recommendations for Phase 2

### Refactoring Targets (by priority)
1. **High**: Simplify `internal/app` - too many dependencies
2. **High**: Break down `display` package - too large
3. **Medium**: Consolidate session management into single location
4. **Medium**: Extract tool registration to separate loader
5. **Low**: Create explicit interfaces for provider abstraction

### Dependency Directions to Maintain
- ✅ Always from higher level → lower level
- ✅ Tools should not import agent
- ✅ Display should not import tools
- ✅ Models should not import internal/llm
- ⚠️ internal/app currently violates layering (acceptable for now)

## How to Regenerate This Analysis

```bash
cd code_agent

# View dependency tree
go mod graph | grep 'code_agent' | sort | uniq

# Analyze import cycles
go mod tidy
go list -m all

# Check for unused dependencies
go mod tidy
```
