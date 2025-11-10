# Edit/Patch Tool Improvements - Based on Real Execution Trace Analysis

**Date**: November 10, 2025  
**Issue**: Real-world execution revealed critical fragility in edit tools

---

## Problem Analysis from Execution Trace

### Critical Issues Identified

#### 1. Whitespace Sensitivity in String Replacement ❌
**Problem**: Tool fails when strings contain literal newlines or escaped newlines

```
replace_in_file with:
  old_text: "return 1;\n    }"
  new_text: "return 1;\n    }"
→ Result: "Text to replace not found"
```

**Root Cause**: Trying to match literal `\n` characters in JSON strings instead of actual newlines

#### 2. Patch Application Doesn't Verify Fix ❌
**Problem**: Patch was applied but syntax error remained

```
Tool call: apply_patch with closing brace fix
→ Result: lines_added: 2, lines_removed: 1, success: true
→ Compile result: STILL HAS ERROR (expected '}')
```

**Root Cause**: Patch parser's fuzzy matching didn't properly align with actual file structure

#### 3. No Context for String Replacement ❌
**Problem**: When multiple similar lines exist, wrong ones can be replaced

```
Multiple "return 1;" statements in file
→ replace_in_file can't target specific one
→ May replace wrong occurrence
```

#### 4. Catastrophic Replace Without Confirmation ❌
**Problem**: Agent issued `replace_in_file` with empty new_text, deleting 5 occurrences!

```
replace_in_file:
  old_text: "        free(expression_copy); // Free allocated memory"
  new_text: ""
→ Result: "Successfully replaced 5 occurrence(s)"
→ File corrupted - memory leak introduced!
```

**Root Cause**: No safeguards, validation, or preview for dangerous operations

#### 5. Line-by-Line Edits Are Impossible ❌
**Problem**: Can't easily fix syntax errors that require precise line modifications

```
Problem: Missing closing brace at end of file
Solution needed: Add "}" on new line at line 108
Current tools: No way to do this precisely
```

---

## Recommended Improvements

### Phase 2A: Critical Safety Improvements (Immediate)

#### 1. Add Replace Safeguards ⚠️
```go
type ReplaceInFileInput struct {
    Path             string `json:"path"`
    OldText          string `json:"old_text"`
    NewText          string `json:"new_text"`
    MaxReplacements  *int   `json:"max_replacements,omitempty"` // NEW
    ExpectExactMatch *bool  `json:"expect_exact_match,omitempty"` // NEW
    Preview          *bool  `json:"preview,omitempty"` // NEW
}
```

**Implementation**:
- Limit replacements to prevent catastrophic deletes
- Reject if replacements exceed expected count
- Optional preview-only mode
- Warnings for empty new_text

#### 2. Enhanced Whitespace Handling 🔧
```go
// Current problem: old_text with "\n" doesn't match actual newlines
// Solution: Auto-normalize whitespace

func normalizeWhitespace(s string) string {
    // Replace \\n with actual \n
    // Trim trailing whitespace
    // Handle different line endings
}
```

#### 3. Line-Based Editing Tool (NEW) ✨
```go
type EditLinesInput struct {
    FilePath  string `json:"file_path"`
    StartLine int    `json:"start_line"`    // 1-indexed
    EndLine   int    `json:"end_line"`      // Inclusive
    NewLines  string `json:"new_lines"`     // Replacement content
    Mode      string `json:"mode,omitempty"` // "replace", "insert", "delete"
}

type EditLinesOutput struct {
    Success      bool   `json:"success"`
    LinesModified int  `json:"lines_modified"`
    Message      string `json:"message"`
    Preview      string `json:"preview,omitempty"`
    Error        string `json:"error,omitempty"`
}
```

**Benefits**:
- Precise line-number-based edits
- No ambiguity about which lines to modify
- Perfect for syntax errors and structure changes
- Easy to preview

#### 4. Improved Patch Parser 🎯
**Current Issues**:
- Fuzzy matching is too loose
- Context lines not properly validated
- Can't detect when patch is misaligned

**Improvements**:
```go
type ApplyPatchInput struct {
    FilePath  string `json:"file_path"`
    Patch     string `json:"patch"`
    DryRun    *bool  `json:"dry_run"`
    Strict    *bool  `json:"strict"`        // ENHANCED
    Verbose   *bool  `json:"verbose"`       // NEW: Show detailed matching
    FuzzyFuzz *int   `json:"fuzz_factor"`   // NEW: Allow N lines offset
}

// Validation: Check that patch actually compiles/parses after apply
type ApplyPatchOutput struct {
    Success       bool   `json:"success"`
    LinesAdded    int    `json:"lines_added"`
    LinesRemoved  int    `json:"lines_removed"`
    Preview       string `json:"preview"`
    ValidationMsg string `json:"validation_msg,omitempty"` // NEW
    Error         string `json:"error"`
}
```

