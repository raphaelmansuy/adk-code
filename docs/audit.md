# Code Agent Architecture Audit & Refactoring Plan
**Date:** November 12, 2025  
**Auditor:** AI Code Analysis  
**Risk Level:** Zero-Regression Required  
**Status:** Phase 3 In Progress - Display Package Relocation Complete

---

## Executive Summary

This audit analyzes the `code_agent/` codebase (~23K LOC, 167 Go files) to identify opportunities for improved modularity, organization, and maintainability while adhering to Go best practices. The analysis reveals a fundamentally sound architecture with specific areas requiring attention.

### Key Findings

**Strengths:**
- Well-organized tool system with category-based organization (4610 LOC)
- Clean orchestrator pattern for component initialization
- Strong error handling system
- No circular dependencies
- Good test coverage foundation (31 test files)

**Critical Issues:**
1. **Display Package Monolith** (5440 LOC, 23% of codebase) - needs decomposition
2. **Deprecated Code Cleanup** - old facades still present
3. **Command Handler Duplication** - 3 different command locations
4. **Session Management Split** - logic across 3 packages

**Risk Assessment:** Medium overall. Recommended changes are incremental and reversible.

---

## Current Architecture Overview

### Package Structure & LOC Breakdown

```
code_agent/                         Total: ~23,167 LOC
├── main.go                         32 LOC - CLEAN ✓
├── agent_prompts/                  2220 LOC - Good structure ✓
│   ├── coding_agent.go
│   ├── dynamic_prompt.go
│   ├── xml_prompt_builder.go
│   └── prompts/
├── cmd/                            ~100 LOC
│   └── commands/
│       └── handlers.go             Special CLI commands
├── display/                        5440 LOC - TOO LARGE ⚠️
│   ├── banner*.go
│   ├── event_handler.go
│   ├── paginator.go
│   ├── renderer.go
│   ├── spinner.go
│   ├── streaming_display.go
│   ├── tool_*.go
│   ├── typewriter.go
│   └── formatters/
├── internal/                       5352 LOC
│   ├── app/                        Application orchestration
│   │   ├── app.go                  Main application struct
│   │   ├── components.go           Type aliases
│   │   ├── factories.go            Component factories
│   │   ├── orchestration.go        DEPRECATED - just facades ⚠️
│   │   ├── repl.go                 DEPRECATED - just facade ⚠️
│   │   ├── session.go
│   │   ├── signals.go
│   │   └── utils.go
│   ├── cli/                        CLI utilities
│   │   └── commands/               CLI command implementations
│   ├── commands/                   DUPLICATE? ⚠️
│   │   └── handlers.go             Just delegates to cli/commands
│   ├── config/                     Configuration management ✓
│   ├── llm/                        LLM client factories ✓
│   ├── orchestration/              NEW orchestrator pattern ✓
│   │   ├── builder.go              Fluent builder API
│   │   ├── components.go
│   │   ├── agent.go
│   │   ├── display.go
│   │   ├── model.go
│   │   └── session.go
│   ├── repl/                       REPL implementation ✓
│   ├── runtime/                    Signal handling ✓
│   └── session/                    Session persistence ✓
├── pkg/                            2686 LOC
│   ├── errors/                     Error handling ✓
│   ├── models/                     Model registry & adapters ✓
│   └── testutil/                   Test utilities ✓
├── tools/                          4610 LOC - WELL ORGANIZED ✓
│   ├── base/                       Common types & registry
│   ├── display/                    Display tools
│   ├── edit/                       Edit tools
│   ├── exec/                       Execution tools
│   ├── file/                       File tools
│   ├── search/                     Search tools
│   ├── v4a/                        V4A patch tools
│   ├── workspace/                  Workspace tools
│   └── tools.go                    Public facade
├── tracking/                       561 LOC - Simple & focused ✓
└── workspace/                      2093 LOC - Good design ✓
```

### Architecture Patterns Used

1. **Orchestrator Pattern** (`internal/orchestration/`) - ✓ Good
   - Fluent builder API for component initialization
   - Clear dependency management
   - Replaces old initialization patterns

2. **Registry Pattern** (`tools/base/registry.go`) - ✓ Good
   - Auto-registration via init() functions
   - Category-based tool organization
   - Centralized tool management

