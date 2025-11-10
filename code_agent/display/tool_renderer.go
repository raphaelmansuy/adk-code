package display

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ToolRenderer provides specialized rendering for tool calls and results.
type ToolRenderer struct {
	renderer *Renderer
}

// NewToolRenderer creates a new tool renderer.
func NewToolRenderer(renderer *Renderer) *ToolRenderer {
	return &ToolRenderer{
		renderer: renderer,
	}
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
