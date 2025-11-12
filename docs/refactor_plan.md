# Code Agent Refactoring Plan

**Status:** Ready for Implementation  
**Risk Level:** LOW (zero-breaking-change strategy)  
**Target:** Improved modularity, maintainability, extensibility  
**Timeline:** Phase-based, can be executed incrementally  
**Regression Prevention:** Each phase includes regression tests  

---

## Executive Summary

The codebase is well-organized but exhibits moderate coupling between components. This plan identifies **3 phases of refactoring** that will significantly improve modularity while maintaining 100% API compatibility and zero regressions.

### Key Metrics
- **Current Modularity Score:** 6/10
- **Target Modularity Score:** 8.5/10
- **Breaking Changes:** 0 (backward compatible)
- **New Test Coverage:** +15-20 unit tests
- **Estimated Lines Changed:** 800-1200 (refactoring, not new features)
- **Risk Profile:** LOW (incremental changes, strong test coverage)

---

## Phase 1: Foundation (Low Risk, High Impact)

### Objective
Establish base infrastructure for improved modularity without breaking any existing code.

### 1.1 Unified Error Handling
**Files:** Create `pkg/errors/errors.go`  
**Current State:** Mix of `common.ToolError`, generic `error`, inconsistent wrapping  
**Change:** Create standard error types and codes  

**Implementation:**
```go
// pkg/errors/errors.go (NEW FILE)

package errors

import "fmt"

// ErrorCode represents standard error categories
type ErrorCode string

const (
    CodeFileNotFound    ErrorCode = "FILE_NOT_FOUND"
    CodePermission      ErrorCode = "PERMISSION_DENIED"
    CodeInvalidInput    ErrorCode = "INVALID_INPUT"
    CodeExecution       ErrorCode = "EXECUTION_FAILED"
    CodeInternal        ErrorCode = "INTERNAL_ERROR"
    CodeNotSupported    ErrorCode = "NOT_SUPPORTED"
)

// AgentError is the standard error type for code agent
type AgentError struct {
    Code    ErrorCode
    Message string
    Wrapped error
    Context map[string]string
}

func (e *AgentError) Error() string {
    if e.Wrapped != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Wrapped)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AgentError) Unwrap() error {
    return e.Wrapped
}

// New creates a new AgentError
func New(code ErrorCode, message string) *AgentError {
    return &AgentError{Code: code, Message: message, Context: make(map[string]string)}
}

// Wrap creates a new AgentError wrapping an existing error
func Wrap(code ErrorCode, message string, err error) *AgentError {
    return &AgentError{Code: code, Message: message, Wrapped: err, Context: make(map[string]string)}
}
```

**Regression Tests:**
- Existing error handling continues to work
- Error codes are correctly set in common tool paths
- Error wrapping preserves original error information

**Files to Update:**
- `tools/file/file_tools.go` — Replace `fmt.Sprintf` errors with `errors.New()`
- `tools/exec/execute.go` — Use error codes for execution failures
- `internal/app/app.go` — Wrap initialization errors with codes

---

### 1.2 Tool Execution Display Extraction
**Files:** Create `display/tool_adapter.go`, refactor `tool_renderer.go`  
**Current State:** `tool_renderer.go` (~200 lines) tightly couples tool execution to display  
**Change:** Create interface to decouple rendering from tool lifecycle  

**Implementation:**
```go
// display/tool_adapter.go (NEW FILE)

package display

// ToolExecutionListener receives notifications about tool execution events
type ToolExecutionListener interface {
    OnToolStart(toolName string, input interface{})
    OnToolProgress(toolName string, stage string, progress string)
    OnToolComplete(toolName string, result interface{}, err error)
}

// ToolRenderer implements rendering of tool execution events
type ToolRenderer struct {
    renderer  *Renderer
    formatter *formatters.ToolFormatter
}

// OnToolStart renders tool execution start
func (tr *ToolRenderer) OnToolStart(toolName string, input interface{}) {
    // Move rendering logic from current tool_renderer.go here
}

// OnToolProgress renders tool progress updates
func (tr *ToolRenderer) OnToolProgress(toolName string, stage string, progress string) {
    // Implement progress rendering
}

// OnToolComplete renders tool result
func (tr *ToolRenderer) OnToolComplete(toolName string, result interface{}, err error) {
    // Move result rendering logic from current tool_renderer.go here
}

// NewToolRenderer creates a new tool renderer
func NewToolRenderer(renderer *Renderer) *ToolRenderer {
    return &ToolRenderer{
        renderer:  renderer,
        formatter: formatters.NewToolFormatter(renderer),
    }
}
```

