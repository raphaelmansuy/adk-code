// Package search provides code search and diff preview tools for the coding agent.
package search

// init registers all search tools automatically at package initialization.
func init() {
	// Auto-register all search tools
	_, _ = NewPreviewReplaceTool()
}
