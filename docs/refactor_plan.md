# Code Agent Refactoring Plan - Phase 5: Modularization

**Version**: 1.0  
**Date**: November 12, 2025  
**Objective**: Improve code organization and modularity while maintaining 0% regression  
**Duration**: 5-7 days  
**Risk Level**: LOW (comprehensive test coverage as safety net)

---

## Executive Summary

Transform the code_agent codebase from good → excellent by:
1. **Reducing file sizes** to improve testability (files >500 LOC → <400 LOC)
2. **Reorganizing packages** for clarity and reduced cognitive load
3. **Formalizing interface contracts** for extensibility
4. **Maintaining zero regressions** with comprehensive testing throughout

**Key Constraints**:
- ✅ 250+ existing tests must continue to pass
- ✅ No breaking changes to public APIs
- ✅ All functionality preserved
- ✅ Performance characteristics unchanged

---

## Phase 5 Breakdown

### Phase 5A: File Size Reduction (High Priority)

**Goal**: Split large files (>500 LOC) into focused modules

#### Task 5A.1: Split `tools/file/file_tools.go` (562 LOC)

**Current State**:
- Single file with 5 tool implementations
- Multiple validation utilities
- Atomic write operations mixed with tool code

**Target Structure**:
```
tools/file/
├── file_tools.go           (keep: shared types, errors)
├── read_tool.go            (new: ReadFile tool)
├── write_tool.go           (new: WriteFile tool)
├── list_tool.go            (new: ListDirectory tool)
├── search_tool.go          (new: SearchFiles tool)
└── validation.go           (moved: validation helpers)
```

**Changes**:
1. Extract `NewReadFileTool()` to read_tool.go
2. Extract `NewWriteFileTool()` to write_tool.go
3. Extract `NewListDirectoryTool()` to list_tool.go
4. Extract `NewSearchFilesTool()` to search_tool.go
5. Move validation helpers to validation.go
6. Update imports in parent tools.go facade
7. Maintain init() registration in each file

**Testing Strategy**:
- Existing file tests should continue to work (no behavior change)
- May need to adjust test file organization
- Run full test suite after each split

**Effort**: 2-3 hours  
**Risk**: MEDIUM (tool registration via init)  
**Regression Risk**: LOW (structure change only)  

**Validation**:
```bash
make check                          # All tests pass
go list ./tools/file/...           # All packages resolve
grep -r "NewReadFileTool"           # Verify imports work
```

---

#### Task 5A.2: Split `pkg/models/openai_adapter.go` (716 LOC)

**Current State**:
- OpenAI adapter implementation (716 LOC)
- Streaming logic mixed with API calls
- Error handling interleaved with client code

**Target Structure**:
```
pkg/models/
├── openai_adapter.go       (keep: OpenAI model factory, type definitions)
├── openai_adapter_client.go (new: Client implementation)
├── openai_adapter_stream.go (new: Streaming logic)
└── openai_adapter_errors.go (new: Error handling)
```

**Changes**:
1. Extract streaming implementation to openai_adapter_stream.go
2. Extract client creation/management to openai_adapter_client.go
3. Extract error handling to openai_adapter_errors.go
4. Keep type definitions and factory in openai_adapter.go
5. Maintain internal interfaces between files
6. Update imports in pkg/models re-exports

**Testing Strategy**:
- OpenAI model tests should continue to pass
- May need to add unit tests for extracted functions
- Run model registry tests

**Effort**: 3-4 hours  
**Risk**: MEDIUM (affects provider interface)  
**Regression Risk**: LOW (refactoring only)  

**Validation**:
```bash
make check                                 # All tests pass
grep -r "CreateOpenAIModel"               # Verify API works
pkg/models tests pass                      # Model tests
```

---

#### Task 5A.3: Split `persistence/sqlite.go` (570 LOC) & `persistence/models.go` (627 LOC)

**Current State**:
- models.go: GORM model definitions + business logic (627 LOC)
- sqlite.go: Database layer, schema, CRUD operations (570 LOC)
- Responsibilities unclear, hard to test

**Target Structure**:
```
persistence/
├── models.go              (keep: GORM model definitions only)
├── schema.go              (new: Schema creation, initialization)
├── service.go             (new: CRUD service implementations)
├── migrations.go          (new: Database migrations)
└── sqlite.go              (keep: SQLite connection management)
```

**Changes**:
1. Extract schema creation logic to schema.go
2. Extract CRUD operations to service.go
3. Extract migrations to migrations.go
4. Keep GORM model definitions in models.go
5. Keep connection management in sqlite.go
6. Update manager.go to use new structure

