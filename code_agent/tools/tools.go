// Package tools provides a comprehensive collection of file, editing, execution,
// and workspace management tools for the coding agent. This package re-exports
// all tool constructors and types from subpackages for backward compatibility.
//
// Subpackages:
//   - file: File read/write/list operations
//   - edit: Code editing tools (patches, search/replace, line edits)
//   - search: Search and discovery tools
//   - exec: Command and program execution
//   - workspace: Workspace management
//   - v4a: V4A patch format tools
//   - common: Shared types and registry
package tools

// Re-export all key functions from subpackages
import (
	"code_agent/tools/common"
	"code_agent/tools/edit"
	"code_agent/tools/exec"
	"code_agent/tools/file"
	"code_agent/tools/search"
	"code_agent/tools/v4a"
	"code_agent/tools/workspace"
)

// File tools
var (
	NewReadFileTool      = file.NewReadFileTool
	NewWriteFileTool     = file.NewWriteFileTool
	NewReplaceInFileTool = file.NewReplaceInFileTool
	NewListDirectoryTool = file.NewListDirectoryTool
	NewSearchFilesTool   = file.NewSearchFilesTool
)

// Search tools
var (
	NewPreviewReplaceTool = search.NewPreviewReplaceTool
)

// Edit tools
var (
	NewApplyPatchTool    = edit.NewApplyPatchTool
	NewEditLinesTool     = edit.NewEditLinesTool
	NewSearchReplaceTool = edit.NewSearchReplaceTool
)

// Execution tools
var (
	NewExecuteCommandTool = exec.NewExecuteCommandTool
	NewExecuteProgramTool = exec.NewExecuteProgramTool
	NewGrepSearchTool     = exec.NewGrepSearchTool
)

// Workspace tools
var (
	NewWorkspaceTools = workspace.NewWorkspaceTools
)

// V4A tools
var (
	NewApplyV4APatchTool = v4a.NewApplyV4APatchTool
)

// Re-export types from subpackages
type (
	// Common types
	ErrorCode    = common.ErrorCode
	ToolError    = common.ToolError
	ToolMetadata = common.ToolMetadata
	ToolCategory = common.ToolCategory
	ToolRegistry = common.ToolRegistry

	// File types
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

	// Search types
	PreviewReplaceInput  = search.PreviewReplaceInput
	PreviewReplaceOutput = search.PreviewReplaceOutput

	// Edit types
	ApplyPatchInput     = edit.ApplyPatchInput
	ApplyPatchOutput    = edit.ApplyPatchOutput
	EditLinesInput      = edit.EditLinesInput
	EditLinesOutput     = edit.EditLinesOutput
	SearchReplaceInput  = edit.SearchReplaceInput
	SearchReplaceOutput = edit.SearchReplaceOutput

	// Execution types
	ExecuteCommandInput  = exec.ExecuteCommandInput
	ExecuteCommandOutput = exec.ExecuteCommandOutput
	ExecuteProgramInput  = exec.ExecuteProgramInput
	ExecuteProgramOutput = exec.ExecuteProgramOutput
	GrepSearchInput      = exec.GrepSearchInput
	GrepSearchOutput     = exec.GrepSearchOutput
	GrepMatch            = exec.GrepMatch

	// Workspace types
	WorkspaceTools = workspace.WorkspaceTools

	// V4A types
	ApplyV4APatchInput  = v4a.ApplyV4APatchInput
	ApplyV4APatchOutput = v4a.ApplyV4APatchOutput
	V4APatch            = v4a.V4APatch
	V4AHunk             = v4a.V4AHunk
)

// Re-export constants from common package
const (
	CategoryFileOperations  = common.CategoryFileOperations
	CategorySearchDiscovery = common.CategorySearchDiscovery
	CategoryCodeEditing     = common.CategoryCodeEditing
	CategoryExecution       = common.CategoryExecution
	CategoryWorkspace       = common.CategoryWorkspace
)

// Registry functions
var (
	GetRegistry = common.GetRegistry
	Register    = common.Register
)
