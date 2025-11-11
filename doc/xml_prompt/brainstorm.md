# XML-Tagged Dynamic Prompt Construction

## Core Principle

Use XML tags to create **hierarchical, contextual, and scannable** prompts that LLMs can parse more effectively than flat text.

## Key Benefits

1. **Structure visibility** - LLM can identify sections at a glance
2. **Conditional inclusion** - Show/hide sections based on context
3. **Priority signaling** - Critical vs optional information
4. **Easier parsing** - LLM pattern-matches on tags vs prose
5. **Maintainability** - Modular, testable prompt components

## Proposed Structure

```xml
<agent_identity>
  Core role and capabilities
</agent_identity>

<workspace_context>
  <file_system>
    Current directory, recent files, structure
  </file_system>
  <vcs_state>
    Git branch, uncommitted changes
  </vcs_state>
</workspace_context>

<tools>
  <tool_category name="File Operations" priority="high">
    <tool name="read_file">
      <description>...</description>
      <when_to_use>...</when_to_use>
      <parameters>...</parameters>
    </tool>
  </tool_category>
</tools>

<decision_trees>
  <scenario type="file_editing">
    <if condition="new file">→ write_file</if>
    <if condition="small change">→ search_replace</if>
    <if condition="large refactor">→ apply_patch</if>
  </scenario>
</decision_trees>

<critical_rules priority="must_follow">
  <rule id="completeness">Never truncate file content</rule>
  <rule id="safety">Always read before editing</rule>
</critical_rules>

<best_practices>
  <practice context="search_replace">
    Keep SEARCH blocks exact, include whitespace
  </practice>
</best_practices>

<common_pitfalls>
  <pitfall issue="auto-formatting">
    Tool responses show post-format state - use that as reference
  </pitfall>
</common_pitfalls>
```

## Dynamic Composition Strategy

### 1. Base Template (Always Included)

```go
type PromptTemplate struct {
    Identity      string // Role definition
    Tools         []ToolSection
    CriticalRules []Rule
}
```

### 2. Contextual Sections (Conditionally Included)

```go
type ContextualPrompt struct {
    WorkspaceInfo  *WorkspaceSection // If in a workspace
    RecentErrors   *ErrorSection     // If previous errors
    TaskHistory    *HistorySection   // If multi-turn session
    SpecialGuidance *GuidanceSection  // If specific task type detected
}
```

### 3. Priority-Based Rendering

```go
func RenderPrompt(base, contextual, priority string) string {
    // High priority items go in <critical_rules>
    // Medium priority in <best_practices>
    // Low priority omitted if token budget tight
}
```

## Practical Implementation

### Current Code

```go
// Static concatenation
const EnhancedSystemPrompt = ToolsSection + "\n" + GuidanceSection + ...
```

### Proposed Refactor

```go
type PromptBuilder struct {
    sections map[string]Section
}

func (pb *PromptBuilder) Build(ctx Context) string {
    var buf strings.Builder
    
    buf.WriteString("<agent_identity>\n")
    buf.WriteString(pb.sections["identity"].Render())
    buf.WriteString("</agent_identity>\n\n")
    
    if ctx.HasWorkspace {
        buf.WriteString("<workspace_context>\n")
        buf.WriteString(renderWorkspace(ctx.Workspace))
        buf.WriteString("</workspace_context>\n\n")
    }
    
    buf.WriteString("<tools>\n")
    for _, category := range ctx.ToolCategories {
        buf.WriteString(renderToolCategory(category))
    }
    buf.WriteString("</tools>\n\n")
    
    buf.WriteString("<decision_trees>\n")
    buf.WriteString(renderDecisionTrees(ctx.TaskType))
    buf.WriteString("</decision_trees>\n\n")
    
    buf.WriteString("<critical_rules priority=\"must_follow\">\n")
    for _, rule := range getCriticalRules(ctx) {
        buf.WriteString(renderRule(rule))
    }
    buf.WriteString("</critical_rules>\n")
    
    return buf.String()
}
```

## Tag Design Principles

### 1. Use Semantic Tags

✅ `<when_to_use>` - Clear intent
❌ `<section_3>` - Meaningless structure

### 2. Add Attributes for Metadata

```xml
<rule priority="critical" enforcement="strict">
<tool category="file_ops" complexity="low">
<guidance context="refactoring" skill_level="intermediate">
```