3. **Repository Pattern** (`internal/session/`) - ✓ Good
   - Abstract data access
   - SQLite and in-memory implementations
   - Clear separation of concerns

4. **Facade Pattern** (`tools/tools.go`, deprecated facades) - Mixed
   - Good: tools.go provides clean public API
   - Bad: Deprecated facades still present (confusion)

5. **Factory Pattern** (`internal/app/factories.go`, `pkg/models/factories/`) - ✓ Good
   - Consistent object creation
   - Dependency injection ready

---

## Detailed Issue Analysis

### Issue #1: Display Package Monolith (CRITICAL)

**Current State:**
- **5440 LOC** in a single package (23% of entire codebase)
- Multiple responsibilities mixed together
- Hard to navigate, test, and maintain

**Files & Responsibilities:**

| File | LOC | Responsibility | Should Be In |
|------|-----|----------------|--------------|
| banner.go, banner_renderer.go | ~500 | Banner rendering | display/banners/ |
| event_handler.go | ~200 | Event handling | display/events/ |
| paginator.go | ~300 | Pagination UI | display/components/ |
| renderer.go | ~600 | Core rendering | display/terminal/ |
| spinner.go | ~400 | Spinner UI | display/components/ |
| streaming_display.go | ~800 | Streaming output | display/streaming/ |
| tool_adapter.go | ~400 | Tool event adaptation | display/tools/ |
| tool_renderer.go | ~300 | Tool formatting | display/tools/ |
| typewriter.go | ~300 | Typewriter effect | display/components/ |
| formatters/ | ~800 | Text formatting | display/formatting/ |
| Other files | ~840 | Various | Appropriate subpackages |

**Impact:**
- High cognitive load for developers
- Testing complexity
- Tight coupling between unrelated components
- Difficult to reuse individual pieces

**Recommended Structure:**

```
internal/display/
├── terminal/          # Low-level terminal operations (ANSI, colors, cursor)
│   ├── renderer.go
│   └── ansi.go
├── components/        # Reusable UI components
│   ├── spinner.go
│   ├── typewriter.go
│   └── paginator.go
├── streaming/         # Streaming display logic
│   ├── display.go
│   └── handler.go
├── formatting/        # Text formatting (move from formatters/)
│   ├── registry.go
│   └── formatters.go
├── banners/           # Banner rendering
│   └── banner.go
├── events/            # Event handling
│   └── handler.go
└── tools/             # Tool-specific rendering
    ├── adapter.go
    └── renderer.go
```

**Migration Strategy:**
1. Create new `internal/display/` structure with subpackages
2. Move implementations to new locations
3. Keep old `display/` with deprecated facades that delegate
4. Update internal imports gradually
5. Remove old facades once all imports updated
6. Update tests to use new structure

**Risk:** Medium. Many imports to update, but changes are mechanical.

**Validation:**
- All tests must pass after each file move
- Build must succeed
- No functional changes, only reorganization

---

### Issue #2: Deprecated Facades (HIGH PRIORITY)

**Current State:**

Two deprecated facade files still exist:

#### 2a. `internal/app/orchestration.go`
```go
// All functions just delegate to internal/orchestration/
func initializeDisplayComponents(cfg *config.Config) (*DisplayComponents, error) {
    return orchestration.InitializeDisplayComponents(cfg)
}
// ... 3 more similar facades
```

**Status:** Already fully replaced by `internal/orchestration/` package.  
**Usage:** Only used in old tests (`internal/app/app_init_test.go`).  
**Risk:** LOW - Safe to remove after updating tests.

#### 2b. `internal/app/repl.go`
```go
// Type aliases with deprecation comments
type REPL = intrepl.REPL
type REPLConfig = intrepl.Config
func NewREPL(config REPLConfig) (*REPL, error) {
    return intrepl.New(config)
}
```

**Status:** Already fully replaced by `internal/repl/` package.  
**Usage:** Used by `internal/app/app.go` (can be updated easily).  
**Risk:** LOW - One file to update.

**Recommended Action:**

**Phase 2.1: Update Tests & Remove orchestration.go**
```bash
# Update internal/app/app_init_test.go to import orchestration directly
# Remove internal/app/orchestration.go
```

**Phase 2.2: Update REPL Import & Remove repl.go**
```bash
# Update internal/app/app.go to import internal/repl directly
# Remove internal/app/repl.go
```

