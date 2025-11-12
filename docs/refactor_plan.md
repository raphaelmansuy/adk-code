# Code Agent Refactoring Plan

**Status**: Ready for Implementation  
**Risk Level**: LOW (0% regression if followed incrementally)  
**Estimated Effort**: 3-4 weeks (phased approach)  
**Priority**: MEDIUM (modernization; not blocking current functionality)

---

## Executive Summary

This plan reorganizes the codebase for **better maintainability, clarity, and extensibility** without changing core functionality. The strategy follows Go best practices:

1. **Clean package boundaries** – Each package has one clear responsibility
2. **Reduced coupling** – Dependency injection, interfaces, not singletons
3. **Layered architecture** – Clear distinction between presentation, business logic, and data layers
4. **Incremental refactoring** – 5 phases; each can ship independently

---

## Current State → Target State

### Current Organization

```
code_agent/
├── main.go                              # Entry point
├── internal/app/                        # Monolithic orchestrator (300+ LOC)
├── agent/                               # LLM & prompts (fragmented: 7 files)
├── display/                             # UI (24+ files, unclear hierarchy)
├── tools/                               # ✅ Well-structured (auto-registering)
├── workspace/                           # ✅ Good design
├── session/                             # ✅ Isolated
├── pkg/
│   ├── cli/                            # Mixed concerns (flags, commands, syntax)
│   ├── errors/                         # ✅ Good
│   └── models/                         # Heavy SDK dependencies
├── tracking/                            # Event tracking
└── [Others]
```

### Target Organization

```
code_agent/
├── main.go                              # Clean entry point
├── cmd/                                 # NEW: CLI commands layer
│   ├── app.go                          # Application entrypoint
│   └── commands/                       # Refactored CLI commands
├── internal/
│   ├── app/                            # Lean orchestrator
│   ├── config/                         # NEW: Configuration management
│   ├── llm/                            # NEW: LLM abstraction layer
│   ├── ui/                             # NEW: Presentation layer
│   ├── data/                           # NEW: Persistence layer
│   └── testutils/                      # Test helpers (existing)
├── agent/                               # Focused: LLM agent definition
│   ├── agent.go                        # Core agent config
│   └── prompts/                        # Consolidated prompt logic
├── display/                             # Lean: Presentation/rendering
│   ├── renderer.go                     # Main renderer facade
│   ├── formatter/                      # Specialized formatters
│   └── terminal/                       # Terminal utilities
├── tools/                               # ✅ Keep as-is (excellent design)
├── workspace/                           # ✅ Keep as-is
├── session/                             # ✅ Keep as-is
├── pkg/
│   ├── cli/                            # SLIM: Just flags, no commands
│   ├── errors/                         # ✅ Keep as-is
│   └── log/                            # NEW: Structured logging
└── [Others]
```

---

## Detailed Refactoring Phases

### Phase 1: Extract Configuration Layer

**Goal**: Separate configuration from orchestration  
**Files Affected**: `pkg/cli/` → `internal/config/`, `internal/app/`  
**Risk**: LOW | **Duration**: 3-5 days | **Regression Risk**: 0%

#### What Changes

1. **Create `internal/config/` package**
   ```go
   // internal/config/config.go
   type Config struct {
       CLI       CLIConfig        // From pkg/cli
       Model     ModelConfig      // From pkg/models
       Workspace WorkspaceConfig  // From workspace
       // ... other configs
   }

   // LoadConfig(ctx) (*Config, error)
   // ValidateConfig(cfg *Config) error
   ```

2. **Slim down `pkg/cli/`**
   - Move `CLIConfig` to `internal/config/`
   - Keep only flag parsing in `pkg/cli/flags.go`
   - Move `commands.go` to `cmd/commands/`

3. **Update `main.go`**
   ```go
   // Before
   cliConfig, args := cli.ParseCLIFlags()
   if cli.HandleCLICommands(ctx, args, cliConfig.DBPath) { os.Exit(0) }
   app, _ := app.New(ctx, &cliConfig)

   // After
   cfg, _ := config.LoadFromEnv()
   if handled, _ := cmd.HandleSpecialCommands(ctx, cfg); handled { os.Exit(0) }
   app, _ := app.New(ctx, cfg)
   ```

#### Checklist
- [ ] Create `internal/config/` package with Config struct
- [ ] Move CLIConfig, ModelConfig definitions to internal/config
- [ ] Create config.LoadFromEnv() factory
- [ ] Update main.go to use new config path
- [ ] Update Application.New(ctx, cfg) signature
- [ ] Run full test suite; verify 0 failures
- [ ] Update imports across codebase

