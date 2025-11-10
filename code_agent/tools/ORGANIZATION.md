# Tools Package Organization

This document describes the reorganization of the `code_agent/tools` package into logical subpackages by functionality.

## Directory Structure

```
tools/
├── common/              # Shared utilities and registry
│   ├── error_types.go   # Error types and constructors
│   └── registry.go      # Tool registry for categorization
│
├── file/                # File operations (read, write, list, search)
│   ├── file_tools.go    # Core file I/O operations
│   ├── file_validation.go # Path validation and security
│   ├── atomic_write.go  # Atomic write operations
│   └── file_tools_test.go # Tests for file tools
│
├── edit/                # Code editing tools
│   ├── patch_tools.go   # Unified diff patch application
│   ├── search_replace_tools.go # SEARCH/REPLACE block editing
│   └── edit_lines.go    # Line-based edits
│
├── exec/                # Command and program execution
│   └── terminal_tools.go # Shell commands, programs, grep
│
├── search/              # Search and discovery tools
│   └── diff_tools.go    # Preview replace operations
│
├── v4a/                 # V4A semantic patch format
│   ├── v4a_types.go     # V4A data structures
│   ├── v4a_parser.go    # V4A patch parser
│   ├── v4a_applier.go   # V4A patch application
│   ├── v4a_tools.go     # V4A tool definition
│   └── v4a_tools_test.go # V4A tests
│
├── workspace/           # Workspace management
│   └── workspace_tools.go # Workspace-aware path resolution
│
└── tools.go             # Package-level re-exports (backward compatibility)
```

## Package Organization

### `/common`
**Purpose**: Shared types and registry for all tools.

**Contents**:
- `ToolError` and error types for structured error handling
- `ToolRegistry` for dynamic tool registration and categorization
- Error constructors (FileNotFoundError, PatchFailedError, etc.)

**Key Functions**:
- `GetRegistry()` - Get the global tool registry
- `Register()` - Register a tool with metadata
- Error constructors for common failure scenarios

---

### `/file`
**Purpose**: Core file operations - reading, writing, listing, and searching files.

**Contents**:
- `ReadFileTool` - Read file contents with line ranges
- `WriteFileTool` - Write files with atomic operations
- `ReplaceInFileTool` - Simple text replacement
- `ListDirectoryTool` - List directory contents
- `SearchFilesTool` - Find files by wildcard pattern
- Path validation for security

**Key Types**:
- `ReadFileInput` / `ReadFileOutput`
- `WriteFileInput` / `WriteFileOutput`
- `ListDirectoryInput` / `ListDirectoryOutput`
- `ReplaceInFileInput` / `ReplaceInFileOutput`
- `SearchFilesInput` / `SearchFilesOutput`

**Safety Features**:
- Atomic writes prevent data corruption
- Path traversal protection
- Size validation to prevent accidental truncation

---

### `/edit`
**Purpose**: Advanced code editing tools for applying complex changes.

**Contents**:
- `ApplyPatchTool` - Apply unified diff patches
- `SearchReplaceTool` - SEARCH/REPLACE block editing (multi-block support)
- `EditLinesTool` - Line-based edits (replace/insert/delete by line number)
- Diff generation utilities

**Key Types**:
- `ApplyPatchInput` / `ApplyPatchOutput`
- `SearchReplaceInput` / `SearchReplaceOutput`
- `EditLinesInput` / `EditLinesOutput`
- `SearchReplaceBlock` - SEARCH/REPLACE operation

**Tool Priority Order** (within CategoryCodeEditing):
0. SearchReplaceTool (PREFERRED - whitespace-tolerant)
1. EditLinesTool (Line-based)
2. (reserved)
3. ApplyPatchTool (Unified diffs)
4. ApplyV4APatchTool (Semantic patches)
5. PreviewReplaceTool (Previews)

---

### `/exec`
**Purpose**: Command and program execution tools.

**Contents**:
- `ExecuteCommandTool` - Run shell commands with pipes/redirects
- `ExecuteProgramTool` - Execute programs with argument arrays (no quoting issues)
- `GrepSearchTool` - Search files for patterns with regex support

**Key Types**:
- `ExecuteCommandInput` / `ExecuteCommandOutput`
- `ExecuteProgramInput` / `ExecuteProgramOutput`
- `GrepSearchInput` / `GrepSearchOutput`
- `GrepMatch` - Single match from grep

**Use Cases**:
- Shell commands: `ls | grep`, `make build`, `git status`
- Program execution: compilers, interpreters, build tools
- Pattern searching: find specific code patterns in files

---

### `/search`
**Purpose**: Search and discovery tools.

**Contents**:
- `PreviewReplaceTool` - Preview changes before applying
- Diff generation for human-readable previews

**Key Types**:
- `PreviewReplaceInput` / `PreviewReplaceOutput`

---

### `/v4a`
**Purpose**: V4A semantic patch format - patches that use context markers instead of line numbers.

**Contents**:
- V4A data structures (V4APatch, V4AHunk)
- Parser for V4A format
- Applier for V4A patches
- ApplyV4APatchTool definition

**Key Types**:
- `V4APatch` - Complete patch document
- `V4AHunk` - Individual patch operation
- `ApplyV4APatchInput` / `ApplyV4APatchOutput`

**Benefits of V4A**:
- More resilient to code changes than line numbers
- Uses semantic context markers (class names, function names)
- Better for refactoring scenarios

---

### `/workspace`
**Purpose**: Workspace-aware path resolution and management.