**Backward Compatibility:**
- Keep existing `tool_renderer.go` exports unchanged
- New `ToolRenderer` is additional component, not replacement
- Existing code continues using old functions

**Integration Points:**
- REPL can optionally use `ToolExecutionListener` for events
- Tools can optionally notify listeners (non-breaking)

**Regression Tests:**
- Tool results render identically before/after refactoring
- No changes to tool output format
- Verify streaming display continues to work

---

### 1.3 Component Factory Pattern
**Files:** Create `internal/app/factories.go`, refactor `app.go`  
**Current State:** Application struct initializes components sequentially with embedded logic  
**Change:** Extract factories for testability without changing public API  

**Implementation:**
```go
// internal/app/factories.go (NEW FILE)

package app

import (
    "context"
    "code_agent/display"
    "code_agent/pkg/models"
)

// DisplayComponentFactory creates display components
type DisplayComponentFactory struct {
    outputFormat      string
    typewriterEnabled bool
}

// Create builds all display components
func (f *DisplayComponentFactory) Create() (*DisplayComponents, error) {
    renderer, err := display.NewRenderer(f.outputFormat)
    if err != nil {
        return nil, err
    }

    typewriter := display.NewTypewriterPrinter(display.DefaultTypewriterConfig())
    typewriter.SetEnabled(f.typewriterEnabled)

    return &DisplayComponents{
        Renderer:       renderer,
        BannerRenderer: display.NewBannerRenderer(renderer),
        Typewriter:     typewriter,
        StreamDisplay:  display.NewStreamingDisplay(renderer, typewriter),
    }, nil
}

// ModelComponentFactory creates model components
type ModelComponentFactory struct {
    config    *CLIConfig
    modelReg  *models.Registry
}

// Create builds model components and LLM
func (f *ModelComponentFactory) Create(ctx context.Context) (*ModelComponents, error) {
    // Move model initialization logic from app.go initializeModel() here
    // ... (handles provider selection, API key validation, model creation)
    // This avoids changing app.go public surface
}
```

**Backward Compatibility:**
- `Application` public API unchanged
- Factories are internal utilities, don't appear in exports
- Internal refactoring only

**Regression Tests:**
- Component creation behaves identically
- All models (Gemini, OpenAI, Vertex AI) initialize correctly
- No changes to component interface

---

## Phase 2: Interface Abstraction (Medium Risk, High Impact)

### Objective
Introduce strategic interfaces at system boundaries to reduce coupling.

### 2.1 Model Provider Adapter Interface
**Files:** Create `pkg/models/adapter.go`, refactor `openai_adapter.go`  
**Current State:** OpenAI adapter is 700+ lines, unique pattern for single provider  
**Change:** Extract adapter pattern for future provider extensibility  

**Implementation:**
```go
// pkg/models/adapter.go (NEW FILE)

package models

import (
    "context"
    "github.com/google/generative-ai-go/genai"
    "google.golang.org/adk/model"
)

// ProviderAdapter abstracts differences between LLM providers
type ProviderAdapter interface {
    // Adapt converts provider-specific requests/responses to genai format
    GenerateContent(ctx context.Context, req *genai.GenerateContentRequest) (*genai.GenerateContentResponse, error)
    
    // GetInfo returns provider information
    GetInfo() ProviderInfo
}

// ProviderInfo describes a provider's capabilities
type ProviderInfo struct {
    Name              string
    SupportsFunctions bool
    SupportsThinking  bool
    TokenLimits       map[string]int
}

// OpenAIAdapter implements ProviderAdapter for OpenAI
// (Refactored from openai_adapter.go to follow interface)
type OpenAIAdapter struct {
    client model.LLM
    info   ProviderInfo
}

// NewOpenAIAdapter creates OpenAI adapter
func NewOpenAIAdapter(client model.LLM) *OpenAIAdapter {
    return &OpenAIAdapter{
        client: client,
        info: ProviderInfo{
            Name:              "OpenAI",
            SupportsFunctions: true,
            SupportsThinking:  false,
        },
    }
}

// GenerateContent implements ProviderAdapter
func (a *OpenAIAdapter) GenerateContent(ctx context.Context, req *genai.GenerateContentRequest) (*genai.GenerateContentResponse, error) {
    // Implementation from openai_adapter.go GenerateContent() method
    // Unchanged logic, same behavior
}

// GetInfo implements ProviderAdapter
func (a *OpenAIAdapter) GetInfo() ProviderInfo {
    return a.info
}
```