**Benefits:**
- Reduces confusion for new developers
- Cleaner codebase
- Enforces use of new patterns

---

### Issue #3: Command Handler Duplication (MEDIUM PRIORITY)

**Current State:**

Three different locations for command-related code:

1. **`cmd/commands/handlers.go`**
   - Special CLI commands (new-session, list-sessions, delete-session)
   - Called from `main.go` before app starts

2. **`internal/commands/handlers.go`**
   - Exact same file as #1!
   - Just delegates to `internal/cli/commands/`
   - Unnecessary middle layer

3. **`internal/cli/commands/`**
   - Actual implementations (session.go, repl.go, model.go)
   - Has the real logic

**Problem:**
- Confusing to navigate
- `internal/commands/` is redundant
- Not clear which to import
- Duplicate code paths

**Recommended Structure:**

```
internal/cli/
├── commands/              # All command implementations
│   ├── session.go        # Session commands (new, list, delete)
│   ├── repl.go           # REPL commands
│   ├── model.go          # Model commands
│   └── interface.go      # Command interfaces
└── parser.go             # Command line parsing
```

**Migration Steps:**

1. **Remove** `cmd/commands/` directory entirely
2. **Remove** `internal/commands/` directory entirely
3. **Keep** `internal/cli/commands/` as single source of truth
4. **Update** `main.go` to import `internal/cli/commands` directly
5. **Update** any other imports

**Risk:** LOW - Mechanical change, no logic modification.

**Before:**
```go
// main.go
import "code_agent/internal/commands"
commands.HandleSpecialCommands(...)
```

**After:**
```go
// main.go
import clicommands "code_agent/internal/cli/commands"
clicommands.HandleSpecialCommands(...)
```

---

### Issue #4: Session Management Split (LOW-MEDIUM PRIORITY)

**Current State:**

Session-related code exists in multiple locations:

1. **`internal/session/`**
   - Session models (session.go)
   - SQLite persistence (sqlite.go)
   - Manager implementation

2. **`internal/app/session.go`**
   - Session component initialization
   - SessionComponents struct

3. **`internal/orchestration/session.go`**
   - InitializeSessionComponents function
   - Part of orchestrator pattern

**Analysis:**

This split is actually **acceptable** as each serves a different layer:
- `internal/session/` = Data layer (persistence)
- `internal/orchestration/session.go` = Initialization layer
- `internal/app/session.go` = Application layer (wiring)

**Recommendation:** KEEP AS-IS, but improve documentation.

**Action Items:**
1. Add package-level documentation to `internal/session/`
2. Add comments explaining the layer separation
3. Consider renaming `internal/session/` to `internal/sessionstore/` for clarity

**Risk:** NONE if kept as-is.

---

## Go Best Practices Assessment

### ✅ Strengths

1. **Clear Package Boundaries** (mostly)
   - `pkg/` for public reusable code
   - `internal/` for private application code
   - Good separation of concerns

2. **No Circular Dependencies**
   - Verified via import analysis
   - Clean dependency graph

3. **Interface Usage**
   - Repository interfaces
   - Workspace interfaces
   - LLM provider interfaces

4. **Error Handling**
   - Centralized error codes
   - Consistent AgentError type
   - Proper error wrapping

5. **Testing**
   - Tests co-located with code
   - Table-driven tests used
   - ~31% coverage by file count

6. **Context Usage**
   - Proper context propagation
   - Cancellation support
   - Timeout handling

### ⚠️ Areas for Improvement

1. **Package Size**
   - Display package too large (5440 LOC)
   - Should be split into focused subpackages

2. **Single Responsibility Principle**
   - Display package violates SRP (multiple responsibilities)
   - Some coupling between UI components

3. **Documentation**
   - Some packages lack package-level docs
   - Public APIs need better documentation
   - Architecture docs need updating

4. **Dependency Injection**
   - Heavy use of init() for tool registration
   - Could be more explicit for better testability

5. **Test Coverage**
   - Could use more integration tests
   - Some edge cases not covered

---

## Recommended Refactoring Plan

### Phase 1: Foundation & Documentation (Week 1) - ZERO RISK

**Objective:** Document current state, no code changes.

**Tasks:**
1. ✅ Complete this audit document
2. Document current architecture in `docs/architecture_overview.md`
3. Create dependency graph diagrams
4. Document migration strategy for each phase
5. Create validation checklist

