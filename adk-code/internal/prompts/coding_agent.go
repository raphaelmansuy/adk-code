// Package agent provides the coding agent configuration and system prompt.
package agent_prompts

import (
	"context"
	"fmt"
	"os"

	agentiface "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"

	pkgerrors "adk-code/pkg/errors"
	"adk-code/pkg/workspace"
	"adk-code/tools"
)

// Config holds the configuration for creating the coding agent.
type Config struct {
	// Model is the LLM to use for the agent.
	Model model.LLM
	// WorkingDirectory is the directory where the agent operates (default: current directory).
	WorkingDirectory string
	// EnableMultiWorkspace enables multi-workspace support (feature flag)
	EnableMultiWorkspace bool
	// EnableThinking enables the model's thinking/reasoning output (default: true)
	EnableThinking bool
	// ThinkingBudget sets the token budget for thinking (only used if EnableThinking is true)
	ThinkingBudget int32
	// MCPToolsets are external MCP server toolsets to be added to the agent
	MCPToolsets []tool.Toolset
}

// GetProjectRoot traverses to find the project root,
// identified by the presence of a "go.mod" file.
// It searches: current path, immediate subdirectories, and parent directories.
// Deprecated: Use workspace.GetProjectRoot instead.
func GetProjectRoot(startPath string) (string, error) {
	return workspace.GetProjectRoot(startPath)
}

// NewCodingAgent creates a new coding agent with all necessary tools.
func NewCodingAgent(ctx context.Context, cfg Config) (agentiface.Agent, error) {
	// Validate required configuration
	if cfg.Model == nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInvalidInput, "model is required", nil)
	}

	// Most tools auto-register via init() functions in their packages.
	// V4A patch tool requires working directory parameter, so we register it explicitly.
	if _, err := tools.NewApplyV4APatchTool(cfg.WorkingDirectory); err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create apply_v4a_patch tool", err)
	}

	// Determine the project root - use the working directory directly
	// This allows adk-code to work as a global CLI tool in any directory
	var err error
	projectRoot := cfg.WorkingDirectory
	if projectRoot == "" {
		projectRoot, err = os.Getwd()
		if err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to get current working directory", err)
		}
	}

	// Load subagent tools using ADK's agent-as-tool pattern
	// This discovers agent definitions and converts them to tools
	// Pass MCP toolsets so subagents can access MCP tools if specified
	subagentTools, subagentErr := tools.LoadSubAgentToolsWithMCP(ctx, projectRoot, cfg.Model, cfg.MCPToolsets)
	if subagentErr != nil {
		// Don't fail if subagents can't be loaded, just log a warning
		fmt.Fprintf(os.Stderr, "Warning: Failed to load subagent tools: %v\n", subagentErr)
	} else if len(subagentTools) > 0 {
		fmt.Fprintf(os.Stderr, "✓ Loaded %d subagent(s) as tools (with MCP support)\n", len(subagentTools))
	}

	// Get all registered tools from the registry (includes subagent tools)
	registry := tools.GetRegistry()
	registeredTools := registry.GetAllTools()

	// Use the working directory directly as the project root
	// No need to search for go.mod - adk-code works in any project type
	actualProjectRoot := projectRoot

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
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to initialize workspace manager", err)
		}
	} else {
		// Use single-directory mode (backward compatible)
		wsManager, err = workspace.FromSingleDirectory(actualProjectRoot)
		if err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create workspace manager", err)
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
		HasMCPTools:          len(cfg.MCPToolsets) > 0, // Indicate if MCP tools are available
	}

	instruction := BuildEnhancedPromptWithContext(registry, promptCtx)

	// Build GenerateContentConfig with optional thinking support
	generateConfig := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.7)),
	}

	// Add thinking config if enabled
	if cfg.EnableThinking {
		generateConfig.ThinkingConfig = &genai.ThinkingConfig{
			IncludeThoughts: true,
			ThinkingBudget:  genai.Ptr(cfg.ThinkingBudget),
		}
	}

	// Create the coding agent with dynamically registered tools and MCP toolsets
	codingAgent, err := llmagent.New(llmagent.Config{
		Name:                  "coding_agent",
		Model:                 cfg.Model,
		Description:           "An expert coding assistant that can read, write, and modify code, execute commands, and solve programming tasks.",
		Instruction:           instruction,
		Tools:                 registeredTools, // Use tools from registry
		Toolsets:              cfg.MCPToolsets, // Add MCP toolsets
		GenerateContentConfig: generateConfig,
	})
	if err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create coding agent", err)
	}

	return codingAgent, nil
}
