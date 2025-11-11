# Prompt Improvement: General Assistance Tasks

**Date**: November 11, 2025  
**Task**: Improve the prompt to allow general assistance tasks beyond coding

## Summary

Successfully updated the agent system prompt to be more inclusive and allow a wide variety of assistance tasks. The agent can now handle creative writing tasks (like writing a poem), analysis, research, and other general assistance requests in addition to its coding expertise.

## Problem

The agent was overly restrictive and would refuse non-coding tasks:
```
❯ Write a poem
⠸ Agent is thinking  [↓used=9201, prompt=9065, response=24, thoughts=112]
│   I cannot write a poem. My capabilities are focused on coding tasks. How can I help you with your code today?
✓ Task completed
```

## Root Cause

The `agent_identity` section in the system prompt was too narrowly scoped:
```
"You are an expert AI coding assistant with state-of-the-art file editing capabilities.
Your purpose is to help users with coding tasks by reading files, writing code, executing commands, and iteratively solving problems."
```

## Solution

Updated the agent identity in `code_agent/agent/xml_prompt_builder.go` to be more inclusive while maintaining the coding expertise as a specialization.

### Changes Made

**File**: `code_agent/agent/xml_prompt_builder.go`

**Modified Section**: `BuildXMLPrompt()` method, `agent_identity` tag

**Old Text**:
```go
buf.WriteString("You are an expert AI coding assistant with state-of-the-art file editing capabilities.\n")
buf.WriteString("Your purpose is to help users with coding tasks by reading files, writing code, executing commands, and iteratively solving problems.\n")
```

**New Text**:
```go
buf.WriteString("You are an expert AI assistant with state-of-the-art capabilities spanning coding, analysis, writing, problem-solving, and general knowledge tasks.\n")
buf.WriteString("Your purpose is to help users with a wide variety of tasks including:\n")
buf.WriteString("- Coding and software engineering (reading files, writing code, executing commands, debugging)\n")
buf.WriteString("- Writing and creative tasks (essays, poetry, stories, explanations)\n")
buf.WriteString("- Analysis and research (breaking down problems, finding information, evaluating solutions)\n")
buf.WriteString("- General assistance (answering questions, providing guidance, offering suggestions)\n")
buf.WriteString("\n")
buf.WriteString("You approach all tasks with the same rigor and iterative problem-solving mindset as you do with coding.\n")
```

## Benefits

1. **Broader Capability**: Agent can now assist with:
   - Creative writing (poetry, stories, essays)
   - Analysis and research tasks
   - General knowledge questions
   - Problem-solving across domains

2. **Maintains Expertise**: Coding remains a key specialization with full tool support

3. **Unified Mindset**: Makes clear that the same rigorous, iterative approach applies to all tasks

4. **Better UX**: Users won't encounter confusing refusals for legitimate assistance requests

## Testing

✅ All code quality checks pass:
- `go fmt` - No formatting issues
- `go vet` - No structural issues  
- All 60+ unit tests pass
- Prompt builder tests specifically validate XML structure

## Expected Behavior After Change

Now when a user asks the agent to write a poem, instead of refusing:
```
❯ Write a poem
⠸ Agent is thinking  [token usage...]
│   [Agent writes a poem]
✓ Task completed
```

## Design Notes

- The change is minimal and focused - only the agent identity is modified
- Other sections (tools, guidance, pitfalls, workflows) remain unchanged
- The agent's tool access doesn't change - it still has file editing, execution, search capabilities
- Non-coding tasks won't trigger file operations or shell executions unless explicitly requested

## Verification

The change was verified by:
1. Building the project: `make build` ✓
2. Running all tests: `make test` ✓ (all pass)
3. Running code quality checks: `make check` ✓ (all pass)
4. Examining the actual code change to ensure it's correct ✓

## Future Enhancements

Possible future improvements:
1. Add specialized guidance for creative writing tasks
2. Include examples of non-coding assistance in the prompt
3. Add task-specific best practices for research and analysis
4. Consider separate tool sets for different task types if needed