**Testing Strategy**:
- Persistence layer tests should continue to work
- May add unit tests for schema/service
- Integration tests for database operations

**Effort**: 4-5 hours  
**Risk**: MEDIUM (database interaction patterns)  
**Regression Risk**: MEDIUM (more complexity here)  

**Validation**:
```bash
make check                                 # All tests pass
sqlite tests pass                          # Database tests
persistence/service tests pass             # New service tests
```

---

#### Task 5A.4: Reorganize `display/tool_renderer.go` (425 LOC)

**Current State**:
- Single file with all tool rendering logic
- Multiple render methods for different tool types
- Complex decision logic mixed throughout

**Target Structure**:
```
display/
├── tool_renderer.go              (keep: Factory, routing)
├── tool_renderers/               (new: Subpackage)
│   ├── file_renderer.go         (ReadFile, WriteFile, List)
│   ├── edit_renderer.go         (Replace, Patch, Search)
│   ├── exec_renderer.go         (Execute, Grep)
│   └── other_renderer.go        (Workspace, Display tools)
└── tool_result_parser.go        (keep: Result parsing)
```

**Changes**:
1. Create display/tool_renderers/ subpackage
2. Extract file operation rendering to file_renderer.go
3. Extract editing operation rendering to edit_renderer.go
4. Extract execution rendering to exec_renderer.go
5. Extract others to other_renderer.go
6. Keep factory/routing in tool_renderer.go
7. Update imports and maintain public API

**Testing Strategy**:
- Display tool tests should pass
- May add unit tests for each renderer
- Tool approval/execution tests still work

**Effort**: 3-4 hours  
**Risk**: MEDIUM (display contract)  
**Regression Risk**: LOW (UI logic only)  

**Validation**:
```bash
make check                                 # All tests pass
display tests pass                         # Display tests
tool_renderer tests pass                   # Renderer tests
```

---

### Phase 5B: CLI & REPL Reorganization (Medium Priority)

**Goal**: Decompose `pkg/cli/commands/repl.go` (448 LOC)

#### Task 5B.1: Split `pkg/cli/commands/repl.go` (448 LOC)

**Current State**:
- Single file with REPL event loop + command handling
- User input → command dispatch → output generation
- Multiple responsibilities interleaved

**Target Structure**:
```
pkg/cli/commands/
├── repl.go                    (keep: REPL struct, main Run loop)
├── repl_commands.go           (new: Command dispatch, routing)
├── repl_formatter.go          (new: Output formatting)
├── repl_input.go              (new: User input handling)
└── commands/                  (existing: Individual commands)
```

**Changes**:
1. Extract command dispatch logic to repl_commands.go
2. Extract output formatting to repl_formatter.go
3. Extract input handling to repl_input.go
4. Keep main REPL loop in repl.go
5. Update imports and maintain API

**Testing Strategy**:
- REPL tests should continue to pass
- Add unit tests for command dispatch
- Add unit tests for formatting

**Effort**: 3-4 hours  
**Risk**: MEDIUM (REPL loop logic)  
**Regression Risk**: LOW (command routing)  

**Validation**:
```bash
make check                                 # All tests pass
cli/commands tests pass                    # REPL tests
```

---

### Phase 5C: Interface Formalization (Lower Priority)

**Goal**: Define explicit interfaces for key abstractions

#### Task 5C.1: Formalize Tool Interfaces

**Current State**:
- Tools registered via registry pattern
- No explicit interface contract
- Tool input/output types only discoverable at runtime

**Changes**:
1. Create `tools/common/contracts.go`
2. Define `Tool` interface:
   ```go
   type Tool interface {
       Name() string
       Description() string
       Execute(ctx context.Context, input interface{}) (interface{}, error)
   }
   ```
3. Document input/output types for each tool
4. Update registry to enforce interface
5. Add validation at registration time

**Testing Strategy**:
- Registry tests verify interface compliance
- Tool registration tests pass

**Effort**: 1-2 hours  
**Risk**: LOW (additive)  
**Regression Risk**: VERY LOW  

---

#### Task 5C.2: Formalize Provider Interface

**Current State**:
- Model providers (Gemini, OpenAI, VertexAI)
- Similar implementation, no formal interface
- Factory function returns generic model.LLM

**Changes**:
1. Create explicit `Provider` interface
2. Define clear contract for all providers
3. Document provider capabilities
4. Update factory to enforce interface

**Testing Strategy**:
- Model tests verify provider compliance
- Provider tests pass

**Effort**: 1-2 hours  
**Risk**: LOW (additive)  
**Regression Risk**: VERY LOW  

---

#### Task 5C.3: Formalize Renderer Interface

