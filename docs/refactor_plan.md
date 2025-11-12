# Code Agent Refactoring Plan

**Date**: November 12, 2025  
**Project**: `adk_training_go/code_agent`  
**Version**: 1.0.0  
**Status**: Draft for Review

---

## Executive Summary

This refactoring plan aims to improve code organization, modularity, and maintainability of the `code_agent` codebase while ensuring **zero regression** and maintaining complete backward compatibility.

**Current State**: 138 Go files, 30 test files (21.7% coverage), all tests passing ✅

**Key Findings**:
- ✅ Clean main.go and internal/app architecture
- ✅ Well-designed tool registry system
- ✅ Good session persistence layer
- ⚠️ Display package has too many responsibilities (24 files)
- ⚠️ Error handling is inconsistent (pkg/errors underutilized)
- ⚠️ Prompt management could be better organized

**Approach**: Incremental, testable refactorings with continuous integration validation

---

## Guiding Principles

1. **Zero Regression**: Every change must pass all existing tests
2. **Pragmatic First**: Favor simple, working solutions over perfect abstractions
3. **Backward Compatible**: Maintain all public APIs and behaviors
4. **Incremental**: Small, verifiable steps with rollback capability
5. **Test Coverage**: Add tests before refactoring risky areas
6. **Go Best Practices**: Follow official Go project layout and idioms

---

## Phase 1: Foundation (Low Risk, High Value)

### 1.1 Standardize Error Handling

**Objective**: Adopt `pkg/errors` consistently across all packages

**Current State**:
- `pkg/errors` has comprehensive error types (ErrorCode, AgentError, helper functions)
- Only `tools/file` and `tools/exec` use it
- Most code uses `fmt.Errorf()` or plain errors

**Refactoring Steps**:

1. **Create error adoption guide** (docs/error-handling-guide.md)
   - When to use `pkg/errors` vs standard errors
   - Error code selection guidelines
   - Examples for each error category

2. **Update display package** (Priority: High)
   - Replace `fmt.Errorf()` with `pkg/errors` in display/
   - Add error codes for: rendering errors, parsing errors, format errors
   - Estimated files: 8-10

3. **Update agent package** (Priority: High)
   - Agent creation errors
   - Prompt building errors
   - Tool registration errors
   - Estimated files: 3-4

4. **Update remaining packages** (Priority: Medium)
   - workspace/
   - session/
   - tracking/
   - Estimated files: 5-6

**Success Criteria**:
- All packages use `pkg/errors` for application-level errors
- Standard library errors only for truly unexpected conditions
- All tests pass
- Zero behavioral changes

**Estimated Effort**: 3-4 hours
**Risk Level**: Low (additive changes, no API modifications)

---

### 1.2 Organize Agent Prompt Management

**Objective**: Consolidate prompt-related code into a dedicated subpackage

**Current State**:
- Prompt logic scattered across 5 files in `agent/`:
  - `dynamic_prompt.go`
  - `xml_prompt_builder.go`
  - `prompt_guidance.go`
  - `prompt_pitfalls.go`
  - `prompt_workflow.go`
- `agent/prompts/` directory exists but is underutilized

**Refactoring Steps**:

1. **Create prompt package structure**:
   ```
   agent/
   ├── coding_agent.go          # Keep agent factory here
   ├── coding_agent_test.go     # Keep agent tests here
   └── prompts/
       ├── builder.go            # XML prompt builder (was xml_prompt_builder.go)
       ├── dynamic.go            # Dynamic prompt generation (was dynamic_prompt.go)
       ├── guidance.go           # Guidance section (was prompt_guidance.go)
       ├── pitfalls.go           # Pitfalls section (was prompt_pitfalls.go)
       ├── workflow.go           # Workflow section (was prompt_workflow.go)
       ├── templates/            # Existing template directory
       │   └── ...
       └── prompts_test.go       # Consolidated prompt tests
   ```

2. **Move files with backward compatibility**:
   - Create `agent/prompts/` versions
   - Add re-exports in original files (deprecation notice)
   - Update imports in `coding_agent.go`
   - Verify all tests pass

