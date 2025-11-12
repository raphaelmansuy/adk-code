# Phase 2 Quick Reference Guide

## New Interfaces Added

### 1. Model Provider Adapter (`pkg/models/adapter.go`)

**Interface**: `ProviderAdapter`

```go
type ProviderAdapter interface {
    GetInfo() ProviderInfo
    ValidateConfig(config map[string]string) error
    Name() string
}
```

**Usage**: To add a new AI provider (e.g., Claude):

```go
type ClaudeProviderAdapter struct {
    info ProviderInfo
}

func (a *ClaudeProviderAdapter) GetInfo() ProviderInfo {
    return ProviderInfo{
        Name: "Claude",
        SupportsFunctions: true,
        // ... other fields
    }
}

func (a *ClaudeProviderAdapter) ValidateConfig(config map[string]string) error {
    if _, ok := config["api_key"]; !ok {
        return errors.New(errors.CodeAPIKey, "API key required")
    }
    return nil
}

func (a *ClaudeProviderAdapter) Name() string {
    return "Claude"
}
```

---

### 2. REPL Command Interface (`pkg/cli/commands/interface.go`)

**Interface**: `REPLCommand`

```go
type REPLCommand interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args []string) error
}
```

**Usage**: To add a new REPL command (e.g., `/config`):

```go
type ConfigCommand struct {
    renderer *display.Renderer
}

func (c *ConfigCommand) Name() string {
    return "config"
}

func (c *ConfigCommand) Description() string {
    return "Display or modify agent configuration"
}

func (c *ConfigCommand) Execute(ctx context.Context, args []string) error {
    // Implementation here
    return nil
}

// Register it
registry.Register(NewConfigCommand(renderer))
```

**Available Commands** (Already Implemented):
- `/help` - Help information
- `/prompt` - System prompt
- `/tools` - List tools
- `/models` - Available models
- `/current-model` - Active model details
- `/providers` - Provider information
- `/tokens` - Token usage stats
- `/set-model` - Validate model switch

---

### 3. Workspace Manager Interfaces (`workspace/interfaces.go`)

#### PathResolver
```go
type PathResolver interface {
    ResolvePath(path string, workspaceHint *string) (*ResolvedPath, error)
    GetWorkspaceForPath(path string) string
    ResolvePathString(pathWithHint string) (*ResolvedPath, error)
}
```

#### ContextBuilder
```go
type ContextBuilder interface {
    BuildEnvironmentContext() (string, error)
    BuildWorkspaceContext(workspace *WorkspaceRoot) (string, error)
    SetIncludeStructure(include bool)
    SetMaxDepth(depth int)
}
```

#### VCSDetector
```go
type VCSDetector interface {
    Detect(path string) (VCSType, error)
    GetCommitHash(path string) (string, error)
    GetRemoteURLs(path string) ([]string, error)
    GetBranch(path string) (string, error)
    IsClean(path string) (bool, error)
    GetStatus(path string) (string, error)
}
```

**Usage**: Implement these for custom workspace operations or VCS systems.

---

## Files Modified: NONE

No existing files were modified. All changes are additive:
- New file: `pkg/models/adapter.go` (190 lines)
- New file: `pkg/cli/commands/interface.go` (353 lines)
- New file: `workspace/interfaces.go` (248 lines)

## Backward Compatibility: 100%

All existing code continues to work unchanged:
- Existing REPL commands work identically
- Existing provider implementations work the same way
- Existing workspace operations unchanged
- No breaking changes to any public APIs

## Testing Status

✅ All tests pass (100+ tests)
✅ All packages compile
✅ Full build succeeds
✅ No regressions detected

## Getting Started with New Features

### Adding a REPL Command

1. Define command struct:
```go
type MyCommand struct {
    // dependencies
}
```

2. Implement REPLCommand interface:
```go
func (c *MyCommand) Name() string { return "mycommand" }
func (c *MyCommand) Description() string { return "..." }
func (c *MyCommand) Execute(ctx context.Context, args []string) error { /* ... */ }
```

3. Register in NewDefaultCommandRegistry():
```go
registry.Register(NewMyCommand(...))
```

### Adding a Provider

1. Implement ProviderAdapter:
```go
type MyProviderAdapter struct { /* ... */ }
func (a *MyProviderAdapter) GetInfo() ProviderInfo { /* ... */ }
func (a *MyProviderAdapter) ValidateConfig(...) error { /* ... */ }
func (a *MyProviderAdapter) Name() string { /* ... */ }
```

2. Create model.LLM implementation (as before)
3. Update factory to create instances

### Custom Workspace Operations

1. Implement one or more of:
   - `PathResolver`
   - `ContextBuilder`
   - `VCSDetector`

2. Pass to Manager or use independently
3. Leverage for custom workspace-aware functionality

---

## Architecture Decision Records

See `docs/decisions/` for detailed rationale on:
- Provider adapter pattern
- Command interface design
- Workspace abstraction layers

---

## Future Work (Phase 3+)

- Metrics collection in ContextBuilder
- Caching support in PathResolver
- Extended VCS operations (commit, push, pull)
- Pluggable context builders
- Custom workspace marker detection

---

*Updated: 2025-11-12*
*Phase 2 Status: ✅ Complete*
