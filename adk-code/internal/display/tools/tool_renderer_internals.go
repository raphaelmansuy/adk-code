package tools

import (
	"fmt"
	"strings"

	"adk-code/internal/display/components"
)

// generateToolHeader generates a contextual header based on the tool and verb tense
// Returns plain text that will be styled by the renderer
func (tr *ToolRenderer) generateToolHeader(toolName string, args map[string]any, verbTense string) string {
	var action string
	var path string

	// Extract path from args if present
	if p, ok := args["path"].(string); ok {
		path = components.ShortenPath(p, 60)
	}

	switch toolName {
	case "read_file":
		if verbTense == "wants to" {
			action = "wants to read"
		} else {
			action = "is reading"
		}
		if path != "" {
			return fmt.Sprintf("Agent %s %s", action, tr.renderer.Dim(path))
		}
		return fmt.Sprintf("Agent %s file", action)

	case "write_file":
		if verbTense == "wants to" {
			action = "wants to write"
		} else {
			action = "is writing"
		}
		if path != "" {
			return fmt.Sprintf("Agent %s %s", action, tr.renderer.Dim(path))
		}
		return fmt.Sprintf("Agent %s file", action)

	case "replace_in_file", "search_replace":
		if verbTense == "wants to" {
			action = "wants to edit"
		} else {
			action = "is editing"
		}
		if path != "" {
			return fmt.Sprintf("Agent %s %s", action, tr.renderer.Dim(path))
		}
		return fmt.Sprintf("Agent %s file", action)

	case "delete_directory":
		if verbTense == "wants to" {
			action = "wants to delete"
		} else {
			action = "is deleting"
		}
		if path != "" {
			return fmt.Sprintf("Agent %s %s", action, tr.renderer.Dim(path))
		}
		return fmt.Sprintf("Agent %s directory", action)

	case "execute_command":
		if verbTense == "wants to" {
			action = "wants to execute"
		} else {
			action = "is executing"
		}
		if cmd, ok := args["command"].(string); ok {
			return fmt.Sprintf("Agent %s %s", action, tr.renderer.Dim(components.ShortenPath(cmd, 60)))
		}
		return fmt.Sprintf("Agent %s command", action)

	case "list_directory":
		if verbTense == "wants to" {
			action = "wants to list"
		} else {
			action = "is listing"
		}
		if path != "" {
			return fmt.Sprintf("Agent %s %s", action, tr.renderer.Dim(path))
		}
		return fmt.Sprintf("Agent %s directory", action)

	case "search_files":
		if verbTense == "wants to" {
			action = "wants to search"
		} else {
			action = "is searching"
		}
		if pattern, ok := args["pattern"].(string); ok {
			return fmt.Sprintf("Agent %s for %s", action, tr.renderer.Dim(pattern))
		}
		return fmt.Sprintf("Agent %s files", action)

	case "grep_search":
		if verbTense == "wants to" {
			action = "wants to grep"
		} else {
			action = "is grepping"
		}
		if pattern, ok := args["query"].(string); ok {
			return fmt.Sprintf("Agent %s for %s", action, tr.renderer.Dim(pattern))
		}
		return fmt.Sprintf("Agent %s files", action)

	default:
		// Generic tool name formatting
		formattedTool := strings.ToLower(toolName)
		formattedTool = strings.ReplaceAll(formattedTool, "_", " ")
		if verbTense == "wants to" {
			return fmt.Sprintf("Agent wants to %s", formattedTool)
		}
		return fmt.Sprintf("Agent is %sing", formattedTool)
	}
}

// generateToolPreview generates a preview snippet for a tool call
func (tr *ToolRenderer) generateToolPreview(toolName string, args map[string]any) string {
	switch toolName {
	case "write_file":
		// Show content preview for write operations
		if content, ok := args["content"].(string); ok {
			preview := strings.TrimSpace(content)
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			previewMd := fmt.Sprintf("```\n%s\n```", preview)
			if tr.mdRenderer != nil {
				rendered, err := tr.mdRenderer.Render(previewMd)
				if err == nil {
					return rendered
				}
			}
			return previewMd
		}

	case "replace_in_file", "search_replace":
		// Show diff preview for edits
		if oldStr, ok := args["old_string"].(string); ok {
			if newStr, ok2 := args["new_string"].(string); ok2 {
				diff := fmt.Sprintf("- %s\n+ %s", oldStr, newStr)
				diffMd := fmt.Sprintf("```diff\n%s\n```", diff)
				if tr.mdRenderer != nil {
					rendered, err := tr.mdRenderer.Render(diffMd)
					if err == nil {
						return rendered
					}
				}
				return diffMd
			}
		}

	case "execute_command":
		// Show command in code block
		if command, ok := args["command"].(string); ok {
			cmdMd := fmt.Sprintf("```shell\n%s\n```", command)
			if tr.mdRenderer != nil {
				rendered, err := tr.mdRenderer.Render(cmdMd)
				if err == nil {
					return rendered
				}
			}
			return cmdMd
		}
	}

	return ""
}

// createProgressBar creates a progress bar string.
func (tr *ToolRenderer) createProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return bar
}