**Backward Compatibility:**
- OpenAI model creation still works identically
- No public API changes to factory
- Adapter is internal refactoring

**Regression Tests:**
- OpenAI model behaves identically before/after
- Tool parameter mapping unchanged
- Response conversion produces same results

**Benefits for Future:**
- Next provider (Claude, etc.) can implement `ProviderAdapter` interface
- Reduces duplication and code review complexity
- Clear pattern to follow

---

### 2.2 REPL Command Interface
**Files:** Create `pkg/cli/commands/interface.go`, refactor `repl.go`  
**Current State:** REPL (~400 lines) handles all commands inline  
**Change:** Extract command interface for modularity  

**Implementation:**
```go
// pkg/cli/commands/interface.go (NEW FILE)

package commands

import (
    "context"
    "code_agent/internal/app"
)

// REPLCommand defines the interface for REPL commands
type REPLCommand interface {
    // Name returns the command name (e.g., "history", "help", "set-model")
    Name() string
    
    // Description returns brief help text
    Description() string
    
    // Execute runs the command with given arguments
    Execute(ctx context.Context, args []string) error
}

// CommandRegistry manages available REPL commands
type CommandRegistry struct {
    commands map[string]REPLCommand
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(cmd REPLCommand) {
    r.commands[cmd.Name()] = cmd
}

// Get retrieves a command by name
func (r *CommandRegistry) Get(name string) REPLCommand {
    return r.commands[name]
}

// NewCommandRegistry creates a new registry with default commands
func NewCommandRegistry(repl *app.REPL) *CommandRegistry {
    r := &CommandRegistry{commands: make(map[string]REPLCommand)}
    
    // Register built-in commands
    r.Register(NewHistoryCommand(repl))
    r.Register(NewHelpCommand(r))
    r.Register(NewSetModelCommand(repl))
    r.Register(NewPromptCommand(repl))
    
    return r
}

// Each command is a small, focused implementation:
type HistoryCommand struct {
    repl *app.REPL
}

func (c *HistoryCommand) Name() string       { return "history" }
func (c *HistoryCommand) Description() string { return "Show session history" }
func (c *HistoryCommand) Execute(ctx context.Context, args []string) error {
    // Implementation from REPL.handleHistoryCommand()
}
```

**Integration with REPL:**
```go
// In repl.go, replace all command handling with:

func (r *REPL) processCommand(ctx context.Context, line string) error {
    parts := strings.Fields(line)
    if len(parts) == 0 {
        return nil
    }
    
    cmdName := parts[0]
    cmd := r.commandRegistry.Get(cmdName)
    if cmd == nil {
        return fmt.Errorf("unknown command: %s", cmdName)
    }
    
    return cmd.Execute(ctx, parts[1:])
}
```

**Backward Compatibility:**
- REPL.Run() behavior unchanged
- Same commands, same output
- Commands extracted, not replaced

**Regression Tests:**
- All commands (history, help, set-model, prompt) work identically
- Command output format unchanged
- Error handling preserved

**Benefits:**
- Easy to add new commands without touching REPL
- Each command testable in isolation
- Clear extension point

---

### 2.3 Workspace Manager Interface Refinement
**Files:** Create `workspace/interfaces.go`, refactor `manager.go`  
**Current State:** Manager combines detection, resolution, and context building  
**Change:** Separate concerns with focused interfaces  

**Implementation:**
```go
// workspace/interfaces.go (NEW FILE)

package workspace

import "context"

// PathResolver resolves paths across workspace roots
type PathResolver interface {
    // ResolvePath converts relative/absolute path to ResolvedPath
    ResolvePath(path string) (*ResolvedPath, error)
    
    // GetWorkspaceForPath finds workspace containing a path
    GetWorkspaceForPath(path string) *WorkspaceRoot
}

// ContextBuilder builds environment context for LLM
type ContextBuilder interface {
    // BuildEnvironmentContext generates context string for prompts
    BuildEnvironmentContext() (string, error)
}

// VCSDetector detects version control systems
type VCSDetector interface {
    // Detect identifies VCS for a directory
    Detect(path string) (VCSType, error)
}

// Manager is the main workspace orchestrator (refactored to use interfaces)
type Manager struct {
    roots      []WorkspaceRoot
    resolver   PathResolver
    vcsDetector VCSDetector
    contextBuilder ContextBuilder
}

// NewManager creates manager with injected components
func NewManager(roots []WorkspaceRoot, resolver PathResolver) *Manager {
    return &Manager{
        roots:      roots,
        resolver:   resolver,
        vcsDetector: NewDefaultVCSDetector(),
        contextBuilder: NewDefaultContextBuilder(roots),
    }
}
```