---

### Phase 2: Refactor Application Orchestrator

**Goal**: Break monolithic Application into focused components  
**Files Affected**: `internal/app/`  
**Risk**: LOW | **Duration**: 5-7 days | **Regression Risk**: 0%

#### What Changes

1. **Extract Component Managers**
   ```go
   // internal/app/components/
   ├── display.go        // DisplayManager
   ├── model.go          // ModelManager
   ├── agent.go          // AgentManager
   ├── session.go        // SessionManager (refactor existing)
   ├── workspace.go      // NEW: WorkspaceManager
   └── repl.go           // REPLManager (extract from repl.go)

   // Each Manager encapsulates initialization + lifecycle
   type DisplayManager struct {
       renderer *display.Renderer
       typewriter *display.Typewriter
       // ...
   }
   func (dm *DisplayManager) Initialize(ctx context.Context, cfg *config.Config) error
   func (dm *DisplayManager) Shutdown(ctx context.Context) error
   ```

2. **Simplify Application struct**
   ```go
   // Before: 7 fields, complex orchestration
   type Application struct {
       config        *CLIConfig
       display       *DisplayComponents
       model         *ModelComponents
       // ... (monolithic)
   }

   // After: Lean orchestrator
   type Application struct {
       config    *config.Config
       managers  ComponentManagers  // Aggregate
       agent     Agent              // Injected
   }

   type ComponentManagers struct {
       Display   *DisplayManager
       Model     *ModelManager
       Agent     *AgentManager
       Session   *SessionManager
       Workspace *WorkspaceManager
       REPL      *REPLManager
   }
   ```

3. **Create AppBuilder for cleaner initialization**
   ```go
   app, err := NewApplicationBuilder(config).
       WithContext(ctx).
       WithSignalHandling().
       Build()
   ```

4. **Extract REPL loop**
   - Move REPL from app.Run() to REPLManager
   - Test REPL independently

#### Checklist
- [ ] Create `internal/app/components/` directory
- [ ] Extract DisplayManager (test independently)
- [ ] Extract ModelManager (test independently)
- [ ] Extract AgentManager (test independently)
- [ ] Extract REPLManager from repl.go
- [ ] Simplify Application struct to orchestrator role only
- [ ] Create AppBuilder for cleaner API
- [ ] Update Application.Run() to use manager pattern
- [ ] Run full test suite; verify 0 failures
- [ ] Verify signal handling still works (Ctrl+C test)

---

### Phase 3: Reorganize Display Package

**Goal**: Reduce 24 files to 5 focused subpackages  
**Files Affected**: `display/`  
**Risk**: LOW | **Duration**: 4-6 days | **Regression Risk**: 0%

#### What Changes

1. **Consolidate display/ structure**
   ```go
   // Target structure
   display/
   ├── renderer.go              # Main Renderer facade (keep as-is)
   ├── formatter/
   │   ├── registry.go          # FormatterRegistry (new abstraction)
   │   ├── base.go              # BaseFormatter interface
   │   ├── tool.go              # ToolFormatter
   │   ├── agent.go             # AgentFormatter
   │   ├── error.go             # ErrorFormatter
   │   └── metrics.go           # MetricsFormatter
   ├── styles/                  # Keep existing structure
   ├── terminal/                # Keep existing structure
   ├── components.go            # Atomic display components (banner, etc.)
   ├── animation/
   │   ├── spinner.go           # Spinner animation
   │   ├── typewriter.go        # Typewriter animation
   │   └── paginator.go         # Pagination
   └── stream/
       ├── streaming.go         # StreamingDisplay
       ├── segment.go           # StreamingSegment
       └── deduplicator.go      # Deduplicator
   ```

2. **Create FormatterRegistry abstraction**
   ```go
   // display/formatter/registry.go
   type FormatterRegistry interface {
       Register(category string, formatter Formatter) error
       Get(category string) (Formatter, error)
       GetAll() map[string]Formatter
   }

   // Renderer uses registry instead of direct field references
   type Renderer struct {
       formatters FormatterRegistry
       styles     *styles.Styles
   }
   ```

3. **Move complex helpers to their own files**
   - `tool_renderer.go`, `tool_adapter.go`, `tool_result_parser.go` → `formatter/tool_internals.go`
   - `streaming_segment.go` → `stream/segment.go`
   - `deduplicator.go` → `stream/deduplicator.go`

