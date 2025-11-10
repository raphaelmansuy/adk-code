# Tool Architecture: code_agent vs Cline

Deep dive into how each system structures, discovers, and presents tools.

## Tool Inventory

### code_agent Tools (Go)

| Tool | Input | Output | Special Features |
|------|-------|--------|------------------|
| `read_file` | path, offset, limit | content, total_lines, returned_lines | Line range support |
| `write_file` | path, content, create_dirs, atomic, allow_size_reduce | success, message, error | Size validation, atomic writes |
| `search_replace` | path, diff (SEARCH/REPLACE blocks), preview | success, blocks_applied, preview | Whitespace-tolerant matching |
| `edit_lines` | file_path, start_line, end_line, new_lines, mode | success, lines_modified, preview | Line-based operations |
| `apply_patch` | file_path, patch (unified diff), dry_run, strict | success, lines_added, lines_removed, preview | RFC 3881 format |
| `list_directory` | path, recursive | files (with metadata), success | Recursive traversal |
| `search_files` | path, pattern, max_results | matches, count, success | Glob pattern matching |
| `execute_command` | command, working_dir, timeout | output, exit_code, stderr | Shell execution |
| `grep_search` | pattern, files/directories | matches (with line numbers) | Regex search |

**Source:** `code_agent/tools/file_tools.go`, `edit_lines.go`, `search_replace_tools.go`, `patch_tools.go`, `terminal_tools.go`

### Cline Tools (TypeScript)

| Tool | Input | Output | Model Variants |
|------|-------|--------|-----------------|
| `read_file` | path, task_progress | content | GENERIC, NATIVE_GPT_5, NATIVE_NEXT_GEN |
| `write_to_file` | path/absolutePath, content, task_progress | success | ✅ (note: absolutePath for native) |
| `replace_in_file` | path/absolutePath, diff (SEARCH/REPLACE), task_progress | success | ✅ (note: absolutePath for native) |
| `apply_patch` | input (bash command), task_progress | success | NATIVE_GPT_5, GPT_5 only |
| `list_files` | path, recursive, task_progress | files | GENERIC, NATIVE variants |
| `search_files` | path, regex, file_pattern, task_progress | matches | All variants |
| `execute_command` | command, requires_approval, timeout, task_progress | output | All variants |
| `browser_action` | action, args | screenshot, logs | Browser automation |
| `ask_followup_question` | question | response | User interaction |

**Source:** `research/cline/src/core/prompts/system-prompt/tools/`

## Design Pattern Comparison

### code_agent: Direct Go Implementation

```go
// Tool Input Structure
type SearchReplaceInput struct {
    Path    string `json:"path" jsonschema:"description"`
    Diff    string `json:"diff" jsonschema:"description"`
    Preview *bool  `json:"preview,omitempty" jsonschema:"optional"`
}

// Tool Output Structure
type SearchReplaceOutput struct {
    Success       bool     `json:"success"`
    BlocksApplied int      `json:"blocks_applied"`
    TotalBlocks   int      `json:"total_blocks"`
    PreviewContent string `json:"preview_content,omitempty"`
    Message       string   `json:"message,omitempty"`
    Error         string   `json:"error,omitempty"`
}

// Handler Function
handler := func(ctx tool.Context, input SearchReplaceInput) SearchReplaceOutput {
    // Implementation...
    return SearchReplaceOutput{
        Success: true,
        BlocksApplied: len(blocks),
        Message: fmt.Sprintf("Applied %d blocks", len(blocks)),
    }
}

// Tool Registration
return functiontool.New(functiontool.Config{
    Name:        "search_replace",
    Description: "Request to replace sections of content...",
}, handler)
```

**Characteristics:**

- ✅ Simple and direct
- ✅ Type-safe (Go generics)
- ✅ Single implementation per tool
- ❌ No model-aware variants
- ❌ All tools in single code file

### Cline: Config-Based Registry

