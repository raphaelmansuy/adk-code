// Common pitfalls and solutions for ADK Code Agent
package prompts

const PitfallsSection = `## Common Pitfalls & Solutions

### Pitfall 1: Wrong Parameter Names for builtin_read_file
❌ Wrong: Using line_start, line_end parameters → Schema validation error
✅ Right: Use offset (start line, 1-indexed) and limit (max lines to read)
→ Example: offset=10, limit=50 reads lines 10-59

### Pitfall 2: Wrong Parameter Format for search_replace
❌ Wrong: Sending SEARCH and REPLACE in uppercase JSON keys → Schema validation error
✅ Right: Use diff parameter with text blocks in the format:
------- SEARCH
[content]
=======
[replacement]
+++++++ REPLACE

### Pitfall 3: Incomplete Content in write_file
❌ Wrong: Providing partial file content, truncating or omitting lines → DATA LOSS
✅ Right: Always provide COMPLETE file content including all sections
→ Use read_file first to get the complete content, then write_file with full content

### Pitfall 4: Shell Argument Parsing with execute_command
❌ Wrong: ./calculate 2 + 2 → Shell sees 4 args: ["./calculate", "2", "+", "2"]
✅ Right: Use execute_program("./calculate", ["2 + 2"]) → One arg to program

### Pitfall 5: File Size Reduction
❌ Wrong: Accidentally overwrite large file with small content → DATA LOSS
✅ Right: write_file has size validation, will reject >90% reduction
→ Use allow_size_reduce=true only if intentional

### Pitfall 6: Not Reading Before Editing
❌ Wrong: Assume code structure, make blind edits → Duplicate code, wrong location
✅ Right: read_file first, understand context, then make precise edits

### Pitfall 7: search_replace Block Not Found
❌ Wrong: SEARCH content doesn't match (whitespace issue, wrong indentation)
✅ Right: Copy EXACT content from file (including all whitespace and indentation)
→ Tool has whitespace-tolerant fallback, but exact match is more reliable

### Pitfall 8: Not Testing After Compile
❌ Wrong: Compile, assume success, run immediately → Runtime errors
✅ Right: Check exit_code=0 and stderr empty before running
`
