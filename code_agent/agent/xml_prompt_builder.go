// XML-tagged prompt builder for improved LLM parsing
package agent

import (
	"fmt"
	"strings"

	"code_agent/tools"
)

// PromptContext holds contextual information for dynamic prompt generation
type PromptContext struct {
	HasWorkspace         bool
	WorkspaceSummary     string
	WorkspaceRoot        string
	EnvironmentMetadata  string
	TaskType             string
	EnableMultiWorkspace bool
}

// PromptBuilder builds XML-tagged prompts from registered tools and context
type PromptBuilder struct {
	registry *tools.ToolRegistry
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder(registry *tools.ToolRegistry) *PromptBuilder {
	return &PromptBuilder{
		registry: registry,
	}
}

// BuildXMLPrompt generates a complete XML-tagged system prompt
func (pb *PromptBuilder) BuildXMLPrompt(ctx PromptContext) string {
	var buf strings.Builder

	buf.WriteString("<agent_system_prompt>\n\n")

	// Agent identity
	buf.WriteString("<agent_identity>\n")
	buf.WriteString("You are an expert AI coding assistant with state-of-the-art file editing capabilities.\n")
	buf.WriteString("Your purpose is to help users with coding tasks by reading files, writing code, executing commands, and iteratively solving problems.\n")
	buf.WriteString("</agent_identity>\n")

	// Workspace context (conditional)
	if ctx.HasWorkspace {
		buf.WriteString("\n<workspace_context>\n")
		buf.WriteString(pb.renderWorkspaceContext(ctx))
		buf.WriteString("</workspace_context>\n")
	}

	// Tools section
	buf.WriteString("\n<tools>\n")
	buf.WriteString(pb.renderToolsXML())
	buf.WriteString("</tools>\n")

	// Guidance section (decision trees and best practices)
	// Note: Don't escape this content as it contains intentional markdown formatting
	buf.WriteString("\n<guidance><![CDATA[\n")
	buf.WriteString(GuidanceSection)
	buf.WriteString("\n]]></guidance>\n")

	// Critical rules (extracted from pitfalls)
	buf.WriteString("\n<critical_rules priority=\"must_follow\"><![CDATA[\n")
	buf.WriteString(PitfallsSection)
	buf.WriteString("\n]]></critical_rules>\n")

	// Workflow patterns
	buf.WriteString("\n<workflow_patterns><![CDATA[\n")
	buf.WriteString(WorkflowSection)
	buf.WriteString("\n]]></workflow_patterns>\n")

	buf.WriteString("</agent_system_prompt>")

	return buf.String()
}

// renderWorkspaceContext generates workspace context information
func (pb *PromptBuilder) renderWorkspaceContext(ctx PromptContext) string {
	var buf strings.Builder

	buf.WriteString("<file_system>\n")
	buf.WriteString(fmt.Sprintf("Primary workspace: %s\n\n", ctx.WorkspaceRoot))
	buf.WriteString(ctx.WorkspaceSummary)
	buf.WriteString("\n</file_system>\n")

	if ctx.EnvironmentMetadata != "" {
		buf.WriteString("\n<environment>\n")
		buf.WriteString(ctx.EnvironmentMetadata)
		buf.WriteString("\n</environment>\n")
	}

	// Path usage instructions
	buf.WriteString("\n<path_usage>\n")
	buf.WriteString("All file paths should be relative to the primary workspace directory.\n")
	buf.WriteString("Examples:\n")
	buf.WriteString("- Current directory file: \"./filename.ext\" or \"filename.ext\"\n")
	buf.WriteString("- Subdirectory file: \"./subdir/filename.ext\" or \"subdir/filename.ext\"\n")
	buf.WriteString("- Do NOT prefix paths with the working directory name\n")

	if ctx.EnableMultiWorkspace {
		buf.WriteString("\n### Multi-workspace Mode\n")
		buf.WriteString("Use @workspace:path syntax to target specific workspaces:\n")
		buf.WriteString("- @frontend:src/index.ts - targets the frontend workspace\n")
		buf.WriteString("- @backend:api/server.go - targets the backend workspace\n")
	}
	buf.WriteString("</path_usage>\n")

	return buf.String()
}

// escapeXMLContent escapes special XML characters to prevent tag confusion
func escapeXMLContent(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// renderToolsXML generates XML-tagged tool documentation
func (pb *PromptBuilder) renderToolsXML() string {
	var buf strings.Builder

	// Get all categories in the predefined order
	categories := pb.registry.GetCategories()

	for _, category := range categories {
		buf.WriteString(fmt.Sprintf("<tool_category name=%q>\n", category))

		// Get tools in this category (already sorted by priority)
		toolsInCategory := pb.registry.GetByCategory(category)

		for _, metadata := range toolsInCategory {
			tool := metadata.Tool
			buf.WriteString(fmt.Sprintf("  <tool name=%q>\n", tool.Name()))
			buf.WriteString("    <description>\n")
			buf.WriteString(escapeXMLContent(tool.Description()))
			buf.WriteString("\n    </description>\n")

			if metadata.UsageHint != "" {
				buf.WriteString("    <usage_hint>")
				buf.WriteString(escapeXMLContent(metadata.UsageHint))
				buf.WriteString("</usage_hint>\n")
			}

			buf.WriteString("  </tool>\n")
		}

		buf.WriteString("</tool_category>\n\n")
	}

	return buf.String()
}

// ValidatePromptStructure validates that XML tags are properly balanced
// Note: This is a simple validator that handles basic XML structure.
// It processes tags inline and handles escaped content (&lt;, &gt;, etc.) and CDATA sections
func ValidatePromptStructure(prompt string) error {
	stack := []string{}
	pos := 0

	for pos < len(prompt) {
		// Find next '<'
		tagStart := strings.IndexByte(prompt[pos:], '<')
		if tagStart == -1 {
			break // No more tags
		}
		tagStart += pos

		// Check for CDATA section
		if strings.HasPrefix(prompt[tagStart:], "<![CDATA[") {
			// Find end of CDATA
			cdataEnd := strings.Index(prompt[tagStart:], "]]>")
			if cdataEnd == -1 {
				return fmt.Errorf("unclosed CDATA section starting at position %d", tagStart)
			}
			pos = tagStart + cdataEnd + 3
			continue
		}

		// Find matching '>'
		tagEnd := strings.IndexByte(prompt[tagStart:], '>')
		if tagEnd == -1 {
			return fmt.Errorf("unclosed tag starting at position %d", tagStart)
		}
		tagEnd += tagStart

		// Extract tag content (between < and >)
		tagContent := prompt[tagStart+1 : tagEnd]
		tagContent = strings.TrimSpace(tagContent)

		// Skip empty tags, XML declarations, or comments
		if tagContent == "" || strings.HasPrefix(tagContent, "?") || strings.HasPrefix(tagContent, "!") {
			pos = tagEnd + 1
			continue
		}

		// Check if it's a closing tag
		if strings.HasPrefix(tagContent, "/") {
			// Closing tag
			tagName := strings.TrimSpace(tagContent[1:])
			// Extract just the tag name (before any space)
			if idx := strings.IndexAny(tagName, " \t\n"); idx != -1 {
				tagName = tagName[:idx]
			}

			if len(stack) == 0 {
				return fmt.Errorf("unexpected closing tag </%s> with no matching opening tag at position %d", tagName, tagStart)
			}
			lastTag := stack[len(stack)-1]
			if lastTag != tagName {
				return fmt.Errorf("mismatched tag at position %d - expected </%s> but got </%s>", tagStart, lastTag, tagName)
			}
			stack = stack[:len(stack)-1]
		} else if !strings.HasSuffix(tagContent, "/") {
			// Opening tag (ignore self-closing tags ending with /)
			// Extract tag name (before space or special chars)
			tagName := tagContent
			if idx := strings.IndexAny(tagName, " \t\n"); idx != -1 {
				tagName = tagName[:idx]
			}
			stack = append(stack, tagName)
		}

		pos = tagEnd + 1
	}

	if len(stack) > 0 {
		return fmt.Errorf("unclosed tags: %v", stack)
	}

	return nil
}

// BuildEnhancedPromptV2 is a backward-compatible wrapper that uses XML tagging.
// This was created during Phase 1 migration but is now deprecated.
//
// Deprecated: Use BuildEnhancedPromptWithContext() directly with a proper PromptContext
func BuildEnhancedPromptV2(registry *tools.ToolRegistry) string {
	builder := NewPromptBuilder(registry)

	// Build with minimal context for backward compatibility
	ctx := PromptContext{
		HasWorkspace: false, // Will be filled in by caller
	}

	return builder.BuildXMLPrompt(ctx)
}

// BuildEnhancedPromptWithContext builds an XML-structured prompt with full context information.
// This is the primary function for generating system prompts with:
// - XML tags for hierarchical structure
// - Conditional sections based on context (workspace, environment, etc.)
// - CDATA sections for content with special characters
// - Proper escaping for tool descriptions
//
// Use this function instead of the deprecated BuildEnhancedPrompt or BuildEnhancedPromptV2.
func BuildEnhancedPromptWithContext(registry *tools.ToolRegistry, ctx PromptContext) string {
	builder := NewPromptBuilder(registry)
	return builder.BuildXMLPrompt(ctx)
}
