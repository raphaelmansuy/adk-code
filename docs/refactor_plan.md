# Code Agent Refactoring Plan

**Date**: November 12, 2025  
**Version**: 1.0  
**Risk Level**: LOW (0% regression target)  
**Approach**: Incremental, backwards-compatible changes  

## Executive Summary

This plan addresses structural issues in the `code_agent/` codebase while maintaining 100% backwards compatibility and ensuring zero regression. The refactoring focuses on improving modularity, reducing coupling, and following Go best practices.

**Current State:**
- 77 Go files, ~14,000 LOC
- 13 test files (~17% coverage)
- Mixed package organization
- Some God objects and procedural code

**Target State:**
- Clean layered architecture
- >80% test coverage
- No duplicate code
- Clear separation of concerns
- Easy to maintain and extend

## Guiding Principles

1. **Zero Regression**: Every change must maintain existing functionality
2. **Incremental Migration**: No big-bang rewrites
3. **Backwards Compatibility**: Existing code continues to work during transition
4. **Test Coverage**: Add tests before refactoring
5. **Go Conventions**: Follow standard Go project layout and idioms

## Phase 1: Foundation & Cleanup (Low Risk)

**Goal**: Remove dead code, fix obvious issues, improve project structure

### 1.1 Remove Legacy Model Package

**Issue**: Duplicate model packages causing confusion

**Actions:**
```bash
# Current structure:
code_agent/model/          # LEGACY - 721 lines
├── openai.go             # Full OpenAI adapter
└── vertexai.go           # Deprecated, moved to pkg/models/factory.go

code_agent/pkg/models/    # NEW - proper location
└── factory.go            # Imports legacy code_agent/model
```

**Steps:**
1. Move `model/openai.go` adapter code to `pkg/models/openai.go` (the existing file only has definitions)
2. Update `pkg/models/factory.go` to use local OpenAI adapter
3. Remove `import "code_agent/model"` from factory.go
4. Delete entire `code_agent/model/` directory
5. Run tests: `make test`
6. Verify build: `make build`

**Risk**: LOW - Only internal imports affected  
**Effort**: 1-2 hours  
**Tests Required**: Existing model tests should pass

### 1.2 Extract Main Package Business Logic

**Issue**: `main.go` (410 lines) contains too much orchestration logic

**Actions:**
```bash
# Create internal package for application code
code_agent/internal/app/
├── app.go           # Application struct and Run() method
├── repl.go          # REPL loop logic
├── signals.go       # Signal handling
└── session.go       # Session setup logic
```

**Steps:**
1. Create `internal/app/` directory structure
2. Extract REPL loop to `repl.go`
3. Extract signal handling to `signals.go`
4. Extract session initialization to `session.go`
5. Create `Application` struct with `Run()` method
6. Update `main.go` to be thin entry point (~50 lines)
7. Move `cli_commands.go` content to `pkg/cli/`
8. Move `utils.go` to `internal/app/utils.go`

**Risk**: LOW - Pure code movement  
**Effort**: 3-4 hours  
**Tests Required**: Integration test for main flow

### 1.3 Split Display Package

**Issue**: `display/renderer.go` (879 lines) is a God object

**Actions:**
```bash
# Split renderer into focused components:
display/
├── renderer.go           # Thin facade (~150 lines)
├── styles/
│   └── styles.go        # ANSI color styles
├── formatters/
│   ├── banner.go        # Already exists
│   ├── event.go         # Event formatting
│   ├── token.go         # Token metrics formatting
│   └── error.go         # Error formatting
└── components/
    ├── spinner.go       # Already exists
    ├── paginator.go     # Already exists
    └── typewriter.go    # Already exists
```

**Steps:**
1. Create new subdirectory structure
2. Extract color styles to `styles/styles.go`
3. Extract formatting functions to `formatters/`
4. Update `renderer.go` to delegate to sub-components
5. Keep all public APIs unchanged (facade pattern)
6. Update imports across codebase incrementally

**Risk**: LOW - Facade pattern maintains compatibility  
**Effort**: 4-5 hours  
**Tests Required**: Existing display tests should pass

## Phase 2: Architecture Improvements (Medium Risk)

**Goal**: Improve modularity and reduce coupling

### 2.1 Introduce Internal Packages

**Issue**: Application-specific code mixed with library code

**Actions:**
```bash
# New structure:
code_agent/
├── internal/           # Application-specific, not importable
│   ├── app/           # Application lifecycle
│   ├── repl/          # REPL implementation
│   └── config/        # Configuration management
├── pkg/               # Library code, importable
│   ├── cli/
│   ├── models/
│   └── ...
└── cmd/               # Multiple binaries (future)
    └── code-agent/
        └── main.go
```

