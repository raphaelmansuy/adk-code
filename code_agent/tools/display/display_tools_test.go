package display

import (
	"testing"
)

// TestDisplayMessageTool tests the display_message tool creation and basic functionality.
func TestDisplayMessageTool(t *testing.T) {
	tool, err := NewDisplayMessageTool()
	if err != nil {
		t.Fatalf("Failed to create display_message tool: %v", err)
	}

	if tool == nil {
		t.Fatal("Expected non-nil tool")
	}

	// Test that the tool has the correct name
	if tool.Name() != "display_message" {
		t.Errorf("Expected tool name 'display_message', got '%s'", tool.Name())
	}
}

// TestUpdateTaskListTool tests the update_task_list tool creation and basic functionality.
func TestUpdateTaskListTool(t *testing.T) {
	tool, err := NewUpdateTaskListTool()
	if err != nil {
		t.Fatalf("Failed to create update_task_list tool: %v", err)
	}

	if tool == nil {
		t.Fatal("Expected non-nil tool")
	}

	// Test that the tool has the correct name
	if tool.Name() != "update_task_list" {
		t.Errorf("Expected tool name 'update_task_list', got '%s'", tool.Name())
	}
}

// TestDisplayMessageInputOutput tests the display message input/output structures.
func TestDisplayMessageInputOutput(t *testing.T) {
	// Create the tool
	displayTool, err := NewDisplayMessageTool()
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Test basic info message
	t.Run("BasicInfoMessage", func(t *testing.T) {
		input := DisplayMessageInput{
			Title:       "Test Title",
			Content:     "This is a test message",
			MessageType: "info",
		}

		// We can't directly invoke the handler, but we can verify the tool is properly configured
		_ = input
		_ = displayTool
	})

	// Test task message
	t.Run("TaskMessage", func(t *testing.T) {
		input := DisplayMessageInput{
			Title:       "Current Tasks",
			Content:     "- [ ] Task 1\n- [x] Task 2\n- [ ] Task 3",
			MessageType: "task",
		}

		_ = input
	})

	// Test different message types
	messageTypes := []string{"info", "task", "update", "warning", "success", "plan"}
	for _, msgType := range messageTypes {
		t.Run("MessageType_"+msgType, func(t *testing.T) {
			icon := getIconForMessageType(msgType)
			if icon == "" {
				t.Errorf("Expected non-empty icon for message type '%s'", msgType)
			}
		})
	}
}

// TestUpdateTaskListInputOutput tests the update task list input/output structures.
func TestUpdateTaskListInputOutput(t *testing.T) {
	// Create the tool
	updateTool, err := NewUpdateTaskListTool()
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	_ = updateTool

	// Test task list parsing
	t.Run("TaskListParsing", func(t *testing.T) {
		taskList := `- [ ] Task 1
- [x] Task 2
- [ ] Task 3
- [x] Task 4`

		input := UpdateTaskListInput{
			TaskList: taskList,
			Title:    "My Tasks",
		}

		_ = input

		// Expected: 4 total tasks, 2 completed
		// This would be verified when the tool is actually invoked
	})

	// Test empty task list
	t.Run("EmptyTaskList", func(t *testing.T) {
		input := UpdateTaskListInput{
			TaskList: "",
			Title:    "Empty List",
		}

		_ = input
	})

	// Test all completed
	t.Run("AllCompleted", func(t *testing.T) {
		input := UpdateTaskListInput{
			TaskList: "- [x] Task 1\n- [x] Task 2\n- [x] Task 3",
			Title:    "Completed",
		}

		_ = input
	})

	// Test all pending
	t.Run("AllPending", func(t *testing.T) {
		input := UpdateTaskListInput{
			TaskList: "- [ ] Task 1\n- [ ] Task 2\n- [ ] Task 3",
			Title:    "Pending",
		}

		_ = input
	})
}

// TestGetIconForMessageType tests the icon selection logic.
func TestGetIconForMessageType(t *testing.T) {
	tests := []struct {
		messageType string
		wantIcon    string
	}{
		{"info", "ℹ️"},
		{"task", "📋"},
		{"update", "🔄"},
		{"warning", "⚠️"},
		{"success", "✅"},
		{"plan", "🎯"},
		{"unknown", "ℹ️"}, // Default to info
		{"", "ℹ️"},        // Empty defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.messageType, func(t *testing.T) {
			got := getIconForMessageType(tt.messageType)
			if got != tt.wantIcon {
				t.Errorf("getIconForMessageType(%q) = %q, want %q", tt.messageType, got, tt.wantIcon)
			}
		})
	}
}
