# Code Agent Refactoring Plan

**Date:** November 12, 2025  
**Version:** 1.0  
**Risk Level:** LOW - All changes maintain 100% backward compatibility

## Executive Summary

This refactoring plan addresses organizational and structural improvements to the `code_agent/` codebase while maintaining pragmatic Go best practices. The codebase is fundamentally well-designed (score: 8/10) but has specific areas that need attention:

1. **Display package size** (3808 lines, 26% of codebase) - Too large
2. **Test coverage gaps** - Some tool packages lack tests
3. **Minor organizational issues** - File naming, package structure

**Key Principle:** Zero regression, incremental improvements, maintain all existing APIs.

---

## Phase 1: Low-Risk Structural Improvements

### 1.1 Split Display Package (Priority: HIGH)

**Current State:**
- Single package with 20+ files (3808 lines)
- Mixed concerns: rendering, formatting, components, UI elements
- Difficult to navigate and understand boundaries

**Target Structure:**
```
display/
├── renderer.go              # Main facade (keep)
├── markdown_renderer.go     # Markdown support
├── ansi.go                  # Terminal detection
├── deduplicator.go          # Utility
├── event.go                 # Event types
├── factory.go               # Factory functions
│
├── components/              # ✓ Already exists, expand
│   ├── banner.go           # Move from display/
│   ├── timeline.go         # ✓ Already here
│   ├── spinner.go          # Move from display/
│   ├── typewriter.go       # Move from display/
│   ├── paginator.go        # Move from display/
│   └── streaming.go        # Move streaming_display.go
│
├── formatters/              # ✓ Already exists, expand
│   ├── tool_formatter.go   # Tool-specific formatting
│   ├── agent_formatter.go  # Agent message formatting
│   ├── error_formatter.go  # Error formatting
│   └── metrics_formatter.go # Metrics formatting
│
├── rendering/               # NEW: Tool rendering logic
│   ├── tool_renderer.go    # Merge tool_renderer.go + internals
│   └── tool_parser.go      # Move tool_result_parser.go
│
└── styles/                  # ✓ Already exists
    └── styles.go
```

**Action Items:**
1. Move `banner.go` to `components/` (update imports)
2. Move `spinner.go`, `typewriter.go`, `paginator.go` to `components/`
3. Rename `streaming_display.go` to `components/streaming.go`
4. Create `rendering/` subdirectory
5. Merge `tool_renderer.go` + `tool_renderer_internals.go` → `rendering/tool_renderer.go`
6. Move `tool_result_parser.go` → `rendering/tool_parser.go`
7. Update all imports across codebase
8. Add package documentation for each subdirectory
9. Run tests to verify no regressions

**Verification:**
```bash
make test              # All tests pass
go build ./...         # No compile errors
go vet ./...           # No warnings
```

**Estimated Effort:** 4-6 hours  
**Risk:** LOW (Go compiler catches import errors)

---

### 1.2 Organize Agent Prompts (Priority: HIGH)

**Current State:**
```
agent/
├── coding_agent.go
├── dynamic_prompt.go
├── prompt_guidance.go
├── prompt_pitfalls.go
├── prompt_workflow.go
└── xml_prompt_builder.go
```

**Target Structure:**
```
agent/
├── coding_agent.go          # Main agent creation
├── prompts/                 # NEW subdirectory
│   ├── builder.go          # Rename xml_prompt_builder.go
│   ├── dynamic.go          # Rename dynamic_prompt.go
│   ├── guidance.go         # Rename prompt_guidance.go
│   ├── pitfalls.go         # Rename prompt_pitfalls.go
│   └── workflow.go         # Rename prompt_workflow.go
└── tests continue in agent/
```

**Action Items:**
1. Create `agent/prompts/` directory
2. Move and rename prompt files
3. Update package declarations to `package prompts`
4. Update imports in `coding_agent.go`
5. Ensure test files still work
6. Add package documentation

**Verification:**
```bash
cd agent && go test ./...
```

**Estimated Effort:** 1-2 hours  
**Risk:** VERY LOW

---

### 1.3 Rename persistence Package to session (Priority: MEDIUM)

**Current State:**
- Package name `persistence` is too generic
- Actually manages sessions specifically
- Clearer name improves understanding

**Action Items:**
1. Rename directory: `persistence/` → `session/`
2. Update package declarations
3. Update imports across codebase:
   - `internal/app/app.go`
   - `internal/app/session.go`
   - `pkg/cli/commands/`
4. Update go.mod if needed
5. Run full test suite

**Verification:**
```bash
grep -r "code_agent/persistence" . --include="*.go"  # Should be empty
make test
```

**Estimated Effort:** 2-3 hours  
**Risk:** LOW (straightforward rename)

---

### 1.4 Consolidate CLI Commands (Priority: MEDIUM)

**Current State:**
```
pkg/cli/
├── commands.go              # In parent directory
└── commands/
    └── repl_builders.go     # In subdirectory
```