**Steps:**
1. Create `internal/` directory
2. Move application-specific code from main package
3. Keep `pkg/` for reusable library code
4. Update import paths incrementally
5. Ensure `internal/` cannot be imported by external packages

**Risk**: MEDIUM - Import path changes  
**Effort**: 5-6 hours  
**Tests Required**: All tests must pass

### 2.2 Refactor Tool Registration

**Issue**: Explicit tool instantiation in coding_agent.go is fragile

**Current Pattern:**
```go
// agent/coding_agent.go
func NewCodingAgent(...) {
    // Manually call 15+ tool constructors
    NewReadFileTool()
    NewWriteFileTool()
    // ... 13 more
}
```

**New Pattern:**
```go
// tools/file/file_tools.go
func init() {
    // Auto-register on import
    tools.Register(tools.ToolMetadata{
        Tool: NewReadFileTool(),
        Category: tools.CategoryFileOperations,
        Priority: 1,
    })
}

// agent/coding_agent.go
func NewCodingAgent(...) {
    // Just get registered tools
    registry := tools.GetRegistry()
    registeredTools := registry.GetAllTools()
    // ... use tools
}
```

**Steps:**
1. Add `init()` functions to each tool file
2. Register tools automatically on package import
3. Remove explicit tool construction from coding_agent.go
4. Import tool packages with blank identifier if needed
5. Ensure tool order doesn't matter

**Risk**: LOW - Registry pattern already exists  
**Effort**: 2-3 hours  
**Tests Required**: Tool registration tests

### 2.3 CLI Command Consolidation

**Issue**: CLI logic scattered across multiple files and packages

**Target Structure:**
```bash
pkg/cli/
├── cli.go              # Main CLI coordinator
├── flags.go            # Flag definitions
├── config.go           # Configuration
├── commands/           # All command handlers
│   ├── session.go      # new-session, list-sessions, delete-session
│   ├── repl.go         # REPL commands (/help, /tools, etc.)
│   └── model.go        # /set-model, /current-model
├── display.go          # Display helpers
└── syntax.go           # Parsing utilities
```

**Steps:**
1. Create `pkg/cli/commands/` directory
2. Move session handlers from `handlers.go` to `commands/session.go`
3. Move REPL commands from `cli_commands.go` (main package) to `commands/repl.go`
4. Create unified command dispatcher
5. Remove duplicate code from main package

**Risk**: LOW - Pure reorganization  
**Effort**: 3-4 hours  
**Tests Required**: CLI integration tests

## Phase 3: Testing & Documentation (Low Risk)

**Goal**: Achieve comprehensive test coverage

### 3.1 Add Missing Tests

**Packages Needing Tests:**
- ❌ `tools/edit/` (search_replace, patches)
- ❌ `tools/exec/` (terminal execution)
- ❌ `tools/search/` (grep, diff)
- ❌ `display/renderer` (comprehensive tests)
- ❌ `internal/app/` (new package)

**Test Strategy:**
1. **Unit Tests**: Test individual functions in isolation
2. **Integration Tests**: Test tool execution end-to-end
3. **Table-Driven Tests**: Use Go's testing patterns
4. **Mock External Dependencies**: File system, command execution

**Steps:**
1. Create test files for each missing package
2. Aim for >80% code coverage per package
3. Use `go test -cover` to verify
4. Add integration tests for critical paths
5. Set up CI to enforce coverage thresholds

**Risk**: NONE - Tests don't affect production code  
**Effort**: 8-10 hours (spread across other phases)  
**Tests Required**: All new tests must pass

### 3.2 Documentation Updates

**Actions:**
1. Update README.md with new structure
2. Add package-level documentation (godoc)
3. Create architecture diagram
4. Document refactoring decisions (ADR format)
5. Update `.github/copilot-instructions.md`

**Risk**: NONE  
**Effort**: 4-5 hours

## Phase 4: Advanced Improvements (Future)

**Goal**: Prepare for future extensibility (not in scope for 0% regression)

### 4.1 Interface-Based Design (Future)

**Current**: Tight coupling to concrete types  
**Target**: Depend on interfaces, not implementations

```go
// Example refactoring (not in scope)
type ModelProvider interface {
    CreateModel(ctx context.Context, config any) (model.LLM, error)
    SupportedModels() []string
}

type ToolRegistry interface {
    Register(tool Tool) error
    GetAllTools() []Tool
}
```

### 4.2 Plugin System (Future)

**Target**: Allow external tools to be registered at runtime

### 4.3 gRPC API (Future)

**Target**: Expose agent as a service for integration

## Migration Strategy

### Incremental Rollout

Each phase follows this pattern:

