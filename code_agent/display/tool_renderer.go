package display

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ToolRenderer provides specialized rendering for tool calls and results.
type ToolRenderer struct {
	renderer   *Renderer
	mdRenderer *MarkdownRenderer
}

// NewToolRenderer creates a new tool renderer.
func NewToolRenderer(renderer *Renderer) *ToolRenderer {
	mdRenderer, _ := NewMarkdownRenderer()
	return &ToolRenderer{
		renderer:   renderer,
		mdRenderer: mdRenderer,
	}
}

// RenderToolApproval renders a tool approval request ("Agent wants to...")
func (tr *ToolRenderer) RenderToolApproval(toolName string, args map[string]any) string {
	header := tr.generateToolHeader(toolName, args, "wants to")
	preview := tr.generateToolPreview(toolName, args)

	var output strings.Builder
	output.WriteString("\n")
	// Add a subtle bullet point and render as styled text (not markdown)
	output.WriteString("  ")
	output.WriteString(tr.renderer.Yellow("▸"))
	output.WriteString(" ")
	output.WriteString(header)
	output.WriteString("\n")

	if preview != "" {
		output.WriteString(preview)
		output.WriteString("\n")
	}

	return output.String()
}

// RenderToolExecution renders a tool execution announcement ("Agent is ...ing")
func (tr *ToolRenderer) RenderToolExecution(toolName string, args map[string]any) string {
	header := tr.generateToolHeader(toolName, args, "is")

	var output strings.Builder
	output.WriteString("\n")
	// Add a subtle bullet point and render as styled text (not markdown)
	output.WriteString("  ")
	output.WriteString(tr.renderer.Blue("▸"))
	output.WriteString(" ")
	output.WriteString(header)
	output.WriteString("\n")

	return output.String()
}

// generateToolHeader generates a contextual header based on the tool and verb tense
// Returns plain text that will be styled by the renderer
func (tr *ToolRenderer) generateToolHeader(toolName string, args map[string]any, verbTense string) string {
	var action string
	var path string

	// Extract path from args if present
	if p, ok := args["path"].(string); ok {
		path = tr.renderer.truncatePath(p, 60)
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

	case "list_directory":
		if verbTense == "wants to" {
			action = "wants to list files in"
		} else {
			action = "is listing files in"
		}
		if path != "" {
			return fmt.Sprintf("Agent %s %s", action, tr.renderer.Dim(path))
		}
		return fmt.Sprintf("Agent %s directory", action)

	case "execute_command", "execute_program":
		if command, ok := args["command"].(string); ok {
			if verbTense == "wants to" {
				return fmt.Sprintf("Agent wants to run %s", tr.renderer.Cyan("`"+command+"`"))
			}
			return fmt.Sprintf("Agent is running %s", tr.renderer.Cyan("`"+command+"`"))
		}
		if program, ok := args["program"].(string); ok {
			if verbTense == "wants to" {
				return fmt.Sprintf("Agent wants to run %s", tr.renderer.Cyan("`"+program+"`"))
			}
			return fmt.Sprintf("Agent is running %s", tr.renderer.Cyan("`"+program+"`"))
		}
		if verbTense == "wants to" {
			return "Agent wants to run command"
		}
		return "Agent is running command"

	case "grep_search", "search_files":
		if pattern, ok := args["pattern"].(string); ok {
			if verbTense == "wants to" {
				return fmt.Sprintf("Agent wants to search for %s", tr.renderer.Cyan("`"+pattern+"`"))
			}
			return fmt.Sprintf("Agent is searching for %s", tr.renderer.Cyan("`"+pattern+"`"))
		}
		if verbTense == "wants to" {
			return "Agent wants to search files"
		}
		return "Agent is searching files"

	default:
		if verbTense == "wants to" {
			return fmt.Sprintf("Agent wants to use %s", tr.renderer.Cyan("`"+toolName+"`"))
		}
		return fmt.Sprintf("Agent is using %s", tr.renderer.Cyan("`"+toolName+"`"))
	}
}

// generateToolPreview generates preview content for tool approvals
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

// RenderToolCallDetailed renders a tool call with detailed argument display.
func (tr *ToolRenderer) RenderToolCallDetailed(toolName string, args map[string]any) string {
	var output strings.Builder

	// Contextual header
	header := tr.renderer.getToolHeader(toolName, args)
	rendered := tr.renderer.RenderMarkdown(header)
	output.WriteString("\n")
	output.WriteString(rendered)

	// Render arguments (if not already shown in header)
	switch toolName {
	case "read_file", "write_file", "list_directory", "replace_in_file", "execute_command", "grep_search":
		// Already shown in header
	default:
		// Show arguments for other tools
		if len(args) > 0 {
			output.WriteString(tr.renderer.Dim("\n  Arguments:\n"))
			for k, v := range args {
				output.WriteString(tr.renderer.Dim(fmt.Sprintf("    %s: %v\n", k, v)))
			}
		}
	}

	output.WriteString("\n")
	return output.String()
}