**Deliverables:**
- Comprehensive architecture documentation
- Refactoring roadmap
- Validation procedures

**Risk:** NONE - Documentation only

---

### Phase 2: Dead Code Removal (Week 1-2) - LOW RISK

**Objective:** Remove deprecated code and reduce confusion.

**Tasks:**

#### 2.1: Remove Deprecated Orchestration Facades
- [ ] Update `internal/app/app_init_test.go` to import `orchestration` directly
- [ ] Remove `internal/app/orchestration.go`
- [ ] Run `make test` - all tests must pass
- [ ] Commit with message: "refactor: remove deprecated orchestration facades"

#### 2.2: Remove Deprecated REPL Facade
- [ ] Update `internal/app/app.go` to import `internal/repl` directly
- [ ] Remove `internal/app/repl.go`
- [ ] Run `make test` - all tests must pass
- [ ] Commit with message: "refactor: remove deprecated REPL facade"

#### 2.3: Consolidate Command Handlers
- [ ] Update `main.go` to import `internal/cli/commands` directly
- [ ] Remove `cmd/commands/` directory
- [ ] Remove `internal/commands/` directory
- [ ] Run `make test` - all tests must pass
- [ ] Commit with message: "refactor: consolidate command handlers"

**Validation:**
```bash
make clean
make check  # Must pass: fmt, vet, test
make build  # Must succeed
./bin/code-agent --help  # Must work
```

**Risk:** LOW - Mechanical changes, no logic modification

**Estimated Time:** 2-4 hours

---

### Phase 3: Display Package Decomposition (Week 2-3) - MEDIUM RISK

**Objective:** Break display package into focused subpackages.

**Part A: Create New Structure**

1. Create new package structure:
```bash
mkdir -p internal/display/{terminal,components,streaming,formatting,banners,events,tools}
```

2. Move files to new locations:

| Source | Destination | LOC |
|--------|-------------|-----|
| display/renderer.go | internal/display/terminal/ | 600 |
| display/spinner.go | internal/display/components/ | 400 |
| display/typewriter.go | internal/display/components/ | 300 |
| display/paginator.go | internal/display/components/ | 300 |
| display/streaming_display.go | internal/display/streaming/ | 800 |
| display/banner*.go | internal/display/banners/ | 500 |
| display/event_handler.go | internal/display/events/ | 200 |
| display/tool_*.go | internal/display/tools/ | 700 |
| display/formatters/ | internal/display/formatting/ | 800 |

3. Update package declarations and imports within moved files

4. Create compatibility facades in old `display/` package:

```go
// display/renderer.go
package display

import "code_agent/internal/display/terminal"

// Deprecated: Use internal/display/terminal.Renderer instead
type Renderer = terminal.Renderer

// Deprecated: Use internal/display/terminal.NewRenderer instead
func NewRenderer() *Renderer {
    return terminal.NewRenderer()
}
```

**Part B: Update Internal Imports**

5. Update all `internal/` packages to use new structure:
   - `internal/app/` imports
   - `internal/orchestration/` imports
   - `internal/repl/` imports

6. Run tests after each package update:
```bash
# After updating each package
go test ./internal/app/...
go test ./internal/orchestration/...
# etc.
```

**Part C: Update External Imports**

7. Update `agent_prompts/` to use new structure
8. Update any remaining imports
9. Run full test suite: `make test`

**Part D: Remove Old Facades**

10. Remove old `display/` package entirely
11. Final validation: `make check && make build`

**Validation Checklist:**
- [ ] All tests pass: `make test`
- [ ] No lint errors: `make lint`
- [ ] Build succeeds: `make build`
- [ ] Manual testing: Run agent with sample commands
- [ ] All imports resolved
- [ ] No deprecated warnings

**Risk:** MEDIUM - Many imports to update, but changes are mechanical

**Estimated Time:** 1-2 weeks

**Rollback Strategy:** Revert to previous commit if validation fails

---

### Phase 4: Documentation Updates (Week 3-4) - ZERO RISK

**Objective:** Update all documentation to reflect new structure.

**Tasks:**
1. Update README.md with new package structure
2. Update architecture diagrams
3. Add package-level documentation to all packages
4. Document public APIs with godoc comments
5. Update CHANGELOG.md
6. Create migration guide for external users (if any)