### 3. Nest Logically

```xml
<tools>
  <category name="Files">
    <tool name="read_file">
      <when_to_use>Need file contents</when_to_use>
    </tool>
  </category>
</tools>
```

### 4. Keep Tags Brief

✅ `<critical_rules>`
❌ `<these_are_the_critical_rules_that_must_be_followed>`

## Migration Path

### Phase 1: Wrapper (No Breaking Changes)

```go
func WrapInXML(existingContent string) string {
    return fmt.Sprintf("<guidance>\n%s\n</guidance>", existingContent)
}
```

### Phase 2: Granular Sections

```go
func BuildToolsXML(registry *ToolRegistry) string {
    var buf strings.Builder
    buf.WriteString("<tools>\n")
    for _, cat := range registry.Categories {
        buf.WriteString(fmt.Sprintf("  <category name=%q>\n", cat.Name))
        // ... render tools
        buf.WriteString("  </category>\n")
    }
    buf.WriteString("</tools>\n")
    return buf.String()
}
```

### Phase 3: Full Dynamic Composition

- Context-aware section selection
- Priority-based inclusion
- Token budget optimization

## Key Insights from Claude's Approach

1. **Hierarchical thinking** - LLMs parse nested structures naturally
2. **Explicit boundaries** - Clear start/end of instructions
3. **Self-documenting** - Tag names convey purpose
4. **Selective attention** - LLM can "zoom in" on relevant tags
5. **Parallel processing** - Multiple `<tool>` blocks can be scanned simultaneously

## Anti-Patterns to Avoid

❌ Over-nesting (>4 levels deep)
❌ Inconsistent tag naming (camelCase vs snake_case)
❌ Mixing XML with markdown headings (choose one style)
❌ Verbose content in tags (defeats scanability)
❌ Custom XML schema (keep it simple, not full XML spec)

## Validation Strategy

```go
func ValidatePromptStructure(prompt string) error {
    // Simple tag matching (not full XML parsing)
    stack := []string{}
    for _, tag := range extractTags(prompt) {
        if isOpenTag(tag) {
            stack = append(stack, tag)
        } else if isCloseTag(tag) {
            if len(stack) == 0 || !matches(stack[len(stack)-1], tag) {
                return fmt.Errorf("mismatched tag: %s", tag)
            }
            stack = stack[:len(stack)-1]
        }
    }
    if len(stack) > 0 {
        return fmt.Errorf("unclosed tags: %v", stack)
    }
    return nil
}
```

## Testing Approach

```go
func TestPromptRendering(t *testing.T) {
    ctx := Context{
        HasWorkspace: true,
        TaskType: "refactoring",
    }
    
    prompt := builder.Build(ctx)
    
    // Verify structure
    assert.Contains(t, prompt, "<workspace_context>")
    assert.Contains(t, prompt, "</workspace_context>")
    
    // Verify no unclosed tags
    assert.NoError(t, ValidatePromptStructure(prompt))
    
    // Verify priority content included
    assert.Contains(t, prompt, "<critical_rules")
}
```

## Example: Before vs After

### Before (Current)

```
## Tool Selection Guide

### When to Edit Files (by what you know):

1. **Creating new file?** → use write_file
2. **Know exact line numbers?** → use edit_lines
```

### After (XML-Tagged)

```xml
<decision_trees>
  <scenario name="file_editing" context="when_you_know">
    <decision>
      <condition>Creating new file</condition>
      <action>write_file</action>
    </decision>
    <decision>
      <condition>Know exact line numbers</condition>
      <action>edit_lines</action>
    </decision>
  </scenario>
</decision_trees>
```

## Recommendation

**Start simple**: Wrap existing sections in semantic XML tags without changing content.

```go
func BuildEnhancedPromptV2(registry *ToolRegistry) string {
    return fmt.Sprintf(`<agent_system_prompt>
<tools>
%s
</tools>

<guidance>
%s
</guidance>

<critical_rules>
%s
</critical_rules>

<workflows>
%s
</workflows>
</agent_system_prompt>`,
        BuildToolsSection(registry),
        GuidanceSection,
        extractCriticalRules(),
        WorkflowSection,
    )
}
```

**Iterate**: Add attributes, refine nesting, implement conditional inclusion based on actual usage patterns.

**Measure**: Compare agent performance with flat vs XML prompts on real tasks.
