# Dynamic Tool Registration & Categorized System Prompt

**Date:** November 10, 2025  
**Status:** ✅ Completed & Tested

## Summary

Implemented a dynamic tool registration system that automatically categorizes tools and generates structured system prompts. This improves LLM comprehension by organizing tools into logical categories (File Operations, Code Editing, Search & Discovery, Execution, Workspace Management).

## Architecture

### 1. Tool Registry (`tools/registry.go`)

**Core Components:**
- **ToolCategory** enum: Defines 5 functional categories
  - `CategoryFileOperations` - Basic file read/write/list operations
  - `CategorySearchDiscovery` - Finding files and content
  - `CategoryCodeEditing` - Advanced editing (patches, search/replace)
  - `CategoryExecution` - Command and program execution
  - `CategoryWorkspace` - Workspace management (future expansion)

- **ToolMetadata** struct:
  ```go
  type ToolMetadata struct {
      Tool      tool.Tool      // The actual ADK tool
      Category  ToolCategory   // Functional category
      Priority  int            // Sort order within category (0 = highest)
      UsageHint string         // Brief usage tip for LLM
  }
  ```

- **ToolRegistry** struct:
  - Thread-safe operations (sync.RWMutex)
  - `Register(metadata)` - Add tool with metadata
  - `GetByCategory(cat)` - Get sorted tools in category
  - `GetAllTools()` - Flat list of all tools
  - `GetCategories()` - Ordered list of categories with tools
  - `Count()` - Total registered tools

- **Global Registry Pattern:**
  ```go
  var globalRegistry = NewToolRegistry()
  
  func Register(metadata ToolMetadata) error {
      return globalRegistry.Register(metadata)
  }
  
  func GetRegistry() *ToolRegistry {
      return globalRegistry
  }
  ```

### 2. Tool Self-Registration

**Pattern Applied to All 13 Tools:**

Each tool constructor now registers itself:

```go
func NewReadFileTool() (tool.Tool, error) {
    // ... handler implementation ...
    
    t, err := functiontool.New(functiontool.Config{
        Name:        "read_file",
        Description: "Reads the content of a file...",
    }, handler)
    
    if err == nil {
        Register(ToolMetadata{
            Tool:      t,
            Category:  CategoryFileOperations,
            Priority:  0,
            UsageHint: "Examine code, read configs, supports line ranges",
        })
    }
    
    return t, err
}
```

**Registered Tools by Category:**

**File Operations (Priority 0-3):**
- `read_file` - Read files with line range support
- `write_file` - Atomic writes with safety checks
- `replace_in_file` - Simple text replacement
- `list_directory` - Directory exploration
- `search_files` - File pattern matching

**Code Editing (Priority 0-5):**
- `search_replace` - SEARCH/REPLACE blocks (PREFERRED)
- `edit_lines` - Line-based edits
- `apply_patch` - Unified diff patches
- `apply_v4a_patch` - Semantic context patches
- `preview_replace_in_file` - Preview changes

**Search & Discovery (Priority 0-1):**
- `search_files` - Find files by pattern
- `grep_search` - Search file contents

**Execution (Priority 0-1):**
- `execute_command` - Shell commands with pipes
- `execute_program` - Direct program execution (no shell)

### 3. Dynamic Prompt Builder (`agent/dynamic_prompt.go`)

**BuildToolsSection(registry):**
Generates categorized tool listing:

```
## Available Tools

### File Operations

**read_file** - Reads the content of a file...
  → *Usage tip: Examine code, read configs, supports line ranges*

**write_file** - Writes content to a file...
  → *Usage tip: Create or overwrite files with safety checks*

### Code Editing
...
```

**BuildEnhancedPrompt(registry):**
Combines dynamic tools section with existing static guidance:
- Tool listings (dynamic)
- Guidance section (static - decision trees, best practices)
- Pitfalls section (static - common mistakes)
- Workflow section (static - response styles)

### 4. Refactored Agent (`agent/coding_agent.go`)

**Before (Manual):**
```go
readFileTool, err := tools.NewReadFileTool()
if err != nil { return nil, err }

writeFileTool, err := tools.NewWriteFileTool()
if err != nil { return nil, err }

// ... 11 more tools ...

Tools: []tool.Tool{
    readFileTool,
    writeFileTool,
    // ... manual array construction
}
```

**After (Dynamic):**
```go
// Initialize tools (they register themselves)
if _, err := tools.NewReadFileTool(); err != nil { return nil, err }
if _, err := tools.NewWriteFileTool(); err != nil { return nil, err }
// ... remaining tools ...

// Get all registered tools
registry := tools.GetRegistry()
registeredTools := registry.GetAllTools()

// Generate dynamic prompt
dynamicPrompt := BuildEnhancedPrompt(registry)

// Create agent with dynamic configuration
codingAgent, err := llmagent.New(llmagent.Config{
    Name:        "coding_agent",
    Instruction: dynamicPrompt + workspaceContext,
    Tools:       registeredTools,
})
```

