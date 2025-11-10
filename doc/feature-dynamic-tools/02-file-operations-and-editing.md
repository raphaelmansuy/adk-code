# Feature Comparison: File Operations and Code Editing Tools

## Overview
This document provides a deep comparison of file operation and code editing capabilities across both systems.

---

## Code Agent: File Operations Toolkit

### Core File Tools

#### Read File
- **Tool Name**: `read_file`
- **Purpose**: Read file contents with optional line range support
- **Parameters**:
  - `path` (string): File path
  - `offset` (int, optional): Starting line number (1-indexed)
  - `limit` (int, optional): Number of lines to read
- **Output**: `Content`, `TotalLines`, `ReturnedLines`, `StartLine`
- **Features**:
  - Line range support for large files
  - Memory-efficient partial reads
  - Line counting
  - Safe error handling

**Use Cases**:
```
- Examine specific sections of large files
- Read configuration files
- Preview code before editing
- Check file structure without full load
```

#### Write File
- **Tool Name**: `write_file`
- **Purpose**: Create or overwrite files atomically
- **Parameters**:
  - `path` (string): File path
  - `content` (string): File content (MUST be complete)
  - `create_dirs` (bool): Create parent directories
  - `atomic` (bool): Use atomic writes (default: true)
  - `allow_size_reduce` (bool): Allow size reduction >90%
- **Output**: `Success`, `Message`, `Error`, `BytesWritten`
- **Safety Features**:
  - Atomic writes prevent corruption
  - Size reduction safeguard (>90% = error)
  - Auto-create parent directories
  - Validation before write

**Critical Usage Rule**: ALWAYS provide complete intended content. Never truncate files.

**Use Cases**:
```
- Create new source files
- Update configuration
- Generate documentation
- Create test fixtures
```

#### Replace in File
- **Tool Name**: `replace_in_file`
- **Purpose**: Find and replace text with exact string matching
- **Parameters**:
  - `path` (string): File path
  - `old_text` (string): Exact text to find (whitespace-sensitive)
  - `new_text` (string): Replacement text
  - `max_replacements` (int): Maximum replacements (safety limit)
- **Output**: `Success`, `Message`, `ReplacementCount`, `Error`
- **Features**:
  - Exact string matching (not regex)
  - Whitespace-sensitive
  - Replacement count limiting
  - Detailed error messages

**Use Cases**:
```
- Simple text substitutions
- Update variable names (carefully)
- Modify string literals
- Quick fixes in config files
```

#### List Directory
- **Tool Name**: `list_directory`
- **Purpose**: Explore directory structure
- **Parameters**:
  - `path` (string): Directory path
  - `recursive` (bool): Recurse into subdirectories
  - `max_depth` (int): Max recursion depth
- **Output**: `Items` (with type and path), `Total`, `Error`
- **Features**:
  - File type identification
  - Recursive listing with depth control
  - Count totals
  - Efficient directory traversal

**Use Cases**:
```
- Understand project structure
- Find all files of type
- Navigate monorepo
- Discover configuration files
```

#### Search Files
- **Tool Name**: `search_files`
- **Purpose**: Find files matching patterns
- **Parameters**:
  - `pattern` (string): Glob pattern (*.go, test_*.py, etc.)
  - `path` (string): Search directory
  - `recursive` (bool): Recursive search
- **Output**: `Files` (matching paths), `Count`, `Error`
- **Features**:
  - Glob pattern matching
  - Recursive search
  - Fast file discovery
  - Multiple file types

**Use Cases**:
```
- Find test files
- Locate source files by extension
- Discover documentation
- Find configuration files
```

---

## Code Agent: Advanced Code Editing Tools

### Search and Replace (RECOMMENDED for edits)

- **Tool Name**: `search_replace`
- **Purpose**: Make targeted changes using SEARCH/REPLACE block notation
- **Format**:
```
------- SEARCH
[exact content to find]
=======
[replacement content]
------- END SEARCH/REPLACE
```
- **Features**:
  - Multiple blocks in single call
  - Whitespace-tolerant
  - Context-aware replacement
  - Detailed diff output
