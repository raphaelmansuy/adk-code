// Package tools provides a comprehensive collection of file, editing, execution,
// and workspace management tools for the coding agent.
//
// This package serves as the public interface for tool functionality, re-exporting
// key types and constructors from internal subpackages. Tools are automatically
// registered via init() functions in their respective subpackages.
//
// Subpackages:
//   - common: Shared types and registry for all tools
//   - file: File read/write/list operations
//   - edit: Code editing tools (patches, search/replace, line edits)
//   - search: Search and discovery tools
//   - exec: Command and program execution
//   - display: Message display and task list updates
//   - workspace: Workspace management and analysis
//   - v4a: V4A patch format tools
//   - agents: Agent definition discovery and management tools
package tools

import (
	"adk-code/tools/agents"
	common "adk-code/tools/base"
	"adk-code/tools/discovery"
	"adk-code/tools/display"
	"adk-code/tools/edit"
	"adk-code/tools/exec"
	"adk-code/tools/file"
	"adk-code/tools/search"
	"adk-code/tools/v4a"
	"adk-code/tools/workspace"
)

// Re-export type aliases for all tool input/output types
// This reduces boilerplate for callers while maintaining clear type identity

type (
	// Common types
	ErrorCode    = common.ErrorCode
	ToolError    = common.ToolError
	ToolMetadata = common.ToolMetadata
	ToolCategory = common.ToolCategory
	ToolRegistry = common.ToolRegistry

	// File tool types
	ReadFileInput       = file.ReadFileInput
	ReadFileOutput      = file.ReadFileOutput
	WriteFileInput      = file.WriteFileInput
	WriteFileOutput     = file.WriteFileOutput
	ListDirectoryInput  = file.ListDirectoryInput
	ListDirectoryOutput = file.ListDirectoryOutput
	ReplaceInFileInput  = file.ReplaceInFileInput
	ReplaceInFileOutput = file.ReplaceInFileOutput
	SearchFilesInput    = file.SearchFilesInput
	SearchFilesOutput   = file.SearchFilesOutput
	FileInfo            = file.FileInfo

	// Search tool types
	PreviewReplaceInput  = search.PreviewReplaceInput
	PreviewReplaceOutput = search.PreviewReplaceOutput

	// Edit tool types
	ApplyPatchInput     = edit.ApplyPatchInput
	ApplyPatchOutput    = edit.ApplyPatchOutput
	EditLinesInput      = edit.EditLinesInput
	EditLinesOutput     = edit.EditLinesOutput
	SearchReplaceInput  = edit.SearchReplaceInput
	SearchReplaceOutput = edit.SearchReplaceOutput

	// Execution tool types
	ExecuteCommandInput  = exec.ExecuteCommandInput
	ExecuteCommandOutput = exec.ExecuteCommandOutput
	ExecuteProgramInput  = exec.ExecuteProgramInput
	ExecuteProgramOutput = exec.ExecuteProgramOutput
	GrepSearchInput      = exec.GrepSearchInput
	GrepSearchOutput     = exec.GrepSearchOutput
	GrepMatch            = exec.GrepMatch

	// Workspace tool types
	WorkspaceTools = workspace.WorkspaceTools

	// V4A patch tool types
	ApplyV4APatchInput  = v4a.ApplyV4APatchInput
	ApplyV4APatchOutput = v4a.ApplyV4APatchOutput
	V4APatch            = v4a.V4APatch
	V4AHunk             = v4a.V4AHunk

	// Display tool types
	DisplayMessageInput  = display.DisplayMessageInput
	DisplayMessageOutput = display.DisplayMessageOutput
	UpdateTaskListInput  = display.UpdateTaskListInput
	UpdateTaskListOutput = display.UpdateTaskListOutput

	// Discovery tool types
	ListModelsInput  = discovery.ListModelsInput
	ListModelsOutput = discovery.ListModelsOutput
	ModelInfoInput   = discovery.ModelInfoInput
	ModelInfoOutput  = discovery.ModelInfoOutput
	ModelEntry       = discovery.ModelEntry
	CapabilitiesInfo = discovery.CapabilitiesInfo

	// Agent tool types
	ListAgentsInput  = agents.ListAgentsInput
	ListAgentsOutput = agents.ListAgentsOutput
	AgentEntry       = agents.AgentEntry
)

// Re-export category constants for tool classification
const (
	CategoryFileOperations  = common.CategoryFileOperations
	CategorySearchDiscovery = common.CategorySearchDiscovery
	CategoryCodeEditing     = common.CategoryCodeEditing
	CategoryExecution       = common.CategoryExecution
	CategoryWorkspace       = common.CategoryWorkspace
	CategoryDisplay         = common.CategoryDisplay
)

// Re-export tool constructors for programmatic tool creation
// Most callers should use the registry for tool access
var (
	// File tools
	NewReadFileTool      = file.NewReadFileTool
	NewWriteFileTool     = file.NewWriteFileTool
	NewReplaceInFileTool = file.NewReplaceInFileTool
	NewListDirectoryTool = file.NewListDirectoryTool
	NewSearchFilesTool   = file.NewSearchFilesTool

	// Display tools
	NewDisplayMessageTool = display.NewDisplayMessageTool
	NewUpdateTaskListTool = display.NewUpdateTaskListTool

	// Search tools
	NewPreviewReplaceTool = search.NewPreviewReplaceTool

	// Edit tools
	NewApplyPatchTool    = edit.NewApplyPatchTool
	NewEditLinesTool     = edit.NewEditLinesTool
	NewSearchReplaceTool = edit.NewSearchReplaceTool

	// Execution tools
	NewExecuteCommandTool = exec.NewExecuteCommandTool
	NewExecuteProgramTool = exec.NewExecuteProgramTool
	NewGrepSearchTool     = exec.NewGrepSearchTool

	// Workspace tools
	NewWorkspaceTools = workspace.NewWorkspaceTools

	// V4A tools
	NewApplyV4APatchTool = v4a.NewApplyV4APatchTool

	// Discovery tools
	NewListModelsTool = discovery.NewListModelsTool
	NewModelInfoTool  = discovery.NewModelInfoTool

	// Agent tools
	NewListAgentsTool        = agents.NewListAgentsTool
	LoadSubAgentTools        = agents.InitSubAgentTools
	LoadSubAgentToolsWithMCP = agents.InitSubAgentToolsWithMCP
)

// Re-export registry functions for tool access and registration
var (
	GetRegistry = common.GetRegistry
	Register    = common.Register
)