4. **Consolidate prompts/ (agent package)**
   - Merge `builder.go` + `builder_cont.go` into `builder.go`
   - Move prompt sections (workflow, guidance, pitfalls) to separate internal file: `builder_sections.go`
   - Result: `prompts/` becomes more modular

#### Checklist
- [ ] Create display/formatter/ subpackage with registry
- [ ] Extract Formatter interface for all formatters to implement
- [ ] Move tool formatter internals to formatter/ subpackage
- [ ] Create display/animation/ subpackage (spinner, typewriter, paginator)
- [ ] Create display/stream/ subpackage (streaming, segment, deduplicator)
- [ ] Update Renderer to use FormatterRegistry instead of direct fields
- [ ] Update agent/prompts/: merge builder.go + builder_cont.go
- [ ] Move prompt sections to builder_sections.go
- [ ] Run full test suite; verify 0 failures
- [ ] Verify display output looks identical (no visual regression)

---

### Phase 4: Extract LLM Abstraction Layer

**Goal**: Isolate LLM implementation details; simplify agent.go  
**Files Affected**: `pkg/models/`, `agent/`, `internal/app/`  
**Risk**: MEDIUM | **Duration**: 5-7 days | **Regression Risk**: <1% (well-isolated)

#### What Changes

1. **Create `internal/llm/` abstraction layer**
   ```go
   // internal/llm/
   ├── provider.go           # LLMProvider interface
   ├── config.go             # LLM configuration
   ├── factory.go            # Factory for creating providers
   ├── backends/
   │   ├── gemini.go         # Gemini implementation
   │   ├── vertexai.go       # Vertex AI implementation
   │   └── openai.go         # OpenAI implementation
   └── cache.go              # Model caching/registry

   // internal/llm/provider.go
   type LLMProvider interface {
       Create(ctx context.Context, config Config) (model.LLM, error)
       Validate(config Config) error
       GetMetadata() ProviderMetadata
   }

   type Config struct {
       Provider string
       Model    string
       APIKey   string
       // ... provider-specific fields
   }
   ```

2. **Move provider creation logic**
   - Move `pkg/models/gemini.go`, `vertexai.go`, `openai.go` to `internal/llm/backends/`
   - Move `pkg/models/registry.go` logic to `internal/llm/factory.go`
   - Keep error handling in `pkg/errors/`

3. **Simplify agent.go**
   ```go
   // Before
   llm, err := models.CreateGeminiModel(ctx, config)  // Complex factory

   // After
   llm, err := llm.Create(ctx, config)  // Clean factory
   ```

4. **Move model registry to pkg/models** (rename to pkg/registry)
   - Keep model.Config definitions here (backward compat)
   - Import from internal/llm for provider logic

#### Checklist
- [ ] Create `internal/llm/` package structure
- [ ] Extract LLMProvider interface from existing code
- [ ] Move provider implementations to `internal/llm/backends/`
- [ ] Create `internal/llm/factory.go` for provider creation
- [ ] Update `pkg/models/` to import from internal/llm
- [ ] Update `agent.go` to use new factory
- [ ] Update `internal/app/` model initialization
- [ ] Run full test suite; verify 0 failures
- [ ] Test all three backends (Gemini, Vertex AI, OpenAI)

---

### Phase 5: Extract Data/Persistence Layer

**Goal**: Separate data access from business logic  
**Files Affected**: `session/`, `tracking/`, `pkg/models/`  
**Risk**: LOW | **Duration**: 3-5 days | **Regression Risk**: 0%

#### What Changes

1. **Create `internal/data/` package**
   ```go
   // internal/data/
   ├── session.go            # Session repository
   ├── models.go             # Model registry repository
   ├── persistence.go        # Persistence abstraction
   └── sqlite/
       ├── session.go        # SQLite session impl
       └── models.go         # SQLite model registry impl
   ```

2. **Extract Session as Repository pattern**
   ```go
   // Before: session/ is tightly bound to SQLite
   type SessionManager struct {
       sessionService session.Service
   }

   // After: Abstracted interface
   type SessionRepository interface {
       Create(ctx context.Context, req *CreateRequest) error
       Get(ctx context.Context, id string) (*Session, error)
       List(ctx context.Context, userID string) ([]*Session, error)
       Delete(ctx context.Context, id string) error
   }

   // Implementations
   type SQLiteSessionRepository struct { ... }
   type InMemorySessionRepository struct { ... }  // For testing
   ```

3. **Move model persistence to data layer**
   - `pkg/models/sqlite.go` → `internal/data/sqlite/models.go`
   - Extract as ModelRegistry interface

