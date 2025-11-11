# System Prompt Updates for Display Tools - November 11, 2025

## Summary

Updated the agent's system prompt to incorporate the new `display_message` and `update_task_list` tools, providing practical guidance on when and how to use them for transparent user communication.

## Changes Made

### 1. Added Communication & Transparency Section (prompt_guidance.go)

**Location**: Beginning of GuidanceSection, before "Tool Selection Guide"

**New Content**:
- **When to Use Display Tools**: Clear guidance on both tools with examples
- **Communication Best Practices**: DO/DON'T patterns to prevent overuse
- **Example Pattern**: Complete workflow showing proper tool usage

**Key Guidelines Added**:
- Use `display_message(type="plan")` before complex operations to explain approach
- Use `update_task_list` for multi-step operations with 3+ steps
- Use `display_message(type="warning")` to proactively alert about issues
- Use `display_message(type="success")` to clearly signal task completion
- DON'T overuse for simple 1-2 step operations
- DON'T repeat information already in tool outputs

### 2. Updated Typical Task Flow (prompt_workflow.go)

**Changes**:
- Added "Plan & Communicate" step with display tool usage
- Separated "Typical Task Flow (with User Communication)" for complex tasks
- Added "Simple Task Flow (1-2 steps)" for straightforward operations
- Integrated progress updates into the iteration cycle

**Key Additions**:
- Display plan before starting complex tasks
- Show task list upfront for multi-step operations
- Update task list as each major step completes
- Use warnings when detecting issues during execution
- Confirm with success message when all tests pass

### 3. Enhanced Example Workflows (prompt_workflow.go)

**Added Two Types of Examples**:

1. **Simple Task Example** (unchanged - existing pattern)
   - No display tools needed
   - Direct: read → edit → test

2. **Complex Task Example** (NEW)
   - Full workflow: "Refactor authentication system to add JWT support"
   - Shows display_message for initial plan
   - Shows update_task_list with progressive updates
   - Shows warning message for important user impact
   - Shows success confirmation at completion

### 4. Updated Response Style Guidelines (prompt_workflow.go)

**Additions**:
- "Be transparent": Use display tools for plans and progress
- "Track progress": Use update_task_list for multi-step operations
- "Handle errors gracefully": Proactively use warnings

## Design Philosophy

The updates follow these principles:

1. **Pragmatic, Not Prescriptive**: Guidelines focus on WHEN to use tools, not forcing their use
2. **Context-Aware**: Different approaches for simple vs complex tasks
3. **User-Centric**: Emphasize transparency and keeping users informed
4. **Anti-Spam**: Explicitly warn against overuse to prevent verbose output
5. **Pattern-Based**: Provide concrete examples rather than abstract rules

## Integration Points

The new guidance integrates seamlessly with existing sections:

- **Before**: Tool Selection Guide, File Editing Patterns, Execution Guidance
- **With**: Workflow patterns showing tool usage in context
- **After**: Safety features, best practices, key differences

## Expected Impact

### Positive Behaviors to Expect:

1. **Better Planning Communication**: Agent will explain approach before diving in
2. **Progress Visibility**: Users see advancement through multi-step tasks
3. **Proactive Issue Detection**: Warnings before problems become errors
4. **Clear Completion Signals**: Explicit success confirmations

### Prevented Behaviors:

1. **Over-Communication**: Guidance prevents excessive display tool usage
2. **Redundancy**: Warns against repeating tool output information
3. **Task List Spam**: Only for 3+ step operations, not simple tasks

## Examples from Prompt

### Good Pattern (Complex Task):
```
Task: "Refactor UserService class"

1. display_message(type="plan"): "I will extract validation logic, add error handling, and improve naming"
2. update_task_list: Show all 5 steps with checkboxes
3. [Do work, updating task list after each step]
4. display_message(type="success"): "Refactoring complete! All tests pass."
```

### Good Pattern (Simple Task):
```
Task: "Fix typo in error message"

1. read_file → find typo
2. search_replace → fix it
3. Done (no display tools needed)
```

## Testing

- ✅ Code compiles successfully
- ✅ All agent tests pass (12/12)
- ✅ Prompt structure validates correctly
- ✅ No breaking changes to existing behavior
- ✅ Backward compatible with existing prompts

## Files Modified

1. **code_agent/agent/prompt_guidance.go**
   - Added 50+ lines for Communication & Transparency section
   - Positioned at the beginning for high visibility

2. **code_agent/agent/prompt_workflow.go**
   - Updated Typical Task Flow with communication steps
   - Added Simple vs Complex task differentiation
   - Enhanced example workflows with display tool usage
   - Updated Response Style guidelines

## Token Impact

Estimated additional tokens in system prompt: ~800-1000 tokens
- Communication guidance: ~400 tokens
- Updated workflows: ~300 tokens
- Enhanced examples: ~200 tokens
- Response style updates: ~100 tokens

This is a reasonable increase for significantly improved agent transparency.

## Next Steps

The agent now has clear guidance on:
- ✅ When to use display tools
- ✅ When NOT to use them (equally important)
- ✅ How to structure communication
- ✅ Patterns for simple vs complex tasks

Future enhancements could include:
- Adding more example patterns based on real usage
- Fine-tuning based on user feedback
- Adjusting token budget if needed
- Adding specialized patterns for specific task types

## Conclusion

The system prompt updates successfully integrate the new display tools with pragmatic, balanced guidance that encourages transparency without spam. The agent now has clear patterns for when and how to communicate with users during task execution.
