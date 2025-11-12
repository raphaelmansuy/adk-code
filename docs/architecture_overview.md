# Code Agent - Architecture Overview

**Current Architecture Analysis** - November 12, 2025

## System Architecture (Current)

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                              │
│                    (Application Entry)                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    internal/app/                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   app.go     │  │   repl.go    │  │  session.go  │     │
│  │ (lifecycle)  │  │   (REPL)     │  │ (sessions)   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────────────────────────────────────────┐     │
│  │  init_*.go files (6 component initializers)      │     │
│  └──────────────────────────────────────────────────┘     │
└──────┬───────────────┬───────────────┬────────────────────┘
       │               │               │
       ▼               ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐
│   agent/     │ │   display/   │ │    session/          │
│              │ │              │ │                      │
│ • coding_    │ │ • renderer   │ │ • manager            │
│   agent.go   │ │ • spinner    │ │ • models             │
│ • dynamic_   │ │ • typewriter │ │ • sqlite             │
│   prompt.go  │ │ • tool_*     │ │                      │
│ • xml_       │ │ • streaming  │ │                      │
│   builder    │ │ • banner     │ │                      │
│              │ │ • events     │ │                      │
│              │ │ (~25 files)  │ │                      │
└──────┬───────┘ └──────────────┘ └──────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│                        tools/                                │
│  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐          │
│  │  file  │  │  edit  │  │  exec  │  │ search │          │
│  └────────┘  └────────┘  └────────┘  └────────┘          │
│  ┌────────┐  ┌─────────┐  ┌────────┐                      │
│  │ display│  │workspace│  │  v4a   │                      │
│  └────────┘  └─────────┘  └────────┘                      │
│                                                              │
│           common/ (registry, error_types)                   │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow (Current)

```
User Input
    │
    ▼
  REPL (internal/app/repl.go)
    │
    ├─── Built-in Commands (/help, /tokens, etc.)
    │    └─── pkg/cli/commands/
    │
    └─── Agent Requests
         │
         ▼
    Agent Runner (ADK)
         │
         ├─── Model (LLM Backend)
         │    │
         │    └─── internal/llm/backends/
         │         ├─── gemini.go
         │         ├─── openai.go
         │         └─── vertexai.go
         │
         └─── Tool Calls
              │
              ▼
         Tool Registry (tools/common/)
              │
              ├─── file tools
              ├─── edit tools
              ├─── exec tools
              ├─── search tools
              ├─── workspace tools
              ├─── display tools
              └─── v4a tools
              │
              ▼
         Tool Results
              │
              ▼
         Display Rendering
              │
              └─── display/ package
                   ├─── renderer
                   ├─── spinner
                   ├─── tool_renderer
                   └─── streaming
              │
              ▼
         Terminal Output
```

## Package Dependencies (Simplified)

```
main.go
  └─── cmd/commands/
  └─── internal/app/
       ├─── internal/config/
       ├─── internal/llm/
       │    └─── internal/llm/backends/
       ├─── agent/
       │    ├─── agent/prompts/
       │    ├─── tools/
       │    └─── workspace/
       ├─── display/
       │    ├─── display/banner/
       │    ├─── display/components/
       │    ├─── display/formatters/
       │    ├─── display/renderer/
       │    ├─── display/styles/
       │    └─── display/terminal/
       ├─── session/
       ├─── tracking/
       └─── pkg/
            ├─── pkg/cli/
            ├─── pkg/errors/
            └─── pkg/models/
```

## Problem Areas (Visual)

### 1. Display Package - Too Large

```
Current (display/ = 4000+ LOC, 25+ files):
┌─────────────────────────────────────────────┐
│              display/                        │
│  • ansi, styles, terminal                   │
│  • renderer, markdown, tool_renderer        │
│  • spinner, typewriter, paginator           │
│  • streaming, events, banner                │
│  • formatters, components                   │
│  • tool_adapter, result_parser              │
│  • factory, facade                          │
│  ⚠️ Mixed abstraction levels                │
│  ⚠️ Tight coupling                          │
└─────────────────────────────────────────────┘

Proposed (organized subpackages):
┌──────────────────────────────────────────────┐
│         display/                             │
│  ├─ core/       (primitives)                │
│  ├─ components/ (UI widgets)                │
│  ├─ renderers/  (content rendering)         │
│  ├─ streaming/  (streaming display)         │
│  ├─ events/     (event handling)            │
│  └─ facade.go   (backward compat)           │
│  ✅ Clear boundaries                         │
│  ✅ Testable in isolation                   │
└──────────────────────────────────────────────┘
```

### 2. App Package - Too Many Responsibilities

```
Current (internal/app/ = "God Object"):
┌─────────────────────────────────────────────┐
│           internal/app/                      │
│  • Application lifecycle                    │
│  • REPL implementation                      │
│  • Session management                       │
│  • Signal handling                          │
│  • Component initialization (6 files)       │
│  • Display setup                            │
│  • Model setup                              │
│  • Agent setup                              │
│  ⚠️ Single Responsibility Principle violated│
└─────────────────────────────────────────────┘

Proposed (split into focused packages):
┌──────────────────────────────────────────────┐
│  internal/app/           (lifecycle only)    │
│  internal/repl/          (REPL logic)        │
│  internal/runtime/       (signals, context)  │
│  internal/orchestration/ (component builder) │
│  ✅ Clear responsibilities                   │
│  ✅ Easier to test                           │
└──────────────────────────────────────────────┘
```

