// Package v4a provides V4A patch parsing and application tools for the coding agent.
package v4a

import (
	"fmt"
	"path/filepath"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"code_agent/tools/base"
)

// ApplyV4APatchInput defines the input for the apply_v4a_patch tool.
type ApplyV4APatchInput struct {
	// Path is the relative path to the file to patch
	Path string `json:"path" jsonschema:"Relative path to the file to patch"`

	// Patch is the V4A format patch content
	Patch string `json:"patch" jsonschema:"V4A format patch content with @@ context markers"`

	// DryRun if true returns a preview without modifying the file
	DryRun *bool `json:"dry_run,omitempty" jsonschema:"If true returns a preview without applying changes (default: false)"`
}

// ApplyV4APatchOutput defines the output of the apply_v4a_patch tool.
type ApplyV4APatchOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NewApplyV4APatchTool creates the apply_v4a_patch tool.
//
// This tool applies V4A format patches that use semantic context markers
// (class/function names) instead of line numbers. V4A is more resilient to
// code changes and provides better readability for refactoring tasks.
func NewApplyV4APatchTool(workingDir string) (tool.Tool, error) {
	handler := func(ctx tool.Context, input ApplyV4APatchInput) ApplyV4APatchOutput {
		// Resolve path - use absolute path if provided, otherwise join with workingDir
		fullPath := input.Path
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(workingDir, fullPath)
		}

		// Default dry_run to false
		dryRun := false
		if input.DryRun != nil {
			dryRun = *input.DryRun
		}

		// Parse the V4A patch
		patch, err := ParseV4APatch(input.Patch)
		if err != nil {
			return ApplyV4APatchOutput{
				Success: false,
				Error:   fmt.Sprintf("Failed to parse V4A patch: %v", err),
			}
		}

		// If patch doesn't specify file path, use the input path
		if patch.FilePath == "" {
			patch.FilePath = fullPath
		} else {
			// Patch specifies path, resolve it relative to working dir
			if !filepath.IsAbs(patch.FilePath) {
				patch.FilePath = filepath.Join(workingDir, patch.FilePath)
			}
		}

		// Apply the patch
		result, err := ApplyV4APatch(patch.FilePath, patch, dryRun)
		if err != nil {
			return ApplyV4APatchOutput{
				Success: false,
				Error:   fmt.Sprintf("Failed to apply V4A patch: %v", err),
			}
		}

		return ApplyV4APatchOutput{
			Success: true,
			Message: result,
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name: "apply_v4a_patch",
		Description: `Apply a V4A format patch to a file.

V4A is a semantic patch format that uses context markers (class/function names) instead of line numbers.
This makes patches more resilient to code changes and easier to understand.

V4A Format:
*** Update File: <filepath>          (optional)
@@ <context1>                         (e.g., class ClassName)
@@     <context2>                     (e.g., def method_name, indented for nesting)
-<line_to_remove>                     (lines to remove)
+<line_to_add>                        (lines to add)

Example (Python):
*** Update File: src/models/user.py
@@ class User
@@     def validate():
-          return True
+          if not self.email:
+              raise ValueError("Email required")
+          return True

Example (Go):
@@ func HandleRequest
-    return nil
+    return processRequest(req)

Use apply_v4a_patch when:
- Refactoring within functions/classes (semantic context is stable)
- The file is frequently modified (line numbers change often)
- Better readability is needed (class/function names are clear)

Use apply_patch (unified diff) when:
- Patching multiple files at once
- Need exact line number control
- External collaboration (standard format)

Always use dry_run=true first to preview changes.`,
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategoryCodeEditing,
			Priority:  4,
			UsageHint: "Semantic patches using context markers (class/function names), more resilient than line-based",
		})
	}

	return t, err
}