```typescript
// Tool Spec for GENERIC variant
const generic: ClineToolSpec = {
    variant: ModelFamily.GENERIC,
    id: ClineDefaultTool.FILE_EDIT,
    name: "replace_in_file",
    description: "Request to replace sections...",
    parameters: [
        {
            name: "path",
            required: true,
            instruction: "The path of the file to modify...",
            usage: "File path here",
        },
        {
            name: "diff",
            required: true,
            instruction: "One or more SEARCH/REPLACE blocks...",
            usage: "Search and replace blocks here",
        },
    ],
}

// Tool Spec for NATIVE_GPT_5 (different parameters!)
const NATIVE_GPT_5: ClineToolSpec = {
    variant: ModelFamily.NATIVE_GPT_5,
    id: ClineDefaultTool.FILE_EDIT,
    name: "replace_in_file",
    description: "[IMPORTANT: Always output the absolutePath first]...",
    parameters: [
        {
            name: "absolutePath",  // <-- Different parameter name
            required: true,
            instruction: "The absolute path to the file to write to.",
        },
        // ...
    ],
}

export const replace_in_file_variants = [generic, NATIVE_GPT_5]
```

**Characteristics:**

- ✅ Model-aware (multiple variants)
- ✅ Easy to add new LLM families
- ✅ Centralized configuration
- ✅ Parameter customization per variant
- ❌ More boilerplate per tool
- ❌ Runtime variant selection (not compiled)

## Key Technical Differences

### 1. Whitespace Matching Strategy

**code_agent:**

```go
// Try exact match first
matchIdx := findExactMatch(result, block.SearchContent, currentOffset)

// Fall back to line-trimmed match if exact fails
if matchIdx == -1 {
    matchIdx = lineTrimmedMatch(result, block.SearchContent, currentOffset)
}

// lineTrimmedMatch: Compare lines with strings.TrimSpace()
// Handles minor indentation differences gracefully
```

**Cline:**

- Only exact matching documented
- Relies on LLM to provide exact content
- Auto-formatter warning acknowledges real-world issues

**Winner:** code_agent for robustness, Cline for simplicity.

### 2. Patch Format Implementation

**code_agent (Unified Diff):**

```go
// RFC 3881 format parsing
func ParseUnifiedDiff(patch string) ([]PatchHunk, error) {
    // Parses @@ -origStart,origCount +newStart,newCount @@ format
    // Returns hunks with context lines
}

// Can use standard git diff tools
```

**Cline (Custom V4A Format):**

```typescript
// Custom format with class/function context
/*
*** Update File: path/to/file
@@ class BaseClass
@@     def method():
-          pass
+          raise NotImplementedError()
*/
```

**Winner:** code_agent for interoperability, Cline for semantic context.

### 3. Safety Mechanisms

**code_agent:**

```go
// Size validation (prevents accidental truncation)
if currentSize > 1000 && newSize < currentSize/10 {
    if !allowSizeReduce {
        return WriteFileOutput{
            Success: false,
            Error: fmt.Sprintf(
                "SAFETY CHECK FAILED: Refusing to reduce file size from %d to %d bytes...",
                currentSize, newSize,
            ),
        }
    }
}

// Atomic write
err = AtomicWrite(input.Path, []byte(newContent), 0644)
```

**Cline:**

```typescript
// Auto-formatting warning in documentation
// "After using either write_to_file or replace_in_file, 
//  the user's editor may automatically format the file"
// Use this final state as reference for SEARCH blocks
```

**Analysis:**

- code_agent: Built-in safety checks
- Cline: Workflow awareness (document real-world behavior)

### 4. Tool Registry Pattern

**code_agent:**

```go
// Tools inline in coding_agent.go
func NewCodingAgent(model model.Model, ...) (agentiface.Agent, error) {
    readFile, _ := NewReadFileTool()
    writeFile, _ := NewWriteFileTool()
    searchReplace, _ := NewSearchReplaceTool()
    // ...
    
    tools := []tool.Tool{
        readFile,
        writeFile,
        searchReplace,
        editLines,
        applyPatch,
        // ...
    }
    
    return llmagent.New(llmagent.Config{
        SystemPrompt: SystemPrompt,
        Tools: tools,
        // ...
    })
}
```