#### 5. Verification Step (NEW) ✨
```go
type VerifyEditInput struct {
    FilePath  string `json:"file_path"`
    TestCmd   string `json:"test_cmd,omitempty"`   // compile/run command
    ExpectPattern string `json:"expect_pattern,omitempty"` // regex
    CheckSyntax *bool `json:"check_syntax,omitempty"`
}

type VerifyEditOutput struct {
    Success      bool   `json:"success"`
    Message      string `json:"message"`
    Changes      string `json:"changes,omitempty"`  // What was changed
    Verification string `json:"verification,omitempty"` // What was verified
    Error        string `json:"error,omitempty"`
}
```

---

### Phase 2B: Advanced Features

#### 6. Semantic Patch Support (Advanced)
```go
// Allow higher-level operations like:
// "add closing brace before line N"
// "convert function parameter type from X to Y"
// "add error check for function call"

type SemanticEditInput struct {
    FilePath    string `json:"file_path"`
    Language    string `json:"language"` // "go", "c", "python", etc
    Operation   string `json:"operation"` // "add_brace", "convert_type", etc
    Target      string `json:"target"` // specific location/pattern
    Params      map[string]string `json:"params"`
}
```

#### 7. Multi-File Patch Support (Advanced)
```go
// Support unified diff format that modifies multiple files
type ApplyMultiPatchInput struct {
    BasePath string `json:"base_path"`
    Patch    string `json:"patch"` // Multi-file unified diff
    DryRun   *bool  `json:"dry_run"`
}
```

---

## Implementation Priority

### Immediate (Week 1) - CRITICAL
1. ✅ Add `max_replacements` to prevent catastrophic deletes
2. ✅ Add empty `new_text` validation
3. ✅ Improve whitespace normalization
4. ✅ Add preview mode to replace_in_file
5. ✅ Better error messages with suggestions

### Short-term (Week 2) - IMPORTANT
1. ✅ Implement `edit_lines` tool for line-based edits
2. ✅ Enhance patch parser with validation
3. ✅ Add verification step support
4. ✅ Implement context-aware replacements

### Medium-term (Weeks 3-4) - NICE-TO-HAVE
1. ⭕ Semantic edit support
2. ⭕ Multi-file patch support
3. ⭕ Smart language-specific editing

---

## Specific Fixes from Trace

### Issue 1: Whitespace-sensitive replacement
**Current Behavior**:
```
old_text: "return 1;\n    }" (literal \n)
→ Doesn't match actual newline in file
```

**Fix**:
```go
func NormalizeOldText(text string) string {
    // Convert \\n to actual \n
    text = strings.ReplaceAll(text, "\\n", "\n")
    text = strings.ReplaceAll(text, "\\t", "\t")
    return text
}
```

### Issue 2: Catastrophic replacement
**Current Behavior**:
```
replace_in_file with new_text=""
→ Deletes all matching lines without warning
→ Result: 5 occurrences deleted!
```

**Fix**:
```go
if input.NewText == "" && !input.AllowEmpty {
    return ReplaceInFileOutput{
        Success: false,
        Error: "Refusing to delete lines: new_text is empty. " +
               "Use mode='delete' in edit_lines tool for intentional deletion.",
    }
}

if input.MaxReplacements != nil && replacementCount > *input.MaxReplacements {
    return ReplaceInFileOutput{
        Success: false,
        Error: fmt.Sprintf(
            "Too many replacements (%d). " +
            "Expected max %d. Use preview mode first.",
            replacementCount, *input.MaxReplacements,
        ),
    }
}
```

### Issue 3: Syntax error remains after patch
**Current**: Patch applied but file still doesn't compile

**Solution**: Add post-apply validation
```go
if input.VerifyCmd != "" {
    // Run compile/syntax check after applying patch
    result, _ := executeCommand(input.VerifyCmd)
    if result.ExitCode != 0 {
        return ApplyPatchOutput{
            Success: false,
            Error: fmt.Sprintf(
                "Patch applied but verification failed: %s",
                result.Stderr,
            ),
        }
    }
}
```

