# Display Tool Event Formatting Improvements

**Date:** 2025-11-11  
**Task:** Improve CLI event display for `update_task_list` and `display_message` tools

## Problem

When the agent used `display_message` and `update_task_list` tools, the CLI was showing raw JSON output instead of the formatted message content:

```json
{
  "message": "📋 Task List\n- [x] Task 1\n- [ ] Task 2\n...",
  "success": true
}
```

Users wanted to see the beautifully formatted message directly, not the JSON structure.

## Solution

Added special handling for display tools in two key places:

### 1. Tool Result Parser (`display/tool_result_parser.go`)

Added two new parser methods to extract and display the pre-formatted `message` field:

```go
// parseDisplayMessage extracts and displays the pre-formatted message content
func (trp *ToolResultParser) parseDisplayMessage(result map[string]any) string {
    if message, ok := result["message"].(string); ok {
        return message
    }
    return trp.parseGeneric(result)
}

// parseUpdateTaskList extracts and displays the pre-formatted task list
func (trp *ToolResultParser) parseUpdateTaskList(result map[string]any) string {
    if message, ok := result["message"].(string); ok {
        return message
    }
    return trp.parseGeneric(result)
}
```

Updated the `ParseToolResult()` switch statement to route these tools to their specialized parsers:

```go
case "display_message":
    return trp.parseDisplayMessage(result)
case "update_task_list":
    return trp.parseUpdateTaskList(result)
```

### 2. Event Handler Spinner Messages (`events.go`)

Added context-aware spinner messages for the display tools in `getToolSpinnerMessage()`:

```go
case "display_message":
    if messageType, ok := args["message_type"].(string); ok {
        return fmt.Sprintf("%s Displaying %s message", icon, messageType)
    }
    return fmt.Sprintf("%s Displaying message", icon)
case "update_task_list":
    return fmt.Sprintf("%s Updating task list", icon)
```

## Results

Now when the agent uses these tools, users see:

### Before

```text
✓ Tool completed: update_task_list
```

```json
{
  "completed_tasks": 3,
  "message": "\n📋 Overall Progress\n...",
  "success": true,
  "total_tasks": 5
}
```

### After

```text
✓ Tool completed: update_task_list

📋 Overall Improvement Progress
────────────────────────────────────────

- [x] Re-verify and complete Chapter 1 improvements
- [x] Re-verify and complete Chapter 2 improvements
- [x] Improve Chapter 3: Control Flow
- [ ] Re-verify and complete Chapter 4 improvements
- [ ] Re-verify and complete Chapter 5 improvements

────────────────────────────────────────
📊 Progress: 3/5 tasks completed (60%)
[██████████████████░░░░░░░░░░░░]
```

## Benefits

1. **Better UX** - Users see clean, formatted messages instead of JSON structures
2. **Visual Progress** - Task lists display with progress bars and checkboxes
3. **Consistency** - Display tools follow the same parsing pattern as other tools
4. **Maintainability** - Clear separation of concerns with specialized parser methods

## Testing

Verified with live agent session showing:

- ✅ display_message showing formatted content
- ✅ update_task_list showing progress bars and checkboxes
- ✅ All builds passing (`make check`)
- ✅ Proper spinner messages during tool execution

## Files Modified

- `code_agent/display/tool_result_parser.go` - Added parseDisplayMessage and parseUpdateTaskList methods
- `code_agent/events.go` - Added display tool cases to getToolSpinnerMessage

## Related Work

This builds on the display tools implementation from:

- `2025-11-11-display-tools-implementation.md`
- `2025-11-11-system-prompt-communication-enhancement.md`
