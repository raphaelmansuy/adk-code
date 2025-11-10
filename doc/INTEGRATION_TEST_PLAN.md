# Integration Test Plan: ADK Code Agent vs Cline

## Purpose
Validate that the enhanced ADK Code Agent with Cline-inspired improvements:
1. Reduces tool call count by ~55% (from 58 to ~25 calls)
2. Eliminates shell quoting confusion (21 wasted attempts → 0)
3. Prevents catastrophic file overwrites (size validation)
4. Provides superior editing capabilities vs Cline

## Test Scenario: calculate.c Improvement

### Original Trace Analysis
**Problem**: Agent took 58 tool calls to improve calculate.c, including:
- ❌ 21 wasted calls due to shell quoting confusion ("./calculate '5+3'", "./calculate \"5+3\"", etc.)
- ❌ Catastrophic file overwrite (3400 bytes → 2 bytes)
- ❌ Multiple syntax errors requiring iteration
- ❌ Duplicate code insertion attempts

### Expected Improvements with New Tools

#### Tool Call Reduction
**Old Approach** (58 calls):
1. read_file: calculate.c
2-22. execute_command: 21 attempts trying different quote combinations
23-40. Multiple replace_in_file calls with syntax errors
41-58. More debugging and corrections

**New Approach** (~25 calls expected):
1. read_file: calculate.c → understand code
2. search_replace: Make targeted improvements using SEARCH/REPLACE blocks
3. execute_command: gcc -o demo/calculate demo/calculate.c
4. execute_program: Test with structured argv (NO quoting confusion)
5-25. Iteration if needed (much cleaner)

**Key Efficiency Gains**:
- execute_program eliminates 21 shell quoting attempts
- search_replace is more reliable than multiple replace_in_file calls
- Whitespace-tolerant matching reduces failed replacements
- Better system prompt guides tool selection

### Test Cases

#### Test 1: Basic Code Improvement
**Task**: "Improve demo/calculate.c to handle expressions with spaces"

**Success Criteria**:
- ✅ Total tool calls < 30 (vs 58 baseline)
- ✅ No shell quoting confusion (zero "./calculate '...' vs \"...\"" attempts)
- ✅ No catastrophic file overwrites (size validation blocks dangerous writes)
- ✅ Code compiles successfully
- ✅ All test cases pass

**Expected Tool Sequence**:
1. `read_file(demo/calculate.c)` → understand current code
2. `search_replace` with SEARCH/REPLACE blocks → improve parsing logic
3. `execute_command("gcc -o demo/calculate demo/calculate.c")` → compile
4. `execute_program(program="./demo/calculate", args=["5 + 3"])` → test (NO quoting issues!)
5. Iterate if needed with more search_replace calls

#### Test 2: Shell Quoting Stress Test
**Task**: "Test the calculate program with various input formats"

**Old Behavior** (21 failed attempts):
```
execute_command('./calculate "5+3"')     → fails (quotes become part of arg)
execute_command('./calculate '5+3'')     → fails (single quotes)
execute_command('./calculate 5+3')       → works but doesn't handle spaces
execute_command('./calculate "5 + 3"')   → fails
... (17 more attempts)
```

**New Behavior** (works immediately):
```
execute_program(program="./demo/calculate", args=["5+3"])        → ✅ works
execute_program(program="./demo/calculate", args=["5 + 3"])      → ✅ works
execute_program(program="./demo/calculate", args=["-5 * 2"])     → ✅ works
execute_program(program="./demo/calculate", args=["10 / 0"])     → ✅ handles error
```

**Success Criteria**:
- ✅ Zero quote-related failures
- ✅ All test inputs handled correctly on first attempt
- ✅ Clear error messages for invalid inputs

#### Test 3: File Safety Validation
**Task**: Attempt operations that would have caused catastrophic overwrites

**Test Scenarios**:
1. **Attempt to write 2 bytes to 3400-byte file** (from original trace)
   - Old: Would succeed, destroying file
   - New: Should be rejected with "ERROR: Rejecting write that would reduce file size by X%"

2. **Attempt to overwrite with empty content**
   - Old: Would create empty file
   - New: Should be rejected (>90% reduction)

3. **Legitimate size reduction** (e.g., removing 50% of code)
   - Should succeed (not >90% reduction)

**Success Criteria**:
- ✅ Dangerous writes blocked (>90% size reduction)
- ✅ Clear error message explaining rejection
- ✅ Legitimate writes still work

#### Test 4: SEARCH/REPLACE Block Editing
**Task**: "Add error handling to a function using SEARCH/REPLACE blocks"

**Test Input**:
```
search_replace(
  file_path="demo/calculate.c",
  diff="""
------- SEARCH
func divide(a, b int) int {
    return a / b
}
=======
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
+++++++ REPLACE
"""
)
```

**Success Criteria**:
- ✅ Exact match found despite whitespace variations
- ✅ Replacement applied correctly
- ✅ Multiple SEARCH/REPLACE blocks in one call supported
- ✅ Clear error if search block not found

#### Test 5: Whitespace Tolerance Test
**Task**: "Test SEARCH/REPLACE matching with different whitespace"

**Scenarios**:
1. Search block has tabs, file has spaces → Should match (line-trimmed fallback)
2. Search block has extra blank lines → Should match (flexible)
3. Search block has wrong indentation → Should match (whitespace-tolerant)

**Success Criteria**:
- ✅ Whitespace variations don't break matching
- ✅ Exact match tried first (for precision)
- ✅ Line-trimmed fallback works when needed

