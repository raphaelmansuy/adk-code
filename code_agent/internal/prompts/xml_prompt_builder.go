// XML-tagged prompt builder for improved LLM parsing
package agent_prompts

import (
	"fmt"
	"strings"

	"code_agent/internal/prompts/prompts"
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
	HasMCPTools          bool // Indicates if MCP (Model Context Protocol) tools are available
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
	buf.WriteString("You are an expert AI assistant with state-of-the-art capabilities spanning coding, analysis, writing, problem-solving, and general knowledge tasks.\n")
	buf.WriteString("Your purpose is to help users with a wide variety of tasks including:\n")
	buf.WriteString("- Coding and software engineering (reading files, writing code, executing commands, debugging)\n")
	buf.WriteString("- Writing and creative tasks (essays, poetry, stories, explanations)\n")
	buf.WriteString("- Analysis and research (breaking down problems, finding information, evaluating solutions)\n")
	buf.WriteString("- General assistance (answering questions, providing guidance, offering suggestions)\n")
	buf.WriteString("\n")
	buf.WriteString("You approach all tasks with the same rigor and iterative problem-solving mindset as you do with coding.\n")
	buf.WriteString("</agent_identity>\n")

	// Workspace context (conditional)
	if ctx.HasWorkspace {
		buf.WriteString("\n<workspace_context>\n")
		buf.WriteString(pb.renderWorkspaceContext(ctx))
		buf.WriteString("</workspace_context>\n")
	}

	// Tools section
	buf.WriteString("\n<tools>\n")
	if ctx.HasMCPTools {
		buf.WriteString("<tool_sources>\n")
		buf.WriteString("You have access to two types of tools:\n\n")
		buf.WriteString("1. **Built-in tools** (prefixed with `builtin_`): Native code-agent tools optimized for coding tasks\n")
		buf.WriteString("   - Use these for standard file operations, code editing, and command execution\n")
		buf.WriteString("   - Examples: builtin_read_file, builtin_write_file, builtin_list_directory\n\n")
		buf.WriteString("2. **MCP tools** (no prefix): External Model Context Protocol server tools\n")
		buf.WriteString("   - Use these when explicitly requested (e.g., \"use MCP\", \"use filesystem tool\")\n")
		buf.WriteString("   - Examples: read_text_file, write_file, list_directory, edit_file\n")
		buf.WriteString("   - Provide additional capabilities and integrations\n\n")
		buf.WriteString("**When to use which:**\n")
		buf.WriteString("- By default, use built-in tools for faster and more integrated operations\n")
		buf.WriteString("- Use MCP tools when the user explicitly requests them or when they provide specific capabilities not available in built-in tools\n")
		buf.WriteString("- MCP tools may have different parameter formats and response structures\n")
		buf.WriteString("</tool_sources>\n\n")
	}
	buf.WriteString(pb.renderToolsXML())
	buf.WriteString("</tools>\n")

	// Guidance section (decision trees and best practices)
	// Note: Don't escape this content as it contains intentional markdown formatting
	buf.WriteString("\n<guidance><![CDATA[\n")
	buf.WriteString(prompts.GuidanceSection)
	buf.WriteString("\n]]></guidance>\n")

	// Critical rules (extracted from pitfalls)
	buf.WriteString("\n<critical_rules priority=\"must_follow\"><![CDATA[\n")
	buf.WriteString(prompts.PitfallsSection)
	buf.WriteString("\n]]></critical_rules>\n")

	// Workflow patterns
	buf.WriteString("\n<workflow_patterns><![CDATA[\n")
	buf.WriteString(prompts.WorkflowSection)
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
	return prompts.ValidatePromptStructure(prompt)
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
