// Package edit provides code editing tools for the coding agent.
package edit

// init registers all edit tools automatically at package initialization.
func init() {
	// Auto-register all edit tools
	_, _ = NewApplyPatchTool()
	_, _ = NewEditLinesTool()
	_, _ = NewSearchReplaceTool()
}