3. **Update documentation**:
   - Update godoc comments
   - Add migration guide for external importers (if any)

**Success Criteria**:
- All prompt logic in `agent/prompts/`
- Backward compatible (old imports still work with deprecation notices)
- All tests pass
- No behavioral changes

**Estimated Effort**: 2-3 hours
**Risk Level**: Low (internal restructuring only)

---

## Phase 2: Display Package Refactoring (Medium Risk, High Value)

### 2.1 Analyze Display Package Responsibilities

**Current State**: 24 files with mixed responsibilities

**Files by Category**:

**Core Rendering** (Keep at root):
- `renderer.go`, `factory.go`, `facade.go`

**Streaming & Output** (Keep at root or move to `output/`):
- `streaming_display.go`, `streaming_segment.go`, `typewriter.go`

**UI Components** (Already in subpackages ✅):
- `banner/banner.go`
- `components/banner.go`, `components/timeline.go`
- `styles/colors.go`, `styles/formatting.go`

**Tool Integration** (Move to `tools/` or new `display/tools/`):
- `tool_adapter.go` (188 lines)
- `tool_renderer.go` (276 lines)
- `tool_renderer_internals.go`
- `tool_result_parser.go` (361 lines)

**Interactive UI** (Keep at root or move to `interactive/`):
- `paginator.go`, `spinner.go`

**Utilities**:
- `ansi.go`, `deduplicator.go`, `event.go`

### 2.2 Refactoring Strategy

**Option A: Conservative (Recommended)**
- Move tool-related files to `display/tooling/`
- Keep existing structure for everything else
- Minimal disruption, focused improvement

**Option B: Aggressive**
- Restructure into `display/{rendering, output, tooling, interactive, util}`
- Higher risk, more churn

**Recommendation**: Option A

**Steps for Option A**:

1. **Create `display/tooling/` package**:
   ```
   display/
   ├── tooling/
   │   ├── adapter.go         # Tool call adaptation
   │   ├── renderer.go        # Tool output rendering
   │   ├── parser.go          # Tool result parsing
   │   └── internal.go        # Internal helpers
   ```

2. **Move files incrementally**:
   - Move `tool_adapter.go` → `tooling/adapter.go`
   - Move `tool_renderer.go` → `tooling/renderer.go`
   - Move `tool_result_parser.go` → `tooling/parser.go`
   - Move `tool_renderer_internals.go` → `tooling/internal.go`

3. **Add facade for backward compatibility**:
   ```go
   // display/tool_adapter.go (deprecated)
   package display
   
   import "code_agent/display/tooling"
   
   // Deprecated: Use display/tooling package instead
   type ToolAdapter = tooling.Adapter
   
   // Deprecated: Use tooling.NewAdapter instead
   func NewToolAdapter(...) *ToolAdapter {
       return tooling.NewAdapter(...)
   }
   ```

4. **Update imports**:
   - Update `internal/app/` to import `display/tooling`
   - Keep facade for external compatibility

**Success Criteria**:
- Tool-related code isolated in `display/tooling/`
- All tests pass
- Backward compatible via facades
- Display package reduced from 24 to ~18 files

**Estimated Effort**: 4-5 hours
**Risk Level**: Medium (touches critical display logic)

---

## Phase 3: Model Provider Organization (Low Risk, Medium Value)

### 3.1 Reorganize Models by Provider

**Current State**: All providers mixed in `pkg/models/`

**Proposed Structure**:
```
pkg/models/
├── types.go              # Common types (Config, Provider, etc.)
├── registry.go           # Model registry
├── factory.go            # Factory interface
├── gemini/
│   ├── gemini.go        # Gemini client (was pkg/models/gemini.go)
│   └── factory.go       # Gemini factory
├── openai/
│   ├── openai.go        # OpenAI client
│   ├── adapter.go       # ADK adapter
│   ├── adapter_helpers.go # Tool conversion helpers
│   └── factory.go       # OpenAI factory
├── vertexai/
│   ├── vertexai.go      # VertexAI client
│   └── factory.go       # VertexAI factory
└── factories/           # Existing factory implementations (keep for now)
    ├── interface.go
    ├── registry.go
    ├── gemini.go
    ├── openai.go
    └── vertexai.go
```

