// Package tools provides a comprehensive collection of agent tools for file operations,
// code editing, execution, and workspace management.
//
// This package serves as the public interface for tool functionality, re-exporting
// key types and constructors from internal subpackages. Tools are categorized by
// their purpose and are registered with the tool registry for use by the agent.
//
// Tool categories:
// - file: File read/write/list operations
// - edit: Code editing tools (patches, search/replace, line edits)
// - search: Search and discovery tools
// - exec: Command and program execution
// - display: Message display and task list updates
// - workspace: Workspace management and analysis
// - v4a: V4A patch format tools
// - base: Shared registry and error types
//
// Each tool implements the Tool interface and provides:
// - Input/Output types with JSON schema
// - A handler function
// - Metadata (name, description, category)
// - Error handling
//
// Tools are registered with the global tool registry during package initialization.
// The registry can be used to discover available tools, get tools by name/ID, and
// invoke tools with type-safe input/output.
//
// Example:
//
//	// Get a tool from the registry
//	tool, err := registry.GetTool("read_file")
//	if err != nil {
//		return err
//	}
//
//	// Invoke the tool
//	output, err := tool.Execute(ctx, input)
//	if err != nil {
//		return err
//	}
//
// For tool development, refer to the individual subpackages:
// - tools/file: File operation tools
// - tools/edit: Code editing tools
// - tools/exec: Execution tools
// - tools/base: Registry and shared types
package tools
