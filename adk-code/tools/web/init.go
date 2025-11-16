// Package web provides web content fetching tools for the coding agent.
package web

// init registers all web tools automatically at package initialization.
func init() {
	// Auto-register Fetch Web tool
	_, _ = NewFetchWebTool()
}
