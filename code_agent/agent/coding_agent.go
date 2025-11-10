// Package agent provides the coding agent configuration and system prompt.
package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	agentiface "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"

	"code_agent/tools"
	"code_agent/workspace"
)

// Use the enhanced system prompt from enhanced_prompt.go
var SystemPrompt = EnhancedSystemPrompt

// Config holds the configuration for creating the coding agent.
type Config struct {
	// Model is the LLM to use for the agent.
	Model model.LLM
	// WorkingDirectory is the directory where the agent operates (default: current directory).
	WorkingDirectory string
	// EnableMultiWorkspace enables multi-workspace support (feature flag)
	EnableMultiWorkspace bool
}

// GetProjectRoot traverses upwards from the given path to find the project root,
// identified by the presence of a "go.mod" file.
func GetProjectRoot(startPath string) (string, error) {
	currentPath := startPath
	for {
		goModPath := fmt.Sprintf("%s/go.mod", currentPath)
		if _, err := os.Stat(goModPath); err == nil {
			return currentPath, nil
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			return "", fmt.Errorf("go.mod not found in %s or any parent directories", startPath)
		}
		currentPath = parentPath
	}
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

	applyV4APatchTool, err := tools.NewApplyV4APatchTool(cfg.WorkingDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create apply_v4a_patch tool: %w", err)
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

	// Determine the project root based on go.mod file
	projectRoot := cfg.WorkingDirectory
	if projectRoot == "" {
		projectRoot, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	actualProjectRoot, err := GetProjectRoot(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to determine project root: %w", err)
	}

	// Create workspace manager with smart initialization
	// This will:
	// 1. Try loading from .workspace.json config file
	// 2. Auto-detect multiple workspaces if no config exists
	// 3. Fall back to single-directory mode if detection fails
	var wsManager *workspace.Manager
	if cfg.EnableMultiWorkspace {
		// Use smart initialization for multi-workspace support
		wsManager, err = workspace.SmartWorkspaceInitialization(actualProjectRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize workspace manager: %w", err)
		}
	} else {
		// Use single-directory mode (backward compatible)
		wsManager, err = workspace.FromSingleDirectory(actualProjectRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to create workspace manager: %w", err)
		}
	}

	// Build environment context for LLM
	envContext, err := wsManager.BuildEnvironmentContext()
	if err != nil {
		// Don't fail if we can't build context, just log and continue
		envContext = ""
	}

	// Create enhanced instruction with workspace context
	instruction := SystemPrompt
	workspaceSummary := wsManager.GetSummary()

	instruction = fmt.Sprintf(`%s

## Workspace Environment

%s

Primary workspace: %s

`, SystemPrompt, workspaceSummary, actualProjectRoot)

	// Add environment context if available
	if envContext != "" {
		instruction += fmt.Sprintf(`### Workspace Metadata

%s

`, envContext)
	}

	instruction += `### Path Usage

All file paths should be relative to the primary workspace directory. For example:
- To access a file in the current directory: "./filename.ext" or "filename.ext"
- To access a file in a subdirectory: "./subdir/filename.ext" or "subdir/filename.ext"
- Do NOT prefix paths with the working directory name.

### Workspace Hints (Future Feature)

In multi-workspace mode, you can use @workspace:path syntax to explicitly target a workspace:
- @frontend:src/index.ts - targets the frontend workspace
- @backend:api/server.go - targets the backend workspace
`

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
			applyV4APatchTool, // NEW: V4A semantic patch format
			previewReplaceTool,
			editLinesTool,
			searchReplaceTool,  // NEW: Cline-inspired SEARCH/REPLACE blocks
			executeProgramTool, // NEW: Direct program execution without shell
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
