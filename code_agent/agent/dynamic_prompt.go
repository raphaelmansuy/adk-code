// Dynamic prompt generation from tool registry
//
// DEPRECATED: This file contains legacy prompt builders that generate flat text prompts.
// New code should use xml_prompt_builder.go which provides XML-structured prompts
// that LLMs can parse more effectively.
//
// These functions are kept only for backward compatibility with existing tests.
package agent

import (
	"fmt"
	"strings"

	"code_agent/tools"
)

// BuildToolsSection generates the tools section of the system prompt from the registry.
// It organizes tools by category and includes name, description, and usage hints.
//
// Deprecated: Use PromptBuilder.renderToolsXML() from xml_prompt_builder.go for XML-structured output
func BuildToolsSection(registry *tools.ToolRegistry) string {
	var builder strings.Builder

	builder.WriteString(`You are an expert AI coding assistant with state-of-the-art file editing capabilities. Your purpose is to help users with coding tasks by reading files, writing code, executing commands, and iteratively solving problems.

## Available Tools

`)

	// Get all categories in the predefined order
	categories := registry.GetCategories()

	for _, category := range categories {
		// Write category header
		builder.WriteString(fmt.Sprintf("### %s\n\n", category))

		// Get tools in this category (already sorted by priority)
		toolsInCategory := registry.GetByCategory(category)

		for _, metadata := range toolsInCategory {
			tool := metadata.Tool
			// Format: **tool_name** - description
			builder.WriteString(fmt.Sprintf("**%s** - %s", tool.Name(), tool.Description()))

			// Add usage hint if provided
			if metadata.UsageHint != "" {
				builder.WriteString(fmt.Sprintf("\n  → *Usage tip: %s*", metadata.UsageHint))
			}
			builder.WriteString("\n\n")
		}
	}

	return builder.String()
}

// BuildEnhancedPrompt combines the dynamic tools section with existing guidance sections.
//
// Deprecated: Use BuildEnhancedPromptWithContext() from xml_prompt_builder.go for XML-structured prompts
func BuildEnhancedPrompt(registry *tools.ToolRegistry) string {
	toolsSection := BuildToolsSection(registry)

	// Combine with existing guidance (keep the static sections)
	return toolsSection + "\n" + GuidanceSection + "\n" + PitfallsSection + "\n" + WorkflowSection
}
