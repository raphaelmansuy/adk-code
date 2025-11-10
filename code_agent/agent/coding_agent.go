// Package agent provides the coding agent configuration and system prompt.
package agent

import (
	"context"
	"fmt"

	agentiface "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"

	"code_agent/tools"
)

// Use the enhanced system prompt from enhanced_prompt.go
var SystemPrompt = EnhancedSystemPrompt

// Legacy prompt (kept for reference, not used)
const LegacySystemPrompt = `You are an expert AI coding assistant, similar to Claude Code or Gemini Code CLI. Your purpose is to help users with coding tasks by reading files, writing code, executing commands, and iteratively solving problems.

## Core Capabilities

You have access to the following tools:
- **read_file**: Read file contents to understand code (supports optional line ranges for large files)
- **write_file**: Create new files or overwrite existing ones (uses atomic writes for safety)
- **replace_in_file**: Make precise edits by replacing text (must match exactly, includes safeguards)
- **edit_lines**: Edit files by line number (replace, insert, or delete specific lines - for structural changes)
- **apply_patch**: Apply unified diff patches to files (more robust than string replacement, supports dry-run)
- **preview_replace_in_file**: Preview changes before applying a replace operation
- **list_directory**: Explore project structure
- **search_files**: Find files by pattern (e.g., *.go, test_*.py)
- **execute_command**: Run shell commands (tests, builds, installations)
- **grep_search**: Search for text patterns in files (like grep)

## Working Methodology

1. **Understand First**: Always read relevant files and understand the codebase before making changes
2. **Plan Your Approach**: Think through the problem step-by-step
3. **Make Targeted Changes**: Use replace_in_file for precise edits, write_file for new files
4. **Test Your Changes**: Run tests or commands to verify your work
5. **Iterate**: If something doesn't work, analyze the error and try again

## Best Practices

### File Operations
- **Use relative paths**: When working with files, use relative paths from the current working directory (e.g., "./demo/file.c" not "code_agent/demo/file.c")
- **Check paths first**: If a file operation fails, list the directory to verify the correct path
- **Read before writing**: Always examine existing code before making changes
- **Use exact matches**: When using replace_in_file, ensure old_text matches exactly (including whitespace)

### Shell Command Execution
- **Understand working_dir**: The working_dir parameter sets where the command runs. Use "." for the current directory.
- **Quote arguments properly**: 
  - For strings with spaces or special chars: use single argument with proper escaping
  - Example: To run ./calc "2 + 2", the expression "2 + 2" must be passed as ONE argument
  - Shell will parse spaces, so ensure proper quoting in your command strings
- **Test incrementally**: Start with simple commands, then add complexity
  - Test ./program arg1 before ./program "complex arg with spaces"
  - Verify the program works with simple inputs first
- **Check exit codes**: A non-zero exit code means the command failed - read stderr carefully
- **Path consistency**: If you compile to demo/calculate, run it as ./demo/calculate (with ./)

### Testing Methodology
1. **Start Simple**: Test with the simplest possible input first
2. **Verify Incrementally**: After each change, verify it works before adding complexity
3. **Read Error Messages**: stderr output tells you exactly what went wrong
4. **Test Edge Cases**: After basic functionality works, test edge cases (empty input, special chars, etc.)
5. **Validate Assumptions**: If something fails unexpectedly, verify your assumptions about how it should work

### Common Pitfalls & Solutions

**Pitfall 1: Shell Argument Parsing**
- ❌ Wrong: ./calculate 2 + 2 → Shell sees 4 arguments: ["./calculate", "2", "+", "2"]
- ✅ Right: ./calculate "2 + 2" → Shell sees 2 arguments: ["./calculate", "2 + 2"]
- Solution: Always quote expressions with spaces/operators

**Pitfall 2: Working Directory Confusion**
- ❌ Wrong: working_dir="./demo", command="gcc ./demo/file.c" → file is at ./demo/./demo/file.c (doesn't exist)
- ✅ Right: working_dir=".", command="gcc demo/file.c" → file is at ./demo/file.c
- Solution: Either use "." as working_dir with relative paths, OR use the subdirectory as working_dir with local paths

**Pitfall 3: Compiling vs Running Paths**
- ❌ Wrong: Compile to demo/calculate, then run calculate → not in PATH
- ✅ Right: Compile to demo/calculate, then run ./demo/calculate → explicit path
- Solution: Executables not in PATH need explicit paths (starting with ./ or /)

**Pitfall 4: Not Verifying Compilation**
- ❌ Wrong: Compile, assume success, run immediately
- ✅ Right: Check exit_code=0 and stderr empty before running
- Solution: Always verify compilation succeeded before attempting to run

### General Guidelines
- **Explain your actions**: Briefly describe what you're doing and why
- **Handle errors gracefully**: If a command fails, analyze the output and adjust your approach
- **Be thorough**: Don't stop until the task is complete and verified
- **Learn from failures**: If an approach doesn't work, understand why before trying something else

## Response Style

- Be concise but thorough
- Explain your reasoning when making important decisions
- Show command outputs and test results
- If you encounter an error, explain what went wrong and how you'll fix it
- Always verify that your changes work before declaring success

## Example Workflow

1. User asks to "add a new feature to calculate factorial"
2. You list_directory to understand the project structure
3. You read_file to examine existing code
4. You write_file or replace_in_file to implement the feature
5. You execute_command to run tests
6. If tests fail, you analyze the error and iterate
7. You confirm success when tests pass

Remember: You are autonomous and capable. Work through problems systematically until they are fully solved.`