// RenderToolResultDetailed renders a tool result with detailed output.
func (tr *ToolRenderer) RenderToolResultDetailed(toolName string, result map[string]any) string {
	var output strings.Builder

	// Extract error if present
	if errStr, ok := result["error"].(string); ok && errStr != "" {
		output.WriteString(tr.renderer.ErrorX(fmt.Sprintf("  Tool failed: %s\n", errStr)))
		return output.String()
	}

	// Success indicator
	output.WriteString(tr.renderer.Dim("  ✓ Completed"))

	// Tool-specific result rendering
	switch toolName {
	case "read_file":
		if content, ok := result["content"].(string); ok {
			lines := strings.Count(content, "\n") + 1
			bytes := len(content)
			output.WriteString(tr.renderer.Dim(fmt.Sprintf(" - %d lines, %d bytes", lines, bytes)))
		}

	case "write_file":
		if bytesWritten, ok := result["bytes_written"].(int); ok {
			output.WriteString(tr.renderer.Dim(fmt.Sprintf(" - %d bytes written", bytesWritten)))
		}

	case "list_directory":
		if entries, ok := result["entries"].([]any); ok {
			output.WriteString(tr.renderer.Dim(fmt.Sprintf(" - %d entries", len(entries))))
		}

	case "execute_command":
		if exitCode, ok := result["exit_code"].(int); ok {
			if exitCode == 0 {
				output.WriteString(tr.renderer.Dim(" - exit code 0"))
			} else {
				output.WriteString(tr.renderer.Red(fmt.Sprintf(" - exit code %d", exitCode)))
			}
		}

	case "grep_search":
		if matches, ok := result["matches"].([]any); ok {
			output.WriteString(tr.renderer.Dim(fmt.Sprintf(" - %d matches", len(matches))))
		}
	}

	output.WriteString("\n")
	return output.String()
}

// RenderToolCallJSON renders a tool call as JSON.
func (tr *ToolRenderer) RenderToolCallJSON(toolName string, args map[string]any) string {
	data := map[string]any{
		"type": "tool_call",
		"tool": toolName,
		"args": args,
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"type":"tool_call","tool":"%s","error":"failed to marshal"}`, toolName)
	}

	return string(jsonBytes) + "\n"
}

// RenderToolResultJSON renders a tool result as JSON.
func (tr *ToolRenderer) RenderToolResultJSON(toolName string, result map[string]any) string {
	data := map[string]any{
		"type":   "tool_result",
		"tool":   toolName,
		"result": result,
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"type":"tool_result","tool":"%s","error":"failed to marshal"}`, toolName)
	}

	return string(jsonBytes) + "\n"
}

// RenderDiff renders a file diff with syntax highlighting.
func (tr *ToolRenderer) RenderDiff(diff string) string {
	var output strings.Builder

	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			output.WriteString(tr.renderer.Green(line))
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			output.WriteString(tr.renderer.Red(line))
		} else if strings.HasPrefix(line, "@@") {
			output.WriteString(tr.renderer.Cyan(line))
		} else {
			output.WriteString(tr.renderer.Dim(line))
		}
		output.WriteString("\n")
	}

	return output.String()
}

// RenderFileTree renders a file tree structure.
func (tr *ToolRenderer) RenderFileTree(entries []map[string]any, indent int) string {
	var output strings.Builder

	for i, entry := range entries {
		name, _ := entry["name"].(string)
		isDir, _ := entry["is_dir"].(bool)

		// Tree characters
		prefix := strings.Repeat("  ", indent)
		if i == len(entries)-1 {
			prefix += "└── "
		} else {
			prefix += "├── "
		}

		if isDir {
			output.WriteString(tr.renderer.Blue(prefix + name + "/"))
		} else {
			output.WriteString(tr.renderer.Dim(prefix + name))
		}
		output.WriteString("\n")

		// Recursively render children if present
		if children, ok := entry["children"].([]map[string]any); ok {
			output.WriteString(tr.RenderFileTree(children, indent+1))
		}
	}

	return output.String()
}

// RenderProgress renders a progress indicator.
func (tr *ToolRenderer) RenderProgress(current, total int, message string) string {
	percent := float64(current) / float64(total) * 100
	bar := tr.createProgressBar(percent, 40)

	return fmt.Sprintf("\r%s %s %.0f%% (%d/%d)",
		tr.renderer.Cyan(bar),
		message,
		percent,
		current,
		total,
	)
}

// createProgressBar creates a progress bar string.
func (tr *ToolRenderer) createProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return bar
}

// RenderTable renders data as a simple table.
func (tr *ToolRenderer) RenderTable(headers []string, rows [][]string) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var output strings.Builder

	// Header row
	for i, header := range headers {
		padding := widths[i] - len(header) + 2
		output.WriteString(tr.renderer.Bold(header))
		output.WriteString(strings.Repeat(" ", padding))
	}
	output.WriteString("\n")

	// Separator
	for _, width := range widths {
		output.WriteString(strings.Repeat("─", width+2))
	}
	output.WriteString("\n")

	// Data rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				padding := widths[i] - len(cell) + 2
				output.WriteString(tr.renderer.Dim(cell))
				output.WriteString(strings.Repeat(" ", padding))
			}
		}
		output.WriteString("\n")
	}

	return output.String()
}