**Contents**:
- `WorkspaceTools` - Resolves paths with workspace hints
- Extends file operation types with workspace support

**Key Types**:
- `WorkspaceTools` - Main resolver
- `WorkspaceReadFileInput` / `WorkspaceWriteFileInput` / `WorkspaceListDirectoryInput`

**Features**:
- Supports multi-workspace projects
- Workspace hints: `@workspace:path/to/file`
- Automatic disambiguation for relative paths

---

## Backward Compatibility

The `tools.go` file at the package root re-exports all public functions and types from subpackages. This ensures **100% backward compatibility** with existing code:

```go
// Old style (still works):
tools.NewReadFileTool()
tools.NewApplyPatchTool()
tools.RegisterTool(metadata)

// Under the hood, these call the subpackage versions:
// tools.NewReadFileTool -> file.NewReadFileTool()
// tools.NewApplyPatchTool -> edit.NewApplyPatchTool()
// etc.
```

---

## Package Imports

### Common Pattern in Subpackages

```go
package edit

import (
    "code_agent/tools/common"  // For registry and error types
    "code_agent/tools/file"    // For AtomicWrite
)
```

### From Agent (coding_agent.go)

```go
import "code_agent/tools"

// All calls still work as before:
tools.NewReadFileTool()
tools.NewApplyPatchTool()
tools.GetRegistry()
```

---

## Tool Categories

Tools are registered with categories for organization:

- **CategoryFileOperations** (Priority 0-3)
  - ReadFile, WriteFile, ReplaceInFile, ListDirectory, SearchFiles

- **CategorySearchDiscovery** (Priority 0-1)
  - SearchFiles, GrepSearch, PreviewReplace

- **CategoryCodeEditing** (Priority 0-5)
  - SearchReplace, EditLines, ApplyPatch, ApplyV4APatch, PreviewReplace

- **CategoryExecution** (Priority 0-1)
  - ExecuteCommand, ExecuteProgram

- **CategoryWorkspace** (Priority 0)
  - WorkspaceTools

---

## Migration Guide

If you're working with these tools, nothing changes from the user perspective:

### Usage (Unchanged)
```go
import "code_agent/tools"

// Create tools (all work the same)
readTool, err := tools.NewReadFileTool()
editTool, err := tools.NewApplyPatchTool()
execTool, err := tools.NewExecuteCommandTool()

// Access registry (all work the same)
registry := tools.GetRegistry()
allTools := registry.GetAllTools()
```

### For Developers Adding New Tools

New tools should go in the appropriate subpackage:
1. Create input/output types in the subpackage
2. Implement the tool handler function
3. Register with the appropriate category
4. Update `tools.go` to re-export (if public API)

Example:
```go
// In tools/lint/lint_tools.go
package lint

func NewLintTool() (tool.Tool, error) {
    // Implementation
}

// In tools/tools.go
var (
    NewLintTool = lint.NewLintTool
)
```

---

## Benefits of This Organization

1. **Clear Separation of Concerns**: Each subpackage has a focused responsibility
2. **Easier Navigation**: Find related tools quickly
3. **Better Maintainability**: Smaller files, logical grouping
4. **Cleaner Imports**: Know where each tool lives
5. **Reduced Conflicts**: Changes to one tool don't affect others
6. **Backward Compatibility**: Existing code continues to work without changes
7. **Scalability**: Easy to add new tool categories or tools

---

## Testing

Each subpackage can have its own `*_test.go` files:
- `file/file_tools_test.go`
- `v4a/v4a_tools_test.go`

Run tests by package:
```bash
cd code_agent
go test ./tools/file/...
go test ./tools/edit/...
go test ./tools/...  # All tests
```

---

## Future Enhancements

Potential new subpackages:
- `/analysis` - Code analysis tools
- `/refactor` - Refactoring operations
- `/build` - Build system integration
- `/version_control` - Git/VCS operations
- `/documentation` - Doc generation/analysis

Each would follow the same pattern:
1. Define input/output types
2. Implement tool handlers
3. Register with appropriate category
4. Re-export in `tools.go`

---

## Files Changed Summary

### Reorganized (moved to subpackages)
- `file_tools.go` → `file/file_tools.go`
- `file_tools_test.go` → `file/file_tools_test.go`
- `file_validation.go` → `file/file_validation.go`
- `atomic_write.go` → `file/atomic_write.go`
- `patch_tools.go` → `edit/patch_tools.go`
- `search_replace_tools.go` → `edit/search_replace_tools.go`
- `edit_lines.go` → `edit/edit_lines.go`
- `terminal_tools.go` → `exec/terminal_tools.go`
- `diff_tools.go` → `search/diff_tools.go`
- `workspace_tools.go` → `workspace/workspace_tools.go`
- `v4a_*.go` (all) → `v4a/v4a_*.go`
- `error_types.go` → `common/error_types.go`
- `registry.go` → `common/registry.go`

### New Files
- `tools.go` - Root package re-exports for backward compatibility

### Modified Files
- Updated all `package tools` declarations to `package <subpackage>`
- Updated imports to reference new locations
- Added cross-package imports where needed (common, file)

### No Changes Required
- `coding_agent.go` - Still imports `"code_agent/tools"`
- Any code using `tools.NewXxxTool()` - All still work
- Registry usage - Unchanged

---

## Contact & Questions

For questions about the organization or tool development, refer to:
- Individual subpackage files for implementation details
- `tools.go` for public API surface
- `common/registry.go` for tool registration