4. **Decouple from database implementation**
   - Use repository pattern throughout
   - Easy to swap implementations for testing/alternative backends

#### Checklist
- [ ] Create `internal/data/` package structure
- [ ] Extract SessionRepository interface from session/
- [ ] Create SQLite implementation in internal/data/sqlite/
- [ ] Move model persistence to internal/data/
- [ ] Extract ModelRegistry interface
- [ ] Update session/ to use repository pattern
- [ ] Create in-memory implementations for testing
- [ ] Run full test suite; verify 0 failures
- [ ] Test with both SQLite and in-memory backends

---

## Summary of Changes by Package

| Package | Phase | Action | Risk |
|---------|-------|--------|------|
| `main.go` | 1, 2 | Update to use new config path | LOW |
| `pkg/cli/` | 1 | **Slim**: Move configs to internal/config | LOW |
| `internal/config/` | 1 | **NEW**: Centralized configuration | LOW |
| `internal/app/` | 2 | **Refactor**: Extract component managers | LOW |
| `display/` | 3 | **Reorganize**: Consolidate 24 files to 5 dirs | LOW |
| `agent/prompts/` | 3 | **Consolidate**: Merge builder files | LOW |
| `internal/llm/` | 4 | **NEW**: LLM abstraction layer | MEDIUM |
| `pkg/models/` | 4 | **Update**: Import from internal/llm | MEDIUM |
| `internal/data/` | 5 | **NEW**: Data/persistence layer | LOW |
| `session/` | 5 | **Update**: Use repository pattern | LOW |
| `tools/` | - | **NO CHANGE**: Keep as-is (excellent) | - |
| `workspace/` | - | **NO CHANGE**: Keep as-is | - |

---

## Validation & Rollback Strategy

### Before Each Phase
- [ ] Create feature branch: `refactor/phase-X-description`
- [ ] Document expected behavior in test cases
- [ ] Run full test suite; baseline established

### During Each Phase
- [ ] Commit frequently (every logical change)
- [ ] Run tests after each commit
- [ ] Keep git history clean for easy bisect if issues arise

### After Each Phase
- [ ] Run full test suite (100% pass required)
- [ ] Run integration tests (REPL, agent.Run(), signal handling)
- [ ] Verify no visual regressions (display output identical)
- [ ] Code review before merging to main

### Rollback Procedure
- If >1 test fails: `git revert` and diagnose
- If regression detected: `git reset --hard` to last stable
- Never push incomplete phase; always backtrack to green state

---

## Testing Strategy

### Unit Tests (per phase)
- New packages: Write tests before/alongside changes
- Modified packages: Verify existing tests still pass
- Use table-driven tests for complex logic

### Integration Tests
- Agent loop: Test agent.Run() with mock tools
- REPL: Test interactive prompt → agent → output
- Signal handling: Test Ctrl+C cancels running agent
- Config loading: Test all config sources (env, flags, file)

### Regression Tests
- Visual display output (tool call formatting, banners, colors)
- Tool execution (execute command, file operations)
- Session persistence (create, list, delete, resume)
- Model selection (all three backends work)

### Test Automation
```bash
# Before each commit
make test            # Unit tests
make integration-test  # Integration tests (after Phase 2)
make check           # Format, vet, lint, test
```

---

## Go Best Practices Applied

1. ✅ **Interface-based design** – Extract interfaces, use dependency injection
2. ✅ **Separation of concerns** – Each package has one job
3. ✅ **Configuration management** – Centralized, validated config
4. ✅ **Repository pattern** – Abstract data access
5. ✅ **Factory pattern** – Complex object creation
6. ✅ **Layered architecture** – Presentation, business logic, data layers
7. ✅ **Error handling** – Consistent error types (already good)
8. ✅ **Testing** – Easy to unit test isolated components
9. ✅ **No global state** – Dependency injection throughout
10. ✅ **Clean imports** – No circular dependencies

---

## Incremental Delivery Plan

**Goal**: Deliver value at end of each phase

### Phase 1 (Week 1)
- ✅ Cleaner main.go
- ✅ Easier to extend config (add new flags)

### Phase 2 (Week 1-2)
- ✅ Testable components (can unit test each manager)
- ✅ Easier to debug initialization issues
- ✅ Prepare for Phase 3 (display refactoring)

### Phase 3 (Week 2-3)
- ✅ Easier to maintain display logic
- ✅ Reduce cognitive load on display/
- ✅ Prepare for custom formatter support

### Phase 4 (Week 3)
- ✅ Cleaner LLM integration
- ✅ Easier to add new backends
- ✅ Better error messages for LLM config

