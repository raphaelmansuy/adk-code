# Test Scenarios for Enhanced Prompt Improvements

## Scenario 1: Auto-Formatting Awareness Test

**Test Case:** Add an import to a Go file, then add code that uses it

**Expected Behavior (with improvements):**
1. Agent uses search_replace to add import
2. Editor auto-formats (reorders imports)
3. Agent reads final state from tool response
4. Agent uses REORDERED state for next SEARCH block
5. Second edit succeeds

**Previous Behavior (without improvements):**
- Agent would use original pre-formatted state
- SEARCH block would fail (content not found)
- Agent would be confused

**Test File:** Create a simple Go file with imports:
```go
package main

import (
    "os"
)

func main() {
    // TODO: add code here
}
```

**Agent Task:** "Add an import for 'fmt' and use fmt.Println to print hello"

**Expected Tool Calls:**
1. read_file to see current state
2. search_replace with 2 blocks:
   - Block 1: Add fmt import
   - Block 2: Add fmt.Println call
3. Should succeed in single call (batching optimization)

---

## Scenario 2: Batching Optimization Test

**Test Case:** Add import + add code that uses it

**Expected Behavior (with improvements):**
- Agent uses ONE search_replace call with 2 blocks
- More efficient, atomic operation

**Previous Behavior:**
- Agent might make 2 separate search_replace calls
- Inefficient, risk of inconsistency

**Test File:**
```go
package calculator

func Add(a, b int) int {
    return a + b
}
```

**Agent Task:** "Add error handling: import 'errors', check if inputs are negative"

**Expected Tool Calls:**
1. read_file
2. search_replace with ONE call, 2-3 blocks:
   - Add import errors
   - Modify function signature to return error
   - Add validation logic

---

## Scenario 3: Scope-Based Tool Selection Test

**Test Case:** Agent chooses correct tool based on change scope

**Small change (< 20 lines):**
- Task: "Fix typo in variable name 'mesage' → 'message'"
- Expected: search_replace (not write_file or edit_lines)

**Medium change (20-100 lines, <50% of file):**
- Task: "Refactor calculateTotal function to use a switch statement"
- Expected: search_replace with multiple blocks

**Large change (>50% of file):**
- Task: "Restructure entire file to follow MVC pattern"
- Expected: write_file OR apply_patch with dry_run first

**Structural change:**
- Task: "Add import at line 5"
- Expected: edit_lines (knows exact line number)

---

## Scenario 4: Real-World Workflow

**Test Case:** Complete feature implementation with auto-formatting

**Initial File (main.go):**
```go
package main

import "os"

func main() {
    args := os.Args
    // TODO: process args
}
```

**Agent Task:** "Add flag parsing using the flag package"

**Expected Workflow:**
1. read_file to understand structure
2. search_replace with batched blocks:
   - Add import "flag" (will auto-format with imports)
   - Add flag definitions
   - Add flag.Parse() call
3. Tool returns final state with auto-formatted imports
4. Agent notes the formatted state
5. If further edits needed, uses formatted state

**Success Criteria:**
- Uses batched search_replace (optimization)
- Handles auto-formatted import ordering correctly
- No failed SEARCH blocks due to formatting mismatch

---

## Verification Commands

Run these to test the enhanced prompt:

```bash
cd /Users/raphaelmansuy/Github/03-working/adk_training_go/code_agent

# Build the agent
make build

# Test scenario 1: Auto-formatting
./code-agent
# User prompt: "Create a Go file test.go with os import, then add fmt import and use it"

# Test scenario 2: Batching
# User prompt: "Create calculator.go, then add error handling with errors package"

# Test scenario 3: Tool selection
# User prompt: "Fix typo in test.go: 'mesage' → 'message'"
# User prompt: "Restructure test.go to split into multiple functions"

# Test scenario 4: Real workflow
# User prompt: "Create a CLI tool main.go that uses flag package for args"
```

---

## Expected Improvements

### Before Enhancements:
- ❌ SEARCH blocks fail after auto-formatting
- ❌ Multiple inefficient tool calls for related changes
- ❌ Poor tool selection (using write_file for small edits)
- ❌ No guidance on change scope

### After Enhancements:
- ✅ Agent aware of auto-formatting, uses final state
- ✅ Batches multiple changes in one call
- ✅ Better tool selection based on scope
- ✅ Clear decision trees for when to use each tool
- ✅ More efficient workflows
- ✅ Fewer failed operations

---

## Metrics to Track

1. **Success Rate:** Percentage of SEARCH blocks that succeed on first try
2. **Tool Call Efficiency:** Average tool calls per task
3. **Batching Adoption:** Percentage of multi-edit scenarios using single call
4. **Tool Selection Accuracy:** Correct tool chosen for task scope
5. **Auto-Format Handling:** No failures due to formatting mismatches

---

## Next Steps

If these improvements prove successful:

1. **Phase 2:** Implement model-aware variants
   - Detect model family (Gemini, GPT, Claude)
   - Adjust tool descriptions per model
   - Use absolutePath for native models

2. **Phase 3:** Add V4A patch format support
   - Implement custom patch parser
   - Add @@ class/function context markers
   - Provide both unified diff and V4A options

3. **Phase 4:** Component-based prompt system
   - Split enhanced_prompt.go into modules
   - Easier to maintain and customize
   - Can enable/disable components
