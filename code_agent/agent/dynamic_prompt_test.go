package agent

import (
	"strings"
	"testing"

	"code_agent/tools"
)

func TestDynamicPromptGeneration(t *testing.T) {
	// Initialize tools (they register themselves)
	tools.NewReadFileTool()
	tools.NewWriteFileTool()
	tools.NewSearchReplaceTool()
	tools.NewExecuteCommandTool()
	tools.NewGrepSearchTool()

	// Get registry
	registry := tools.GetRegistry()

	// Verify tools are registered
	if registry.Count() < 5 {
		t.Errorf("Expected at least 5 tools registered, got %d", registry.Count())
	}

	// Generate prompt
	prompt := BuildToolsSection(registry)

	// Verify prompt contains expected sections
	expectedSections := []string{
		"Available Tools",
		"File Operations",
		"Code Editing",
		"Search & Discovery",
		"Execution",
		"read_file",
		"write_file",
		"search_replace",
		"execute_command",
		"grep_search",
	}

	for _, section := range expectedSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("Prompt missing expected section: %s", section)
		}
	}

	// Verify tools are grouped by category
	if !strings.Contains(prompt, "### File Operations") {
		t.Error("Prompt missing File Operations category header")
	}

	if !strings.Contains(prompt, "### Code Editing") {
		t.Error("Prompt missing Code Editing category header")
	}

	// Print prompt for visual inspection (useful during development)
	t.Logf("Generated prompt:\n%s", prompt)
}

func TestToolCategorization(t *testing.T) {
	// Initialize tools
	tools.NewReadFileTool()
	tools.NewWriteFileTool()
	tools.NewSearchReplaceTool()
	tools.NewExecuteCommandTool()

	registry := tools.GetRegistry()

	// Verify categories
	fileOpsTools := registry.GetByCategory(tools.CategoryFileOperations)
	if len(fileOpsTools) < 2 {
		t.Errorf("Expected at least 2 file operation tools, got %d", len(fileOpsTools))
	}

	codeEditTools := registry.GetByCategory(tools.CategoryCodeEditing)
	if len(codeEditTools) < 1 {
		t.Errorf("Expected at least 1 code editing tool, got %d", len(codeEditTools))
	}

	execTools := registry.GetByCategory(tools.CategoryExecution)
	if len(execTools) < 1 {
		t.Errorf("Expected at least 1 execution tool, got %d", len(execTools))
	}

	// Verify tools have proper metadata
	for _, metadata := range fileOpsTools {
		if metadata.Tool == nil {
			t.Error("Tool is nil")
		}
		if metadata.Category != tools.CategoryFileOperations {
			t.Errorf("Wrong category for file ops tool: %v", metadata.Category)
		}
	}
}