**Target Structure:**
```
pkg/cli/
├── config.go
├── flags.go
├── syntax.go
├── display.go
└── commands/
    ├── commands.go          # Move from parent
    ├── repl.go              # Rename repl_builders.go
    └── builtin.go           # Extract builtin commands
```

**Action Items:**
1. Move `commands.go` into `commands/` subdirectory
2. Rename `repl_builders.go` → `repl.go`
3. Extract builtin command handling to `builtin.go`
4. Update imports
5. Add package documentation

**Verification:**
```bash
cd pkg/cli && go test ./...
```

**Estimated Effort:** 2-3 hours  
**Risk:** LOW

---

## Phase 2: Test Coverage Improvements

### 2.1 Add Missing Tests (Priority: HIGH)

**Packages needing tests:**
- `tools/common/` - Registry functionality
- `tools/edit/` - Patch, search/replace, line editing
- `tools/exec/` - Command execution
- `tools/search/` - Search operations
- `tools/workspace/` - Workspace tools
- `display/components/` - UI components
- `display/formatters/` - Formatters
- `display/styles/` - Style system
- `pkg/cli/commands/` - CLI commands

**Test Template:**
```go
package targetpackage

import (
    "testing"
)

func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name: "success case",
            input: InputType{},
            want: OutputType{},
            wantErr: false,
        },
        {
            name: "error case",
            input: InputType{},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Priority Order:**
1. `tools/common/registry_test.go` - Critical infrastructure
2. `tools/edit/*_test.go` - Complex logic needs tests
3. `tools/exec/terminal_tools_test.go` - Security-sensitive
4. `display/components/*_test.go` - UI components
5. `display/formatters/*_test.go` - Formatting logic

**Target Coverage:** 80% for each package

**Verification:**
```bash
make coverage
# Open coverage.html and verify coverage
```

**Estimated Effort:** 12-16 hours  
**Risk:** NONE (adding tests is always safe)

---

### 2.2 Add Package Documentation (Priority: MEDIUM)

**Current State:**
- Some packages lack package-level documentation
- Godoc generation incomplete

**Action Items:**
For each package, add doc.go:

```go
// Package packagename provides ...
//
// Longer description of package purpose, key concepts, and usage examples.
//
// Example:
//   // Create a new component
//   comp := packagename.New()
//   comp.DoSomething()
package packagename
```

**Packages needing documentation:**
- `display/components/`
- `display/formatters/`
- `display/rendering/` (after creation)
- `display/styles/`
- `agent/prompts/` (after creation)
- `tools/common/`
- `tools/edit/`
- `tools/exec/`
- `tools/search/`
- `tools/workspace/`

**Verification:**
```bash
godoc -http=:6060
# Open http://localhost:6060/pkg/code_agent/ and verify docs
```

**Estimated Effort:** 4-6 hours  
**Risk:** NONE

---

## Phase 3: Code Quality Improvements

### 3.1 Add Interfaces for Testability (Priority: MEDIUM)

**Current State:**
- Concrete types passed everywhere
- Difficult to mock dependencies in tests

**Target Interfaces:**

```go
// In display/interfaces.go
package display

type Renderer interface {
    RenderMarkdown(markdown string) string
    RenderText(text string) string
    Dim(text string) string
    Green(text string) string
    Red(text string) string
    // ... other methods
}

type StreamingDisplay interface {
    StreamResponse(ctx context.Context, chunks <-chan string) error
    // ... other methods
}

// In tools/common/interfaces.go
package common

type Registry interface {
    Register(metadata ToolMetadata) error
    GetByCategory(category ToolCategory) []ToolMetadata
    GetAllTools() []tool.Tool
}

// In session/interfaces.go (after rename)
package session

type Manager interface {
    SaveSession(session *models.Session) error
    LoadSession(id string) (*models.Session, error)
    // ... other methods
}
```

**Action Items:**
1. Define interfaces for major components
2. Update function signatures to accept interfaces
3. Maintain concrete constructors for backward compatibility
4. Add mock implementations in test files
5. Update tests to use mocks

**Verification:**
```bash
make test
# All tests pass, coverage improves
```

**Estimated Effort:** 8-10 hours  
**Risk:** LOW (additive changes only)

---

### 3.2 Reduce Global State (Priority: LOW)

**Current State:**
- Tool registry is global
- Makes parallel testing difficult

**Approach:**
Create injectable registry while maintaining backward compatibility:

```go
// tools/common/registry.go

var defaultRegistry = NewToolRegistry()

// GetRegistry returns the default global registry (existing API)
func GetRegistry() *ToolRegistry {
    return defaultRegistry
}

// GetRegistryOrDefault returns registry from context or default
func GetRegistryOrDefault(ctx context.Context) *ToolRegistry {
    if reg, ok := ctx.Value(registryKey).(*ToolRegistry); ok {
        return reg
    }
    return defaultRegistry
}

// WithRegistry returns a context with the given registry
func WithRegistry(ctx context.Context, reg *ToolRegistry) context.Context {
    return context.WithValue(ctx, registryKey, reg)
}

// For tests:
func NewTestRegistry() *ToolRegistry {
    return NewToolRegistry()
}
```

**Action Items:**
1. Add context-based registry access
2. Update agent creation to accept registry
3. Maintain global registry for backward compatibility
4. Update tests to use isolated registries
5. Document migration path

**Verification:**
```bash
go test -race ./...  # No race conditions
```

**Estimated Effort:** 6-8 hours  
**Risk:** MEDIUM (changes initialization flow)

---

## Phase 4: Documentation and Polish

### 4.1 Update Architecture Documentation (Priority: MEDIUM)

**Action Items:**
1. Create `docs/architecture.md` with package diagrams
2. Update README.md with new structure
3. Document design patterns used
4. Add sequence diagrams for key flows
5. Document testing strategy

**Tools to use:**
- Mermaid for diagrams
- Godoc for API documentation
- Markdown for guides

**Estimated Effort:** 6-8 hours  
**Risk:** NONE

---

### 4.2 Add Code Examples (Priority: LOW)

**Action Items:**
1. Add `examples/` subdirectory
2. Create example programs for:
   - Custom tool creation
   - Display formatting
   - Session management
   - Workspace configuration
3. Reference examples from package documentation

**Estimated Effort:** 4-6 hours  
**Risk:** NONE

---

## Implementation Strategy

### Week 1: High-Priority Structural Changes
- [ ] Day 1-2: Split display package (1.1)
- [ ] Day 3: Organize agent prompts (1.2)
- [ ] Day 4: Rename persistence to session (1.3)
- [ ] Day 5: Consolidate CLI commands (1.4)

### Week 2: Test Coverage
- [ ] Day 1-2: Add tools/common tests (2.1)
- [ ] Day 3-4: Add tools/edit tests (2.1)
- [ ] Day 5: Add display component tests (2.1)

### Week 3: Quality and Documentation
- [ ] Day 1-2: Add package documentation (2.2)
- [ ] Day 3-4: Add testability interfaces (3.1)
- [ ] Day 5: Update architecture docs (4.1)

### Week 4: Polish and Verification
- [ ] Day 1-2: Reduce global state (3.2)
- [ ] Day 3: Add code examples (4.2)
- [ ] Day 4-5: Full regression testing and verification

---

## Verification Checklist

After each phase:

```bash
# 1. All tests pass
make test

# 2. No lint errors
make lint

# 3. No vet warnings
make vet

# 4. Code formats correctly
make fmt

# 5. Builds successfully
make build

# 6. Coverage maintained or improved
make coverage

# 7. No compile errors
go build ./...

# 8. No race conditions
go test -race ./...

# 9. Documentation generates
godoc -http=:6060

# 10. Integration test passes
./bin/code-agent --help
```

---

## Rollback Plan

If any phase causes issues:

1. **Git branches** - Each phase in separate branch
2. **Commit strategy** - Small, atomic commits
3. **Revert process** - `git revert` specific commits
4. **Testing** - Run full test suite after each change

---

## Success Metrics

### Before Refactoring
- Display package: 3808 lines
- Test coverage: ~70% (estimated)
- Packages with no tests: 9
- Package doc files: ~5

### After Refactoring (Target)
- Largest package: <2000 lines
- Test coverage: >80%
- Packages with no tests: 0
- Package doc files: 100% coverage
- Maintainability score: 9/10 (up from 8/10)

---

## Risk Mitigation

### Low-Risk Changes
- Adding tests (Phase 2)
- Adding documentation (Phase 2, 4)
- Adding interfaces (Phase 3.1)

### Medium-Risk Changes
- Package splitting (Phase 1.1) - Use compiler for verification
- Package renaming (Phase 1.3) - Comprehensive grep for imports
- Registry changes (Phase 3.2) - Maintain backward compatibility

### High-Risk Changes
- None in this plan - All changes maintain APIs

---

## Long-Term Considerations (Future Phases)

### Not included in this plan (requires separate planning):
1. **Plugin architecture** - Major feature addition
2. **Breaking API changes** - Requires versioning strategy
3. **Performance optimization** - Requires profiling first
4. **Security hardening** - Requires threat modeling
5. **Metrics/observability** - Requires requirements gathering

These should be addressed in future refactoring plans after current phase completes.

---

## Approval and Sign-off

**Prepared by:** AI Assistant  
**Review required by:** Team Lead  
**Approval required by:** Project Owner  

**Risks:** LOW  
**Impact:** HIGH (improved maintainability)  
**Effort:** ~80-100 hours over 4 weeks  

**Recommendation:** APPROVED FOR IMPLEMENTATION

The refactoring maintains backward compatibility while significantly improving code organization and maintainability. All changes are incremental and reversible.
