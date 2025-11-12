// Package display - Event handling and display logic
package display

import (
	"fmt"
	"path/filepath"
	"strings"

	"google.golang.org/adk/session"

	"code_agent/tracking"
)

// PrintEventEnhanced processes and displays agent events
func PrintEventEnhanced(renderer *Renderer, streamDisplay *StreamingDisplay,
	event *session.Event, spinner *Spinner, activeToolName *string, toolRunning *bool,
	sessionTokens *tracking.SessionTokens, requestID string, timeline *EventTimeline) {

	if event.Content == nil || len(event.Content.Parts) == 0 {
		return
	}

	// Record token metrics if available and update spinner with metrics
	if event.UsageMetadata != nil {
		sessionTokens.RecordMetrics(event.UsageMetadata, requestID)
		// Create token metrics for spinner display
		metric := &tracking.TokenMetrics{
			PromptTokens:   event.UsageMetadata.PromptTokenCount,
			CachedTokens:   event.UsageMetadata.CachedContentTokenCount,
			ResponseTokens: event.UsageMetadata.CandidatesTokenCount,
			ThoughtTokens:  event.UsageMetadata.ThoughtsTokenCount,
			ToolUseTokens:  event.UsageMetadata.ToolUsePromptTokenCount,
			TotalTokens:    event.UsageMetadata.TotalTokenCount,
		}
		// Update spinner with metrics if it's actively running
		if *toolRunning {
			spinner.UpdateWithMetrics("Processing", metric)
		} else {
			spinner.UpdateWithMetrics("Agent is thinking", metric)
		}
	}

	// Create tool renderer with enhanced features
	toolRenderer := NewToolRenderer(renderer)
	toolResultParser := NewToolResultParser(nil)

	for _, part := range event.Content.Parts {
		// Handle text content
		if part.Text != "" {
			// Only stop spinner for actual agent responses (not tool-related text)
			text := part.Text
			isToolRelated := strings.Contains(text, "read_file") ||
				strings.Contains(text, "write_file") ||
				strings.Contains(text, "execute_command") ||
				strings.Contains(text, "list_directory") ||
				strings.Contains(text, "grep_search") ||
				strings.Contains(text, "search_replace") ||
				strings.Contains(text, "edit_lines") ||
				strings.Contains(text, "apply_patch")

			if !isToolRelated {
				// This is actual agent response text, stop spinner
				spinner.Stop()

				// Detect if this is thinking/reasoning text
				isThinking := strings.Contains(strings.ToLower(text), "thinking") ||
					strings.Contains(strings.ToLower(text), "analyzing") ||
					strings.Contains(strings.ToLower(text), "considering")

				if isThinking {
					// Update spinner message with thinking indicator and set thinking mode
					spinner.SetMode(SpinnerModeThinking)
					spinner.Update(EventTypeIcon(EventTypeThinking) + " Agent is thinking")
					spinner.Start()
				} else {
					// Render the actual text content with result indicator
					prefix := EventTypeIcon(EventTypeResult) + " "
					output := prefix + renderer.RenderPartContent(part)
					fmt.Print(output)
				}
			}
		}

		// Handle function calls - show what tool is being executed
		if part.FunctionCall != nil {
			// First, stop the current spinner to print the tool banner
			spinner.Stop()

			*activeToolName = part.FunctionCall.Name
			*toolRunning = true

			// Track in timeline
			timeline.AppendEvent(part.FunctionCall.Name, "executing")

			args := make(map[string]any)
			for k, v := range part.FunctionCall.Args {
				args[k] = v
			}

			// Show what tool is being executed
			output := toolRenderer.RenderToolExecution(part.FunctionCall.Name, args)
			fmt.Print(output)

			// Start spinner with context-aware message for the tool execution
			spinnerMessage := GetToolSpinnerMessage(part.FunctionCall.Name, args)
			spinner.Update(spinnerMessage)
			spinner.Start()
		}

		// Handle function responses - show the result
		if part.FunctionResponse != nil {
			// Stop spinner now that tool is complete
			spinner.Stop()
			*toolRunning = false

			// Update timeline status to completed
			timeline.UpdateLastEventStatus("completed")

			result := make(map[string]any)
			if part.FunctionResponse.Response != nil {
				for k, v := range part.FunctionResponse.Response {
					result[k] = v
				}
			}

			// Show success indicator for tool completion
			successIcon := EventTypeIcon(EventTypeSuccess)
			fmt.Printf("\n%s Tool completed: %s\n", successIcon, part.FunctionResponse.Name)

			// Use enhanced result parser for structured output
			parsedResult := toolResultParser.ParseToolResult(part.FunctionResponse.Name, result)
			if parsedResult != "" {
				// Show parsed result
				fmt.Print(parsedResult)
				fmt.Print("\n")
			}

			// Show basic result indicator (compact version)
			resultOutput := renderer.RenderToolResult(part.FunctionResponse.Name, result)
			fmt.Print(resultOutput)

			// Restart spinner for next operation (agent might still be working)
			// Update message and restart
			spinner.Update("Processing")
			spinner.Start()
		}
	}
}

// GetToolSpinnerMessage returns a context-aware spinner message for tool execution
func GetToolSpinnerMessage(toolName string, args map[string]any) string {
	icon := EventTypeIcon(EventTypeExecuting)

	switch toolName {
	case "read_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("%s Reading %s", icon, filepath.Base(path))
		}
		return fmt.Sprintf("%s Reading file", icon)
	case "write_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("%s Writing %s", icon, filepath.Base(path))
		}
		return fmt.Sprintf("%s Writing file", icon)
	case "search_replace", "replace_in_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("%s Editing %s", icon, filepath.Base(path))
		}
		return fmt.Sprintf("%s Editing file", icon)
	case "edit_lines":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("%s Modifying %s", icon, filepath.Base(path))
		}
		return fmt.Sprintf("%s Modifying file", icon)
	case "apply_patch", "apply_v4a_patch":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("%s Applying patch to %s", icon, filepath.Base(path))
		}
		return fmt.Sprintf("%s Applying patch", icon)
	case "list_directory", "list_files":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("%s Listing %s", icon, filepath.Base(path))
		}
		return fmt.Sprintf("%s Listing directory", icon)
	case "search_files":
		if pattern, ok := args["pattern"].(string); ok {
			return fmt.Sprintf("%s Searching for %s", icon, pattern)
		}
		return fmt.Sprintf("%s Searching files", icon)
	case "grep_search":
		if pattern, ok := args["pattern"].(string); ok {
			return fmt.Sprintf("%s Searching for '%s'", icon, pattern)
		}
		return fmt.Sprintf("%s Searching code", icon)
	case "execute_command":
		if command, ok := args["command"].(string); ok {
			// Truncate long commands
			if len(command) > 40 {
				command = command[:37] + "..."
			}
			return fmt.Sprintf("%s Running: %s", icon, command)
		}
		return fmt.Sprintf("%s Running command", icon)
	case "execute_program":
		if program, ok := args["program"].(string); ok {
			return fmt.Sprintf("%s Executing %s", icon, filepath.Base(program))
		}
		return fmt.Sprintf("%s Executing program", icon)
	case "display_message":
		if messageType, ok := args["message_type"].(string); ok {
			return fmt.Sprintf("%s Displaying %s message", icon, messageType)
		}
		return fmt.Sprintf("%s Displaying message", icon)
	case "update_task_list":
		return fmt.Sprintf("%s Updating task list", icon)
	default:
		return fmt.Sprintf("%s Running %s", icon, toolName)
	}
}
