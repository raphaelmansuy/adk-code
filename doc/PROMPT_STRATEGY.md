# System Prompt Strategy: code_agent vs Cline

How each system instructs the LLM to use tools effectively.

## Prompt Architecture Overview

### code_agent: Monolithic Enhanced Prompt

**File:** `agent/enhanced_prompt.go` (~300 lines)

**Structure:**

```
1. Header: "You are an expert AI coding assistant..."
2. Available Tools section (lists all tools)
3. Tool Selection Guide (decision tree for which tool when)
4. Discovery Tools section
5. Execution Tools section
6. Critical Best Practices
7. Correct Tool Usage (anti-patterns)
8. Testing Methodology
9. Common Pitfalls & Solutions
10. Workflow Pattern
11. Response Style
12. Safety Features (Our Advantages)
13. Key Differences from Other Agents
```

**Characteristics:**

- Single monolithic prompt string
- Embedded at compile time
- ~2500 words, comprehensive
- Tool-centric (organized by tool)
- Compares itself to other agents

### Cline: Component-Based Prompt

**Directory:** `src/core/prompts/system-prompt/`

**Architecture:**

```text
PromptRegistry
├── PromptBuilder (assembles components)
├── components/
│   ├── agent_role.ts
│   ├── capabilities.ts
│   ├── editing_files.ts
│   ├── rules.ts
│   ├── feedback.ts
│   └── ... (15+ components)
├── tools/
│   ├── read_file.ts
│   ├── write_to_file.ts
│   ├── replace_in_file.ts
│   └── ... (20+ tools)
└── variants/
    ├── GENERIC
    ├── NATIVE_GPT_5
    └── NATIVE_NEXT_GEN
```

**Characteristics:**

- Modular components (separate files)
- Runtime assembly (PromptBuilder)
- ~500 lines per component
- Component-centric (organized by capability)
- Model-aware variants

## Tool Guidance Approach

### code_agent: Decision Tree

**Philosophy:** "Teach the LLM to choose the right tool"

```markdown
## Tool Selection Guide

### When to Edit Files

1. **Creating new file?** → use write_file
2. **Know exact line numbers?** → use edit_lines
3. **Know exact content to find?** → use search_replace
4. **Have unified diff patch?** → use apply_patch
5. **Want to preview first?** → use preview=true

## Critical Best Practices

### COMPLETENESS (Prevent Truncation)

- When using write_file: ALWAYS provide COMPLETE file content
- NEVER truncate files
- Include ALL sections, even unchanged ones

### SAFETY FIRST

1. Read before edit
2. Validate after edit
3. Use preview modes
4. Start simple

### CORRECT TOOL USAGE

✅ DO:
- Keep blocks concise
- Use multiple small blocks
- List blocks in file order
- Ensure SEARCH matches EXACTLY

❌ DON'T:
- Include long runs of unchanged lines
- Truncate lines mid-way
- Assume whitespace doesn't matter
```

### Cline: Workflow Optimization

**Philosophy:** "Teach the LLM effective workflows"

```markdown
## EDITING FILES

### write_to_file vs replace_in_file Decision

**Default to replace_in_file** for most changes.

**Use write_to_file** when:
- Creating new files
- Changes are so extensive replace_in_file is complex
- Need to completely reorganize
- File is small and mostly changing
- Generating boilerplate

### Workflow Tips

1. Assess scope of changes
2. Apply replace_in_file with carefully crafted blocks
3. **IMPORTANT:** Prefer single replace_in_file with multiple blocks
   - NOT multiple successive replace_in_file calls
4. For major overhauls, use write_to_file
5. Use final state as reference for subsequent edits

### Auto-formatting Considerations

- Editor may auto-format after write_to_file or replace_in_file
- May break lines, adjust indentation, convert quotes
- **Use this final state as reference point**
- Especially important for SEARCH blocks (must match exactly)
```

**Insight:** code_agent optimizes for correctness, Cline optimizes for efficiency.

## Safety and Error Handling

### code_agent: Built-In Safeguards

**Size Validation:**

```go
// Prevents accidental truncation
if currentSize > 1000 && newSize < currentSize/10 {
    if !allowSizeReduce {
        return WriteFileOutput{
            Success: false,
            Error: fmt.Sprintf(
                "SAFETY CHECK FAILED: Refusing to reduce file size from %d to %d bytes (%.1f%% reduction)",
                currentSize, newSize, reduction_percent,
            ),
        }
    }
}
```

**Whitespace Tolerance:**

```go
// Try exact match first, then line-trimmed fallback
matchIdx := findExactMatch(result, block.SearchContent, currentOffset)
if matchIdx == -1 {
    matchIdx = lineTrimmedMatch(result, block.SearchContent, currentOffset)
}
```

**Prompt Guidance:**

```markdown
## Common Pitfalls & Solutions

### Pitfall 1: Shell Argument Parsing
❌ Wrong: ./calculate 2 + 2 → ["./calculate", "2", "+", "2"]
✅ Right: execute_program("./calculate", ["2 + 2"]) → One arg

### Pitfall 2: File Size Reduction
❌ Wrong: Overwrite large file with small content
✅ Right: write_file has size validation
→ Use allow_size_reduce=true only if intentional

### Pitfall 3: Not Reading Before Editing
❌ Wrong: Assume code structure, make blind edits
✅ Right: read_file first, understand context

### Pitfall 4: search_replace Block Not Found
❌ Wrong: SEARCH doesn't match (whitespace issue)
✅ Right: Copy exact content from file (including indentation)

### Pitfall 5: Not Testing After Compile
❌ Wrong: Compile, assume success, run immediately
✅ Right: Check exit_code=0 and stderr empty before running
```

### Cline: Workflow Awareness

