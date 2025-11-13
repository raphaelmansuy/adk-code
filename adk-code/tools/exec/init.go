// Package exec provides command execution tools for the coding agent.
package exec

// init registers all execution tools automatically at package initialization.
func init() {
	// Auto-register all execution tools
	_, _ = NewExecuteCommandTool()
	_, _ = NewExecuteProgramTool()
	_, _ = NewGrepSearchTool()
}
