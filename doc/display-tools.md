# Display Tools Documentation

## Overview

The display tools allow the agent to communicate with users through formatted messages and task lists. These tools are designed to provide clear, structured communication about the agent's plans, progress, and updates.

## Tools

### 1. display_message

Display formatted messages to the user with support for different message types and markdown formatting.

**Purpose:**
- Communicate plans before executing tasks
- Show structured information or summaries
- Display warnings or important notices
- Provide general updates

**Input Parameters:**

```json
{
  "title": "Optional title/header for the message",
  "content": "Message content in markdown format (supports lists, formatting, etc.)",
  "message_type": "Type of message: info, task, update, warning, success, plan (default: info)",
  "show_timestamp": "Show timestamp with message (default: true)"
}
```

**Output:**

```json
{
  "success": true,
  "message": "Formatted message that was displayed to the user",
  "error": "Error message if operation failed"
}
```

**Message Types & Icons:**
- `info` (ℹ️): General information
- `task` (📋): Task-related information
- `update` (🔄): Progress updates
- `warning` (⚠️): Warnings or cautions
- `success` (✅): Success messages
- `plan` (🎯): Plans or strategies

**Examples:**

1. **Communicate a plan:**
```json
{
  "title": "Execution Plan",
  "content": "I will now:\n1. Search for the function\n2. Analyze its usage\n3. Suggest improvements",
  "message_type": "plan"
}
```

2. **Show a task list:**
```json
{
  "title": "Current Tasks",
  "content": "- [ ] Read configuration file\n- [ ] Validate settings\n- [x] Update dependencies",
  "message_type": "task"
}
```

3. **Provide an update:**
```json
{
  "content": "Currently processing files 1-10 of 50...",
  "message_type": "update"
}
```

4. **Show a warning:**
```json
{
  "title": "Potential Issue Detected",
  "content": "The function `processData` may cause performance issues with large datasets. Consider using streaming instead.",
  "message_type": "warning"
}
```

### 2. update_task_list

Display and update a task list with automatic progress tracking and visualization.

**Purpose:**
- Show multi-step operation progress
- Track task completion status
- Visualize progress with a progress bar
- Provide clear status updates

**Input Parameters:**

```json
{
  "task_list": "Task list in markdown format with - [ ] or - [x] checkboxes",
  "title": "Optional title for the task list",
  "show_progress": "Show progress summary (default: true)"
}
```

**Output:**

```json
{
  "success": true,
  "message": "Formatted task list with progress information",
  "completed_tasks": 2,
  "total_tasks": 4,
  "error": "Error message if operation failed"
}
```

**Task Format:**
- `- [ ]` Pending task
- `- [x]` Completed task
- `- [X]` Completed task (alternative)

**Examples:**

1. **Initial task list:**
```json
{
  "task_list": "- [ ] Read configuration file\n- [ ] Validate settings\n- [ ] Update code\n- [ ] Run tests",
  "title": "Setup Tasks"
}
```

Output will show:
```
📋 Setup Tasks
────────────────────────────────────────

- [ ] Read configuration file
- [ ] Validate settings
- [ ] Update code
- [ ] Run tests

────────────────────────────────────────
📊 Progress: 0/4 tasks completed (0%)
[░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]
```

2. **Updated task list with progress:**
```json
{
  "task_list": "- [x] Read configuration file\n- [x] Validate settings\n- [ ] Update code\n- [ ] Run tests",
  "title": "Setup Tasks"
}
```

Output will show:
```
📋 Setup Tasks
────────────────────────────────────────

- [x] Read configuration file
- [x] Validate settings
- [ ] Update code
- [ ] Run tests

────────────────────────────────────────
📊 Progress: 2/4 tasks completed (50%)
[███████████████░░░░░░░░░░░░░░░]
```

3. **Completed task list:**
```json
{
  "task_list": "- [x] Read configuration file\n- [x] Validate settings\n- [x] Update code\n- [x] Run tests",
  "title": "Setup Tasks"
}
```

Output will show:
```
📋 Setup Tasks
────────────────────────────────────────

- [x] Read configuration file
- [x] Validate settings
- [x] Update code
- [x] Run tests

────────────────────────────────────────
📊 Progress: 4/4 tasks completed (100%)
[██████████████████████████████]
```

