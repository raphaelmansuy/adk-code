package display

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ToolResultParser provides specialized parsing for tool results
type ToolResultParser struct {
	mdRenderer *MarkdownRenderer
}

// NewToolResultParser creates a new tool result parser
func NewToolResultParser(mdRenderer *MarkdownRenderer) *ToolResultParser {
	return &ToolResultParser{
		mdRenderer: mdRenderer,
	}
}

// ParseToolResult parses and formats tool results for display
func (trp *ToolResultParser) ParseToolResult(toolName string, result map[string]any) string {
	// Check for errors first
	if errStr, ok := result["error"].(string); ok && errStr != "" {
		return trp.formatError(errStr)
	}

	// Tool-specific parsing
	switch toolName {
	case "list_directory":
		return trp.parseListDirectory(result)
	case "grep_search", "search_files":
		return trp.parseSearchResults(result)
	case "execute_command", "execute_program":
		return trp.parseCommandOutput(result)
	case "read_file":
		return trp.parseFileContent(result)
	case "write_file":
		return trp.parseWriteFile(result)
	default:
		return trp.parseGeneric(result)
	}
}

// formatError formats an error message
func (trp *ToolResultParser) formatError(errMsg string) string {
	return fmt.Sprintf("❌ Error: %s", errMsg)
}

// parseListDirectory formats directory listing results
func (trp *ToolResultParser) parseListDirectory(result map[string]any) string {
	var output strings.Builder

	// Extract entries
	entriesRaw, ok := result["entries"].([]any)
	if !ok {
		entriesRaw, ok = result["items"].([]any)
	}
	if !ok {
		return trp.parseGeneric(result)
	}

	output.WriteString(fmt.Sprintf("📁 Found %d items:\n\n", len(entriesRaw)))

	// Group by type
	var dirs, files []string
	for _, entry := range entriesRaw {
		if entryMap, ok := entry.(map[string]any); ok {
			name, _ := entryMap["name"].(string)
			isDir, _ := entryMap["is_directory"].(bool)
			if !isDir {
				isDir, _ = entryMap["is_dir"].(bool)
			}

			if isDir {
				dirs = append(dirs, name)
			} else {
				files = append(files, name)
			}
		} else if name, ok := entry.(string); ok {
			// Simple string entry
			if strings.HasSuffix(name, "/") {
				dirs = append(dirs, strings.TrimSuffix(name, "/"))
			} else {
				files = append(files, name)
			}
		}
	}

	// Render directories
	if len(dirs) > 0 {
		output.WriteString("**Directories:**\n")
		for _, dir := range dirs {
			output.WriteString(fmt.Sprintf("- 📂 %s/\n", dir))
		}
		output.WriteString("\n")
	}

	// Render files
	if len(files) > 0 {
		output.WriteString("**Files:**\n")
		for _, file := range files {
			output.WriteString(fmt.Sprintf("- 📄 %s\n", file))
		}
	}

	// Render as markdown if available
	result_str := output.String()
	if trp.mdRenderer != nil {
		rendered, err := trp.mdRenderer.Render(result_str)
		if err == nil {
			return rendered
		}
	}
	return result_str
}

// parseSearchResults formats search results
func (trp *ToolResultParser) parseSearchResults(result map[string]any) string {
	var output strings.Builder

	// Extract matches
	matchesRaw, ok := result["matches"].([]any)
	if !ok {
		return trp.parseGeneric(result)
	}

	if len(matchesRaw) == 0 {
		return "🔍 No matches found"
	}

	output.WriteString(fmt.Sprintf("🔍 Found %d matches:\n\n", len(matchesRaw)))

	// Render matches
	for i, match := range matchesRaw {
		if i >= 20 {
			output.WriteString(fmt.Sprintf("\n...and %d more matches", len(matchesRaw)-20))
			break
		}

		if matchMap, ok := match.(map[string]any); ok {
			file, _ := matchMap["file"].(string)
			line, _ := matchMap["line"].(float64)
			content, _ := matchMap["content"].(string)

			if file != "" {
				output.WriteString(fmt.Sprintf("**%s:%d**\n", file, int(line)))
				if content != "" {
					output.WriteString(fmt.Sprintf("```\n%s\n```\n\n", strings.TrimSpace(content)))
				}
			}
		}
	}

	result_str := output.String()
	if trp.mdRenderer != nil {
		rendered, err := trp.mdRenderer.Render(result_str)
		if err == nil {
			return rendered
		}
	}
	return result_str
}

// parseCommandOutput formats command execution output
func (trp *ToolResultParser) parseCommandOutput(result map[string]any) string {
	var output strings.Builder

	// Extract exit code
	exitCode := 0
	if code, ok := result["exit_code"].(float64); ok {
		exitCode = int(code)
	} else if code, ok := result["exit_code"].(int); ok {
		exitCode = code
	}

	// Extract stdout/stderr
	stdout, _ := result["stdout"].(string)
	stderr, _ := result["stderr"].(string)
	combinedOutput, _ := result["output"].(string)

	// Show exit status
	if exitCode == 0 {
		output.WriteString("✅ Command completed successfully\n\n")
	} else {
		output.WriteString(fmt.Sprintf("❌ Command failed with exit code %d\n\n", exitCode))
	}

	// Show output
	if combinedOutput != "" {
		output.WriteString("**Output:**\n")
		output.WriteString(fmt.Sprintf("```\n%s\n```\n", strings.TrimSpace(combinedOutput)))
	} else {
		if stdout != "" {
			output.WriteString("**stdout:**\n")
			output.WriteString(fmt.Sprintf("```\n%s\n```\n", strings.TrimSpace(stdout)))
		}
		if stderr != "" {
			output.WriteString("**stderr:**\n")
			output.WriteString(fmt.Sprintf("```\n%s\n```\n", strings.TrimSpace(stderr)))
		}
	}

	result_str := output.String()
	if trp.mdRenderer != nil {
		rendered, err := trp.mdRenderer.Render(result_str)
		if err == nil {
			return rendered
		}
	}
	return result_str
}

// parseFileContent formats file content results
func (trp *ToolResultParser) parseFileContent(result map[string]any) string {
	content, ok := result["content"].(string)
	if !ok {
		return trp.parseGeneric(result)
	}

	lines := strings.Count(content, "\n") + 1
	bytes := len(content)

	return fmt.Sprintf("📄 Read %d lines (%d bytes)", lines, bytes)
}

// parseWriteFile formats write file results
func (trp *ToolResultParser) parseWriteFile(result map[string]any) string {
	bytesWritten := 0
	if b, ok := result["bytes_written"].(float64); ok {
		bytesWritten = int(b)
	} else if b, ok := result["bytes_written"].(int); ok {
		bytesWritten = b
	}

	if bytesWritten > 0 {
		return fmt.Sprintf("✅ Wrote %d bytes", bytesWritten)
	}

	return "✅ File written successfully"
}

// parseGeneric formats generic tool results
func (trp *ToolResultParser) parseGeneric(result map[string]any) string {
	// Try to pretty print as JSON
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", result)
	}

	return fmt.Sprintf("```json\n%s\n```", string(jsonBytes))
}