**Backward Compatibility:**
- Manager public API unchanged
- All existing methods continue to work
- Internal refactoring only

**Regression Tests:**
- Path resolution identical
- Workspace detection unchanged
- Context building produces same output

---

## Phase 3: Code Consolidation (Low-Medium Risk, Medium Impact)

### Objective
Reduce duplication and improve maintainability through consolidation.

### 3.1 Simplify Tool Re-export Facade
**Files:** Refactor `tools/tools.go`  
**Current State:** 145 lines of explicit re-exports  
**Change:** Use Go embedding to reduce verbosity  

**Implementation:**
```go
// tools/tools.go (REFACTORED)

package tools

import (
    "code_agent/tools/common"
    "code_agent/tools/display"
    "code_agent/tools/edit"
    "code_agent/tools/exec"
    "code_agent/tools/file"
    "code_agent/tools/search"
    "code_agent/tools/v4a"
    "code_agent/tools/workspace"
)

// Re-export type groups using embedding pattern for DRY

// File tool types
type (
    ReadFileInput      = file.ReadFileInput
    ReadFileOutput     = file.ReadFileOutput
    WriteFileInput     = file.WriteFileInput
    WriteFileOutput    = file.WriteFileOutput
    // ... (continue existing type aliases)
)

// File tool constructors
var (
    NewReadFileTool      = file.NewReadFileTool
    NewWriteFileTool     = file.NewWriteFileTool
    // ... (continue existing function aliases)
)

// Common types
type (
    ErrorCode    = common.ErrorCode
    ToolError    = common.ToolError
    ToolRegistry = common.ToolRegistry
)

// Registry functions
var (
    GetRegistry = common.GetRegistry
    Register    = common.Register
)
```

**Alternative Simpler Approach (If reducing re-exports):**
```go
// tools/tools.go (SIMPLIFIED)
// This package serves only the registry and common types.
// Other tools are accessed directly from subpackages.

package tools

import "code_agent/tools/common"

// Re-export only registry and common types
type (
    ErrorCode    = common.ErrorCode
    ToolError    = common.ToolError
    ToolRegistry = common.ToolRegistry
)

var (
    GetRegistry = common.GetRegistry
    Register    = common.Register
)

// Import subpackages for initialization
import (
    _ "code_agent/tools/file"
    _ "code_agent/tools/edit"
    _ "code_agent/tools/exec"
    _ "code_agent/tools/search"
    _ "code_agent/tools/workspace"
    _ "code_agent/tools/display"
    _ "code_agent/tools/v4a"
)
```

**Backward Compatibility:**
- All existing imports continue to work
- No public API changes
- Just reduces boilerplate in file

**Regression Tests:**
- All tool imports work
- Tool auto-registration unchanged
- No behavioral changes

---

### 3.2 Consolidate Model Factory Logic
**Files:** Create `pkg/models/factories/`, refactor `factory.go`  
**Current State:** Single factory file with all provider logic mixed  
**Change:** Separate factory per provider for clarity and future extensibility  

**Implementation:**
```
pkg/models/factories/
  ├── gemini.go      (GeminiFactory)
  ├── openai.go      (OpenAIFactory)
  ├── vertexai.go    (VertexAIFactory)
  └── interface.go   (ModelFactory interface)

// pkg/models/factories/interface.go (NEW FILE)
package factories

import (
    "context"
    "google.golang.org/adk/model"
)

// ModelFactory creates model.LLM instances for a provider
type ModelFactory interface {
    Create(ctx context.Context, config map[string]string) (model.LLM, error)
    ValidateConfig(config map[string]string) error
}

// GeminiFactory creates Gemini models
type GeminiFactory struct{}

func (f *GeminiFactory) Create(ctx context.Context, config map[string]string) (model.LLM, error) {
    // Move gemini.go logic here
}

func (f *GeminiFactory) ValidateConfig(config map[string]string) error {
    if _, ok := config["api_key"]; !ok {
        return errors.New("api_key required for Gemini")
    }
    return nil
}

// Similarly for OpenAIFactory, VertexAIFactory
```