### 3. Session Management - Split Across 3 Locations

```
Current:
┌─────────────────┐  ┌──────────────────┐  ┌────────────────┐
│   session/      │  │  internal/data/  │  │ internal/data/ │
│                 │  │                  │  │   sqlite/      │
│ • manager.go    │  │ • repository.go  │  │ • session.go   │
│ • models.go     │  │                  │  │ • models.go    │
│ • sqlite.go     │  │                  │  │ • adapter.go   │
└─────────────────┘  └──────────────────┘  └────────────────┘
  ⚠️ Confusing ownership and boundaries

Proposed:
┌───────────────────────────────────────────────┐
│         internal/session/                     │
│  ├─ session.go    (domain models)            │
│  ├─ manager.go    (high-level API)           │
│  ├─ repository.go (interface)                │
│  └─ storage/                                  │
│      ├─ sqlite/   (SQLite impl)              │
│      └─ memory/   (in-memory impl)           │
│  ✅ Single location for session logic        │
└───────────────────────────────────────────────┘
```

## Tool Registration Pattern

### Current (init-based, fragile):

```go
// tools/file/read_tool.go
func init() {
    _, _ = NewReadFileTool()  // Side effect on import
}

func NewReadFileTool() (tool.Tool, error) {
    // ... create tool
    common.Register(metadata)  // Auto-register
    return t, err
}
```

**Issues**:
- Init order dependencies
- Side effects on package import
- Hard to test in isolation
- Cannot control registration in tests

### Proposed (explicit, testable):

```go
// tools/registry/loader.go
func LoadAllTools() (*common.ToolRegistry, error) {
    reg := common.NewToolRegistry()
    
    // Explicit registration
    if err := registerFileTools(reg); err != nil {
        return nil, err
    }
    if err := registerEditTools(reg); err != nil {
        return nil, err
    }
    // ... more tools
    
    return reg, nil
}

func registerFileTools(reg *common.ToolRegistry) error {
    // Explicit, testable, controllable
    tools := []func() (tool.Tool, error){
        file.NewReadFileTool,
        file.NewWriteFileTool,
        // ...
    }
    
    for _, factory := range tools {
        t, err := factory()
        if err != nil {
            return err
        }
        if err := reg.Register(t); err != nil {
            return err
        }
    }
    return nil
}
```

**Benefits**:
- ✅ Explicit control flow
- ✅ Testable in isolation
- ✅ No side effects
- ✅ Better error handling

## Refactoring Impact Summary

| Component | Current LOC | Current Files | Proposed Structure | Impact |
|-----------|-------------|---------------|-------------------|--------|
| display/ | ~4000 | 25+ | Split into 5 subpackages | HIGH |
| internal/app/ | ~2000 | 15+ | Split into 4 packages | HIGH |
| session/* | ~800 | 8 | Consolidate to 1 location | MEDIUM |
| tools/ | ~5000 | 30+ | Add explicit loader | MEDIUM |
| Other | ~15664 | 50+ | Standardize organization | LOW |

## Architecture Quality Metrics

### Before Refactoring
- ✅ No circular dependencies
- ⚠️ Large package (display/)
- ⚠️ God Object (app/)
- ⚠️ Split responsibilities (session/)
- ⚠️ Fragile registration (init())

### After Refactoring (Target)
- ✅ No circular dependencies (maintained)
- ✅ Reasonable package sizes (<1000 LOC each)
- ✅ Single Responsibility Principle adhered
- ✅ Clear ownership and boundaries
- ✅ Explicit, testable initialization
- ✅ Comprehensive documentation
- ✅ Improved test coverage (>70%)

## Implementation Approach

```
Phase 1: Foundation (2 days)
   └─── Baseline metrics, documentation

Phase 2: Display (4 days)
   └─── Restructure into subpackages
        └─── Facade for backward compat

Phase 3: App (5 days)
   └─── Split into focused packages
        └─── Builder pattern for init

Phase 4: Session (3 days)
   └─── Consolidate to single location

Phase 5: Tools (4 days)
   └─── Explicit registration pattern

Phase 6: Org (3 days)
   └─── Standardize pkg structure

Phase 7: Testing (4 days)
   └─── Enhanced test utilities

Phase 8: Docs (3 days)
   └─── Comprehensive documentation

Phase 9: Quality (4 days, optional)
   └─── Performance, linting, benchmarks
```

## Risk Mitigation

```
┌──────────────────────────────────────┐
│     Feature Branch (per phase)       │
│                                      │
│  ┌────────────────────────────┐    │
│  │  1. Implement changes       │    │
│  └───────────┬────────────────┘    │
│              │                      │
│  ┌───────────▼────────────────┐    │
│  │  2. Run tests frequently    │    │
│  └───────────┬────────────────┘    │
│              │                      │
│  ┌───────────▼────────────────┐    │
│  │  3. Validate before merge   │    │
│  └───────────┬────────────────┘    │
│              │                      │
│  ┌───────────▼────────────────┐    │
│  │  4. Tag release point       │    │
│  └───────────┬────────────────┘    │
│              │                      │
│              ▼                      │
│     ┌────────────────┐             │
│     │ Merge to main  │             │
│     └────────────────┘             │
│                                      │
│  If issues: git revert / rollback   │
└──────────────────────────────────────┘
```

---

**See also**:
- `docs/refactor_plan.md` - Detailed implementation guide
- `docs/refactoring_summary.md` - Executive summary
- `docs/draft.md` - Analysis working notes
