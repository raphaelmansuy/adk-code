# Code Agent Refactoring Plan

**Date**: November 12, 2025  
**Objective**: Improve code organization, modularity, and maintainability while ensuring 0% regression  
**Current Scale**: ~23,464 LOC across ~100 Go files, 31 test files

## Executive Summary

This plan addresses architectural concerns while maintaining backward compatibility and zero regression. The focus is on pragmatic improvements that enhance code quality without disrupting existing functionality.

## Guiding Principles

1. **Zero Regression**: All existing tests must pass after each phase
2. **Incremental Changes**: Small, testable changes with clear rollback points
3. **Backward Compatibility**: Maintain existing APIs through deprecation/facade patterns
4. **Pragmatic Over Perfect**: Focus on high-impact improvements
5. **Test-Driven**: Add tests before refactoring, ensure coverage remains stable

## Current Architecture Assessment

### Strengths to Preserve
- ✅ Clean tool registration system with category-based organization
- ✅ No circular dependencies
- ✅ Well-defined error handling patterns
- ✅ Repository pattern for data access
- ✅ Interface-driven design in key areas

### Issues to Address
- ❌ Display package too large (~4000+ LOC, 25+ files)
- ❌ Internal/app package has "God Object" characteristics
- ❌ Mixed abstraction levels in several packages
- ❌ Session management split across 3 locations
- ❌ Inconsistent package organization patterns
- ❌ Heavy reliance on init() for registration (fragile, hard to test)

---

## Phase 1: Foundation & Documentation (Low Risk)

**Goal**: Establish testing baseline and document current architecture without changing code

### 1.1 Test Coverage Analysis
- [ ] Run `make test` with coverage reporting
- [ ] Document current test coverage percentage
- [ ] Identify packages with <50% coverage
- [ ] Create coverage baseline report: `docs/test_coverage_baseline.md`

### 1.2 Dependency Graph
- [ ] Generate package dependency graph using `go mod graph`
- [ ] Document key dependency flows
- [ ] Identify tight coupling points
- [ ] Output: `docs/architecture/dependency_graph.md`

### 1.3 API Surface Documentation
- [ ] Document all exported functions/types in each package
- [ ] Identify public APIs that must remain stable
- [ ] Mark internal-only exports that can change freely
- [ ] Output: `docs/architecture/api_surface.md`

**Validation**: No code changes, documentation only  
**Duration**: 1-2 days  
**Risk**: None

---

## Phase 2: Display Package Restructuring (Medium Risk)

**Goal**: Break down the monolithic display package into focused sub-packages

### Current Structure (display/)
```
display/
  ├── ansi.go, event.go, facade.go, factory.go
  ├── renderer.go, streaming_display.go, typewriter.go, spinner.go
  ├── tool_*.go (adapter, renderer, result_parser)
  ├── paginator.go, deduplicator.go
  ├── banner/ (banner.go, banner_test.go)
  ├── components/ (banner.go)
  ├── formatters/ (...)
  ├── renderer/ (renderer.go, markdown_renderer.go)
  ├── styles/ (colors.go, formatting.go)
  ├── terminal/ (...)
  └── tooling/ (...)
```

### Proposed Structure
```
display/
  ├── facade.go          # Public API facade (maintain backward compatibility)
  ├── factory.go         # Component factory
  │
  ├── core/              # Core rendering primitives
  │   ├── ansi.go        # ANSI codes
  │   ├── terminal.go    # Terminal utilities
  │   └── styles.go      # Style definitions
  │
  ├── components/        # UI components
  │   ├── spinner.go
  │   ├── typewriter.go
  │   ├── paginator.go
  │   └── banner.go
  │
  ├── renderers/         # Content renderers
  │   ├── renderer.go
  │   ├── markdown.go
  │   └── tool_renderer.go
  │
  ├── streaming/         # Streaming display
  │   ├── display.go
  │   ├── segment.go
  │   └── deduplicator.go
  │
  ├── events/            # Event handling
  │   ├── event.go
  │   ├── timeline.go
  │   └── adapter.go
  │
  └── formatters/        # Content formatters
      └── (existing formatters)
```