1. **Create New Structure**: Add new packages/files without changing existing code
2. **Add Tests**: Ensure new code is tested
3. **Migrate Incrementally**: Move code piece by piece
4. **Deprecate Old Code**: Mark old code as deprecated but keep it working
5. **Remove Dead Code**: Only after migration is complete and tested

### Rollback Plan

For each phase:
- Create feature branch
- Commit after each logical change
- Run full test suite after each commit
- If tests fail, revert that commit
- Squash-merge to main only when phase is complete

### Validation Checklist

Before merging each phase:

- [ ] All existing tests pass
- [ ] New tests added and passing
- [ ] `make check` passes (fmt, vet, lint)
- [ ] Build succeeds: `make build`
- [ ] Manual smoke test: Run CLI and execute common commands
- [ ] Code review completed
- [ ] Documentation updated

## Risk Mitigation

### Low-Risk Changes
- Pure code movement (no logic changes)
- Adding new packages/files without deleting old ones
- Extracting functions (original calls unchanged)
- Adding tests

### Medium-Risk Changes
- Changing import paths
- Modifying public APIs (even with backwards compatibility)
- Refactoring God objects

### High-Risk Changes (Not in This Plan)
- Rewriting core algorithms
- Changing data structures
- Modifying external interfaces

## Success Metrics

### Phase 1 (Foundation)
- [ ] Legacy `model/` package deleted
- [ ] `main.go` reduced to <100 lines
- [ ] `display/renderer.go` split into <300 lines per file
- [ ] All tests passing

### Phase 2 (Architecture)
- [ ] `internal/` package structure in place
- [ ] Tool registration automated
- [ ] CLI commands consolidated
- [ ] No circular dependencies

### Phase 3 (Testing)
- [ ] >80% code coverage
- [ ] All critical paths have integration tests
- [ ] Documentation complete

### Overall Success
- [ ] Zero regression in functionality
- [ ] Faster development velocity (easier to add features)
- [ ] Easier onboarding for new contributors
- [ ] Cleaner, more maintainable codebase

## Timeline Estimate

**Phase 1**: 2-3 days (8-10 hours work)  
**Phase 2**: 3-4 days (10-12 hours work)  
**Phase 3**: 4-5 days (12-15 hours work)  

**Total**: 2 weeks of focused work

## Immediate Next Steps

1. Review and approve this plan
2. Create tracking issue for refactoring project
3. Set up feature branch: `refactor/phase-1-foundation`
4. Start with Phase 1.1 (remove legacy model package)
5. Commit and test each change incrementally

## Appendix: File Changes

### Phase 1 File Changes

**Created:**
- `internal/app/app.go`
- `internal/app/repl.go`
- `internal/app/signals.go`
- `internal/app/session.go`
- `internal/app/utils.go`
- `display/styles/styles.go`
- `display/formatters/banner.go`
- `display/formatters/event.go`
- `display/formatters/token.go`
- `display/formatters/error.go`

**Modified:**
- `pkg/models/openai.go` (merge from `model/openai.go`)
- `pkg/models/factory.go` (remove legacy import)
- `main.go` (reduce to thin entry point)
- `display/renderer.go` (become facade)

**Deleted:**
- `code_agent/model/` (entire directory)
- `cli_commands.go` (moved to pkg/cli)
- `utils.go` (moved to internal/app)

**Impact**: ~15 files created/modified, ~700 lines moved

### Phase 2 File Changes

**Created:**
- `pkg/cli/commands/session.go`
- `pkg/cli/commands/repl.go`
- `pkg/cli/commands/model.go`
- `cmd/code-agent/main.go` (optional)

**Modified:**
- All tool files (add init() functions)
- `agent/coding_agent.go` (use registry)
- Import paths throughout codebase

**Impact**: ~25 files modified

## Questions & Answers

**Q: Why not use a different directory layout (e.g., hexagonal architecture)?**  
A: The current structure is close to Go conventions. Incremental improvement is safer than radical change.

**Q: Should we use dependency injection?**  
A: Not in this phase. Focus on modularity first, DI can be added later if needed.

**Q: What about breaking changes to internal APIs?**  
A: Internal packages (in `internal/`) can have breaking changes since they're not importable by external code.

**Q: How do we ensure 0% regression?**  
A: Comprehensive testing, incremental changes, and careful code review. Each commit should leave the codebase in a working state.

## Conclusion

This refactoring plan provides a clear, incremental path to improving the `code_agent` codebase while maintaining stability and backwards compatibility. By following Go best practices and focusing on modularity, we'll create a more maintainable foundation for future development.

**Recommendation**: Start with Phase 1.1 (remove legacy model package) as a low-risk, high-value first step.