**Validation:**
```bash
godoc -http=:6060  # Browse documentation
```

**Risk:** NONE - Documentation only

**Estimated Time:** 1 week

---

### Phase 5: Testing Improvements (Week 4-5) - LOW RISK

**Objective:** Improve test coverage and quality.

**Tasks:**
1. Add integration tests for new package structure
2. Increase unit test coverage for critical paths
3. Add edge case tests
4. Document testing strategy
5. Set up code coverage tracking

**Target Coverage:** 60-70% by line coverage

**Risk:** LOW - Only adding tests, not changing code

**Estimated Time:** 1 week

---

### Phase 6: Optional Enhancements (Week 5+) - VARIES

**Objective:** Address lower-priority improvements.

**6.1: Improve Tool Registration (Optional)**

Current: Auto-registration via init()
Proposed: Explicit registration for better testability

**Trade-offs:**
- More verbose but more testable
- Better for dependency injection
- Would require significant changes

**Recommendation:** DEFER - Current pattern works well enough

**6.2: Performance Optimization (Optional)**

Tasks:
- Profile the application
- Optimize hot paths
- Reduce allocations
- Improve startup time

**Recommendation:** DEFER until performance issues identified

**6.3: Enhanced Error Handling (Optional)**

Tasks:
- Add error context
- Improve error messages
- Add error recovery strategies

**Risk:** LOW-MEDIUM

**Estimated Time:** 1-2 weeks

---

## Validation Strategy

### Validation Gates (Required at Each Phase)

#### Gate 1: Code Quality
```bash
make fmt      # Format code
make vet      # Run go vet
make lint     # Run linters (if golangci-lint installed)
```
**Criteria:** All checks must pass, no new warnings

#### Gate 2: Tests
```bash
make test     # Run all tests
```
**Criteria:** All tests must pass, no new failures

#### Gate 3: Build
```bash
make clean
make build
```
**Criteria:** Build must succeed, binary must run

#### Gate 4: Integration Testing
```bash
./bin/code-agent --help
./bin/code-agent new-session test-session
./bin/code-agent list-sessions
./bin/code-agent delete-session test-session
# Test with actual agent prompts
```
**Criteria:** All commands must work as before

#### Gate 5: Code Review
- Review changes against checklist
- Verify no unintended changes
- Check for proper documentation
- Ensure backward compatibility

### Continuous Validation

**After each file change:**
```bash
go test ./path/to/changed/package
```

**After each package change:**
```bash
make test
```

**Before each commit:**
```bash
make check
```

**Before phase completion:**
```bash
make clean && make check && make build
# Manual integration tests
# Code review
```

---

## Risk Assessment & Mitigation

### Risk Matrix

| Phase | Risk Level | Impact if Failed | Mitigation |
|-------|-----------|------------------|------------|
| 1. Foundation | ZERO | None | Documentation only |
| 2. Dead Code | LOW | Build failure | Easy rollback, tests |
| 3. Display Refactor | MEDIUM | Import errors | Incremental, facades, tests |
| 4. Documentation | ZERO | None | Documentation only |
| 5. Testing | LOW | None | Only adding tests |
| 6. Optional | VARIES | Depends | Can skip entirely |

### Mitigation Strategies

1. **Incremental Changes**
   - Small, focused commits
   - One logical change per commit
   - Test after each change

2. **Backward Compatibility**
   - Use facade pattern during migration
   - Deprecate before removing
   - Maintain old APIs temporarily

3. **Comprehensive Testing**
   - Run tests after each change
   - Add regression tests
   - Manual integration testing

4. **Version Control**
   - Work in feature branches
   - Tag each phase completion
   - Easy rollback if issues

5. **Code Review**
   - Review changes before merging
   - Check against requirements
   - Verify validation gates passed

---

## Success Criteria

### Quantitative Metrics

1. **Package Size**
   - ✅ No package > 2000 LOC
   - ✅ Display split into ~6 subpackages of 500-1000 LOC each
   - ✅ Average package size < 1500 LOC

2. **Code Quality**
   - ✅ All tests pass (100%)
   - ✅ No new lint errors
   - ✅ No new go vet warnings
   - ✅ Build succeeds

3. **Test Coverage**
   - ✅ Maintain or increase test coverage
   - ✅ All existing tests still pass
   - ✅ New tests added for refactored code