**Current State**:
- Multiple renderer types (Renderer, BannerRenderer, ToolRenderer)
- No explicit interface contract
- Composition-based approach

**Changes**:
1. Define `Renderer` interface
2. Document rendering contract
3. Define `BannerRenderer`, `ToolRenderer` interfaces
4. Document expected behavior

**Testing Strategy**:
- Display tests verify interface compliance
- Renderer tests pass

**Effort**: 1 hour  
**Risk**: VERY LOW (documentation only)  
**Regression Risk**: VERY LOW  

---

## Implementation Timeline

### Week 1: File Size Reduction & Testing

| Day | Task | Owner | Time | Status |
|-----|------|-------|------|--------|
| Mon | 5A.1: Split tools/file | Dev | 2-3h | 🔄 |
| Mon | Run full test suite | Dev | 0.5h | 🔄 |
| Mon | Document changes | Dev | 0.5h | 🔄 |
| Tue | 5A.2: Split openai_adapter | Dev | 3-4h | 🔄 |
| Tue | Run full test suite | Dev | 0.5h | 🔄 |
| Tue | Document changes | Dev | 0.5h | 🔄 |
| Wed | 5A.3: Split persistence | Dev | 4-5h | 🔄 |
| Wed | Run full test suite | Dev | 0.5h | 🔄 |
| Wed | Document changes | Dev | 0.5h | 🔄 |
| Thu | 5A.4: Reorganize display | Dev | 3-4h | 🔄 |
| Thu | Run full test suite | Dev | 0.5h | 🔄 |
| Thu | Document changes | Dev | 0.5h | 🔄 |
| Fri | 5B.1: Split REPL | Dev | 3-4h | 🔄 |
| Fri | Run full test suite | Dev | 0.5h | 🔄 |
| Fri | Document changes | Dev | 0.5h | 🔄 |

### Week 2: Interface Formalization & Documentation

| Day | Task | Owner | Time | Status |
|-----|------|-------|------|--------|
| Mon | 5C.1: Tool interfaces | Dev | 1-2h | 🔄 |
| Mon | 5C.2: Provider interface | Dev | 1-2h | 🔄 |
| Mon | 5C.3: Renderer interface | Dev | 1h | 🔄 |
| Mon | Full regression testing | Dev | 1h | 🔄 |
| Tue | Documentation update | Dev | 2-3h | 🔄 |
| Tue | Final validation | Dev | 1h | 🔄 |

---

## Detailed Steps for Each Task

### Task 5A.1: Split tools/file/file_tools.go

**Step 1: Create read_tool.go**
```bash
# Copy ReadFile-related code from file_tools.go
# Include: ReadFileInput, ReadFileOutput, NewReadFileTool
# Include: init() function with registration
# Keep error types in file_tools.go
```

**Step 2: Create write_tool.go**
```bash
# Copy WriteFile-related code from file_tools.go
# Include: WriteFileInput, WriteFileOutput, NewWriteFileTool
# Include: init() function with registration
# Keep error types in file_tools.go
```

**Step 3: Create list_tool.go**
```bash
# Copy ListDirectory-related code from file_tools.go
# Include: ListDirectoryInput, ListDirectoryOutput, NewListDirectoryTool
# Include: init() function with registration
```

**Step 4: Create search_tool.go**
```bash
# Copy SearchFiles-related code from file_tools.go
# Include: SearchFilesInput, SearchFilesOutput, NewSearchFilesTool
# Include: init() function with registration
```

**Step 5: Extract validation.go**
```bash
# Move validation helper functions to validation.go
# Keep: ValidateFilePath, ValidateSizeLimit, other helpers
```

**Step 6: Update file_tools.go**
```bash
# Keep: Type definitions (Input/Output structs)
# Keep: Error constants and types
# Remove: Tool implementations (now in separate files)
# Verify: Total file size <300 LOC
```

**Step 7: Update imports in tools.go**
```bash
# Verify re-exports still work
# Check: NewReadFileTool, NewWriteFileTool, etc. accessible
```

**Step 8: Run tests**
```bash
cd /code_agent
make check
# Verify: All tests pass
# Verify: No import errors
```

**Step 9: Document**
```bash
# Update tools/file/README or comments
# Document: File splitting rationale
# Document: Import paths for users
```

---

### Task 5A.2: Split pkg/models/openai_adapter.go

**Step 1: Create openai_adapter_client.go**
```bash
# Move: Client creation functions
# Move: Request/response handling
# Move: API communication logic
# Keep: Clean interface between files
```

**Step 2: Create openai_adapter_stream.go**
```bash
# Move: Streaming implementation
# Move: Stream processing logic
# Move: Token counting for streaming
# Keep: Interface to client functions
```