- **Advantages**:
  - RECOMMENDED for most edits
  - Better readability
  - Less error-prone than string matching
  - Easier for LLM reasoning

**Example**:
```
------- SEARCH
function add(a, b) {
    return a + b;
}
=======
function add(a, b) {
    // Adding two numbers
    return a + b;
}
------- END SEARCH/REPLACE
```

### Edit Lines

- **Tool Name**: `edit_lines`
- **Purpose**: Edit specific lines by line number
- **Parameters**:
  - `file_path` (string): File to edit
  - `start_line` (int): Starting line (1-indexed)
  - `end_line` (int): Ending line (1-indexed)
  - `new_lines` (string): Replacement content
  - `mode` (string): "replace", "insert", or "delete"
- **Modes**:
  - **replace**: Replace lines start to end with new_lines
  - **insert**: Insert new_lines before start_line
  - **delete**: Remove lines start to end
- **Output**: `Success`, `Message`, `NewLineCount`, `Error`

**Use Cases**:
```
- Fix specific line ranges
- Add import statements
- Delete unused code blocks
- Reorder function definitions
```

### Apply Patch (Standard Unified Diff)

- **Tool Name**: `apply_patch`
- **Purpose**: Apply unified diff patches
- **Format**: Standard unified diff format
- **Features**:
  - Line number flexibility
  - Context matching
  - Dry-run mode for testing
  - Detailed error messages
- **Parameters**:
  - `path` (string): File to patch
  - `patch` (string): Unified diff content
  - `dry_run` (bool): Preview without applying

**Example Patch**:
```
--- a/src/main.go
+++ b/src/main.go
@@ -10,6 +10,7 @@
 func main() {
     fmt.Println("Hello")
+    fmt.Println("World")
     os.Exit(0)
 }
```

### Apply V4A Patch (Semantic Patches)

- **Tool Name**: `apply_v4a_patch`
- **Purpose**: Apply context-aware semantic patches
- **Format**: Semantic markers instead of line numbers
```
*** Update File: filepath
@@ <context1>              (e.g., class User, func HandleRequest)
@@     <context2>          (nested: def method, indented)
-<line_to_remove>
+<line_to_add>
```
- **Advantages**:
  - Resilient to code changes
  - Semantic clarity
  - Better for frequently-changing files
- **Parameters**:
  - `path` (string): File to patch
  - `patch` (string): V4A patch content
  - `dry_run` (bool): Preview mode

**Example V4A Patch**:
```
*** Update File: src/User.go
@@ type User struct
-    email string
+    email string
+    verified bool
```

### Preview Replace

- **Tool Name**: `preview_replace` (in diff_tools.go)
- **Purpose**: Show diff before replacing
- **Parameters**:
  - `path` (string): File path
  - `old_string` (string): Text to replace
  - `new_string` (string): Replacement
- **Output**: Unified diff preview

---

## Cline: File Operations in VS Code

### Integrated Editor Experience

Cline's file operations are deeply integrated with VS Code:

1. **File Reading**:
   - Syntax highlighting in editor
   - Integrated with VS Code's file system
   - Respects `.gitignore` and `.vscodeignore`
   - AST analysis via tree-sitter

2. **File Writing**:
   - Creates files in workspace directly
   - Diff view for review before accepting
   - Timeline tracking all changes
   - Automatic linter/compiler error detection

3. **File Editing**:
   - Diff view comparing old vs new
   - Line-by-line approval capability
   - Human can edit in diff view
   - Can accept, reject, or modify

### Context Management

**Large Project Handling**:
- Analyzes file structure & AST
- Runs regex searches for relevant context
- Carefully manages context window
- Avoids overwhelming LLM with irrelevant files
- Uses `@file`, `@folder`, `@url` to add context manually

### Search Capabilities

**Built-in Search Tools**:
- **Ripgrep Integration**: Fast grep-like search
- **Tree-sitter**: Syntax-aware code search
- **File pattern matching**: Find related files
- **AST-based search**: Find code structures