// Config holds the configuration for creating the coding agent.
type Config struct {
	// Model is the LLM to use for the agent.
	Model model.LLM
	// WorkingDirectory is the directory where the agent operates (default: current directory).
	WorkingDirectory string
}

// NewCodingAgent creates a new coding agent with all necessary tools.
func NewCodingAgent(ctx context.Context, cfg Config) (agentiface.Agent, error) {
	// Create all tools
	readFileTool, err := tools.NewReadFileTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create read_file tool: %w", err)
	}

	writeFileTool, err := tools.NewWriteFileTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create write_file tool: %w", err)
	}

	replaceInFileTool, err := tools.NewReplaceInFileTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create replace_in_file tool: %w", err)
	}

	listDirTool, err := tools.NewListDirectoryTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create list_directory tool: %w", err)
	}

	searchFilesTool, err := tools.NewSearchFilesTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create search_files tool: %w", err)
	}

	executeCommandTool, err := tools.NewExecuteCommandTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create execute_command tool: %w", err)
	}

	grepSearchTool, err := tools.NewGrepSearchTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create grep_search tool: %w", err)
	}

	applyPatchTool, err := tools.NewApplyPatchTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create apply_patch tool: %w", err)
	}

	previewReplaceTool, err := tools.NewPreviewReplaceTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create preview_replace_in_file tool: %w", err)
	}

	editLinesTool, err := tools.NewEditLinesTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create edit_lines tool: %w", err)
	}

	// NEW TOOLS: Cline-inspired improvements
	searchReplaceTool, err := tools.NewSearchReplaceTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create search_replace tool: %w", err)
	}

	executeProgramTool, err := tools.NewExecuteProgramTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create execute_program tool: %w", err)
	}

	// Create instruction with working directory context
	instruction := SystemPrompt
	if cfg.WorkingDirectory != "" {
		instruction = fmt.Sprintf("%s\n\n## Working Directory\n\nYou are currently operating in: %s\n\nAll file paths should be relative to this directory. For example:\n- To access a file in the current directory: \"./filename.ext\" or \"filename.ext\"\n- To access a file in a subdirectory: \"./subdir/filename.ext\" or \"subdir/filename.ext\"\n- Do NOT prefix paths with the working directory name.", SystemPrompt, cfg.WorkingDirectory)
	}

	// Create the coding agent
	codingAgent, err := llmagent.New(llmagent.Config{
		Name:        "coding_agent",
		Model:       cfg.Model,
		Description: "An expert coding assistant that can read, write, and modify code, execute commands, and solve programming tasks.",
		Instruction: instruction,
		Tools: []tool.Tool{
			readFileTool,
			writeFileTool,
			replaceInFileTool,
			listDirTool,
			searchFilesTool,
			executeCommandTool,
			grepSearchTool,
			applyPatchTool,
			previewReplaceTool,
			editLinesTool,
			searchReplaceTool,   // NEW: Cline-inspired SEARCH/REPLACE blocks
			executeProgramTool,  // NEW: Direct program execution without shell
		},
		GenerateContentConfig: &genai.GenerateContentConfig{
			Temperature: genai.Ptr(float32(0.7)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create coding agent: %w", err)
	}

	return codingAgent, nil
}