**Step 3: Create openai_adapter_errors.go**
```bash
# Move: Error handling functions
# Move: Error conversion logic
# Move: Error type definitions
# Keep: Shared error interfaces
```

**Step 4: Update openai_adapter.go**
```bash
# Keep: Type definitions
# Keep: Factory function (CreateOpenAIModel)
# Keep: Main Model implementation
# Remove: Implementation details (now in other files)
# Verify: Total file size <300 LOC
```

**Step 5: Run tests**
```bash
make check
# Verify: All model tests pass
# Verify: OpenAI-specific tests pass
```

**Step 6: Document**
```bash
# Document: Function signatures
# Document: Internal interfaces between files
```

---

### Task 5A.3: Split persistence layer

**Step 1: Create persistence/schema.go**
```bash
# Move: Schema creation functions
# Move: Table initialization
# Move: Index creation
# Keep: Schema constants
```

**Step 2: Create persistence/service.go**
```bash
# Move: CRUD operations
# Move: Service implementation
# Move: Database queries
# Keep: Clean interface to schema
```

**Step 3: Create persistence/migrations.go**
```bash
# Move: Migration logic if any
# Move: Version tracking
# Keep: Migration registration
```

**Step 4: Update models.go**
```bash
# Keep: GORM model definitions ONLY
# Remove: Business logic
# Verify: File size <300 LOC
```

**Step 5: Update sqlite.go**
```bash
# Keep: Connection management
# Update: Imports to use new files
```

**Step 6: Run tests**
```bash
make check
# Verify: Persistence tests pass
# Verify: Database operations work
```

---

### Task 5B.1: Split pkg/cli/commands/repl.go

**Step 1: Create repl_commands.go**
```bash
# Move: Command dispatch logic
# Move: Command routing
# Move: Command parsing
# Keep: Function to route commands
```

**Step 2: Create repl_formatter.go**
```bash
# Move: Output formatting functions
# Move: Result formatting
# Move: Error formatting
# Keep: Clean formatting functions
```

**Step 3: Create repl_input.go**
```bash
# Move: Input handling functions
# Move: History management
# Move: Input validation
# Keep: Clean input functions
```

**Step 4: Update repl.go**
```bash
# Keep: REPL struct
# Keep: Run() main loop
# Update: Call to repl_commands for routing
# Update: Call to repl_formatter for output
# Verify: Total file size <250 LOC
```

**Step 5: Run tests**
```bash
make check
# Verify: CLI tests pass
# Verify: REPL tests pass
```

---

## Testing Strategy

### Per-Task Testing

**After Each Major Change**:
1. Run full test suite: `make check`
2. Check for regressions: Compare before/after behavior
3. Verify imports: No circular dependencies
4. Check binary compatibility: Existing code still works

### Regression Testing Checklist

```
For each split file:
□ Original file tests still pass (may need path updates)
□ New files have appropriate init() registration
□ Public API unchanged
□ No new imports that create cycles
□ Tool registration verified with registry test
□ Type definitions match original
□ Error handling preserved
```

### Integration Testing

**After All Changes**:
```bash
cd /code_agent

# Full test suite
make check

# Verify tool registration
go test ./tools/... -v

# Verify model creation
go test ./pkg/models/... -v

# Verify persistence
go test ./persistence/... -v

# Build and smoke test
make build
./code-agent --help
```

---

## Rollback Plan

If any task introduces regressions:

**Immediate Action**:
1. Run `make check` to identify failures
2. Review git log to find last good commit
3. Revert with: `git revert <commit>`
4. Run `make check` to verify revert successful

**Root Cause Analysis**:
1. Identify what broke (test output)
2. Analyze code changes (git diff)
3. Fix issue or take different approach
4. Create targeted test to prevent regression

---

## Documentation Requirements

### For Each Task, Update/Create:

1. **Code Comments**
   - Document why file was split
   - Document responsibilities of each file
   - Document interfaces between files

2. **README Updates**
   - Update package structure documentation
   - Update architecture diagrams if applicable
   - Document new import paths

3. **Task Logs**
   - Create log file: `logs/YYYY-MM-DD-phase5a-task-name.md`
   - Document what was changed
   - Document why the change was made
   - Document any lessons learned
   - Document any challenges overcome

4. **Code Examples**
   - Show how to use new structure
   - Show how imports changed
   - Show how tests should be written

---

## Success Criteria (Per Task)

### 5A.1: Split tools/file
- ✅ All files <400 LOC
- ✅ `make check` passes
- ✅ Tool registration works
- ✅ File tests pass
- ✅ No circular imports

