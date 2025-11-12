// Common pitfalls and solutions for ADK Code Agent
package prompts

const PitfallsSection = `## Common Pitfalls & Solutions

### Pitfall 1: Shell Argument Parsing with execute_command
❌ Wrong: ./calculate 2 + 2 → Shell sees 4 args: ["./calculate", "2", "+", "2"]
✅ Right: Use execute_program("./calculate", ["2 + 2"]) → One arg to program

### Pitfall 2: File Size Reduction
❌ Wrong: Accidentally overwrite large file with small content → DATA LOSS
✅ Right: write_file has size validation, will reject >90% reduction
→ Use allow_size_reduce=true only if intentional

### Pitfall 3: Not Reading Before Editing
❌ Wrong: Assume code structure, make blind edits → Duplicate code, wrong location
✅ Right: read_file first, understand context, then make precise edits

### Pitfall 4: search_replace Block Not Found
❌ Wrong: SEARCH content doesn't match (whitespace issue)
✅ Right: Copy exact content from file (including indentation, newlines)
→ Tool has whitespace-tolerant fallback, but exact is better

### Pitfall 5: Not Testing After Compile
❌ Wrong: Compile, assume success, run immediately → Runtime errors
✅ Right: Check exit_code=0 and stderr empty before running
`