---

## Code Examples

### Example 1: Safe Replace with Safeguards
```go
// GOOD: Safe replacement with limits
Tool: replace_in_file
{
    "path": "demo/calculate.c",
    "old_text": "    printf(\"%f\\n\", result);",
    "new_text": "    printf(\"%.2f\\n\", result);",
    "max_replacements": 1,
    "preview": true
}

// Result: 
// - Preview shows 1 match
// - Confirms it's the right location
// - Proceeds with single replacement
```

### Example 2: Line-Based Editing
```go
// BETTER: Using new edit_lines tool for syntax fixes
Tool: edit_lines
{
    "file_path": "demo/calculate.c",
    "start_line": 107,
    "end_line": 107,
    "new_lines": "    return 0;\n}",
    "mode": "replace"
}

// Result:
// - Replaces lines 107-107 with new content
// - Exact line numbers = no ambiguity
// - Clear what changed
```

### Example 3: Patch with Verification
```go
Tool: apply_patch
{
    "file_path": "demo/calculate.c",
    "patch": "[unified diff]",
    "dry_run": false,
    "verify_cmd": "gcc demo/calculate.c -o /tmp/test 2>&1"
}

// Result:
// - Applies patch
// - Runs verification command
// - Returns error if verification fails
// - Prevents applying broken patches
```

---

## Test Cases for New Tools

### Test: Replace Safeguard - Empty new_text
```go
func TestReplaceInFile_EmptyNewText_Rejected(t *testing.T) {
    // Should reject with helpful error
    input := ReplaceInFileInput{
        Path: tmpFile,
        OldText: "test",
        NewText: "",  // DANGER
    }
    output := handler(ctx, input)
    assert.Equal(t, false, output.Success)
    assert.Contains(t, output.Error, "empty")
}
```

### Test: Replace Safeguard - Too Many Replacements
```go
func TestReplaceInFile_MaxReplacements_Enforced(t *testing.T) {
    // Create file with 5 "return 1;" lines
    // Try to replace with max=1
    // Should fail
    input := ReplaceInFileInput{
        Path: tmpFile,
        OldText: "return 1;",
        NewText: "return 1; // fixed",
        MaxReplacements: intPtr(1),
    }
    output := handler(ctx, input)
    assert.Equal(t, false, output.Success)
    assert.Equal(t, 0, output.ReplacementCount)
}
```

### Test: Edit Lines - Precise Line Replacement
```go
func TestEditLines_ReplaceLines(t *testing.T) {
    input := EditLinesInput{
        FilePath: tmpFile,
        StartLine: 107,
        EndLine: 107,
        NewLines: "    return 0;\n}",
        Mode: "replace",
    }
    output := handler(ctx, input)
    assert.True(t, output.Success)
    assert.Equal(t, 1, output.LinesModified)
}
```

### Test: Patch Verification
```go
func TestApplyPatch_WithVerification(t *testing.T) {
    input := ApplyPatchInput{
        FilePath: tmpFile,
        Patch: patchContent,
        VerifyCmd: "gcc demo/calculate.c -o /tmp/test",
    }
    output := handler(ctx, input)
    // Should fail if patch breaks compilation
    assert.False(t, output.Success)
    assert.Contains(t, output.Error, "verification failed")
}
```

---

## Summary of Changes

| Issue | Current | Improved | Impact |
|-------|---------|----------|--------|
| Empty replacement | ❌ Deletes all | ✅ Rejected | Prevents data loss |
| Whitespace matching | ❌ Too strict | ✅ Normalized | Fixes 80% of failures |
| Catastrophic deletes | ❌ No limit | ✅ Max enforced | Safety |
| Line precision | ❌ String-based | ✅ Line-based | Accuracy |
| Patch verification | ❌ No check | ✅ Post-apply verify | Correctness |
| Failure recovery | ❌ No help | ✅ Context-aware | Usability |

---

## Estimated Effort

- **Immediate Fixes**: 4-6 hours
- **Line-Based Editing Tool**: 6-8 hours
- **Enhanced Validation**: 4-6 hours
- **Verification Framework**: 6-8 hours
- **Testing**: 8-10 hours
- **Total Phase 2A**: ~30-40 hours

---

## Next Steps

1. Implement safeguards in `replace_in_file` (prevent empty new_text)
2. Add `max_replacements` validation
3. Create `edit_lines` tool for line-based operations
4. Enhanced patch parser with post-apply verification
5. Comprehensive test coverage for all edge cases