**Search Optimization**:
- Understands project dependencies
- Locates related files automatically
- Fetches function signatures
- Finds usage patterns

---

## Comparative Analysis: File Operations

### Speed and Efficiency

| Operation | Code Agent | Cline |
|-----------|-----------|-------|
| Read File | Direct I/O (fast) | VS Code API (cached) |
| Write File | Atomic direct write | Diff + editor creation |
| Replace | Text matching (fast) | Diff computation |
| Search | Grep pattern (moderate) | Ripgrep + AST (fast) |
| Directory List | Direct I/O (fast) | VS Code explorer (cached) |

### Safety Features

| Feature | Code Agent | Cline |
|---------|-----------|-------|
| Size reduction check | ✓ (>90% warns) | ✓ (diff preview) |
| Atomic writes | ✓ | ✓ (after approval) |
| Change tracking | Session only | Timeline + git |
| Human approval | None | ✓ Required |
| Linter integration | Manual execution | ✓ Automatic |

### Code Editing Approaches

**Code Agent**:
- Line-based (`edit_lines`)
- Text-based (`search_replace`)
- Diff-based (`apply_patch`)
- Semantic patches (`apply_v4a_patch`)
- Multiple strategies, agent chooses

**Cline**:
- Editor-integrated (single unified approach)
- Always shows diff first
- Requires human approval
- Automatic error detection
- Can manually edit before accepting

### Context Awareness

**Code Agent**:
- Workspace-relative paths
- Multi-workspace support
- Manual context building
- System prompt includes environment

**Cline**:
- IDE-native context
- Automatic file discovery
- AST-based understanding
- Project structure awareness
- Dependency graph analysis

---

## Tool Ecosystem Comparison

### Code Agent's 14+ Tools by Category

**File Operations** (4 tools):
- read_file
- write_file
- replace_in_file
- list_directory

**Search & Discovery** (2 tools):
- search_files
- grep_search

**Code Editing** (4 tools):
- search_replace ⭐ RECOMMENDED
- edit_lines
- apply_patch
- apply_v4a_patch

**Execution** (2 tools):
- execute_command
- execute_program

**Workspace** (2+ tools):
- Workspace-aware path helpers
- Multi-workspace support

### Cline's Integrated Tools

**Native Tools**:
- File creation/editing (diff-based)
- Terminal execution
- Browser automation (Computer Use)
- Context addition (@file, @folder, @url)

**MCP-Extensible Tools**:
- Any MCP server tool
- Community tools
- Custom built tools
- Integration with external systems

---

## Best Practices

### Code Agent File Editing

1. **Use search_replace for most edits** (recommended)
2. **Use edit_lines for structural changes** (line-based)
3. **Use apply_v4a_patch for semantic changes** (context-aware)
4. **Always provide complete content in write_file**
5. **Use preview_replace to check before replacing**
6. **Use line ranges in read_file for large files**

### Cline File Editing

1. **Review diff before approving**
2. **Use @file to include context manually**
3. **Verify linter errors auto-detected**
4. **Use checkpoint system for testing**
5. **Let browser automation test changes**
6. **Leverage AST search for complex refactoring**

---

## Conclusion

**Code Agent** provides a toolkit of specialized file and editing tools that require the agent to choose the right tool for each task. This forces intelligent decision-making but adds complexity.

**Cline** integrates file operations directly into VS Code, providing a unified diff-based review and approval workflow. This is safer for interactive use but requires human involvement.

**For complex edits**: Code Agent's `apply_v4a_patch` offers superior semantic awareness
**For interactive development**: Cline's diff + approval workflow is superior
**For bulk operations**: Code Agent's tool diversity is more powerful
**For safety**: Cline's human-in-the-loop approach is recommended

---

## See Also

- [03-terminal-and-execution.md](./03-terminal-and-execution.md) - Command execution comparison
- [04-extensibility.md](./04-extensibility.md) - Custom tool creation
- [05-browser-and-ui-testing.md](./05-browser-and-ui-testing.md) - Browser automation
