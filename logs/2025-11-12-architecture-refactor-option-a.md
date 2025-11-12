# Architecture Refactoring: Option A - Clear Reusable vs Application Split

**Date**: November 12, 2025  
**Status**: ✅ COMPLETE  
**Test Results**: All 16 packages passing, zero regressions

## Overview

Implemented **Option A** architecture refactoring to establish clear boundaries between:
- **`pkg/`** - Truly reusable packages
- **`internal/`** - Application-specific packages

## Changes Made

### 1. Moved `pkg/cli` → `internal/cli`

**Rationale**: `pkg/cli` is app-specific, not a reusable CLI library.

**Evidence**:
- Hardcoded backend names: "gemini", "vertexai"
- Session-specific configuration fields
- API key handling tied to specific providers
- Dependencies on app-specific modules

**Files Updated**:
```
internal/app/factories.go              - Updated import from pkg/cli → internal/cli
internal/app/orchestration/model.go    - Updated import from pkg/cli → internal/cli
internal/repl/repl.go                  - Updated import from pkg/cli → internal/cli
cmd/commands/handlers.go               - Updated import from pkg/cli/commands → internal/cli/commands
internal/cli/commands.go               - Updated internal import references
```

**Migration Details**:
1. Copied `pkg/cli/` to `internal/cli/`
2. Updated all 4 files that imported from `pkg/cli`
3. Updated internal references in `internal/cli/commands.go`
4. Removed old `pkg/cli/` directory

### 2. Final Architecture Structure

**`pkg/` (Reusable Packages)**:
```
pkg/
├── errors/       # Generic error types (reusable)
├── models/       # Model factory & management (reusable)
└── testutil/     # Test helpers (reusable)
```

**`internal/` (Application-Specific)**:
```
internal/
├── app/          # App initialization & lifecycle
├── cli/          # CLI command dispatching (formerly pkg/cli)
├── config/       # Configuration loading
├── llm/          # LLM provider abstractions
├── orchestration/# Agent orchestration
├── repl/         # REPL/interactive session
├── runtime/      # Execution runtime
└── session/      # Session persistence
```

**Root Level**:
```
display/         # UI/display components (major package at root)
session/         # Session manager facade
tools/           # Tool implementations
tracking/        # Token tracking
workspace/       # Workspace management
```

## Rationale

### Why move `pkg/cli` to `internal/cli`?

1. **App-Specific Config**: `CLIConfig` struct has backend-specific fields
2. **Session Coupling**: Tightly coupled to session management
3. **Non-Transferable**: Can't be extracted as a standalone library
4. **Better Signals**: Signals to users "this isn't for external use"

### Why keep `pkg/models`, `pkg/errors`, `pkg/testutil`?

1. **Generic Types**: `AgentError`, `ValidationError` have no app coupling
2. **Factory Pattern**: Model creation is backend-agnostic pattern
3. **Reusable Helpers**: Test utilities could be used in other projects
4. **Clear Intent**: Signals "safe to import from external code"

## Testing

✅ **All tests passing**:
```
✓ code_agent/agent
✓ code_agent/display
✓ code_agent/display/formatters
✓ code_agent/internal/app          (new import path tested)
✓ code_agent/internal/cli          (new import path tested)
✓ code_agent/internal/orchestration (new import path tested)
✓ code_agent/internal/repl
✓ code_agent/internal/runtime
✓ code_agent/pkg/errors
✓ code_agent/pkg/models
✓ code_agent/session
✓ code_agent/tools/display
✓ code_agent/tools/file
✓ code_agent/tools/v4a
✓ code_agent/tracking
✓ code_agent/workspace

Total: 16 packages passing, 0 failures
```

## Future Considerations

### 1. Consider `internal/llm` → `pkg/llm`?

**Current**: `internal/llm/` has provider abstractions  
**Future**: Could be moved to `pkg/llm` if you want to extract LLM backends as a library

**Recommendation**: Keep in `internal/` unless you need external LLM abstraction library

### 2. Display Package

**Current**: `display/` at root level  
**Status**: Keep as-is (it's a major package and special case is acceptable)

### 3. Session Package

**Current**: Both `session/` (public) and `internal/session/` exist  
**Status**: Already well-designed with facade pattern

## Metrics

- **Lines Moved**: 0 (moved via copy, not refactor)
- **Files Updated**: 4 imports changed
- **Build Time**: No change
- **Test Time**: No change (all cached)
- **Package Count**: 16 passing (same)
- **Regressions**: 0

## Benefits

1. **Clearer Intent**: Users see `pkg/` = reusable, `internal/` = app-specific
2. **Better IDE Support**: Go tools understand import visibility rules
3. **Scalability**: Clear pattern for future packages
4. **Documentation**: Structure documents intent without comments

## Migration Notes

If you're using this codebase:
- Internal consumers: `pkg/cli` → `internal/cli` (automatic via Go build)
- External consumers: Can still import `pkg/models`, `pkg/errors`, `pkg/testutil`
- Breaking change: `import "code_agent/pkg/cli"` becomes `import "code_agent/internal/cli"` (only internal)

## Summary

✅ Successfully implemented **Option A** architecture:
- Clear separation: `pkg/` for reusable, `internal/` for app-specific
- Zero regressions: All 16 test packages passing
- Improved signaling: Structure now documents reusability intent
- Future-ready: Pattern established for adding new packages

**Recommendation**: This architecture is stable and ready for continued development.
