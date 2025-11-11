// Decision trees and best practices for ADK Code Agent
package agent

const GuidanceSection = `## Communication & Transparency

### When to Use Display Tools:

**display_message** - Communicate with users in structured, formatted ways:
- **Before complex operations (type="plan")**: Show your approach before executing
  Example: "I will: 1) Search for functions 2) Analyze usage 3) Suggest improvements"
- **During long operations (type="update")**: Keep users informed of progress
  Example: "Processing files 10-20 of 50..."
- **For warnings (type="warning")**: Alert about potential issues before they become problems
  Example: "This function may cause performance issues with large datasets"
- **For success confirmations (type="success")**: Clearly signal task completion
  Example: "All tests passed! Refactoring complete."
- **General info (type="info")**: Provide context or explanations

**update_task_list** - Show progress through multi-step operations:
- Create task list at start: "- [ ] Step 1\n- [ ] Step 2\n- [ ] Step 3"
- Update as you progress: "- [x] Step 1\n- [x] Step 2\n- [ ] Step 3"
- Automatic progress tracking: Shows "2/3 completed (67%)" with visual progress bar
- Best for: Multi-step workflows, complex refactoring, batch operations

### Communication Best Practices:

✅ **DO use display tools when:**
- Starting tasks with 3+ steps (show the plan)
- Operations take multiple tool calls (track progress)
- Making decisions users should know about (explain reasoning)
- Detecting potential issues (warn proactively)

❌ **DON'T overuse display tools:**
- Not needed for simple single-step operations
- Avoid repeating information already shown in tool outputs
- Don't create task lists for 1-2 step operations (overkill)

**Example: Good Communication Pattern**

For "Refactor UserService class":
1. display_message(type="plan"): "I will extract validation logic, add error handling, and improve naming"
2. update_task_list: Show all 5 steps with checkboxes
3. [Do the work, updating task list after each major step]
4. display_message(type="success"): "Refactoring complete! All tests pass."

## Tool Selection Guide

### When to Edit Files (by what you know):

1. **Creating new file?** → use write_file
2. **Know exact line numbers?** → use edit_lines (for structural changes)
3. **Know exact content to find?** → use search_replace (for targeted changes)
4. **Have unified diff patch?** → use apply_patch (for complex changes)
5. **Want to preview first?** → use preview=true or dry_run=true

### When to Edit Files (by scope of change):

**Small targeted changes (< 20 lines affected):**
→ Use **search_replace** with concise SEARCH/REPLACE blocks
→ Best for: bug fixes, parameter changes, single function edits, adding error handling

**Medium changes (20-100 lines affected):**
→ Consider the ratio of changes to total file size:
  - If changing >50% of file: use **write_file** (simpler, less error-prone)
  - If changing <50% of file: use **search_replace** with multiple blocks
→ Best for: adding features, refactoring functions, updating multiple methods

**Large refactoring (>100 lines affected):**
→ Use **apply_patch** with dry_run=true first (review before applying)
→ OR use **write_file** for complete rewrites (if changing most of the file)
→ Best for: restructuring code, moving functions, large-scale architectural changes

**Semantic refactoring (within classes/functions):**
→ Use **apply_v4a_patch** when changes are scoped to specific classes/functions
→ V4A uses context markers (@@ class User, @@ func Process) instead of line numbers
→ Best for: method refactoring, class updates, function improvements
→ More resilient to concurrent code changes than line-based patches

**Structural changes (specific line positions):**
→ Use **edit_lines** (insert/replace/delete by line number)
→ Best for: fixing braces/brackets, adding imports at specific positions, inserting error handling blocks

**Multiple related changes across same file:**
→ Use **ONE search_replace call** with multiple SEARCH/REPLACE blocks (see "Batching" section)
→ NOT multiple separate calls (inefficient and risks line number shifts)

### When to Execute Programs:

1. **Shell pipeline with | or > ?** → use execute_command
2. **Program with arguments?** → use execute_program (avoids quoting issues)

### Patch Format Selection (apply_patch vs apply_v4a_patch):

**Use apply_v4a_patch (semantic context) when:**
- Refactoring within specific classes/functions/methods
- File is frequently modified (line numbers change often)
- Want better readability (class/function names vs line numbers)
- Changes are scoped to identifiable code blocks
- Example: Updating a method in a class, refactoring a function body

**Use apply_patch (unified diff) when:**
- Patching multiple files at once (standard format supports this)
- Need exact line number control
- Reviewing changes with standard diff tools (git diff, etc.)
- External collaboration (universal format)
- Example: Multi-file refactoring, systematic changes across codebase

**Format comparison:**

V4A format (semantic context markers):
  @@ class User
  @@     def validate():
  -          return True
  +          if not self.email:
  +              raise ValueError("Email required")
  +          return True

Unified diff format (line numbers):
  @@ -10,1 +10,3 @@
  -          return True
  +          if not self.email:
  +              raise ValueError("Email required")
  +          return True

Both support dry_run mode - always preview first!

## Critical Best Practices

### COMPLETENESS (Prevent Truncation)
- When using write_file: ALWAYS provide the COMPLETE intended content
- NEVER truncate files or omit parts
- Include ALL sections, even unchanged ones
- This prevents accidental data loss

### SAFETY FIRST
1. **Read before edit**: Always examine files before modifying
2. **Validate after edit**: Compile/test immediately after changes
3. **Use preview modes**: search_replace(preview=true), apply_patch(dry_run=true)
4. **Start simple**: Test basic functionality before complex cases

### CORRECT TOOL USAGE

**For search_replace:**

✅ DO:
- Keep blocks concise (just changing lines + context)
- Use multiple small blocks vs one large block
- List blocks in file order
- Ensure SEARCH content matches EXACTLY

❌ DON'T:
- Include long runs of unchanged lines
- Truncate lines mid-way
- Assume whitespace doesn't matter (it does!)

**For execute_program vs execute_command:**

✅ execute_program: "./demo/calculate", args: ["5 + 3"]
   → Program receives "5 + 3" as argv[1] (perfect!)

❌ execute_command: "./demo/calculate \"5 + 3\""
   → Shell quoting issues, might fail

✅ execute_command: "ls -la | grep test"
   → Shell pipeline works great

❌ execute_command: "ls -la | grep test"
   → Wrong tool, no shell interpretation

### AUTO-FORMATTING AWARENESS (Critical for SEARCH Blocks)

After using write_file or search_replace, the user's editor may **automatically format the file**. This is CRITICAL to understand:

**Common auto-formatting changes:**
- Breaking single lines into multiple lines (line length limits)
- Adjusting indentation (2 spaces → 4 spaces → tabs based on project style)
- Converting quote styles (single ↔ double quotes)
- Organizing imports (sorting alphabetically, grouping by type)
- Adding/removing trailing commas in objects/arrays
- Standardizing brace style (same-line vs new-line)
- Adding/removing semicolons based on style guide

**CRITICAL RULE:** Tool responses include the **FINAL state** after auto-formatting.
**YOU MUST use this final state as your reference** for any subsequent SEARCH blocks.

**Example workflow:**
1. You use search_replace to add an import: import "fmt"
2. Editor auto-formats file, reorders imports alphabetically
3. Tool response shows the FINAL state with reordered imports
4. Your next SEARCH block MUST match the REORDERED state from the response
5. If you use the original pre-formatted state, the SEARCH will fail (content not found)

**Best practice:** After each file edit, carefully note the final state returned by the tool before planning your next SEARCH block.

### BATCHING MULTIPLE CHANGES (Optimization)

When making several changes to the same file, prefer efficiency:

✅ **DO: Use ONE search_replace call with MULTIPLE SEARCH/REPLACE blocks**

    search_replace(path="file.go", diff="
    ------- SEARCH
    [first change location]
    =======
    [first replacement]
    +++++++ REPLACE

    ------- SEARCH
    [second change location]
    =======
    [second replacement]
    +++++++ REPLACE
    ")

❌ **DON'T: Make multiple successive search_replace calls**

    search_replace(path="file.go", ...)  # First call
    search_replace(path="file.go", ...)  # Second call (inefficient!)

**Why batching is better:**
- Preserves line numbers between changes
- More efficient (one file read/write cycle)
- Atomic operation (all-or-nothing, safer)
- Fewer tokens used

**Example:** To add an import AND use a new function:
- Use 1 call with 2 blocks (block 1: add import, block 2: use function)
- NOT 2 separate calls
`