## Usage Guidelines

### When to Use display_message

1. **Before executing complex operations:**
   - Explain what you're about to do
   - Show the plan or strategy
   - Set expectations

2. **During long-running operations:**
   - Provide progress updates
   - Show intermediate results
   - Keep the user informed

3. **After completing tasks:**
   - Summarize what was done
   - Highlight important outcomes
   - Show success or warnings

### When to Use update_task_list

1. **Multi-step operations:**
   - Show all steps upfront
   - Update as each step completes
   - Provide clear progress indication

2. **Complex workflows:**
   - Break down the process
   - Show dependencies
   - Track completion status

3. **Long-running tasks:**
   - Give visibility into progress
   - Show what's done and what remains
   - Help users understand how far along you are

## Best Practices

1. **Be Clear and Concise:**
   - Use simple language
   - Focus on what matters to the user
   - Avoid technical jargon unless necessary

2. **Use Appropriate Message Types:**
   - Match the message type to the content
   - Use icons to provide visual cues
   - Be consistent in your usage

3. **Update Regularly:**
   - For long operations, provide periodic updates
   - Don't leave users wondering what's happening
   - Show progress incrementally

4. **Structure Information:**
   - Use markdown lists for clarity
   - Break down complex information
   - Highlight key points with formatting

5. **Track Progress:**
   - Use task lists for multi-step operations
   - Update the list as you complete tasks
   - Show the full list so users see the big picture

## Integration with Agent Workflow

These tools are designed to work seamlessly with the agent's workflow:

1. **Planning Phase:** Use `display_message` with type `plan` to show your approach
2. **Execution Phase:** Use `update_task_list` to track progress through steps
3. **Update Phase:** Use `display_message` with type `update` for interim updates
4. **Completion Phase:** Use `display_message` with type `success` to confirm completion

## Examples in Context

### Example 1: File Analysis Task

```
Step 1: Show the plan
{
  "title": "File Analysis Plan",
  "content": "I will analyze the configuration files in the following order:\n1. Read all config files\n2. Parse JSON structure\n3. Validate required fields\n4. Generate summary report",
  "message_type": "plan"
}

Step 2: Show task list
{
  "task_list": "- [ ] Read all config files\n- [ ] Parse JSON structure\n- [ ] Validate required fields\n- [ ] Generate summary report",
  "title": "Analysis Progress"
}

Step 3: Update task list after each step
{
  "task_list": "- [x] Read all config files\n- [x] Parse JSON structure\n- [ ] Validate required fields\n- [ ] Generate summary report",
  "title": "Analysis Progress"
}

Step 4: Final update
{
  "task_list": "- [x] Read all config files\n- [x] Parse JSON structure\n- [x] Validate required fields\n- [x] Generate summary report",
  "title": "Analysis Progress"
}

Step 5: Show summary
{
  "title": "Analysis Complete",
  "content": "Successfully analyzed 5 configuration files. All files are valid. Key findings:\n- 2 files use outdated schema\n- 1 file has deprecated settings\n- Recommended updates available",
  "message_type": "success"
}
```

### Example 2: Code Refactoring Task

```
Step 1: Communicate the plan
{
  "title": "Refactoring Plan",
  "content": "I will refactor the `UserService` class with these improvements:\n\n**Changes:**\n- Extract validation logic to separate validator\n- Add error handling for edge cases\n- Improve naming consistency\n- Add comprehensive documentation\n\n**Impact:**\n- Better code organization\n- Easier to test\n- More maintainable",
  "message_type": "plan"
}

Step 2: Track progress
{
  "task_list": "- [ ] Extract validation logic\n- [ ] Add error handling\n- [ ] Improve naming\n- [ ] Add documentation\n- [ ] Run tests",
  "title": "Refactoring Progress"
}

Step 3: Updates during execution
{
  "task_list": "- [x] Extract validation logic\n- [x] Add error handling\n- [ ] Improve naming\n- [ ] Add documentation\n- [ ] Run tests",
  "title": "Refactoring Progress"
}

Step 4: Completion
{
  "content": "Refactoring complete! All tests pass. The code is now more modular and maintainable.",
  "message_type": "success"
}
```

## Category

These tools are registered under the `CategoryDisplay` category in the tool registry, making them easy to find and use alongside other agent tools.
