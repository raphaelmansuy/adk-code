// Package display provides display and UI tools for the coding agent.
package display

// init registers all display tools automatically at package initialization.
func init() {
	// Auto-register all display tools
	_, _ = NewDisplayMessageTool()
	_, _ = NewUpdateTaskListTool()
}