**Auto-formatting Warning:**

```markdown
# Auto-formatting Considerations

- After using write_to_file or replace_in_file, editor may auto-format
- May modify:
  - Breaking single lines into multiple
  - Adjusting indentation
  - Converting quote style
  - Organizing imports
  - Adding/removing trailing commas
  - Enforcing consistent braces
  - Standardizing semicolons
  
- Tool responses include final state after formatting
- **Use this final state as reference for subsequent SEARCH blocks**
- Especially important for replace_in_file (content must match exactly)
```

**Emphasis on Optimization:**

```markdown
# Workflow Tips

**IMPORTANT:** When making several changes to same file:
- Prefer **single replace_in_file call** with multiple SEARCH/REPLACE blocks
- DO NOT make multiple successive replace_in_file calls
- Reason: More efficient, preserves line numbers

**Example:**
✅ One call with 2 blocks (import + usage)
❌ Two separate calls (first for import, then for usage)
```

## Variant-Specific Guidance

### code_agent Variants

- Single prompt for all models
- Model-agnostic approach
- Works with any LLM

### Cline Variants

Each variant gets specialized guidance:

**GENERIC Variant:**

```typescript
{
    variant: ModelFamily.GENERIC,
    description: "Request to replace sections... (standard phrasing)"
    parameters: [
        {name: "path", required: true},
        {name: "diff", required: true},
    ]
}
```

**NATIVE_GPT_5 Variant:**

```typescript
{
    variant: ModelFamily.NATIVE_GPT_5,
    description: "[IMPORTANT: Always output the absolutePath first]..."
    parameters: [
        {name: "absolutePath", required: true},  // ← Different!
        {name: "diff", required: true},
    ]
}
```

**Impact:** Native models get different parameter names and extra instructions.

## Best Practices Documentation

### code_agent Approach: Comprehensive

```markdown
## Best Practices

### File Operations
- Use relative paths (not absolute)
- Check paths first if operation fails
- Read before writing
- Use exact matches for replace_in_file

### Shell Execution
- Understand working_dir
- Quote arguments properly
- Test incrementally
- Check exit codes

### Testing Methodology
1. Start Simple
2. Verify Incrementally
3. Read Error Messages
4. Test Edge Cases
5. Validate Assumptions

### Common Pitfalls (5 detailed examples with solutions)
```

### Cline Approach: Workflow-Focused

```markdown
## Workflow Tips

1. Before editing, assess scope
2. Apply replace_in_file with crafted blocks
3. Prefer single call with multiple blocks
4. Use write_to_file for major overhauls
5. Use final state as reference

## Auto-formatting Considerations (7 specific scenarios)
```

## Response Style Guidance

### code_agent Approach

```markdown
## Response Style

- **Be concise but thorough**: Explain reasoning for important decisions
- **Show your work**: Display command outputs, test results, errors
- **Handle errors gracefully**: When something fails, explain why and fix
- **Verify success**: Always test changes before declaring victory
- **Iterate systematically**: If approach doesn't work, understand why first
```

### Cline Approach

No explicit response style guidance (implied through task_progress tracking).

## Prompt Performance Insights

| Aspect | code_agent | Cline |
|--------|-----------|-------|
| **Length** | ~2500 words | ~500 per component |
| **Complexity** | High (monolithic) | Medium (modular) |
| **Token count** | Higher baseline | Varies by variant |
| **Compilation** | Static (built-in) | Dynamic (runtime) |
| **Personalization** | Same for all models | Variant-specific |
| **Maintenance** | Single file | 20+ component files |
| **Error messaging** | Embedded in prompt | Component-specific |

## Surprising Insights

### 1. code_agent is More Detailed

Despite being a "simple" Go agent, code_agent's prompt is MORE comprehensive than Cline's components. It covers:

- 5 detailed pitfall examples
- Decision trees for tool selection
- Safety features list
- Comparison to other agents

### 2. Cline is More Workflow-Aware

Cline acknowledges real-world behavior (auto-formatting) that code_agent doesn't mention. This shows:

- Integration with real editor workflows
- Practical experience with user pain points
- Adaptation to real-world LLM usage

### 3. code_agent Has Better Safety Guidance

code_agent documents preventive safety (don't truncate files, use atomic writes, check before writing).
Cline documents reactive safety (watch out for auto-formatting changes).

### 4. Variant-Specific Parameters Are Powerful

Cline's approach of changing parameter names per model variant (path → absolutePath) shows deep understanding that different LLMs have different reasoning patterns.

### 5. Tool Selection > Tool Usage

code_agent spends more time teaching WHEN to use tools, not just HOW.
Cline spends more time teaching workflow patterns, not just tool mechanics.

## Recommendations for Best Prompt Design

### From code_agent

1. ✅ Include decision trees ("when to use each tool")
2. ✅ Document common pitfalls with concrete solutions
3. ✅ Explain safety features built into tools
4. ✅ Compare your approach to alternatives
5. ✅ Provide comprehensive testing methodology
6. ✅ Include response style guidance

### From Cline

1. ✅ Warn about real-world editor behaviors
2. ✅ Optimize for common workflows (prefer single calls)
3. ✅ Create model-aware variants
4. ✅ Organize prompts into reusable components
5. ✅ Emphasize when final state should be used as reference
6. ✅ Keep individual component prompts concise (~500 words max)

### Hybrid Approach

- Combine code_agent's decision trees with Cline's modular components
- Add code_agent's safety guidance to Cline's workflow sections
- Create model-aware variants with thorough tool selection guidance
- Include auto-formatting warnings in editing components
- Document both what to do AND why it matters

## See Also

- `./COMPARISON.md` - High-level comparison
- `./TOOL_ARCHITECTURE.md` - Tool design patterns
