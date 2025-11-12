# Code Agent Refactoring Plan

**Date**: 2025-11-12
**Goal**: Make code_agent/ more organized, modular, and maintainable
**Constraint**: 0% regression - our reputation is at stake

## Executive Summary

The codebase is **fundamentally well-structured** with good separation of concerns. The refactoring focuses on:

1. **Reducing complexity** in the Application struct (15 fields → grouped components)
2. **Adding missing tests** for critical infrastructure (internal/app)
3. **Improving modularity** through better composition patterns
4. **Moving misplaced code** to appropriate packages
5. **Standardizing patterns** across the codebase

**Risk Level**: LOW to MEDIUM (mostly structural improvements, no algorithm changes)

## Current State Assessment

### ✅ What's Working Well

- **Package structure**: Clear separation by feature (tools/, display/, agent/, etc.)
- **Tool registry pattern**: Excellent auto-registration via init()
- **Provider abstraction**: Clean model registry supporting multiple backends
- **Workspace management**: Smart multi-workspace support with VCS detection
- **Error handling**: Generally good with proper error wrapping
- **No globals**: Clean state management (except tool registry, which is fine)

### ⚠️ Areas for Improvement

- **Application complexity**: 15-field struct violates Single Responsibility Principle
- **Missing tests**: internal/app has 0 tests despite being critical infrastructure
- **Parameter explosion**: REPLConfig (10 fields), CLIConfig (11 fields)
- **Misplaced code**: GetProjectRoot() in agent/ should be in workspace/
- **Limited display tests**: Only 1 test file for 3,714 LOC

### 📊 Code Metrics

```
Total: ~14,495 LOC (excluding tests)

By Package:
- display:     3,714 LOC (25.6%)  ⚠️ Only 1 test file
- tools:       3,576 LOC (24.7%)  ✅ Good test coverage
- pkg:         2,449 LOC (16.9%)  ✅ Partially tested
- workspace:   1,332 LOC (9.2%)   ✅ Has tests
- persistence: 1,314 LOC (9.1%)   ✅ Excellent test coverage
- agent:       1,038 LOC (7.2%)   ⚠️ Only builder test
- internal:      706 LOC (4.9%)   ❌ NO TESTS!
- tracking:      335 LOC (2.3%)   ✅ Has tests
```

---

## Refactoring Phases

### Phase 0: Pre-Refactoring Setup (CRITICAL)

**Objective**: Establish safety net before any changes

#### 0.1 Create Comprehensive Integration Tests

**Priority**: P0 (BLOCKER)
**Risk**: None (only adds tests)
**Effort**: 2-3 hours

Create `internal/app/app_integration_test.go`:

```go
// Test application initialization
func TestApplication_New_Success(t *testing.T)
func TestApplication_New_InvalidConfig(t *testing.T)
func TestApplication_InitializeDisplay(t *testing.T)
func TestApplication_InitializeModel(t *testing.T)
func TestApplication_InitializeAgent(t *testing.T)
func TestApplication_InitializeSession(t *testing.T)
```

**Acceptance Criteria**:
- All initialization paths tested
- Error cases covered
- 80%+ coverage of app.go

#### 0.2 Capture Current Behavior

Create `internal/app/testdata/` with:
- Sample configs
- Expected outputs
- Golden files for banner rendering

**Deliverable**: Test suite that passes 100% before any refactoring

---

### Phase 1: Structural Improvements (Low Risk)

**Objective**: Reduce complexity without changing behavior

#### 1.1 Group Application Components

**Priority**: P1
**Risk**: LOW (internal refactoring only)
**Effort**: 1 hour
**Files**: `internal/app/app.go`, `internal/app/components.go` (new)

**Current**:
```go
type Application struct {
    renderer       *display.Renderer
    bannerRenderer *display.BannerRenderer
    typewriter     *display.TypewriterPrinter
    streamDisplay  *display.StreamingDisplay
    // ... 11 more fields
}
```

**Proposed**:
```go
// New file: internal/app/components.go
type DisplayComponents struct {
    Renderer       *display.Renderer
    BannerRenderer *display.BannerRenderer
    Typewriter     *display.TypewriterPrinter
    StreamDisplay  *display.StreamingDisplay
}

type ModelComponents struct {
    Registry *models.Registry
    Selected models.Config
    LLM      model.LLM
}

type SessionComponents struct {
    Manager *persistence.SessionManager
    Runner  *runner.Runner
    Tokens  *tracking.SessionTokens
}

// Updated Application
type Application struct {
    config        *cli.CLIConfig
    ctx           context.Context
    signalHandler *SignalHandler
    
    Display DisplayComponents
    Model   ModelComponents
    Session SessionComponents
    Agent   agent.Agent
    REPL    *REPL
}
```

