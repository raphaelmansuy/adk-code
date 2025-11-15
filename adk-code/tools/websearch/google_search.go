// Package websearch provides web search tools for the coding agent.
package websearch

import (
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"

	common "adk-code/tools/base"
)

// NewGoogleSearchTool creates a Google Search tool using ADK's built-in implementation.
// This tool leverages Google's native search capabilities through the Gemini API.
//
// Requirements:
//   - Only works with Gemini 2.0+ models (gemini-2.0-flash, gemini-2.0-flash-thinking, etc.)
//   - Requires a valid Gemini API key
//   - May return HTML content that needs to be rendered for rich results
//
// The tool enables the agent to:
//   - Search the web for current information
//   - Find documentation and tutorials
//   - Get real-time data and news
//   - Answer questions requiring external knowledge
//
// Usage by agent:
//
//	The agent will automatically invoke this tool when it needs to search
//	for information not in its training data or requiring current web data.
//
// Example queries that trigger this tool:
//   - "What are the latest features in Go 1.24?"
//   - "Find documentation for the ADK framework"
//   - "What's the current weather in San Francisco?"
func NewGoogleSearchTool() (tool.Tool, error) {
	// Use ADK's native Google Search tool implementation
	// This is a zero-configuration tool that integrates directly with Gemini
	searchTool := geminitool.GoogleSearch{}

	// Register the tool in the global registry with appropriate metadata
	common.Register(common.ToolMetadata{
		Tool:      searchTool,
		Category:  common.CategorySearchDiscovery,
		Priority:  0, // Highest priority in Search & Discovery category
		UsageHint: "Search the web for current information, documentation, news, or answers. Requires Gemini 2.0+ models.",
	})

	return searchTool, nil
}