4. **Performance**
   - ✅ No significant regression in startup time
   - ✅ No increase in memory usage
   - ✅ Build time not significantly increased

### Qualitative Metrics

1. **Code Organization**
   - ✅ Clear package boundaries
   - ✅ Single Responsibility Principle followed
   - ✅ Easy to navigate codebase
   - ✅ Logical package grouping

2. **Developer Experience**
   - ✅ Easier to find relevant code
   - ✅ Clearer import paths
   - ✅ Better discoverability
   - ✅ Reduced cognitive load

3. **Maintainability**
   - ✅ Easier to make changes
   - ✅ Better isolation for testing
   - ✅ Clearer dependencies
   - ✅ Reduced coupling

4. **Documentation**
   - ✅ All packages documented
   - ✅ Public APIs documented
   - ✅ Architecture diagrams updated
   - ✅ Migration guides created

---

## Timeline & Effort Estimate

### Detailed Timeline

| Phase | Duration | Effort (hours) | Dependencies |
|-------|----------|----------------|--------------|
| 1. Foundation | 1 week | 8-12 | None |
| 2. Dead Code | 0.5 week | 4-6 | Phase 1 |
| 3. Display Refactor | 2 weeks | 20-30 | Phase 2 |
| 4. Documentation | 1 week | 8-12 | Phase 3 |
| 5. Testing | 1 week | 8-12 | Phase 4 |
| 6. Optional | 1-2 weeks | 10-20 | Phase 5 |
| **Total** | **5-7 weeks** | **58-92 hours** | |

### Resource Requirements

1. **Developer Time:**
   - 1 developer full-time OR
   - 2 developers part-time

2. **Tools Required:**
   - Go 1.24.4+
   - golangci-lint (optional but recommended)
   - Git
   - Make

3. **Infrastructure:**
   - Development environment
   - CI/CD for testing (optional)

---

## What NOT to Change

### Keep As-Is (Works Well) ✅

1. **Tools Package Structure** (`tools/`)
   - Well-organized by category
   - Auto-registration pattern works
   - Clean public API via tools.go
   - **Rationale:** "If it ain't broke, don't fix it"

2. **Workspace Package** (`workspace/`)
   - Good design
   - Clear responsibility
   - VCS awareness well-implemented
   - **Rationale:** Recently designed, works well

3. **Error Handling** (`pkg/errors/`)
   - Clean pattern
   - Consistent usage
   - Good error codes
   - **Rationale:** Standard and effective

4. **Orchestration Builder** (`internal/orchestration/`)
   - New and well-designed
   - Fluent API
   - Clear dependency management
   - **Rationale:** Recently introduced, follows best practices

5. **Core Agent Logic** (`agent_prompts/`)
   - Well-structured
   - Clear separation
   - Good abstraction
   - **Rationale:** Core functionality, well-designed

6. **LLM Integration** (`internal/llm/`, `pkg/models/`)
   - Multiple providers supported
   - Clean abstraction
   - Extensible design
   - **Rationale:** Works well, good patterns

---

## Appendix A: Package Dependency Graph

```
main.go
  └─> internal/app/
       ├─> internal/orchestration/
       │    ├─> display/ (to be internal/display/)
       │    ├─> internal/llm/
       │    ├─> internal/session/
       │    └─> agent_prompts/
       ├─> internal/config/
       └─> internal/commands/ (to be removed)

agent_prompts/
  ├─> tools/
  ├─> workspace/
  └─> pkg/errors/

tools/
  ├─> tools/base/ (common)
  └─> workspace/

display/ (to be internal/display/)
  ├─> tracking/
  └─> pkg/models/

internal/session/
  └─> pkg/models/

pkg/models/
  └─> pkg/errors/
```

---

## Appendix B: File Move Checklist

### Phase 3: Display Package Decomposition

#### Terminal Package
- [ ] Move `display/renderer.go` → `internal/display/terminal/renderer.go`
- [ ] Move ANSI-related code → `internal/display/terminal/ansi.go`
- [ ] Update package declaration to `package terminal`
- [ ] Update imports in moved files
- [ ] Create facade in old location
- [ ] Update tests
- [ ] Run `go test ./internal/display/terminal/...`