### 5. Testing (`agent/dynamic_prompt_test.go`)

**Test Coverage:**
- `TestDynamicPromptGeneration` - Validates prompt structure, category headers, tool inclusion
- `TestToolCategorization` - Verifies tools are properly categorized with correct metadata

**All Tests Pass:**
```
=== RUN   TestDynamicPromptGeneration
--- PASS: TestDynamicPromptGeneration (0.00s)
=== RUN   TestToolCategorization
--- PASS: TestToolCategorization (0.00s)
PASS
ok      code_agent/agent        0.333s
```

## Benefits

### 1. **Better LLM Comprehension**
- Tools organized by functional category
- Clear hierarchy: File Operations → Code Editing → Execution
- Usage tips provide immediate context

### 2. **Maintainability**
- Tool metadata lives with tool implementation (co-location)
- No duplication between tool code and prompt descriptions
- Single source of truth for tool information

### 3. **Extensibility**
- New tools self-register by calling `Register()` in constructor
- New categories easily added to `ToolCategory` enum
- Priority controls ordering within categories

### 4. **Simple & Pragmatic**
- No over-engineering: ~150 lines for registry
- Backward compatible: same tools, better organization
- Global registry pattern (simple, effective)

### 5. **Type-Safe**
- Strong typing for categories (enum)
- Compile-time validation of tool registration
- Clear metadata structure

## Example Generated Prompt (Excerpt)

```markdown
## Available Tools

### File Operations

**read_file** - Reads the content of a file from the filesystem with optional line range support.
  → *Usage tip: Examine code, read configs, supports line ranges (offset/limit) for large files*

**write_file** - Writes content to a file with atomic write support and size validation for safety.
  → *Usage tip: Create or overwrite files with safety checks, atomic writes prevent corruption*

### Code Editing

**search_replace** - Request to replace sections of content in an existing file using SEARCH/REPLACE blocks.
This is the PREFERRED tool for making targeted changes to specific parts of a file.
  → *Usage tip: PREFERRED for targeted edits, supports multiple blocks, whitespace-tolerant*

**edit_lines** - Edit specific lines in a file by line number. Supports replace, insert, and delete operations.
  → *Usage tip: Line-based edits (replace/insert/delete by line number), perfect for structural changes*

### Execution

**execute_command** - Executes a shell command and returns its output.
  → *Usage tip: Run shell commands with pipes/redirects (ls | grep, make build)*

**execute_program** - Execute a program with structured arguments (no shell quoting issues).
  → *Usage tip: Execute programs with arguments (no quoting issues), perfect for compilers/interpreters*
```

## Future Enhancements

1. **Dynamic Tool Discovery:**
   - Auto-discover tools from `tools/` directory
   - Convention-based registration (e.g., `New*Tool()` functions)

2. **Tool Filtering:**
   - Enable/disable categories based on context
   - Workspace-specific tool sets

3. **Advanced Metadata:**
   - Parameter schema generation
   - Example usage snippets
   - Tool dependencies/relationships

4. **Performance Tracking:**
   - Tool usage statistics
   - Success/failure rates per tool
   - Optimization recommendations

## Files Modified

### New Files:
- `code_agent/tools/registry.go` - Registry implementation (150 lines)
- `code_agent/agent/dynamic_prompt.go` - Prompt builder (53 lines)
- `code_agent/agent/dynamic_prompt_test.go` - Tests (97 lines)

### Modified Files:
- `code_agent/agent/coding_agent.go` - Use registry + dynamic prompt
- `code_agent/tools/file_tools.go` - Add registration (5 tools)
- `code_agent/tools/terminal_tools.go` - Add registration (3 tools)
- `code_agent/tools/search_replace_tools.go` - Add registration (1 tool)
- `code_agent/tools/edit_lines.go` - Add registration (1 tool)
- `code_agent/tools/patch_tools.go` - Add registration (1 tool)
- `code_agent/tools/v4a_tools.go` - Add registration (1 tool)
- `code_agent/tools/diff_tools.go` - Add registration (1 tool)

**Total Lines Changed:** ~400 lines  
**Net Addition:** ~300 lines (registry + tests)

## Validation

✅ **Compilation:** `make build` succeeds  
✅ **Tests:** All tests pass (agent + tools)  
✅ **Prompt Quality:** Generated prompt is well-structured and categorized  
✅ **Backward Compatibility:** Same tools, same functionality, better organization  
✅ **Code Quality:** Clean, idiomatic Go with proper error handling

## Conclusion

Successfully implemented a **simple, high-value solution** for dynamic tool registration with automatic categorization. The system:
- Improves LLM comprehension through logical grouping
- Maintains co-location of tool metadata with tool code
- Provides extensibility for future tools and categories
- Remains simple and pragmatic (not over-engineered)

**Status:** ✅ Ready for production use
