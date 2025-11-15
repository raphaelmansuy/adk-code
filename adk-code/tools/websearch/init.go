// Package websearch provides web search tools for the coding agent.
package websearch

// init registers all web search tools automatically at package initialization.
func init() {
	// Auto-register Google Search tool
	_, _ = NewGoogleSearchTool()
}
