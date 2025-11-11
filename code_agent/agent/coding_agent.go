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
	"google.golang.org/genai"

	"code_agent/tools"
	"code_agent/workspace"
)

// Config holds the configuration for creating the coding agent.
type Config struct {
	// Model is the LLM to use for the agent.
	Model model.LLM
	// WorkingDirectory is the directory where the agent operates (default: current directory).
	WorkingDirectory string
	// EnableMultiWorkspace enables multi-workspace support (feature flag)
	EnableMultiWorkspace bool
}

// GetProjectRoot traverses to find the project root,
// identified by the presence of a "go.mod" file.
// It searches: current path, immediate subdirectories, and parent directories.
func GetProjectRoot(startPath string) (string, error) {
	// First, check if go.mod exists in the start path
	if _, err := os.Stat(filepath.Join(startPath, "go.mod")); err == nil {
		return startPath, nil
	}

	// Check if go.mod exists in immediate subdirectories (e.g., code_agent/)
	entries, err := os.ReadDir(startPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subdir := filepath.Join(startPath, entry.Name())
				if _, err := os.Stat(filepath.Join(subdir, "go.mod")); err == nil {
					return subdir, nil
				}
			}
		}
	}

	// Then traverse upwards to find go.mod in parent directories
	currentPath := startPath
	for {
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached the root of the filesystem
			return "", fmt.Errorf("go.mod not found in %s, its subdirectories, or any parent directories", startPath)
		}
		currentPath = parentPath

		goModPath := filepath.Join(currentPath, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentPath, nil
		}
	}
}

// NewCodingAgent creates a new coding agent with all necessary tools.
func NewCodingAgent(ctx context.Context, cfg Config) (agentiface.Agent, error) {
	// Initialize all tools (they register themselves in the global registry)
	// We still need to call these to trigger registration via init side effects
	if _, err := tools.NewReadFileTool(); err != nil {
		return nil, fmt.Errorf("failed to create read_file tool: %w", err)
	}
	if _, err := tools.NewWriteFileTool(); err != nil {
		return nil, fmt.Errorf("failed to create write_file tool: %w", err)
	}
	if _, err := tools.NewReplaceInFileTool(); err != nil {
		return nil, fmt.Errorf("failed to create replace_in_file tool: %w", err)
	}
	if _, err := tools.NewListDirectoryTool(); err != nil {
		return nil, fmt.Errorf("failed to create list_directory tool: %w", err)
	}
	if _, err := tools.NewSearchFilesTool(); err != nil {
		return nil, fmt.Errorf("failed to create search_files tool: %w", err)
	}
	if _, err := tools.NewExecuteCommandTool(); err != nil {
		return nil, fmt.Errorf("failed to create execute_command tool: %w", err)
	}
	if _, err := tools.NewGrepSearchTool(); err != nil {
		return nil, fmt.Errorf("failed to create grep_search tool: %w", err)
	}
	if _, err := tools.NewApplyPatchTool(); err != nil {
		return nil, fmt.Errorf("failed to create apply_patch tool: %w", err)
	}
	if _, err := tools.NewApplyV4APatchTool(cfg.WorkingDirectory); err != nil {
		return nil, fmt.Errorf("failed to create apply_v4a_patch tool: %w", err)
	}
	if _, err := tools.NewPreviewReplaceTool(); err != nil {
		return nil, fmt.Errorf("failed to create preview_replace_in_file tool: %w", err)
	}
	if _, err := tools.NewEditLinesTool(); err != nil {
		return nil, fmt.Errorf("failed to create edit_lines tool: %w", err)
	}
	if _, err := tools.NewSearchReplaceTool(); err != nil {
		return nil, fmt.Errorf("failed to create search_replace tool: %w", err)
	}
	if _, err := tools.NewExecuteProgramTool(); err != nil {
		return nil, fmt.Errorf("failed to create execute_program tool: %w", err)
	}

	// Get all registered tools from the registry
	registry := tools.GetRegistry()
	registeredTools := registry.GetAllTools()

	// Determine the project root based on go.mod file
	projectRoot := cfg.WorkingDirectory
	if projectRoot == "" {
		var err error
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

	// Build dynamic XML-tagged system prompt from registered tools
	promptCtx := PromptContext{
		HasWorkspace:         true,
		WorkspaceRoot:        actualProjectRoot,
		WorkspaceSummary:     wsManager.GetSummary(),
		EnvironmentMetadata:  envContext,
		EnableMultiWorkspace: cfg.EnableMultiWorkspace,
	}

	instruction := BuildEnhancedPromptWithContext(registry, promptCtx)

	// Create the coding agent with dynamically registered tools
	codingAgent, err := llmagent.New(llmagent.Config{
		Name:        "coding_agent",
		Model:       cfg.Model,
		Description: "An expert coding assistant that can read, write, and modify code, execute commands, and solve programming tasks.",
		Instruction: instruction,
		Tools:       registeredTools, // Use tools from registry
		GenerateContentConfig: &genai.GenerateContentConfig{
			Temperature: genai.Ptr(float32(0.7)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create coding agent: %w", err)
	}

	return codingAgent, nil
}
