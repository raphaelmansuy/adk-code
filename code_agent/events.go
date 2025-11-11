// Package main - Event handling and display logic
package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"google.golang.org/adk/session"

	"code_agent/display"
	"code_agent/tracking"
)

// printEventEnhanced processes and displays agent events
func printEventEnhanced(renderer *display.Renderer, streamDisplay *display.StreamingDisplay,
	event *session.Event, spinner *display.Spinner, activeToolName *string, toolRunning *bool,
	sessionTokens *tracking.SessionTokens, requestID string) {

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
	toolRenderer := display.NewToolRenderer(renderer)
	toolResultParser := display.NewToolResultParser(nil)

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
					// Update spinner message instead of stopping
					spinner.Update("Analyzing your request")
				} else {
					// Render the actual text content
					output := renderer.RenderPartContent(part)
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

			args := make(map[string]any)
			for k, v := range part.FunctionCall.Args {
				args[k] = v
			}

			// Show what tool is being executed
			output := toolRenderer.RenderToolExecution(part.FunctionCall.Name, args)
			fmt.Print(output)

			// Start spinner with context-aware message for the tool execution
			spinnerMessage := getToolSpinnerMessage(part.FunctionCall.Name, args)
			spinner.Update(spinnerMessage)
			spinner.Start()
		}

		// Handle function responses - show the result
		if part.FunctionResponse != nil {
			// Stop spinner now that tool is complete
			spinner.Stop()
			*toolRunning = false

			result := make(map[string]any)
			if part.FunctionResponse.Response != nil {
				for k, v := range part.FunctionResponse.Response {
					result[k] = v
				}
			}

			// Use enhanced result parser for structured output
			parsedResult := toolResultParser.ParseToolResult(part.FunctionResponse.Name, result)
			if parsedResult != "" {
				// Show parsed result
				fmt.Print("\n")
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

// getToolSpinnerMessage returns a context-aware spinner message for tool execution
func getToolSpinnerMessage(toolName string, args map[string]any) string {
	switch toolName {
	case "read_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Reading %s", filepath.Base(path))
		}
		return "Reading file"
	case "write_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Writing %s", filepath.Base(path))
		}
		return "Writing file"
	case "search_replace", "replace_in_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Editing %s", filepath.Base(path))
		}
		return "Editing file"
	case "edit_lines":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Modifying %s", filepath.Base(path))
		}
		return "Modifying file"
	case "apply_patch", "apply_v4a_patch":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Applying patch to %s", filepath.Base(path))
		}
		return "Applying patch"
	case "list_directory", "list_files":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Listing %s", filepath.Base(path))
		}
		return "Listing directory"
	case "search_files":
		if pattern, ok := args["pattern"].(string); ok {
			return fmt.Sprintf("Searching for %s", pattern)
		}
		return "Searching files"
	case "grep_search":
		if pattern, ok := args["pattern"].(string); ok {
			return fmt.Sprintf("Searching for '%s'", pattern)
		}
		return "Searching code"
	case "execute_command":
		if command, ok := args["command"].(string); ok {
			// Truncate long commands
			if len(command) > 40 {
				command = command[:37] + "..."
			}
			return fmt.Sprintf("Running: %s", command)
		}
		return "Running command"
	case "execute_program":
		if program, ok := args["program"].(string); ok {
			return fmt.Sprintf("Executing %s", filepath.Base(program))
		}
		return "Executing program"
	default:
		return fmt.Sprintf("Running %s", toolName)
	}
}
