package main

import (
	"fmt"
	"strings"

	"code_agent/tools"
)

// This example demonstrates how to use the display_message and update_task_list tools
func main() {
	fmt.Println("=== Display Tools Example ===")
	fmt.Println()

	// Example 1: Create display_message tool
	displayTool, err := tools.NewDisplayMessageTool()
	if err != nil {
		panic(err)
	}
	fmt.Printf("✓ Created display_message tool: %s\n", displayTool.Name())

	// Example 2: Create update_task_list tool
	taskListTool, err := tools.NewUpdateTaskListTool()
	if err != nil {
		panic(err)
	}
	fmt.Printf("✓ Created update_task_list tool: %s\n", taskListTool.Name())

	fmt.Println()
	fmt.Println("=== Example Usage ===")
	fmt.Println()

	// Example 3: Display a plan message
	fmt.Println("Example 1: Communicating a plan")
	fmt.Println("Input:")
	fmt.Println(`{
  "title": "Execution Plan",
  "content": "I will now:\n1. Search for the function\n2. Analyze its usage\n3. Suggest improvements",
  "message_type": "plan"
}`)
	fmt.Println("\nOutput:")
	fmt.Println("🎯 Execution Plan")
	fmt.Println("────────────────────")
	fmt.Println("\nI will now:")
	fmt.Println("1. Search for the function")
	fmt.Println("2. Analyze its usage")
	fmt.Println("3. Suggest improvements")

	fmt.Println("\n" + strings.Repeat("─", 60) + "\n")

	// Example 4: Display a task list with progress
	fmt.Println("Example 2: Tracking progress with task list")
	fmt.Println("Input:")
	fmt.Println(`{
  "task_list": "- [x] Read configuration\n- [x] Validate settings\n- [ ] Update code\n- [ ] Run tests",
  "title": "Setup Tasks"
}`)
	fmt.Println("\nOutput:")
	fmt.Println("📋 Setup Tasks")
	fmt.Println("────────────────────────────────────────")
	fmt.Println("\n- [x] Read configuration")
	fmt.Println("- [x] Validate settings")
	fmt.Println("- [ ] Update code")
	fmt.Println("- [ ] Run tests")
	fmt.Println("\n────────────────────────────────────────")
	fmt.Println("📊 Progress: 2/4 tasks completed (50%)")
	fmt.Println("[███████████████░░░░░░░░░░░░░░░]")

	fmt.Println("\n" + strings.Repeat("─", 60) + "\n")

	// Example 5: Display a warning
	fmt.Println("Example 3: Showing a warning")
	fmt.Println("Input:")
	fmt.Println(`{
  "title": "Potential Issue Detected",
  "content": "The function 'processData' may cause performance issues with large datasets.",
  "message_type": "warning"
}`)
	fmt.Println("\nOutput:")
	fmt.Println("⚠️  Potential Issue Detected")
	fmt.Println("────────────────────────────────")
	fmt.Println("\nThe function 'processData' may cause performance issues with large datasets.")

	fmt.Println("\n" + strings.Repeat("─", 60) + "\n")

	// Example 6: Display a success message
	fmt.Println("Example 4: Success confirmation")
	fmt.Println("Input:")
	fmt.Println(`{
  "content": "All tests passed! The refactoring is complete.",
  "message_type": "success"
}`)
	fmt.Println("\nOutput:")
	fmt.Println("✅ All tests passed! The refactoring is complete.")

	fmt.Println()
	fmt.Println("=== Available Message Types ===")
	fmt.Println()
	messageTypes := []struct {
		Type        string
		Icon        string
		Description string
	}{
		{"info", "ℹ️", "General information"},
		{"task", "📋", "Task-related information"},
		{"update", "🔄", "Progress updates"},
		{"warning", "⚠️", "Warnings or cautions"},
		{"success", "✅", "Success messages"},
		{"plan", "🎯", "Plans or strategies"},
	}

	for _, mt := range messageTypes {
		fmt.Printf("%-10s %s  %s\n", mt.Type, mt.Icon, mt.Description)
	}

	fmt.Println()
	fmt.Println("=== Integration with Agent ===")
	fmt.Println()
	fmt.Println("These tools are automatically registered and available to the coding agent.")
	fmt.Println("The agent can use them to:")
	fmt.Println("  • Communicate plans before executing tasks")
	fmt.Println("  • Track progress during multi-step operations")
	fmt.Println("  • Provide updates during long-running tasks")
	fmt.Println("  • Display warnings when issues are detected")
	fmt.Println("  • Confirm success after completing operations")
	fmt.Println()
}