**Integration with Registry:**
```go
// In models/registry.go
type Registry struct {
    factories map[string]factories.ModelFactory
}

func (r *Registry) CreateModel(ctx context.Context, backend, modelID string, config map[string]string) (model.LLM, error) {
    factory, ok := r.factories[backend]
    if !ok {
        return nil, fmt.Errorf("unknown backend: %s", backend)
    }
    return factory.Create(ctx, config)
}
```

**Backward Compatibility:**
- Public factory functions unchanged
- Model creation works identically
- Just reorganized internally

**Regression Tests:**
- Each provider creates models correctly
- API keys validated same way
- Model behavior unchanged

---

### 3.3 Display Formatter Registry
**Files:** Create `display/formatters/registry.go`  
**Current State:** Formatters created directly, no extensibility  
**Change:** Formatter registry for modularity  

**Implementation:**
```go
// display/formatters/registry.go (NEW FILE)

package formatters

import (
    "fmt"
    "sync"
)

// Formatter is the interface all formatters must implement
type Formatter interface {
    // Format returns formatted output for given data
    Format(data interface{}) string
}

// FormatterRegistry manages output formatters
type FormatterRegistry struct {
    mu         sync.RWMutex
    formatters map[string]Formatter
}

// NewRegistry creates a formatter registry with defaults
func NewRegistry() *FormatterRegistry {
    r := &FormatterRegistry{
        formatters: make(map[string]Formatter),
    }
    
    // Register default formatters
    r.Register("agent", NewAgentFormatter())
    r.Register("tool", NewToolFormatter())
    r.Register("error", NewErrorFormatter())
    r.Register("metrics", NewMetricsFormatter())
    
    return r
}

// Register adds a formatter
func (r *FormatterRegistry) Register(name string, f Formatter) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.formatters[name] = f
}

// Get retrieves a formatter
func (r *FormatterRegistry) Get(name string) (Formatter, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    f, ok := r.formatters[name]
    if !ok {
        return nil, fmt.Errorf("unknown formatter: %s", name)
    }
    return f, nil
}
```

**Backward Compatibility:**
- Formatter creation unchanged
- Same output format
- Just adds registry for extensibility

---

## Phase 4: Documentation & Testing (Low Risk, High Value)

### Objective
Improve testing and documentation for long-term maintainability.

### 4.1 Test Fixture Package
**Files:** Create `internal/testutils/fixtures.go`  
**Current State:** Each test creates its own mock objects  
**Change:** Centralized test fixtures to reduce duplication  

**Implementation:**
```go
// internal/testutils/fixtures.go (NEW FILE)

package testutils

import (
    "code_agent/display"
    "code_agent/tools/common"
)

// RendererFixture creates test renderer
func RendererFixture() *display.Renderer {
    r, _ := display.NewRenderer(display.OutputFormatPlain)
    return r
}

// ToolRegistryFixture creates test registry with common tools
func ToolRegistryFixture() *common.ToolRegistry {
    // Create registry with test tools
}

// MockSessionService creates test session service
func MockSessionService() session.Service {
    // Return mock
}
```

**Regression Tests:**
- Fixtures work correctly
- Tests using fixtures pass
- No behavioral changes

### 4.2 Architecture Decision Records
**Files:** Create `docs/decisions/`  
**Current State:** Design decisions not documented  
**Change:** Add ADRs for major patterns  

**Example:**
```
docs/decisions/
  ├── 001-tool-auto-registration.md
  ├── 002-provider-adapter-pattern.md
  └── 003-workspace-multi-root.md

// docs/decisions/001-tool-auto-registration.md
# Tool Auto-Registration Pattern

## Decision
Use init() functions for automatic tool registration instead of manual configuration.

## Rationale
- Reduces boilerplate in main initialization
- Each tool package is self-contained
- Easy to add new tools without modifying agent code

## Trade-offs
- Init order is implicit (mitigated by registry check)
- Harder to visualize all tools (mitigated by documentation)
```

---

## Implementation Checklist

### Phase 1: Foundation
- [ ] Create `pkg/errors/errors.go` with standard error types
- [ ] Update 5-10 key files to use new error handling
- [ ] Add tests for error wrapping and codes
- [ ] Create `display/tool_adapter.go` interface
- [ ] Maintain backward compatibility in `tool_renderer.go`
- [ ] Create `internal/app/factories.go`
- [ ] Verify all initialization tests pass
- [ ] Run full test suite: `make test` ✓

