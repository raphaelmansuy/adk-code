// Package display provides tools for displaying formatted messages and task lists to the user.
package display

import (
	"fmt"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"code_agent/tools/base"
)

// DisplayMessageInput defines the input parameters for displaying a message to the user.
type DisplayMessageInput struct {
	// Title is an optional title for the message (displayed as a header).
	Title string `json:"title,omitempty" jsonschema:"Optional title/header for the message"`
	// Content is the main message content in markdown format.
	Content string `json:"content" jsonschema:"Message content in markdown format (supports lists, formatting, etc.)"`
	// MessageType indicates the type of message (info, task, update, warning, success).
	MessageType string `json:"message_type,omitempty" jsonschema:"Type of message: info, task, update, warning, success (default: info)"`
	// ShowTimestamp indicates whether to show a timestamp with the message.
	ShowTimestamp *bool `json:"show_timestamp,omitempty" jsonschema:"Show timestamp with message (default: true)"`
}

// DisplayMessageOutput defines the output of displaying a message.
type DisplayMessageOutput struct {
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`
	// Message contains a confirmation message.
	Message string `json:"message,omitempty"`
	// Error contains error message if the operation failed.
	Error string `json:"error,omitempty"`
}

// NewDisplayMessageTool creates a tool for displaying formatted messages to the user.
func NewDisplayMessageTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input DisplayMessageInput) DisplayMessageOutput {
		// Default message type to "info"
		messageType := input.MessageType
		if messageType == "" {
			messageType = "info"
		}

		// Build the formatted message
		var builder strings.Builder

		// Add header with icon based on message type
		icon := getIconForMessageType(messageType)

		if input.Title != "" {
			builder.WriteString(fmt.Sprintf("\n%s %s\n", icon, input.Title))
			builder.WriteString(strings.Repeat("─", len(input.Title)+4))
			builder.WriteString("\n\n")
		} else {
			builder.WriteString(fmt.Sprintf("\n%s ", icon))
		}

		// Add the content
		builder.WriteString(input.Content)
		builder.WriteString("\n")

		// The message will be displayed to the user via the tool result
		// which is shown in the terminal
		displayedMessage := builder.String()

		return DisplayMessageOutput{
			Success: true,
			Message: displayedMessage,
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name: "display_message",
		Description: `Display formatted messages, task lists, and updates to the user in markdown format.
Use this tool to:
- Communicate your plan of action before executing tasks
- Show task lists with checkboxes (- [ ] for pending, - [x] for completed)
- Provide progress updates during long operations
- Display structured information or summaries
- Show warnings or important notices

Examples:
- Task list: "- [ ] Read configuration file\n- [ ] Validate settings\n- [x] Update dependencies"
- Update: "Currently processing files 1-10 of 50..."
- Plan: "I will now:\n1. Search for the function\n2. Analyze its usage\n3. Suggest improvements"

The message supports full markdown formatting including lists, emphasis, code blocks, etc.`,
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategoryDisplay,
			Priority:  0,
			UsageHint: "Communicate plans, show task lists, provide updates to the user in markdown",
		})
	}

	return t, err
}

// getIconForMessageType returns an appropriate emoji/icon for the message type.
func getIconForMessageType(messageType string) string {
	switch strings.ToLower(messageType) {
	case "task":
		return "📋"
	case "update":
		return "🔄"
	case "warning":
		return "⚠️"
	case "success":
		return "✅"
	case "plan":
		return "🎯"
	case "info":
		fallthrough
	default:
		return "ℹ️"
	}
}

// UpdateTaskListInput defines the input parameters for updating a task list.
type UpdateTaskListInput struct {
	// TaskList is the complete task list in markdown format with checkboxes.
	TaskList string `json:"task_list" jsonschema:"Task list in markdown format with - [ ] or - [x] checkboxes"`
	// Title is an optional title for the task list.
	Title string `json:"title,omitempty" jsonschema:"Optional title for the task list"`
	// ShowProgress indicates whether to show progress summary (X of Y completed).
	ShowProgress *bool `json:"show_progress,omitempty" jsonschema:"Show progress summary (default: true)"`
}

// UpdateTaskListOutput defines the output of updating a task list.
type UpdateTaskListOutput struct {
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`
	// Message contains the formatted task list.
	Message string `json:"message,omitempty"`
	// CompletedTasks is the number of completed tasks.
	CompletedTasks int `json:"completed_tasks"`
	// TotalTasks is the total number of tasks.
	TotalTasks int `json:"total_tasks"`
	// Error contains error message if the operation failed.
	Error string `json:"error,omitempty"`
}

// NewUpdateTaskListTool creates a tool for displaying and updating task lists.
func NewUpdateTaskListTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input UpdateTaskListInput) UpdateTaskListOutput {
		// Parse the task list to count completed and total tasks
		lines := strings.Split(input.TaskList, "\n")
		totalTasks := 0
		completedTasks := 0

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- [ ]") || strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
				totalTasks++
				if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
					completedTasks++
				}
			}
		}

		// Default to showing progress
		showProgress := true
		if input.ShowProgress != nil {
			showProgress = *input.ShowProgress
		}

		// Build the formatted output
		var builder strings.Builder

		builder.WriteString("\n📋 ")
		if input.Title != "" {
			builder.WriteString(input.Title)
		} else {
			builder.WriteString("Task List")
		}
		builder.WriteString("\n")
		builder.WriteString(strings.Repeat("─", 40))
		builder.WriteString("\n\n")

		// Add the task list
		builder.WriteString(input.TaskList)
		builder.WriteString("\n")

		// Add progress summary if requested and there are tasks
		if showProgress && totalTasks > 0 {
			builder.WriteString("\n")
			builder.WriteString(strings.Repeat("─", 40))
			builder.WriteString("\n")

			progressPercentage := 0
			if totalTasks > 0 {
				progressPercentage = (completedTasks * 100) / totalTasks
			}

			builder.WriteString(fmt.Sprintf("📊 Progress: %d/%d tasks completed (%d%%)\n",
				completedTasks, totalTasks, progressPercentage))

			// Add a simple progress bar
			progressBarWidth := 30
			filledWidth := (completedTasks * progressBarWidth) / totalTasks
			if completedTasks > 0 && filledWidth == 0 {
				filledWidth = 1 // Show at least some progress
			}

			builder.WriteString("[")
			builder.WriteString(strings.Repeat("█", filledWidth))
			builder.WriteString(strings.Repeat("░", progressBarWidth-filledWidth))
			builder.WriteString("]\n")
		}

		return UpdateTaskListOutput{
			Success:        true,
			Message:        builder.String(),
			CompletedTasks: completedTasks,
			TotalTasks:     totalTasks,
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name: "update_task_list",
		Description: `Display and update a task list with progress tracking.
Use this tool to show a structured task list with checkboxes and automatic progress calculation.

Task format:
- [ ] Pending task
- [x] Completed task

Example:
task_list: "- [x] Read configuration\n- [x] Validate settings\n- [ ] Update code\n- [ ] Run tests"

This will automatically count and display progress (2/4 tasks completed, 50%).
Perfect for showing multi-step operation progress to the user.`,
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategoryDisplay,
			Priority:  0,
			UsageHint: "Show structured task lists with automatic progress tracking and visualization",
		})
	}

	return t, err
}