#### Components Package
- [ ] Move `display/spinner.go` → `internal/display/components/spinner.go`
- [ ] Move `display/typewriter.go` → `internal/display/components/typewriter.go`
- [ ] Move `display/paginator.go` → `internal/display/components/paginator.go`
- [ ] Update package declarations
- [ ] Update imports in moved files
- [ ] Create facades in old location
- [ ] Move corresponding tests
- [ ] Run `go test ./internal/display/components/...`

#### Streaming Package
- [ ] Move `display/streaming_display.go` → `internal/display/streaming/display.go`
- [ ] Update package declaration
- [ ] Update imports
- [ ] Create facade in old location
- [ ] Move tests
- [ ] Run `go test ./internal/display/streaming/...`

#### Formatting Package
- [ ] Move `display/formatters/` → `internal/display/formatting/`
- [ ] Update package declarations
- [ ] Update imports
- [ ] Create facade in old location
- [ ] Move tests
- [ ] Run `go test ./internal/display/formatting/...`

#### Banners Package
- [ ] Move `display/banner.go` → `internal/display/banners/banner.go`
- [ ] Move `display/banner_renderer.go` → `internal/display/banners/renderer.go`
- [ ] Update package declarations
- [ ] Update imports
- [ ] Create facade in old location
- [ ] Move tests
- [ ] Run `go test ./internal/display/banners/...`

#### Events Package
- [ ] Move `display/event_handler.go` → `internal/display/events/handler.go`
- [ ] Update package declaration
- [ ] Update imports
- [ ] Create facade in old location
- [ ] Move tests
- [ ] Run `go test ./internal/display/events/...`

#### Tools Package
- [ ] Move `display/tool_adapter.go` → `internal/display/tools/adapter.go`
- [ ] Move `display/tool_renderer.go` → `internal/display/tools/renderer.go`
- [ ] Update package declarations
- [ ] Update imports
- [ ] Create facade in old location
- [ ] Move tests
- [ ] Run `go test ./internal/display/tools/...`

#### Final Validation
- [ ] All display tests pass: `go test ./internal/display/...`
- [ ] All other tests pass: `make test`
- [ ] Build succeeds: `make build`
- [ ] No import errors
- [ ] Manual integration test

---

## Appendix C: Quick Reference

### Commands for Each Phase

#### Phase 2: Dead Code Removal
```bash
# Before changes
git checkout -b refactor/remove-deprecated-code
make test  # Baseline

# Make changes
# ... edit files ...

# Validate
make check
make build
./bin/code-agent --help

# Commit
git add .
git commit -m "refactor: remove deprecated facades"
git push origin refactor/remove-deprecated-code
```

#### Phase 3: Display Decomposition
```bash
# Before changes
git checkout -b refactor/display-decomposition
make test  # Baseline

# Create structure
mkdir -p internal/display/{terminal,components,streaming,formatting,banners,events,tools}

# Move files (one at a time)
# ... move and update ...

# Validate after each move
go test ./internal/display/...
make test

# Final validation
make check
make build
./bin/code-agent new-session test
./bin/code-agent list-sessions
./bin/code-agent delete-session test

# Commit
git add .
git commit -m "refactor: decompose display package into subpackages"
git push origin refactor/display-decomposition
```

### Rollback Commands
```bash
# Rollback to previous commit
git reset --hard HEAD~1

# Rollback to specific phase
git checkout <phase-tag>

# Rollback entire branch
git checkout main
git branch -D refactor/<branch-name>
```

---

## Progress Tracking

### Phase 1: Foundation & Documentation ✅ COMPLETE
**Completed:** November 12, 2025

- ✅ Created comprehensive audit document (docs/audit.md)
- ✅ Created validation checklist (docs/validation-checklist.md)
- ✅ Updated working notes (docs/draft.md)
- ✅ Established baseline metrics
- ✅ All validation gates passing

**Outcome:** Foundation established for incremental refactoring with zero-regression guarantee.

### Phase 2: Dead Code Removal ✅ COMPLETE
**Completed:** November 12, 2025

#### Phase 2.1: Remove Orchestration Facade ✅
- ✅ Analyzed usage with grep
- ✅ Updated 3 test files to use direct orchestration
- ✅ Removed internal/app/orchestration.go
- ✅ All tests passing

