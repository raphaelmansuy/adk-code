# code_agent vs Cline: Code Tools & Prompts Comparison

## Executive Summary

Both systems solve the same core problem—enabling AI agents to reliably edit code and files—but take fundamentally different architectural approaches. **code_agent** (Go, ADK) prioritizes safety and explicit tool guidance, while **Cline** (TypeScript, MCP) prioritizes modularity and model awareness.

## Quick Comparison Table

| Aspect | code_agent | Cline |
|--------|-----------|-------|
| **Language** | Go (compiled) | TypeScript (runtime) |
| **Framework** | Google ADK (llmagent) | MCP protocol |
| **Primary Tools** | read, write, search_replace, edit_lines, apply_patch | read, write, replace_in_file, apply_patch |
| **Edit Tool** | SEARCH/REPLACE (+ edit_lines for line-based) | SEARCH/REPLACE (only) |
| **Patch Format** | Unified diff (standard RFC 3881) | Custom V4A format |
| **Safety Feature** | Size validation prevents >90% truncation | Auto-formatting warnings |
| **Whitespace Match** | Exact + line-trimmed fallback | Exact only |
| **Model Variants** | Single prompt | 3+ variants (GENERIC, NATIVE_GPT_5, NATIVE_NEXT_GEN) |
| **Tool Registry** | Inline in code | Config-based registry |
| **Extensibility** | Simple tool.Tool interface | MCP server integration |
| **Prompt Architecture** | Monolithic enhanced_prompt | Component-based (PromptBuilder) |

## Key Differences

### 1. File Editing Strategy

**code_agent:**

- Four distinct edit tools for different scenarios
- `edit_lines`: Line-number based editing (replace/insert/delete)
- `search_replace`: SEARCH/REPLACE blocks with fallback matching
- `apply_patch`: Unified diff patches
- `write_file`: Full-file rewrites with size safety check

**Cline:**

- Two primary edit tools
- `write_to_file`: Full-file rewrites
- `replace_in_file`: SEARCH/REPLACE blocks
- No line-based editing tool
- Applies auto-formatting after edits (important workflow consideration)

**Winner:** code_agent has more specialized tools; Cline is more minimal but sufficient.

### 2. Patch Format Approach

**code_agent** uses standard RFC 3881 unified diff:

```diff
@@ -10,5 +12,7 @@
```

Advantage: Can leverage existing diff tools, familiar to most developers.

**Cline** uses custom V4A format:

```diff
@@ class BaseClass
@@     def method():
```

Advantage: More context-aware, better handles complex refactoring with class/function markers.

### 3. Safety Mechanisms

| Mechanism | code_agent | Cline |
|-----------|-----------|-------|
| Size validation | ✅ Rejects >90% truncation (prevents data loss) | ❌ Warns but doesn't prevent |
| Atomic writes | ✅ Built-in to write_file | ❌ Not mentioned |
| Whitespace tolerance | ✅ Line-trimmed fallback matching | ❌ Exact matching only |
| Preview modes | ✅ search_replace/apply_patch/edit_lines | ❌ Not emphasized |
| Auto-format awareness | ❌ Silent about formatters | ✅ Explicitly warns editors |

**Winner:** code_agent for compiled safety, Cline for real-world workflow awareness.

### 4. Prompt/Tool Documentation

**code_agent** (enhanced_prompt.go):

- Extensive "Tool Selection Guide" (which tool for which job)
- "Critical Best Practices" section with specific rules
- "Common Pitfalls & Solutions" with examples
- "Safety Features (Our Advantages)" section
- Compares vs other agents

**Cline** (components/editing_files.ts):

- Separate "write_to_file vs replace_in_file" decision guide
- Mentions auto-formatting behavior
- Emphasizes single replace_in_file call with multiple blocks
- Less detailed error recovery guidance

**Winner:** code_agent for thoroughness, Cline for workflow clarity.

## Architectural Patterns

### code_agent: Direct Implementation

```go
type ReadFileInput struct {
    Path string
    Offset *int
    Limit *int
}

handler := func(ctx tool.Context, input ReadFileInput) ReadFileOutput { ... }
functiontool.New(functiontool.Config{
    Name: "read_file",
    Description: "...",
}, handler)
```

**Pros:** Simple, explicit, no indirection
**Cons:** Harder to add variants, not model-aware

### Cline: Registry Pattern

```typescript
const generic: ClineToolSpec = {
    variant: ModelFamily.GENERIC,
    id: "read_file",
    description: "...",
    parameters: [...]
}

const NATIVE_GPT_5: ClineToolSpec = {
    variant: ModelFamily.NATIVE_GPT_5,
    id: "read_file",
    description: "...",  // Different wording
    parameters: [...]
}
```

**Pros:** Model-aware, easy to add variants, centralized registry
**Cons:** More boilerplate, indirection through PromptBuilder

## Surprising Findings

1. **code_agent's whitespace tolerance is more sophisticated** than Cline's approach
   - Uses line-trimmed matching as fallback (handles minor indentation differences)
   - Cline requires exact matches, which can fail with auto-formatters

2. **code_agent's size validation is clever but Cline-specific workflows don't have it**
   - Prevents accidental data loss from truncation
   - Cline's auto-formatting warning is the real-world equivalent

3. **Cline's component-based system prompt is more maintainable**
   - Each tool/concept gets own TypeScript file
   - code_agent's monolithic prompt is longer but more cohesive

4. **Both use SEARCH/REPLACE blocks but with different philosophies**
   - code_agent: Emphasize using multiple blocks, keep concise
   - Cline: Emphasize single call with multiple blocks (optimization hint)

5. **code_agent addresses a gap Cline doesn't: line-based editing**
   - Useful for structural changes (fixing braces, adding imports)
   - Cline relies on SEARCH/REPLACE for this

## Winning Strategies from Each

### From code_agent

- ✅ Build safety checks into tools (size validation)
- ✅ Provide detailed "when to use this tool" guidance
- ✅ Document common pitfalls with recovery steps
- ✅ Implement whitespace-tolerant matching as fallback
- ✅ Offer line-based editing for structural changes
- ✅ Use atomic writes for reliability

### From Cline

- ✅ Make tools model-aware (different prompts for different LLMs)
- ✅ Use registry pattern for easy variant management
- ✅ Warn about real-world editor behavior (auto-formatting)
- ✅ Emphasize workflow optimization (single call vs multiple)
- ✅ Integrate with MCP for extensibility
- ✅ Component-based prompt architecture for modularity

## Integration Recommendations

If building a hybrid system:

1. **Use code_agent's safety mechanisms** (size validation, atomic writes)
2. **Adopt Cline's component-based prompt** for maintainability
3. **Support both patch formats** (unified diff + V4A)
4. **Add model-aware variants** for different LLM families
5. **Include line-based editing** alongside SEARCH/REPLACE
6. **Warn about auto-formatting** behavior after edits

## See Also

- `./tool_architecture.md` - Deep dive into tool design patterns
- `./prompt_strategy.md` - Analysis of system prompt approaches
