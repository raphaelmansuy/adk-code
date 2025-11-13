package tools

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
	// Check for errors first - handle multiple error formats
	if err := trp.extractError(result); err != "" {
		return trp.formatError(err)
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
	case "display_message":
		return trp.parseDisplayMessage(result)
	case "update_task_list":
		return trp.parseUpdateTaskList(result)
	default:
		return trp.parseGeneric(result)
	}
}

// extractError extracts error messages from various error formats in tool results
// Handles: string errors, empty objects {}, objects with message/details fields
func (trp *ToolResultParser) extractError(result map[string]any) string {
	errorValue, hasError := result["error"]
	if !hasError {
		return ""
	}

	// Handle string error (most common case)
	if errStr, ok := errorValue.(string); ok && errStr != "" {
		return errStr
	}

	// Handle empty error object {} (common with MCP tools that fail)
	if errorMap, ok := errorValue.(map[string]any); ok {
		// If error object is empty, return generic message
		if len(errorMap) == 0 {
			// Check if there's any other useful information in the result
			if output, ok := result["output"].(string); ok && output != "" {
				return output
			}
			// Try to extract tool name for more context
			if toolName, ok := result["tool"].(string); ok && toolName != "" {
				return fmt.Sprintf("Tool '%s' failed with no error details provided", toolName)
			}
			return "Tool execution failed with no error details provided"
		}

		// Try common error field names in the error object
		if msg, ok := errorMap["message"].(string); ok && msg != "" {
			return msg
		}
		if msg, ok := errorMap["error"].(string); ok && msg != "" {
			return msg
		}
		if msg, ok := errorMap["details"].(string); ok && msg != "" {
			return msg
		}
		if msg, ok := errorMap["text"].(string); ok && msg != "" {
			return msg
		}

		// If error object has fields but none match common patterns, stringify it
		if jsonBytes, err := json.MarshalIndent(errorMap, "", "  "); err == nil {
			return fmt.Sprintf("Error details:\n%s", string(jsonBytes))
		}
	}

	// Fallback: convert any other error type to string
	return fmt.Sprintf("%v", errorValue)
}

// formatError formats an error message
func (trp *ToolResultParser) formatError(errMsg string) string {
	return fmt.Sprintf("❌ Error: %s", errMsg)
}

// parseListDirectory formats directory listing results
func (trp *ToolResultParser) parseListDirectory(result map[string]any) string {
	var output strings.Builder

	// Extract entries - try multiple field names
	var entriesRaw []any
	var ok bool

	// Try "files" first (used by our tools)
	if entriesRaw, ok = result["files"].([]any); !ok {
		// Try "entries"
		if entriesRaw, ok = result["entries"].([]any); !ok {
			// Try "items"
			if entriesRaw, ok = result["items"].([]any); !ok {
				// Return empty if no recognized field
				return ""
			}
		}
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

// parseSearchResults formats search results with file grouping
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

	// Group matches by file
	fileMatches := make(map[string]int)
	files := []string{}
	filesMap := make(map[string][]map[string]any)

	for _, match := range matchesRaw {
		if matchMap, ok := match.(map[string]any); ok {
			file, _ := matchMap["file"].(string)
			if file != "" {
				if _, seen := fileMatches[file]; !seen {
					files = append(files, file)
				}
				fileMatches[file]++
				filesMap[file] = append(filesMap[file], matchMap)
			}
		}
	}

	output.WriteString(fmt.Sprintf("🔍 Found %d matches in %d files:\n\n", len(matchesRaw), len(fileMatches)))

	// Show file summary
	for _, file := range files {
		count := fileMatches[file]
		output.WriteString(fmt.Sprintf("**%s** (%d matches)\n", file, count))
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

	// If there's no output to show, return empty (let the success indicator handle it)
	if combinedOutput == "" && stdout == "" && stderr == "" {
		return ""
	}

	var output strings.Builder

	// Show exit status only if there was an error
	if exitCode != 0 {
		output.WriteString(fmt.Sprintf("❌ Command failed with exit code %d\n\n", exitCode))
	}

	// Show output concisely
	if combinedOutput != "" {
		trimmed := strings.TrimSpace(combinedOutput)
		if trimmed != "" {
			output.WriteString(fmt.Sprintf("```\n%s\n```", trimmed))
		}
	} else {
		if stdout != "" {
			trimmed := strings.TrimSpace(stdout)
			if trimmed != "" {
				output.WriteString(fmt.Sprintf("```\n%s\n```", trimmed))
			}
		}
		if stderr != "" {
			trimmed := strings.TrimSpace(stderr)
			if trimmed != "" {
				if stdout != "" {
					output.WriteString("\n\n**stderr:**\n")
				}
				output.WriteString(fmt.Sprintf("```\n%s\n```", trimmed))
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

// parseFileContent formats file content results with better info display
func (trp *ToolResultParser) parseFileContent(result map[string]any) string {
	// Extract file path and total lines count
	filePath, _ := result["file_path"].(string)
	totalLines := 0
	if tl, ok := result["total_lines"].(float64); ok {
		totalLines = int(tl)
	} else if tl, ok := result["total_lines"].(int); ok {
		totalLines = tl
	}

	// Extract file size if available
	fileSize := 0
	if fs, ok := result["file_size"].(float64); ok {
		fileSize = int(fs)
	} else if fs, ok := result["file_size"].(int); ok {
		fileSize = fs
	}

	// If no total_lines, count from content (fallback)
	if totalLines == 0 {
		if content, ok := result["content"].(string); ok {
			totalLines = strings.Count(content, "\n") + 1
		}
	}

	var output strings.Builder

	// Show path and line count
	if filePath != "" {
		output.WriteString(fmt.Sprintf("📄 %s\n", filePath))
	}

	// Build info line with lines and size
	var infoParts []string
	if totalLines > 0 {
		infoParts = append(infoParts, fmt.Sprintf("%d lines", totalLines))
	}
	if fileSize > 0 {
		infoParts = append(infoParts, formatFileSize(fileSize))
	}

	if len(infoParts) > 0 {
		output.WriteString(fmt.Sprintf("   %s", strings.Join(infoParts, " | ")))
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

// formatFileSize formats bytes into human-readable size
func formatFileSize(bytes int) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)

	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
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

// parseDisplayMessage extracts and displays the pre-formatted message content
func (trp *ToolResultParser) parseDisplayMessage(result map[string]any) string {
	// Extract the "message" field which contains the pre-formatted output
	if message, ok := result["message"].(string); ok {
		return message
	}
	// Fallback to generic parsing if message field is missing
	return trp.parseGeneric(result)
}

// parseUpdateTaskList extracts and displays the pre-formatted task list
func (trp *ToolResultParser) parseUpdateTaskList(result map[string]any) string {
	// Extract the "message" field which contains the pre-formatted task list
	if message, ok := result["message"].(string); ok {
		return message
	}
	// Fallback to generic parsing if message field is missing
	return trp.parseGeneric(result)
}