#### Phase 2.2: Remove REPL Facade ✅
- ✅ Analyzed usage patterns
- ✅ Updated Application struct to use intrepl.REPL
- ✅ Updated internal/app/app.go imports
- ✅ Updated 3 test files
- ✅ Removed internal/app/repl.go
- ✅ All tests passing

#### Phase 2.3: Consolidate Command Handlers ✅
- ✅ Analyzed 3 command locations
- ✅ Enhanced internal/cli/commands/session.go with HandleSpecialCommands
- ✅ Updated main.go to import internal/cli/commands
- ✅ Removed cmd/commands/handlers.go
- ✅ Removed internal/commands/handlers.go
- ✅ Removed empty cmd/ directory
- ✅ All tests passing

**Metrics:**
- Lines removed: ~145 LOC
- Files deleted: 4
- Directories removed: 3
- Build status: ✅ Passing
- Test status: ✅ All tests passing

**Outcome:** Codebase cleaner, fewer deprecated patterns, improved navigation.

### Phase 3: Display Package Decomposition 🔄 IN PROGRESS
**Started:** November 12, 2025

#### Phase 3.1: Move display/ to internal/display/ ✅ COMPLETE
**Completed:** November 12, 2025

**What was done:**
1. ✅ Moved display/ directory to internal/display/ using `mv display internal/display`
2. ✅ Updated imports within internal/display/ package itself (sed replacement)
3. ✅ Updated all imports in internal/ packages:
   - ✅ internal/app/ (17 files updated)
   - ✅ internal/orchestration/ (8 files updated)
   - ✅ internal/repl/ (2 files updated)
   - ✅ internal/cli/ (5 files updated)
4. ✅ Updated all imports in root packages:
   - ✅ agent_prompts/ package
   - ✅ tools/ package
   - ✅ tracking/ package
   - ✅ workspace/ package
5. ✅ Verified no old imports remain outside internal/display/
6. ✅ All tests passing (make test)
7. ✅ Full validation passing (make check)
8. ✅ Binary builds successfully
9. ✅ Integration test verified (--help command works)

**Import Strategy:**
- Used sed with find to systematically replace `"code_agent/display"` with `"code_agent/internal/display"` across all Go files
- Processed packages in order: internal/display → internal/* → root packages
- No manual file editing required - all automated with shell commands

**Validation Results:**
- ✅ go fmt: Passed (some files auto-formatted)
- ✅ go vet: Passed (no issues)
- ✅ golangci-lint: Skipped (not installed - non-blocking)
- ✅ go test: All tests passing
- ✅ Build: Binary created successfully at ../bin/code-agent
- ✅ Integration: --help command displays correctly

**Metrics:**
- Files moved: ~40 Go files (5440 LOC total)
- Import updates: ~80+ files updated across 9 packages
- Build time: ~2 seconds
- Test time: ~10 seconds (all passing)

**Next Steps:**
Phase 3.2-3.6 will involve decomposing the internal/display/ package further into focused subpackages:
- internal/display/terminal/ (renderer, ANSI)
- internal/display/components/ (spinner, typewriter, paginator)
- internal/display/streaming/ (streaming display logic)
- internal/display/formatting/ (formatters)
- internal/display/banners/ (banner rendering)
- internal/display/events/ (event handling)
- internal/display/tools/ (tool-specific rendering)

This decomposition will be done carefully with similar validation at each step.

---

## Conclusion

This refactoring plan provides a **pragmatic, low-risk approach** to improving the code_agent codebase organization and maintainability. The plan is:

1. **Incremental** - Small, focused changes
2. **Reversible** - Easy rollback at any point
3. **Validated** - Comprehensive testing at each step
4. **Zero-regression** - All tests must pass always
5. **Pragmatic** - Focus on high-impact changes, defer optional items

**Estimated Timeline:** 5-7 weeks for complete execution  
**Risk Level:** Low to Medium (with proper validation)  
**Impact:** High - Significantly improved code organization

### Next Steps

1. **Review & Approve** this audit plan
2. **Set up** development environment and validation tools
3. **Begin Phase 1** - Foundation & Documentation
4. **Execute phases** sequentially with validation at each step
5. **Monitor** metrics and adjust as needed

### Questions or Concerns?

Contact the development team for clarification or to discuss alternative approaches.

---

**Document Version:** 1.0  
**Last Updated:** November 12, 2025  
**Status:** Ready for Review