**Migration Strategy**:
1. Create new structs in `components.go`
2. Update Application struct
3. Update all references (IDE refactor)
4. Run tests to verify behavior unchanged

**Verification**:
```bash
make test
make check
```

#### 1.2 Simplify REPL Configuration

**Priority**: P1
**Risk**: LOW
**Effort**: 45 minutes
**Files**: `internal/app/repl.go`

**Current**: 10 separate fields in REPLConfig

**Proposed**:
```go
type REPLConfig struct {
    UserID      string
    SessionName string
    Display     *DisplayComponents  // Group display-related
    Session     *SessionComponents  // Group session-related
    Model       ModelComponents     // Model info for /model command
}
```

**Benefits**:
- 5 fields instead of 10
- Clearer responsibility grouping
- Easier to pass related components

#### 1.3 Group CLI Configuration

**Priority**: P1
**Risk**: LOW
**Effort**: 30 minutes
**Files**: `pkg/cli/config.go`

**Current**: 11 flat fields

**Proposed**:
```go
type CLIConfig struct {
    // Display settings
    Display struct {
        OutputFormat      string
        TypewriterEnabled bool
    }
    
    // Session settings
    Session struct {
        Name             string
        DBPath           string
        WorkingDirectory string
    }
    
    // Model settings
    Model struct {
        Backend          string  // "gemini", "vertexai", "openai"
        Name             string  // Model ID
        APIKey           string
        VertexAIProject  string  // For Vertex AI
        VertexAILocation string  // For Vertex AI
    }
    
    // AI settings
    AI struct {
        EnableThinking bool
        ThinkingBudget int32
    }
}
```

**Benefits**:
- Logical grouping by concern
- Easier to extend
- Clearer documentation

#### 1.4 Extract Builder Pattern for Application

**Priority**: P2 (optional but valuable)
**Risk**: MEDIUM
**Effort**: 2 hours
**Files**: `internal/app/builder.go` (new)

**Current**: 5 separate init methods called in sequence

**Proposed**:
```go
// New file: internal/app/builder.go
type ApplicationBuilder struct {
    config *cli.CLIConfig
    ctx    context.Context
    // ... internal state
}

func NewApplicationBuilder(ctx context.Context, config *cli.CLIConfig) *ApplicationBuilder
func (b *ApplicationBuilder) WithDisplay() *ApplicationBuilder
func (b *ApplicationBuilder) WithModel() *ApplicationBuilder
func (b *ApplicationBuilder) WithAgent() *ApplicationBuilder
func (b *ApplicationBuilder) WithSession() *ApplicationBuilder
func (b *ApplicationBuilder) WithREPL() *ApplicationBuilder
func (b *ApplicationBuilder) Build() (*Application, error)

// Usage in main.go
app, err := app.NewApplicationBuilder(ctx, &cliConfig).
    WithDisplay().
    WithModel().
    WithAgent().
    WithSession().
    WithREPL().
    Build()
```

**Benefits**:
- More testable (can mock each step)
- Clearer initialization flow
- Easier to add optional components

**Note**: This is optional - current approach works fine

---

### Phase 2: Code Organization (Low Risk)

**Objective**: Move code to appropriate packages

#### 2.1 Move GetProjectRoot to workspace/

**Priority**: P1
**Risk**: VERY LOW (simple move)
**Effort**: 15 minutes
**Files**: 
- `agent/coding_agent.go` → `workspace/project_root.go`

**Steps**:
1. Create `workspace/project_root.go`
2. Move `GetProjectRoot()` function
3. Update import in `agent/coding_agent.go`
4. Run tests

**Verification**:
```bash
go test ./workspace/...
go test ./agent/...
```

#### 2.2 Extract Display Component Initialization

**Priority**: P2
**Risk**: LOW
**Effort**: 30 minutes
**Files**: `display/factory.go` (new)

**Current**: Display components created directly in app.go

**Proposed**:
```go
// New file: display/factory.go
type Config struct {
    OutputFormat      string
    TypewriterEnabled bool
}

func NewComponents(cfg Config) (*Components, error) {
    renderer, err := NewRenderer(cfg.OutputFormat)
    if err != nil {
        return nil, err
    }
    
    return &Components{
        Renderer:       renderer,
        BannerRenderer: NewBannerRenderer(renderer),
        Typewriter:     NewTypewriterPrinter(DefaultTypewriterConfig()),
        StreamDisplay:  NewStreamingDisplay(renderer, typewriter),
    }, nil
}
```