**Migration Strategy**:

1. **Phase 3.1a: Create provider packages** (non-breaking)
   - Create `pkg/models/gemini/`, `openai/`, `vertexai/`
   - Copy relevant code
   - Add tests for each package

2. **Phase 3.1b: Add facades** (backward compatible)
   - Keep original files at `pkg/models/` with re-exports
   - Mark as deprecated
   - Internal code switches to new packages

3. **Phase 3.1c: Remove deprecated files** (future version)
   - Remove facades after deprecation period
   - Only if no external consumers

**Success Criteria**:
- Each provider in its own subpackage
- All tests pass
- Backward compatible
- Easier to add new providers

**Estimated Effort**: 3-4 hours
**Risk Level**: Low (internal reorganization with facades)

---

## Phase 4: Testing & Documentation (High Value)

### 4.1 Expand Test Coverage

**Current**: 30 test files (21.7% coverage)

**Targets**:
1. **Add integration tests** for:
   - Full agent workflow (user input → tool execution → response)
   - Session persistence roundtrip
   - Multi-workspace scenarios

2. **Add table-driven tests** for:
   - Error handling paths
   - Edge cases in file operations
   - Path resolution logic

3. **Improve coverage in**:
   - `display/` package (currently light on tests)
   - `workspace/` detection logic
   - CLI command handlers

**Success Criteria**:
- 40+ test files
- Integration test suite exists
- Critical paths have >80% coverage

**Estimated Effort**: 6-8 hours
**Risk Level**: None (additive only)

### 4.2 Documentation Improvements

**Create/Update**:
1. `docs/architecture.md` - System overview, component diagram
2. `docs/error-handling-guide.md` - Error handling standards
3. `docs/testing-guide.md` - How to write/run tests
4. `docs/contributing.md` - Contribution guidelines
5. Package-level godoc improvements

**Estimated Effort**: 3-4 hours
**Risk Level**: None

---

## Phase 5: Code Quality Improvements (Nice to Have)

### 5.1 Extract Large Functions

**Candidates** (functions >50 lines in large files):
- `display/tool_result_parser.go` - Parse* functions
- `tools/edit/search_replace_tools.go` - Search/replace logic
- `workspace/manager.go` - BuildEnvironmentContext
- `internal/app/app.go` - initializeModel

**Approach**:
- Extract helper functions
- Add unit tests for extracted functions
- Verify no behavioral changes

**Estimated Effort**: 2-3 hours per file
**Risk Level**: Low (with proper tests)

### 5.2 Add Code Metrics

**Tools to integrate**:
- `gocyclo` - Cyclomatic complexity
- `golangci-lint` - Comprehensive linting
- `go-critic` - Advanced static analysis

**Add to Makefile**:
```makefile
metrics:
    gocyclo -over 15 .
    golangci-lint run
    
check: fmt vet lint metrics test
```

**Estimated Effort**: 1-2 hours
**Risk Level**: None (CI enhancement)

---

## Implementation Schedule

### Week 1: Foundation
- **Day 1-2**: Phase 1.1 - Error handling standardization
- **Day 3**: Phase 1.2 - Agent prompt organization
- **Day 4-5**: Phase 4.1 - Test coverage expansion (start)

### Week 2: Display Refactoring
- **Day 1-3**: Phase 2.1-2.2 - Display package refactoring
- **Day 4-5**: Phase 4.1 - Test coverage expansion (continue)

### Week 3: Models & Completion
- **Day 1-2**: Phase 3.1 - Model provider organization
- **Day 3**: Phase 4.2 - Documentation improvements
- **Day 4-5**: Phase 5.1-5.2 - Code quality (if time permits)

**Total Estimated Time**: 15-20 hours over 3 weeks

---

## Risk Mitigation

