// Package tools provides terminal execution tools for the coding agent.
package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// ExecuteCommandInput defines the input parameters for executing a command.
type ExecuteCommandInput struct {
	// Command is the command to execute.
	Command string `json:"command" jsonschema:"Command to execute (e.g., 'ls -la', 'go test ./...')"`
	// WorkingDir is the working directory for the command (optional, defaults to current directory).
	WorkingDir string `json:"working_dir,omitempty" jsonschema:"Working directory for the command (optional)"`
	// Timeout is the maximum time in seconds to wait for the command (default: 30).
	Timeout *int `json:"timeout,omitempty" jsonschema:"Maximum time in seconds to wait for the command (default: 30)"`
}

// ExecuteCommandOutput defines the output of executing a command.
type ExecuteCommandOutput struct {
	// Stdout is the standard output of the command.
	Stdout string `json:"stdout"`
	// Stderr is the standard error output of the command.
	Stderr string `json:"stderr"`
	// ExitCode is the exit code of the command.
	ExitCode int `json:"exit_code"`
	// Success indicates whether the command executed successfully (exit code 0).
	Success bool `json:"success"`
	// Error contains error message if the operation failed.
	Error string `json:"error,omitempty"`
}

// NewExecuteCommandTool creates a tool for executing shell commands.
func NewExecuteCommandTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ExecuteCommandInput) ExecuteCommandOutput {
		timeoutSecs := 30 // default
		if input.Timeout != nil {
			timeoutSecs = *input.Timeout
		}
		timeout := time.Duration(timeoutSecs) * time.Second

		cmdCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Parse command into parts
		parts := strings.Fields(input.Command)
		if len(parts) == 0 {
			return ExecuteCommandOutput{
				Success: false,
				Error:   "Command is empty",
			}
		}

		cmd := exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
		if input.WorkingDir != "" {
			cmd.Dir = input.WorkingDir
		}

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				return ExecuteCommandOutput{
					Success: false,
					Error:   fmt.Sprintf("Failed to execute command: %v", err),
				}
			}
		}

		return ExecuteCommandOutput{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: exitCode,
			Success:  exitCode == 0,
		}
	}

	return functiontool.New(functiontool.Config{
		Name:        "execute_command",
		Description: "Executes a shell command and returns its output. Use this to run tests, build code, install dependencies, or run any command-line tools. The command runs in a shell environment with a timeout.",
	}, handler)
}

// GrepSearchInput defines the input parameters for searching text in files.
type GrepSearchInput struct {
	// Path is the directory or file to search in.
	Path string `json:"path" jsonschema:"Directory or file to search in"`
	// Pattern is the text pattern to search for.
	Pattern string `json:"pattern" jsonschema:"Text pattern to search for"`
	// CaseSensitive indicates whether the search should be case-sensitive.
	CaseSensitive *bool `json:"case_sensitive,omitempty" jsonschema:"Whether the search should be case-sensitive (default: false)"`
	// FilePattern is an optional file pattern to limit the search (e.g., '*.go').
	FilePattern string `json:"file_pattern,omitempty" jsonschema:"Optional file pattern to limit the search (e.g., '*.go')"`
}

// GrepMatch represents a single match in a file.
type GrepMatch struct {
	// File is the path to the file containing the match.
	File string `json:"file"`
	// Line is the line number (1-indexed).
	Line int `json:"line"`
	// Content is the content of the matching line.
	Content string `json:"content"`
}

// GrepSearchOutput defines the output of a grep search.
type GrepSearchOutput struct {
	// Matches is the list of matches found.
	Matches []GrepMatch `json:"matches"`
	// Count is the total number of matches found.
	Count int `json:"count"`
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`
	// Error contains error message if the operation failed.
	Error string `json:"error,omitempty"`
}

// NewGrepSearchTool creates a tool for searching text in files (similar to grep).
func NewGrepSearchTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input GrepSearchInput) GrepSearchOutput {
		// Build grep command
		args := []string{"-n", "-r"} // line numbers, recursive

		caseSensitive := false
		if input.CaseSensitive != nil {
			caseSensitive = *input.CaseSensitive
		}
		if !caseSensitive {
			args = append(args, "-i") // case insensitive
		}
		if input.FilePattern != "" {
			args = append(args, "--include="+input.FilePattern)
		}
		args = append(args, input.Pattern, input.Path)

		cmd := exec.CommandContext(ctx, "grep", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		// grep returns exit code 1 if no matches found, which is not an error
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				// No matches found, which is fine
				return GrepSearchOutput{
					Matches: []GrepMatch{},
					Count:   0,
					Success: true,
				}
			}
			return GrepSearchOutput{
				Matches: make([]GrepMatch, 0),
				Count:   0,
				Success: false,
				Error:   fmt.Sprintf("Grep failed: %v - %s", err, stderr.String()),
			}
		}

		// Parse grep output
		lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
		matches := make([]GrepMatch, 0, len(lines))

		for _, line := range lines {
			if line == "" {
				continue
			}

			// Parse format: filename:linenumber:content
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 3 {
				lineNum := 0
				fmt.Sscanf(parts[1], "%d", &lineNum)
				matches = append(matches, GrepMatch{
					File:    parts[0],
					Line:    lineNum,
					Content: parts[2],
				})
			}
		}

		return GrepSearchOutput{
			Matches: matches,
			Count:   len(matches),
			Success: true,
		}
	}

	return functiontool.New(functiontool.Config{
		Name:        "grep_search",
		Description: "Searches for text patterns in files (like grep). Returns matching lines with file paths and line numbers. Useful for finding specific code patterns, function definitions, or error messages.",
	}, handler)
}