**Benefits**:
- Display package owns its initialization
- Easier to test display components
- Reduces app.go complexity

---

### Phase 3: Testing Improvements (HIGH PRIORITY)

**Objective**: Add comprehensive tests for critical paths

#### 3.1 Test Coverage Goals

| Package         | Current | Target | Priority |
|----------------|---------|--------|----------|
| internal/app   | 0%      | 80%+   | P0       |
| display        | ~5%     | 60%+   | P1       |
| agent          | ~20%    | 70%+   | P2       |
| tools          | ~70%    | 80%+   | P3       |

#### 3.2 Critical Test Files to Create

**P0 (Must Have)**:
- `internal/app/app_test.go` - Application initialization
- `internal/app/repl_test.go` - REPL behavior
- `internal/app/session_test.go` - Session initialization

**P1 (Should Have)**:
- `display/renderer_test.go` - Core rendering logic
- `display/formatters/tool_test.go` - Tool formatting
- `agent/coding_agent_test.go` - Agent creation

**P2 (Nice to Have)**:
- `internal/app/signals_test.go` - Signal handling
- `display/components/timeline_test.go` - Timeline rendering

#### 3.3 Test Utilities

Create `internal/app/testing/` with:
- `fixtures.go` - Sample configs, test data
- `mocks.go` - Mock implementations of interfaces
- `helpers.go` - Common test utilities

---

### Phase 4: Code Quality Improvements (Low Risk)

**Objective**: Standardize patterns and improve maintainability

#### 4.1 Standardize Error Handling

**Priority**: P2
**Risk**: LOW
**Effort**: 1 hour

**Current**: Inconsistent error wrapping

**Standard Pattern**:
```go
// Always use %w for wrapping
return fmt.Errorf("failed to create agent: %w", err)

// Always add context
return fmt.Errorf("failed to read file %q: %w", path, err)

// Use sentinel errors for known cases
var (
    ErrSessionNotFound = errors.New("session not found")
    ErrInvalidConfig   = errors.New("invalid configuration")
)
```

**Apply to**: All packages systematically

#### 4.2 Add Missing Documentation

**Priority**: P3
**Risk**: None
**Effort**: 2 hours

**Focus Areas**:
- Package-level documentation for all packages
- Public function documentation
- Complex private functions
- Type documentation with usage examples

**Example**:
```go
// Package app provides the core application lifecycle management
// for the code agent. It orchestrates initialization of display,
// model, agent, session, and REPL components.
//
// Example usage:
//
//     config := &cli.CLIConfig{...}
//     app, err := app.New(ctx, config)
//     if err != nil {
//         log.Fatal(err)
//     }
//     app.Run()
package app
```

#### 4.3 Extract Long Functions

**Priority**: P3
**Risk**: LOW
**Effort**: Variable

**Candidates**:
- `app.initializeModel()` (114 lines) → extract provider creation
- `repl.Run()` (229 lines) → extract input handling
- Long switch statements in command handlers

**Pattern**:
```go
// Before
func (a *Application) initializeModel() error {
    // 114 lines of logic
}

// After
func (a *Application) initializeModel() error {
    if err := a.resolveSelectedModel(); err != nil {
        return err
    }
    if err := a.createLLMModel(); err != nil {
        return err
    }
    return nil
}

func (a *Application) resolveSelectedModel() error { ... }
func (a *Application) createLLMModel() error { ... }
```

---

## Implementation Strategy

### Week 1: Safety & Foundations

**Day 1-2**: Phase 0 - Create comprehensive tests
- Focus on internal/app integration tests
- Create test fixtures and utilities
- Ensure 100% current behavior captured

**Day 3**: Phase 2.1 - Move GetProjectRoot
- Simple, low-risk change
- Verify with tests

**Day 4-5**: Phase 1.1 - Group Application components
- Create components.go
- Update Application struct
- Update all references
- Extensive testing

### Week 2: Improvements

**Day 1**: Phase 1.2 & 1.3 - Simplify configurations
- Group REPL config
- Group CLI config
- Update usage sites

**Day 2**: Phase 2.2 - Display factory
- Create display/factory.go
- Refactor initialization
- Test thoroughly

**Day 3-5**: Phase 3 - Add missing tests
- Display package tests
- Agent package tests
- Integration tests

### Week 3: Polish

**Day 1-2**: Phase 4.1 - Standardize error handling
- Review all error returns
- Apply consistent patterns
- Update error messages

**Day 3-4**: Phase 4.2 - Documentation
- Package docs
- Function docs
- Examples

**Day 5**: Final review & validation
- Run all tests
- Performance check
- Code review

---