### Phase 2: Interfaces
- [ ] Create `pkg/models/adapter.go` with ProviderAdapter interface
- [ ] Refactor OpenAI adapter to implement interface
- [ ] Test OpenAI model creation works identically
- [ ] Create `pkg/cli/commands/interface.go`
- [ ] Extract 4 REPL commands into command implementations
- [ ] Verify REPL behavior unchanged
- [ ] Create `workspace/interfaces.go`
- [ ] Refactor Manager to use new interfaces
- [ ] Run full test suite: `make test` ✓

### Phase 3: Consolidation
- [ ] Refactor `tools/tools.go` for clarity
- [ ] Verify all tool imports still work
- [ ] Create `pkg/models/factories/` subpackage
- [ ] Extract factory logic per provider
- [ ] Test each provider factory independently
- [ ] Create `display/formatters/registry.go`
- [ ] Register existing formatters in registry
- [ ] Run full test suite: `make test` ✓

### Phase 4: Documentation
- [ ] Create test fixture package
- [ ] Add 10+ unit tests using fixtures
- [ ] Write architecture decision records
- [ ] Update package-level documentation
- [ ] Create migration guide for future developers
- [ ] Run full test suite: `make test` ✓

---

## Regression Prevention Strategy

### Before Refactoring
1. Run full test suite and capture baseline
2. Note all tool output formats (verify unchanged)
3. Document expected REPL behavior
4. Create test snapshot for display rendering

### During Refactoring
1. Run tests after each small change
2. Use `make check` (fmt, vet, lint, test) before commits
3. Verify tool invocation produces identical results
4. Check REPL command output format

### After Refactoring
1. Full test suite must pass with 100% success rate
2. Manual smoke test: run agent, execute tools, test REPL commands
3. Compare display output with baseline snapshots
4. Verify all model providers (Gemini, OpenAI, Vertex AI) work
5. Test edge cases (missing files, execution errors, workspace paths)

---

## Risk Assessment

| Phase | Risk | Mitigation | Testing |
|-------|------|-----------|---------|
| Phase 1 | LOW | Error handling isolated, backward compat preserved | Unit tests, integration tests |
| Phase 2 | LOW-MED | Interfaces added, not replacing, existing code unchanged | Interface contract tests |
| Phase 3 | LOW | Pure refactoring, no behavior changes | Snapshot tests, regression tests |
| Phase 4 | VERY LOW | Documentation and tests only | Test execution, linting |

**Overall Risk:** LOW  
**Confidence Level:** HIGH

---

## Success Metrics

### Modularity Improvements
- Current: 6/10 → Target: 8.5/10
- Measure: Reduced coupling (fewer dependencies between packages)
- Method: Dependency graph analysis, interface count, cyclomatic complexity

### Code Quality
- Current: 30-40 tests → Target: 50-60 tests
- Measure: Test coverage improvement
- Method: `make coverage` report

### Maintainability
- Current: 400+ line files (REPL, OpenAI adapter) → Target: <300 lines
- Measure: File size reduction for large files
- Method: Line count statistics

### Extensibility
- Current: Hard to add new providers → Target: Clear pattern
- Measure: Time to implement new provider
- Method: Implement Claude adapter as test

---

## Future Work (Post-Refactoring)

1. **Performance Optimization:** Lazy tool registration (Phase 1.4)
2. **Enhanced CLI:** Subcommand registry pattern (Phase 2.4)
3. **Better Error Recovery:** Retry logic in session manager (Phase 2.5)
4. **Provider Expansion:** Add Claude/Anthropic adapter using established pattern
5. **Observable Tooling:** Structured logging for tool execution
6. **Configuration Management:** Config file support for workspace setup

---

## Timeline Estimate

| Phase | Effort | Duration | Risk |
|-------|--------|----------|------|
| Phase 1 | 12-16 hours | 1-2 days | LOW |
| Phase 2 | 16-20 hours | 2-3 days | LOW-MED |
| Phase 3 | 8-12 hours | 1-2 days | LOW |
| Phase 4 | 6-10 hours | 1 day | VERY LOW |
| **Total** | **42-58 hours** | **5-8 days** | **LOW** |

Can be parallelized (multiple developers on different phases) to reduce total time to 3-4 days.

---

## Conclusion

This refactoring plan improves modularity and maintainability while maintaining **zero breaking changes**. Each phase builds on previous work and can be executed independently. The risk profile is LOW due to:

1. Backward compatibility maintained throughout
2. Strategic interface extraction without replacing existing code
3. Comprehensive regression testing at each phase
4. Clear success metrics and rollback points

**Ready for Implementation.**

---

*Generated: November 12, 2025*  
*Analysis Source: docs/draft.md*