### Pre-Refactoring Checklist
- [ ] All current tests pass
- [ ] Git branch created: `refactor/phase-{N}-{name}`
- [ ] Baseline metrics captured (test coverage, LOC)
- [ ] Backup of current working state

### During Refactoring
- [ ] Run `make check` after every significant change
- [ ] Commit after each working increment
- [ ] Keep changes small and reviewable
- [ ] Update tests before moving code

### Post-Refactoring Validation
- [ ] All tests pass (`make test`)
- [ ] All lints pass (`make check`)
- [ ] Manual smoke test of CLI
- [ ] Documentation updated
- [ ] Changelog entry added

---

## Rollback Strategy

**If issues arise**:
1. **Stop immediately** - Don't compound issues
2. **Identify last known good commit**
3. **Revert using**: `git revert <commit-range>` or `git reset --hard <good-commit>`
4. **Document issue** in rollback notes
5. **Review approach** before retrying

---

## Success Metrics

### Quantitative
- [ ] All 30 existing test files still pass
- [ ] New test files added: target 10+ (total 40+)
- [ ] `make check` passes with zero warnings
- [ ] No increase in cyclomatic complexity of refactored functions
- [ ] LOC per file reduced in refactored packages

### Qualitative
- [ ] Code is easier to navigate
- [ ] Package boundaries are clearer
- [ ] Error messages are more helpful
- [ ] New contributors can understand structure
- [ ] Maintenance burden is reduced

---

## Non-Goals (Out of Scope)

**Explicitly NOT doing**:
- ❌ Changing CLI interface or behavior
- ❌ Modifying tool functionality
- ❌ Rewriting session persistence layer
- ❌ Changing workspace detection logic (unless bugs found)
- ❌ Updating external dependencies (except bug fixes)
- ❌ Performance optimization (unless identified bottleneck)
- ❌ Adding new features

---

## Appendix A: Quick Reference

### File Movement Summary

**Phase 1.2 - Agent Prompts**:
```
agent/dynamic_prompt.go        → agent/prompts/dynamic.go
agent/xml_prompt_builder.go    → agent/prompts/builder.go
agent/prompt_guidance.go       → agent/prompts/guidance.go
agent/prompt_pitfalls.go       → agent/prompts/pitfalls.go
agent/prompt_workflow.go       → agent/prompts/workflow.go
```

**Phase 2.2 - Display Tooling**:
```
display/tool_adapter.go        → display/tooling/adapter.go
display/tool_renderer.go       → display/tooling/renderer.go
display/tool_result_parser.go  → display/tooling/parser.go
display/tool_renderer_internals.go → display/tooling/internal.go
```

**Phase 3.1 - Model Providers**:
```
pkg/models/gemini.go           → pkg/models/gemini/gemini.go
pkg/models/openai.go           → pkg/models/openai/openai.go
pkg/models/openai_adapter.go   → pkg/models/openai/adapter.go
pkg/models/openai_adapter_helpers.go → pkg/models/openai/adapter_helpers.go
```

### Backward Compatibility Facades

All moved files will have facade re-exports in original locations with deprecation notices for at least 2 release cycles.

---

## Appendix B: Go Best Practices Applied

1. **Package Organization** (Effective Go)
   - Packages grouped by functionality, not type
   - Clear single responsibility per package
   - Avoid circular dependencies

2. **Error Handling** (Go Blog)
   - Use custom error types for application errors
   - Wrap errors with context
   - Avoid error codes in normal returns

3. **Testing** (Go Wiki)
   - Table-driven tests for variations
   - Test packages in `_test` package for black-box testing
   - Use testdata/ directories for fixtures

4. **Documentation** (Go Code Review)
   - Every exported symbol has godoc comment
   - Package comment explains purpose
   - Examples in godoc where helpful

---

## Sign-off

**Reviewed by**: [Pending]  
**Approved by**: [Pending]  
**Start Date**: [TBD]  
**Target Completion**: [TBD]

---

**Document Version**: 1.0  
**Last Updated**: November 12, 2025