**Cline:**

```typescript
// Registry pattern with PromptBuilder
export class ClineToolSet {
    static getTools(family: ModelFamily): ToolSpec[] {
        // Returns tools for specific model family
    }
    
    static getToolsForVariantWithFallback(
        family: ModelFamily,
        requestedIds: string[]
    ): ToolSpec[] {
        // Returns requested tools, falling back to GENERIC
    }
}

// Used by PromptBuilder
const enabledTools = PromptBuilder.getEnabledTools(variant, context)
```

**Advantages:**

- code_agent: Simple, straightforward
- Cline: Flexible, extensible, model-aware

## Tool Evolution Patterns

### Adding a New Tool to code_agent

1. Define input struct with JSON schema tags
2. Define output struct
3. Implement handler function
4. Register with `functiontool.New()`
5. Add to tool list in `NewCodingAgent()`
6. Document in `enhanced_prompt.go`
7. Add tests

**Effort:** 4-5 files, straightforward

### Adding a New Tool to Cline

1. Create `tools/new_tool.ts`
2. Define variants (GENERIC, NATIVE_GPT_5, etc.)
3. Export from `tools/index.ts`
4. Add to registry (ClineToolSet)
5. Update prompt components
6. Document in system prompt

**Effort:** 5-6 files, more boilerplate

## Tool Presentation in Prompts

### code_agent Enhanced Prompt

Tool documentation emphasizes **when to use each tool**:

```markdown
## Tool Selection Guide

### When to Edit Files:

1. **Creating new file?** → use write_file
2. **Know exact line numbers?** → use edit_lines (for structural changes)
3. **Know exact content to find?** → use search_replace (for targeted changes)
4. **Have unified diff patch?** → use apply_patch (for complex changes)
5. **Want to preview first?** → use preview=true or dry_run=true

### When to Execute Programs:

1. **Shell pipeline with | or > ?** → use execute_command
2. **Program with arguments?** → use execute_program (avoids quoting issues)
```

### Cline Prompt Components

Tool documentation organized by **file editing strategy**:

```markdown
# EDITING FILES

## write_to_file
### Purpose
- Create a new file, or overwrite the entire contents

### When to Use
- Initial file creation
- Overwriting large boilerplate files
- When replace_in_file would be unwieldy

## replace_in_file
### Purpose
- Make targeted edits to specific parts

### When to Use
- Small, localized changes
- Especially useful for long files
```

**Insight:** code_agent teaches decision-making, Cline teaches best practices.

## Extensibility Comparison

### code_agent

- **Fixed tool set** (defined at compile time)
- **Adding tools** requires code changes + recompile
- **Cannot be extended at runtime**
- **Pro:** Predictable, type-safe
- **Con:** Monolithic, not plugin-friendly

### Cline

- **MCP server integration** for extensibility
- **Tools can be added dynamically** via MCP
- **Plugin architecture** for custom tools
- **Pro:** Extensible, flexible
- **Con:** Runtime overhead, more complex

## Performance Characteristics

### code_agent

- Compiled Go binary (fast startup)
- No variant selection overhead
- Direct function calls (no indirection)
- Memory efficient

### Cline

- TypeScript runtime (slower startup)
- Variant selection on each use
- PromptBuilder overhead
- More flexible but slower

## Lessons for Tool Design

### From code_agent

1. **Keep tools focused** - separate edit_lines from search_replace
2. **Provide structured output** - success/error/metadata fields
3. **Include safety checks** - size validation, atomic writes
4. **Document decision trees** - which tool for which job
5. **Implement fallbacks** - whitespace-tolerant matching

### From Cline

1. **Make tools model-aware** - variants for different LLMs
2. **Use registry pattern** - for easy discovery and management
3. **Support extensibility** - MCP servers, plugin architecture
4. **Document real-world behavior** - auto-formatting, editor quirks
5. **Optimize for parameters** - e.g., absolutePath for native models

## See Also

- `./COMPARISON.md` - High-level comparison
- `./PROMPT_STRATEGY.md` - System prompt analysis