### Comparison Metrics

| Metric | Old Agent | New Agent | Improvement |
|--------|-----------|-----------|-------------|
| Total tool calls | 58 | ~25 | 57% reduction |
| Shell quoting failures | 21 | 0 | 100% elimination |
| Catastrophic overwrites | 1 | 0 | 100% prevention |
| Syntax errors per edit | 3-5 | 1-2 | 50% reduction |
| Time to completion | ~5 min | ~2 min | 60% faster |

### Testing Procedure

#### Setup
```bash
cd /Users/raphaelmansuy/Github/03-working/adk_training_go/code_agent

# Backup original demo files
cp demo/calculate.c demo/calculate.c.backup
cp demo/fibonacci.c demo/fibonacci.c.backup

# Ensure clean build
go build -o code-agent main.go
```

#### Run Test 1: Basic Improvement
```bash
./code-agent

# At prompt, enter:
> Improve demo/calculate.c to handle expressions with spaces around operators

# Monitor tool calls and verify:
# - Total calls < 30
# - No shell quoting issues
# - Uses execute_program for testing
# - Code compiles and runs correctly
```

#### Run Test 2: Shell Quoting
```bash
# After Test 1 completes, test execution:
> Test the calculate program with these inputs: "5+3", "5 + 3", "-5 * 2", "10 / 0"

# Verify execute_program is used (not execute_command)
# Verify all inputs handled correctly on first attempt
```

#### Run Test 3: File Safety
```bash
# Restore backup
cp demo/calculate.c.backup demo/calculate.c

# Create malicious test (simulate the catastrophic overwrite)
> Replace the entire contents of demo/calculate.c with just a closing brace

# Expected: Should be blocked with size validation error
# If accepted, TEST FAILED - safety feature broken
```

#### Run Test 4: SEARCH/REPLACE
```bash
# Test SEARCH/REPLACE block functionality
> Use search_replace to add a comment header to demo/calculate.c

# Verify:
# - Tool uses search_replace (not replace_in_file)
# - SEARCH/REPLACE blocks parsed correctly
# - Changes applied successfully
```

#### Run Test 5: Whitespace Tolerance
```bash
# Create test file with varying whitespace
cat > test_whitespace.go <<'EOF'
func hello() {
	fmt.Println("world")
}
EOF

# Test with different whitespace in search block
> Use search_replace to modify the hello function, using SEARCH/REPLACE blocks

# Verify:
# - Match succeeds despite whitespace differences
# - Line-trimmed fallback works if needed
```

### Success Definition

**Test PASSES if**:
- ✅ All 5 test cases pass
- ✅ Tool call reduction > 50% vs baseline
- ✅ Zero shell quoting failures
- ✅ Zero catastrophic overwrites
- ✅ Agent completes tasks successfully
- ✅ System prompt guides correct tool usage

**Test FAILS if**:
- ❌ Tool calls still > 40 (not enough improvement)
- ❌ Shell quoting issues still occur
- ❌ Catastrophic overwrites not prevented
- ❌ SEARCH/REPLACE blocks don't work correctly
- ❌ Agent gets stuck or fails tasks

### Benchmark Against Cline

To truly validate "better than Cline", we need:

1. **Feature Comparison** (Already documented in TOOL_SET_FINAL.md)
   - ADK: 12 tools (including search_replace, execute_program, edit_lines, apply_patch, preview)
   - Cline: 8 tools
   - Winner: ADK (7-0-1)

2. **Efficiency Comparison** (This test plan)
   - Measure: Tool calls, time to completion, success rate
   - Target: 50%+ efficiency gain vs old approach
   - Comparison: Use Cline's calculate.c improvement as baseline if available

3. **Robustness Comparison**
   - ADK: Size validation, atomic writes, whitespace tolerance, structured argv
   - Cline: Basic functionality
   - Winner: ADK (more safety features)

4. **Usability Comparison**
   - ADK: Enhanced system prompt with clear tool selection guidance
   - Cline: Standard prompt
   - Winner: ADK (better guidance)

### Expected Outcome

After running this integration test plan, we expect:

1. **Quantitative Improvements**:
   - 50-60% reduction in tool calls (58 → ~25)
   - 100% elimination of shell quoting issues (21 → 0)
   - 100% prevention of catastrophic overwrites
   - 50% reduction in syntax errors per edit

2. **Qualitative Improvements**:
   - Smoother task completion
   - Better tool selection by agent
   - More reliable code editing
   - Clearer error messages

3. **Competitive Position**:
   - Feature parity with Cline: ✅ Achieved
   - Performance superiority: ✅ Expected (55% efficiency gain)
   - Safety superiority: ✅ Achieved (size validation, atomic writes)
   - Overall: **Better than Cline** ✅

### Next Steps After Testing

1. **If tests PASS**:
   - Document results in VALIDATION_RESULTS.md
   - Update README with performance metrics
   - Create demo video showing efficiency gains
   - Mark project as "Production Ready"

2. **If tests FAIL**:
   - Analyze failure points
   - Iterate on tool implementations
   - Refine system prompt
   - Retest until all criteria met

3. **Future Enhancements** (Post-validation):
   - Phase 2B+ features (post-edit validation, duplicate detection, auto-backup)
   - Benchmark against other coding agents (Cursor, Aider, etc.)
   - Performance optimization (tool call caching, parallel operations)
   - Extended language support (Python, JavaScript, etc.)