### 5A.2: Split openai_adapter
- ✅ openai_adapter.go <300 LOC
- ✅ Client/stream/error files created
- ✅ `make check` passes
- ✅ Model tests pass
- ✅ CreateOpenAIModel API unchanged

### 5A.3: Split persistence
- ✅ models.go <300 LOC (definitions only)
- ✅ sqlite.go <400 LOC
- ✅ schema.go created
- ✅ service.go created
- ✅ `make check` passes
- ✅ Database operations work

### 5A.4: Reorganize display
- ✅ tool_renderer.go <300 LOC
- ✅ tool_renderers/ subpackage created
- ✅ Specific renderers created
- ✅ `make check` passes
- ✅ Display tests pass

### 5B.1: Split REPL
- ✅ repl.go <250 LOC
- ✅ Commands/formatter/input files created
- ✅ `make check` passes
- ✅ REPL works correctly
- ✅ Commands route correctly

### 5C (All): Interface formalization
- ✅ Interfaces documented
- ✅ Contracts defined
- ✅ No breaking changes
- ✅ `make check` passes

---

## Risk Mitigation

### Risk: Circular Import Introduction

**Likelihood**: MEDIUM (when splitting files)  
**Severity**: HIGH (blocks building)  
**Mitigation**:
- Before each split, document import graph
- After split, verify `go build` succeeds
- Use `go mod graph | grep -E "X→Y.*Y→X"` to detect cycles

### Risk: Tool Registration Failure

**Likelihood**: MEDIUM (when moving init functions)  
**Severity**: HIGH (tools unavailable)  
**Mitigation**:
- Include init() functions in each new tool file
- Test registration explicitly: `tools.GetRegistry().GetAllTools()`
- Verify tool count matches before/after

### Risk: Test Failure

**Likelihood**: LOW (comprehensive test suite)  
**Severity**: HIGH (regressions)  
**Mitigation**:
- Run `make check` after every change
- Review failing tests carefully
- May need minor test adjustments (file paths, etc.)

### Risk: Breaking API Changes

**Likelihood**: LOW (using re-exports in tools.go)  
**Severity**: CRITICAL (external breakage)  
**Mitigation**:
- Keep public API in tools.go facade
- Re-export all types and functions
- Test that existing imports still work

---

## Pragmatic Trade-offs

### Will Do
✅ Split files >500 LOC into <400 LOC modules  
✅ Extract concerns into separate files  
✅ Formalize interface contracts  
✅ Update documentation

### Won't Do
❌ Reorganize every small file (over-modularization)  
❌ Create base classes/interfaces for everything  
❌ Rename things for consistency (breaking changes)  
❌ Reorganize working tests unnecessarily  

### Sweet Spot Achieved
- Files: 150-400 LOC (readable, testable)
- Packages: Clear concern separation
- Interfaces: Explicit where needed
- Tests: 250+ maintained/improved
- Regressions: 0%

---

## Success Metrics (Final)

| Metric | Current | Target | Result |
|--------|---------|--------|--------|
| Max File Size | 716 LOC | <400 LOC | 🎯 |
| Avg File Size | 132 LOC | <100 LOC | 🎯 |
| Test Execution | <3s | <3s | ✅ |
| Test Count | 250+ | 260+ | 🎯 |
| Regressions | 0 | 0 | ✅ |
| Code Coverage | Good | Good+ | 🎯 |

---

## Approval & Sign-off

**Plan Created By**: AI Coding Agent  
**Date**: November 12, 2025  
**Review Status**: ⏳ Awaiting Review  

**Reviewers**:
- [ ] Technical Lead
- [ ] Code Quality Lead
- [ ] Test Lead

**Approval**:
- [ ] Approved for Phase 5A
- [ ] Approved for Phase 5B
- [ ] Approved for Phase 5C

---

## Appendix: File Size Target Summary

### Phase 5A Target Results

```
Current Files (>500 LOC):
- openai_adapter.go:  716 LOC → 300 LOC (split into 3 files)
- persistence/models.go: 627 LOC → 150 LOC (moved to schema/service)
- persistence/sqlite.go: 570 LOC → 300 LOC (split concerns)
- tools/file/file_tools.go: 562 LOC → 250 LOC (split into 4 files)
- commands/repl.go: 448 LOC → 200 LOC (split into 3 files)
- tool_renderer.go: 425 LOC → 150 LOC (split into subpackage)

After Phase 5A:
- Max file size: 300 LOC
- All files <400 LOC
- Each file has single responsibility
- Tests remain comprehensive
```

---

**Status**: ✅ Plan Complete - Ready for Execution