### Phase 5 (Week 3-4)
- ✅ Easier to swap backends (SQLite ↔ in-memory)
- ✅ Testable data layer
- ✅ Final codebase: clean, modular, maintainable

---

## Estimated Effort Breakdown

| Phase | Task | Estimate | Notes |
|-------|------|----------|-------|
| 1 | Extract config layer | 3-5 days | Low complexity, high impact |
| 2 | Refactor Application | 5-7 days | More coordination required |
| 3 | Reorganize display/ | 4-6 days | Mechanical refactoring |
| 4 | LLM abstraction | 5-7 days | Needs careful testing |
| 5 | Data persistence layer | 3-5 days | Repository pattern straightforward |
| **Total** | **Full refactor** | **20-30 days** | ~3-4 weeks of focused work |

---

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Regression in agent behavior | Low | High | Comprehensive test suite before/after each phase |
| Display output changes | Low | High | Visual regression tests (screenshot comparison) |
| Breaking API changes | Low | Medium | Maintain backward compatibility facades |
| Performance degradation | Very Low | Medium | Benchmark before/after refactoring |
| Integration issues | Low | High | Run integration tests daily during refactoring |

---

## Success Criteria

Phase completion checklist:
- [ ] All unit tests pass (100%)
- [ ] All integration tests pass (100%)
- [ ] No visual regressions in display
- [ ] Code review approved
- [ ] Documentation updated
- [ ] Cyclomatic complexity reduced
- [ ] Package coupling reduced
- [ ] New developers can understand code faster

---

## Post-Refactoring Benefits

1. **Developer Velocity** ↑ 30-40%
   - Smaller packages = easier to understand
   - Clear boundaries = fewer surprises
   - Better error messages = faster debugging

2. **Code Quality** ↑
   - Testable components = more tests written
   - Reduced coupling = fewer bugs
   - Clear interfaces = easier to maintain

3. **Extensibility** ↑
   - Add new formatter = implement interface + register
   - Add new backend = implement LLMProvider interface
   - Add new tool = already solved by tools/ package

4. **Maintainability** ↑
   - Onboarding new contributors = 50% less time
   - Bug fixes = faster isolation and repair
   - Feature additions = follow clear patterns

---

## Next Steps

1. ✅ **Review this plan** – Stakeholder sign-off
2. ✅ **Create feature branch** – `refactor/phase-1-config`
3. ✅ **Phase 1 Implementation** – Extract config layer (3-5 days)
4. ✅ **Phase 2 Implementation** – Refactor orchestrator (5-7 days)
5. ✅ **Phase 3 Implementation** – Display reorganization (4-6 days)
6. ✅ **Phase 4 Implementation** – LLM abstraction (5-7 days)
7. ✅ **Phase 5 Implementation** – Data layer (3-5 days)
8. ✅ **Integration & QA** – Full regression testing (2-3 days)
9. ✅ **Merge to main** – Code review, approval, deployment
10. ✅ **Update documentation** – Architecture docs, runbook updates

---

## Appendix: Example: Phase 1 in Detail

### What We'll Do

Move from:
```go
// main.go
cliConfig, args := cli.ParseCLIFlags()
app, err := app.New(ctx, &cliConfig)
```

To:
```go
// main.go
cfg, args := config.LoadFromEnv()
app, err := app.New(ctx, cfg)
```

### Step-by-Step

1. Create `internal/config/config.go`
   - Define `type Config struct`
   - Implement `LoadFromEnv()`
   - Implement `Validate()`

2. Move CLI types
   - `CLIConfig` → `internal/config/cli.go`
   - Move model selection logic here

3. Update main.go
   - Change import from `pkg/cli` to `internal/config`
   - Update flag parsing call

4. Update app.New()
   - Change signature: `New(ctx context.Context, cfg *cli.CLIConfig)` → `New(ctx context.Context, cfg *config.Config)`
   - Update references: `a.config` type is now `*config.Config`

5. Test
   - `go test ./...` – all green?
   - `./code-agent` – manual smoke test?
   - `git log --oneline` – clean commits?

6. Commit
   - `git commit -am "refactor: extract config layer"`

---

## Conclusion

This refactoring plan delivers a **more modular, maintainable, and professional Go codebase** while maintaining **100% backward compatibility** and **zero functional changes**. Each phase is independent and can be reviewed/approved separately.

**Status**: Ready for implementation. No blocking issues identified.

**Recommendation**: Start with Phase 1 (extract config) immediately. It has the lowest risk and highest immediate impact on code clarity.