## Risk Mitigation

### Before Each Change

1. ✅ Ensure tests exist for affected code
2. ✅ Run `make test` - all tests pass
3. ✅ Run `make check` - no lint errors
4. ✅ Create git branch for change

### During Each Change

1. 🔍 Make small, incremental changes
2. 🔍 Test after each logical step
3. 🔍 Keep original behavior unchanged
4. 🔍 Document breaking changes (should be none!)

### After Each Change

1. ✅ Run full test suite
2. ✅ Run `make check`
3. ✅ Manual smoke test
4. ✅ Git commit with clear message
5. ✅ Code review

### Rollback Plan

- Each phase is independent
- Each change is in separate commit
- Can rollback any single change without affecting others
- Tests ensure behavior preservation

---

## Success Metrics

### Code Quality
- [ ] Test coverage: internal/app > 80%
- [ ] Test coverage: display > 60%
- [ ] Zero new lint warnings
- [ ] All existing tests pass
- [ ] `make check` succeeds

### Complexity
- [ ] Application struct: 15 fields → 7 fields
- [ ] REPLConfig: 10 fields → 5 fields
- [ ] CLIConfig: 11 fields → 4 groups
- [ ] Average function length reduced by 20%

### Documentation
- [ ] All public packages documented
- [ ] All public functions documented
- [ ] Package examples added
- [ ] README updated

### Maintainability
- [ ] Clear separation of concerns
- [ ] Easy to add new features
- [ ] Easy to test components
- [ ] Consistent patterns throughout

---

## Non-Goals

### What We're NOT Doing

❌ **Not changing algorithms** - No logic changes
❌ **Not adding features** - Pure refactoring only
❌ **Not changing APIs** - Backward compatible
❌ **Not replacing dependencies** - Keep current stack
❌ **Not rewriting from scratch** - Incremental improvements
❌ **Not optimizing performance** - Structure focus only

---

## Verification Checklist

Before considering refactoring complete:

### Functionality
- [ ] All original features work identically
- [ ] No regression in any user-facing behavior
- [ ] All CLI flags work as before
- [ ] All commands produce same output
- [ ] Session persistence works
- [ ] Tool execution works
- [ ] Model selection works
- [ ] Multi-workspace works

### Code Quality
- [ ] `make test` passes 100%
- [ ] `make check` passes (fmt, vet, lint)
- [ ] Test coverage improved
- [ ] No new warnings
- [ ] Code is more readable
- [ ] Simpler to maintain

### Documentation
- [ ] Code changes documented
- [ ] Tests document behavior
- [ ] Package docs updated
- [ ] README updated if needed

---

## Rollout Plan

### Phase 1: Internal Testing (Week 1-2)
- Develop in feature branch
- Team code review
- Internal testing
- Performance benchmarks

### Phase 2: Alpha Testing (Week 3)
- Merge to main
- Deploy to test environment
- Run full test suite
- Gather feedback

### Phase 3: Production (Week 4)
- Final code review
- Update documentation
- Tag release
- Monitor for issues

---

## Conclusion

This refactoring plan is **pragmatic and incremental**:

### Why This Plan is Safe

1. **Test-First Approach**: Add tests before changing code
2. **Small Changes**: Each phase is independent and reversible
3. **No Behavior Changes**: Pure structural improvements
4. **Extensive Validation**: Multiple checkpoints throughout
5. **Clear Rollback**: Each change is isolated and reversible

### Expected Benefits

- **25% reduction** in Application complexity
- **80%+ test coverage** for critical infrastructure
- **Better organization** through component grouping
- **Improved maintainability** via clearer patterns
- **Easier testing** through better separation

### Time Investment

- **Minimum**: 3 days (P0 + P1 items only)
- **Recommended**: 2-3 weeks (all phases)
- **Return**: Easier maintenance for years to come

**The codebase is already good. These changes make it excellent.**

---

## Appendix: Go Best Practices Applied

### Struct Design
✅ Keep structs focused and small
✅ Group related fields
✅ Use embedding sparingly
✅ Prefer composition over inheritance

### Error Handling
✅ Always wrap errors with context
✅ Use %w for error chains
✅ Define sentinel errors for known cases
✅ Return errors, don't panic

### Testing
✅ Table-driven tests
✅ Test public APIs, not internals
✅ Use testdata/ for fixtures
✅ Mock external dependencies

### Package Design
✅ Package by feature, not layer
✅ Keep packages focused
✅ Minimize dependencies
✅ Clear import hierarchies

### Code Style
✅ Follow go fmt
✅ Pass go vet
✅ Pass golangci-lint
✅ Clear naming conventions