### Implementation Steps

#### 2.1 Create New Package Structure
- [ ] Create subdirectories under display/
- [ ] Move files to appropriate packages (copy first, don't delete)
- [ ] Update package declarations
- [ ] Fix import paths within display package

#### 2.2 Update facade.go for Backward Compatibility
```go
// display/facade.go
package display

import (
    "code_agent/display/components"
    "code_agent/display/renderers"
    "code_agent/display/streaming"
)

// Re-export types for backward compatibility
type Spinner = components.Spinner
type Renderer = renderers.Renderer
type StreamingDisplay = streaming.Display

// Re-export constructors
func NewSpinner(r *Renderer, msg string) *Spinner {
    return components.NewSpinner(r, msg)
}
// ... etc
```

#### 2.3 Update External Imports
- [ ] Update imports in internal/app/
- [ ] Update imports in pkg/cli/
- [ ] Update imports in agent/
- [ ] Update imports in tools/

#### 2.4 Testing & Validation
- [ ] Run all display package tests
- [ ] Run full test suite: `make test`
- [ ] Verify no regression in functionality
- [ ] Remove old files after validation

**Validation Criteria**:
- All tests pass
- No change in external API behavior
- Import paths updated consistently
- Coverage remains ≥ baseline

**Duration**: 3-4 days  
**Risk**: Medium (many files to move, many imports to update)  
**Rollback**: Revert directory changes, restore imports

---

## Phase 3: App Package Decomposition (High Risk)

**Goal**: Break down internal/app/ into focused components with clear responsibilities

### Current Structure (internal/app/)
```
app/
  ├── app.go                    # Main application orchestrator
  ├── components.go             # Component structs
  ├── factories.go              # Generic factories
  ├── factories_test.go
  ├── init_*.go                 # Component initializers (6 files)
  ├── repl.go, repl_test.go     # REPL implementation
  ├── session.go                # Session components
  ├── signals.go                # Signal handling
  └── utils.go                  # Utilities
```

### Proposed Structure
```
internal/
  ├── app/
  │   ├── app.go              # Simplified application entry point
  │   ├── config.go           # App-level configuration
  │   └── lifecycle.go        # Lifecycle management
  │
  ├── repl/                   # REPL implementation
  │   ├── repl.go
  │   ├── commands.go         # Built-in command handlers
  │   └── history.go          # History management
  │
  ├── runtime/                # Runtime components
  │   ├── signal_handler.go
  │   ├── context.go
  │   └── shutdown.go
  │
  └── orchestration/          # Component orchestration
      ├── builder.go          # Application builder pattern
      ├── components.go       # Component definitions
      ├── display.go          # Display component initialization
      ├── model.go            # Model component initialization
      ├── session.go          # Session component initialization
      └── agent.go            # Agent component initialization
```

### Implementation Steps

#### 3.1 Create Runtime Package
- [ ] Create internal/runtime/
- [ ] Move signals.go → signal_handler.go
- [ ] Extract context management logic
- [ ] Add tests for signal handling
- [ ] Update app.go to use internal/runtime

#### 3.2 Create REPL Package
- [ ] Create internal/repl/
- [ ] Move repl.go and repl_test.go
- [ ] Extract command handling logic from pkg/cli to repl/commands.go
- [ ] Update imports in app.go
- [ ] Ensure REPL tests still pass

#### 3.3 Create Orchestration Package
- [ ] Create internal/orchestration/
- [ ] Move init_*.go files to orchestration/
- [ ] Rename to descriptive names (display.go, model.go, etc.)
- [ ] Create builder.go with fluent API for app construction
- [ ] Move components.go to orchestration/

#### 3.4 Simplify App Package
- [ ] Refactor app.go to use builder pattern
- [ ] Remove "God Object" characteristics
- [ ] Keep only high-level lifecycle management
- [ ] Document clear responsibilities

**Example Builder Pattern**:
```go
// internal/orchestration/builder.go
type ApplicationBuilder struct {
    config  *config.Config
    ctx     context.Context
}

func NewBuilder(ctx context.Context, cfg *config.Config) *ApplicationBuilder {
    return &ApplicationBuilder{config: cfg, ctx: ctx}
}

func (b *ApplicationBuilder) BuildRuntime() (*RuntimeComponents, error) { ... }
func (b *ApplicationBuilder) BuildDisplay() (*DisplayComponents, error) { ... }
func (b *ApplicationBuilder) BuildModel() (*ModelComponents, error) { ... }
func (b *ApplicationBuilder) BuildAgent() (agent.Agent, error) { ... }
func (b *ApplicationBuilder) Build() (*app.Application, error) { ... }
```

#### 3.5 Update Tests
- [ ] Move and update test files to match new structure
- [ ] Add integration tests for builder pattern
- [ ] Ensure all app tests pass

**Validation Criteria**:
- All tests pass
- Application starts and runs correctly
- REPL functions identically
- Signal handling works (Ctrl-C tests)
- Coverage remains ≥ baseline

**Duration**: 4-5 days  
**Risk**: High (core application flow changes)  
**Rollback**: Revert all changes in internal/app, internal/repl, internal/runtime, internal/orchestration

---

## Phase 4: Session Management Consolidation (Medium Risk)

**Goal**: Consolidate session-related code into a single, coherent package

### Current Split
1. `session/` - Models and SQLite implementation (manager.go, models.go, sqlite.go)
2. `internal/data/` - Repository interfaces (repository.go)
3. `internal/data/sqlite/` - SQLite session implementation
4. `internal/data/memory/` - In-memory session implementation

### Proposed Structure
```
internal/
  └── session/
      ├── session.go          # Session domain models and interfaces
      ├── manager.go          # Session manager (high-level API)
      ├── repository.go       # Repository interface
      │
      ├── storage/            # Storage implementations
      │   ├── sqlite/
      │   │   ├── adapter.go
      │   │   ├── models.go
      │   │   └── session.go
      │   └── memory/
      │       └── session.go
      │
      └── service.go          # Service layer (bridge to ADK)
```

### Implementation Steps

#### 4.1 Create Unified Session Package
- [ ] Create internal/session/ (new location)
- [ ] Move session/models.go → internal/session/session.go
- [ ] Move internal/data/repository.go → internal/session/repository.go
- [ ] Update package declarations and internal imports

#### 4.2 Reorganize Storage Implementations
- [ ] Create internal/session/storage/sqlite/
- [ ] Move internal/data/sqlite/session.go → storage/sqlite/
- [ ] Move internal/data/sqlite/models.go → storage/sqlite/
- [ ] Move internal/data/memory/session.go → storage/memory/
- [ ] Update imports and package declarations

#### 4.3 Migrate Session Manager
- [ ] Move session/manager.go → internal/session/manager.go
- [ ] Update manager to use new storage package paths
- [ ] Maintain backward compatibility facade if needed

#### 4.4 Update All References
- [ ] Update internal/app/ to use internal/session
- [ ] Update cmd/commands/ to use internal/session
- [ ] Update any other references to old session packages

#### 4.5 Deprecation Path
- [ ] Mark old session/ package as deprecated
- [ ] Add deprecation comments with migration guidance
- [ ] Can remove old package in future release

**Validation Criteria**:
- All session-related tests pass
- Session persistence works (SQLite and memory)
- List/create/delete session commands work
- Coverage remains ≥ baseline

**Duration**: 2-3 days  
**Risk**: Medium (data persistence changes)  
**Rollback**: Revert internal/session changes, restore old imports

---

## Phase 5: Tool Registration Explicit Pattern (Low-Medium Risk)

**Goal**: Replace fragile init() pattern with explicit registration for better testability

### Current Pattern
```go
// tools/file/read_tool.go
func init() {
    _, _ = NewReadFileTool()
}

func NewReadFileTool() (tool.Tool, error) {
    // ... create tool
    common.Register(metadata)
    return t, err
}
```

**Issues**:
- Init order dependencies
- Hard to test in isolation
- Side effects on package import
- Cannot control registration in tests

### Proposed Pattern
```go
// tools/registry/loader.go
package registry

import (
    "code_agent/tools/common"
    "code_agent/tools/file"
    "code_agent/tools/edit"
    // ... other tools
)

// LoadAllTools explicitly registers all tools
func LoadAllTools() (*common.ToolRegistry, error) {
    reg := common.NewToolRegistry()
    
    // File tools
    if err := registerFileTool(reg); err != nil {
        return nil, err
    }
    if err := registerEditTools(reg); err != nil {
        return nil, err
    }
    // ... more tools
    
    return reg, nil
}

func registerFileTools(reg *common.ToolRegistry) error {
    tools := []struct {
        name string
        factory func() (tool.Tool, error)
    }{
        {"read_file", file.NewReadFileTool},
        {"write_file", file.NewWriteFileTool},
        // ...
    }
    
    for _, t := range tools {
        tool, err := t.factory()
        if err != nil {
            return fmt.Errorf("failed to create %s: %w", t.name, err)
        }
        if err := reg.Register(tool); err != nil {
            return err
        }
    }
    return nil
}
```

### Implementation Steps

#### 5.1 Create Loader Package
- [ ] Create tools/registry/loader.go
- [ ] Implement LoadAllTools() function
- [ ] Add helper functions for each tool category
- [ ] Add tests for loader

#### 5.2 Update Tool Constructors
- [ ] Remove init() functions from all tool files
- [ ] Ensure NewXxxTool() functions don't auto-register
- [ ] Keep registration logic but make it explicit
- [ ] Update tool package imports

#### 5.3 Update Agent Initialization
- [ ] Modify agent/coding_agent.go to use explicit loader
- [ ] Replace `tools.GetRegistry()` with `registry.LoadAllTools()`
- [ ] Handle errors properly during registration

#### 5.4 Update Tests
- [ ] Add unit tests for each tool in isolation
- [ ] Add integration tests for full tool registration
- [ ] Ensure no init() side effects in tests

**Example Usage**:
```go
// agent/coding_agent.go
func NewCodingAgent(ctx context.Context, cfg Config) (agentiface.Agent, error) {
    // Explicitly load all tools
    toolRegistry, err := registry.LoadAllTools()
    if err != nil {
        return nil, fmt.Errorf("failed to load tools: %w", err)
    }
    
    allTools := toolRegistry.GetAllTools()
    // ... rest of agent creation
}
```

**Validation Criteria**:
- All tools register successfully
- No init() side effects
- Tools can be tested in isolation
- Full test suite passes
- Coverage improves due to better testability

**Duration**: 3-4 days  
**Risk**: Medium (changes initialization flow)  
**Rollback**: Restore init() functions, revert loader package

---

## Phase 6: Package Organization Standardization (Low Risk)

**Goal**: Standardize package organization patterns across the codebase

### Current Inconsistencies
- Some use subpackages (display/banner, display/renderer)
- Others are flat (session/, tracking/)
- pkg/ vs internal/ distinction unclear

### Standardization Rules

#### 6.1 Internal vs Pkg Guidelines
```
internal/       - Application-specific, not reusable
  ├── app/       - Application orchestration
  ├── config/    - Configuration management
  ├── data/      - Data persistence
  ├── llm/       - LLM provider integration
  └── ...

pkg/            - Reusable, potentially extractable libraries
  ├── errors/    - Error types and utilities
  ├── models/    - LLM model configurations
  ├── cli/       - CLI utilities
  └── ...
```

#### 6.2 When to Use Subpackages
- Package has >10 files → consider splitting
- Clear sub-responsibilities exist
- Different abstraction levels
- Can be tested independently

### Implementation Steps

#### 6.1 Audit Current Packages
- [ ] List all packages and file counts
- [ ] Identify packages that should split or merge
- [ ] Document decisions in `docs/architecture/package_organization.md`

#### 6.2 Move CLI Command Code
- [ ] Evaluate if pkg/cli/commands should move to internal/repl/commands
- [ ] Commands are app-specific, not reusable → move to internal
- [ ] Update imports

#### 6.3 Consolidate Tracking Package
- [ ] Review tracking/ - should it be internal/tracking?
- [ ] If app-specific → move to internal/
- [ ] If reusable → keep as pkg/tracking

#### 6.4 Standardize Naming
- [ ] Ensure consistent naming conventions
- [ ] Package names singular (not plural) unless collection
- [ ] Clear, descriptive names

**Validation Criteria**:
- All tests pass after moves
- Import paths updated consistently
- No functional changes
- Documentation updated

**Duration**: 2-3 days  
**Risk**: Low (mostly file moves)  
**Rollback**: Revert package moves, restore imports

---

## Phase 7: Testing Infrastructure (Low Risk)

**Goal**: Improve test organization and shared testing utilities

### Current State
- Tests co-located with code (good)
- Some shared utilities in testutils/ but underutilized
- No clear pattern for integration tests
- Coverage reporting exists but not standardized

### Proposed Improvements

#### 7.1 Enhance testutils Package
```
internal/testutils/
  ├── fixtures/        # Test data and fixtures
  │   ├── models.go    # Common model configs for tests
  │   ├── sessions.go  # Common session data
  │   └── tools.go     # Tool test helpers
  │
  ├── mocks/           # Mock implementations
  │   ├── display.go   # Mock display components
  │   ├── llm.go       # Mock LLM for testing
  │   └── storage.go   # Mock storage
  │
  ├── assert/          # Custom assertions
  │   └── errors.go    # Error assertion helpers
  │
  └── helpers.go       # General test helpers
```

#### 7.2 Integration Test Suite
```
tests/
  ├── integration/
  │   ├── agent_test.go       # End-to-end agent tests
  │   ├── tools_test.go       # Tool integration tests
  │   ├── session_test.go     # Session persistence tests
  │   └── repl_test.go        # REPL interaction tests
  │
  └── e2e/
      ├── scenarios/          # Full user scenarios
      └── smoke_test.go       # Basic smoke tests
```

### Implementation Steps

#### 7.1 Build testutils Package
- [ ] Create organized testutils structure
- [ ] Extract common test setup code
- [ ] Create reusable fixtures
- [ ] Add mock implementations
- [ ] Document usage patterns

#### 7.2 Add Integration Tests
- [ ] Create tests/integration/ directory
- [ ] Write key integration tests
- [ ] Ensure tests can run in CI
- [ ] Add make target: `make test-integration`

#### 7.3 Coverage Reporting
- [ ] Standardize coverage reporting
- [ ] Set coverage targets per package
- [ ] Add coverage badge to README
- [ ] Create make target: `make coverage`

#### 7.4 Test Documentation
- [ ] Document testing patterns in TESTING.md
- [ ] Provide examples of unit vs integration tests
- [ ] Document how to use testutils

**Validation Criteria**:
- Existing tests still pass
- New integration tests pass
- Coverage reporting works
- Documentation clear

**Duration**: 3-4 days  
**Risk**: Low (additive, doesn't change existing code)  
**Rollback**: Remove new tests, keep existing tests

---

## Phase 8: Documentation & Developer Experience (Low Risk)

**Goal**: Comprehensive documentation for the refactored architecture

### Documentation Deliverables

#### 8.1 Architecture Documentation
- [ ] `docs/architecture/overview.md` - High-level architecture
- [ ] `docs/architecture/package_guide.md` - Package responsibilities
- [ ] `docs/architecture/data_flow.md` - How data flows through system
- [ ] `docs/architecture/extension_guide.md` - How to add new features

#### 8.2 Developer Guides
- [ ] `docs/guides/adding_tools.md` - How to add new tools
- [ ] `docs/guides/testing.md` - Testing best practices
- [ ] `docs/guides/display_components.md` - Using display system
- [ ] `docs/guides/llm_providers.md` - Adding new LLM providers

#### 8.3 API Documentation
- [ ] Generate godoc for all packages
- [ ] Add package-level documentation comments
- [ ] Document key interfaces and types
- [ ] Add usage examples in docs

#### 8.4 Migration Guides
- [ ] Document changes from original structure
- [ ] Provide import path migration guide
- [ ] List deprecated APIs and replacements

### Implementation Steps

#### 8.1 Write Core Documentation
- [ ] Create architecture overview
- [ ] Document each major package
- [ ] Create diagrams (mermaid or similar)

#### 8.2 Add Code Examples
- [ ] Add examples/ directory with runnable code
- [ ] Document common use cases
- [ ] Add troubleshooting guide

#### 8.3 Update README
- [ ] Add architecture section
- [ ] Link to detailed documentation
- [ ] Update build and test instructions

#### 8.4 Generate API Docs
- [ ] Run `go doc` for all packages
- [ ] Host docs (optional: use pkgsite)
- [ ] Add doc links to README

**Validation Criteria**:
- All documentation builds correctly
- Examples run successfully
- Links work correctly
- Clear and comprehensive

**Duration**: 2-3 days  
**Risk**: None (documentation only)  
**Rollback**: Not applicable

---

## Phase 9: Performance & Code Quality (Optional)

**Goal**: Optimize performance and enforce code quality standards

### Areas for Optimization

#### 9.1 Code Quality Tools
- [ ] Add golangci-lint with comprehensive rules
- [ ] Add pre-commit hooks
- [ ] Configure staticcheck
- [ ] Add gofmt/goimports checks

#### 9.2 Performance Profiling
- [ ] Add benchmarks for critical paths
- [ ] Profile tool execution
- [ ] Optimize display rendering
- [ ] Measure and optimize startup time

#### 9.3 Error Handling Audit
- [ ] Review all error returns
- [ ] Ensure proper error wrapping
- [ ] Add context to all errors
- [ ] Consistent error messages

#### 9.4 Concurrency Review
- [ ] Review goroutine usage
- [ ] Check for race conditions
- [ ] Add race detector to CI
- [ ] Document concurrency patterns

### Implementation Steps

#### 9.1 Setup Quality Tools
- [ ] Add .golangci.yml configuration
- [ ] Add pre-commit hook script
- [ ] Update Makefile with quality checks
- [ ] Run and fix all linting issues

#### 9.2 Add Benchmarks
- [ ] Identify critical code paths
- [ ] Write benchmark tests
- [ ] Establish baseline metrics
- [ ] Add benchmark CI job

#### 9.3 Race Detection
- [ ] Run tests with `-race` flag
- [ ] Fix any race conditions found
- [ ] Add race detection to CI

**Validation Criteria**:
- All linters pass
- Benchmarks establish baselines
- No race conditions detected
- Performance not degraded

**Duration**: 3-4 days  
**Risk**: Low (mostly additive)  
**Rollback**: Remove quality tools if they cause issues

---

## Implementation Timeline

### Conservative Estimate (Sequential)
```
Phase 1: Foundation           2 days
Phase 2: Display              4 days
Phase 3: App Decomposition    5 days
Phase 4: Session              3 days
Phase 5: Tool Registration    4 days
Phase 6: Package Org          3 days
Phase 7: Testing              4 days
Phase 8: Documentation        3 days
Phase 9: Quality (Optional)   4 days
----------------------------------------
Total:                        32 days (6-7 weeks)
```

### Aggressive Estimate (Some Parallelization)
```
Phase 1: Foundation           2 days
Phase 2 + 6: Display + Org    5 days  (can parallelize)
Phase 3: App Decomposition    5 days
Phase 4: Session              3 days
Phase 5: Tool Registration    4 days
Phase 7: Testing              4 days  (concurrent with docs)
Phase 8: Documentation        3 days
Phase 9: Quality              4 days  (optional)
----------------------------------------
Total:                        26-30 days (5-6 weeks)
```

---

## Success Criteria

### Must Have (Phase Completion)
- ✅ All existing tests pass (0% regression)
- ✅ Test coverage remains ≥ baseline (ideally improves)
- ✅ No new bugs introduced
- ✅ Application behavior unchanged from user perspective
- ✅ Code compiles without warnings
- ✅ All import paths updated and working

### Nice to Have (Quality Improvements)
- ⭐ Test coverage improves by >10%
- ⭐ Reduced package coupling (fewer interdependencies)
- ⭐ Faster test execution
- ⭐ Better error messages
- ⭐ Comprehensive documentation

### Acceptance Testing Checklist
After each phase:
```bash
# 1. Run full test suite
make test

# 2. Check test coverage
make coverage

# 3. Run linters
make lint

# 4. Build the application
make build

# 5. Manual smoke test
./bin/code-agent --help
./bin/code-agent new-session test-refactor
# Test basic interactions

# 6. Check for regressions
./bin/code-agent list-sessions  # Should show test-refactor
```

---

## Risk Mitigation

### Before Each Phase
1. Create feature branch: `refactor/phase-N-description`
2. Document current behavior/tests
3. Take coverage snapshot
4. Review implementation plan with team

### During Each Phase
1. Commit frequently with clear messages
2. Run tests after each logical change
3. Document any issues or deviations
4. Keep rollback plan ready

### After Each Phase
1. Run full test suite including manual tests
2. Review code coverage changes
3. Update documentation
4. Merge to main only after validation
5. Tag release: `v1.x.x-phase-N`

### Emergency Rollback Plan
```bash
# Rollback to previous phase
git checkout main
git revert <phase-commit-range>

# Or rollback to specific tag
git checkout v1.x.x-phase-N-1

# Verify tests pass
make test
```

---

## Open Questions / Decisions Needed

1. **Display Package**: Should we keep a facade in the root display/ for backward compatibility indefinitely, or deprecate after some time?
   - **Recommendation**: Keep facade for at least 2 releases with deprecation warnings

2. **Tool Registration**: Should we support both init() and explicit patterns during transition?
   - **Recommendation**: Hard cutover (remove init()) to avoid confusion

3. **Session Package Location**: Keep session/ as top-level or move to internal/session?
   - **Recommendation**: Move to internal/session - it's app-specific, not reusable

4. **Testing Strategy**: Should we require minimum coverage % for new code?
   - **Recommendation**: Yes, require 70% coverage for new packages

5. **Performance**: Should Phase 9 (performance) be required or optional?
   - **Recommendation**: Optional unless performance issues identified

---

## Appendix A: File Move Checklist Template

Use this checklist for each file move:

```markdown
### Moving: `<source_path>` → `<dest_path>`

- [ ] Create destination directory
- [ ] Copy file (don't delete original yet)
- [ ] Update package declaration
- [ ] Fix internal imports in the file
- [ ] Update all external imports to new path
- [ ] Run tests for affected packages
- [ ] Verify functionality unchanged
- [ ] Delete original file
- [ ] Commit with descriptive message
```

---

## Appendix B: Backward Compatibility Pattern

When changing public APIs, use this pattern:

```go
// Old location: session/manager.go
package session

import "code_agent/internal/session"

// Deprecated: Use internal/session.Manager instead
type Manager = session.Manager

// Deprecated: Use internal/session.NewManager instead
func NewManager(appName, dbPath string) (*Manager, error) {
    return session.NewManager(appName, dbPath)
}
```

---

## Appendix C: Testing Commands

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run specific package tests
go test ./internal/app/...

# Run with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Check for test cache issues
go clean -testcache
go test ./...
```

---

## Conclusion

This refactoring plan addresses the key architectural issues while maintaining pragmatism and zero regression. Each phase is designed to be:

1. **Independent**: Can be executed separately
2. **Testable**: Clear validation criteria
3. **Reversible**: Defined rollback procedures
4. **Incremental**: Small, manageable changes

The plan prioritizes high-impact changes (Display and App restructuring) while deferring optional improvements (Phase 9) that can be done later. The estimated timeline of 5-7 weeks is realistic for a careful, test-driven refactoring that maintains quality and stability.

**Recommendation**: Start with Phase 1 (Foundation) immediately to establish baseline, then proceed to Phase 2 (Display) as the highest-impact improvement. Evaluate progress after each phase before committing to the next.
