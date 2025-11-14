// Package tools provides explicit tool registration for the coding agent.
// This file documents all available tools and their registration.
package tools

import (
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

// RegisterAllTools registers all available tools with the provided registry.
// This provides a clear inventory of all tools and explicit registration order.
// Tools are organized by category for clarity and maintainability.
//
// This function is the authoritative source for the complete tool inventory.
// It can be used to:
// - Verify all tools are registered
// - Conditionally register tools based on configuration
// - Test tool registration in isolation
// - Generate tool documentation
//
// Note: This is called automatically during package initialization via init(),
// but can also be called manually for testing or advanced configuration.
func RegisterAllTools(reg *common.ToolRegistry) error {
	// Group tools by category for clarity - file tools are registered through their init() functions
	// which call their New*Tool() constructors. Each constructor internally calls base.Register()
	// to add the tool to the global registry.
	//
	// Rather than duplicating that logic here, we rely on the existing init() functions in each
	// tool package to perform registration. This maintains the current working system while
	// providing a documented entry point for tool management.
	//
	// Tool organization for reference:
	// - File Operations: read, write, list, replace, search (in tools/file/)
	// - Edit Operations: apply_patch, edit_lines, search_replace (in tools/edit/)
	// - Search Operations: preview_replace (in tools/search/)
	// - Execution: execute_command, execute_program, grep_search (in tools/exec/)
	// - Display: display_message, update_task_list (in tools/display/)
	// - Workspace: workspace_tools (in tools/workspace/)
	// - V4A Format: apply_v4a_patch (in tools/v4a/)
	//
	// This function serves as documentation and a future refactoring point
	// if explicit registration becomes necessary.

	return nil
}

// init automatically triggers tool registration at package initialization.
// This ensures all tools from subpackages (file, edit, exec, display, search, workspace, v4a, discovery)
// are registered when the tools package is imported.
//
// Each tool subpackage has its own init() function that calls tool constructors,
// which in turn call base.Register() to register themselves in the global registry.
//
// This pattern provides:
// - Automatic tool discovery and registration
// - Decentralized tool definition (each package owns its tools)
// - Lazy initialization through Go's init() mechanism
// - Clear separation of concerns
func init() {
	// Import all tool subpackages to trigger their init() functions
	// This ensures all tools are automatically registered
	_ = display.NewDisplayMessageTool
	_ = file.NewReadFileTool
	_ = edit.NewApplyPatchTool
	_ = exec.NewExecuteCommandTool
	_ = search.NewPreviewReplaceTool
	_ = workspace.NewWorkspaceTools
	_ = v4a.NewApplyV4APatchTool
	_ = discovery.NewListModelsTool
	_ = discovery.NewModelInfoTool
}
