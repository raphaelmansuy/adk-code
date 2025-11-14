// Package discovery - integration tests
package discovery

import (
	"testing"

	common "adk-code/tools/base"
)

// TestDiscoveryToolsRegistered verifies that discovery tools are properly registered
func TestDiscoveryToolsRegistered(t *testing.T) {
	registry := common.GetRegistry()

	if registry == nil {
		t.Fatal("Global tool registry is nil")
	}

	// Get all tools
	allTools := registry.GetAllTools()

	// Check if we have any tools
	if len(allTools) == 0 {
		t.Fatal("No tools registered in global registry")
	}

	// Look for our discovery tools
	toolNames := make(map[string]bool)
	for _, tool := range allTools {
		toolNames[tool.Name()] = true
	}

	expectedTools := []string{
		"list_models",
		"model_info",
	}

	for _, expectedTool := range expectedTools {
		if !toolNames[expectedTool] {
			t.Errorf("Expected tool %s not found in registry", expectedTool)
		}
	}
}

// TestListModelsToolCreation tests that the list_models tool can be created
func TestListModelsToolCreation(t *testing.T) {
	tool, err := NewListModelsTool()
	if err != nil {
		t.Fatalf("Failed to create list_models tool: %v", err)
	}

	if tool == nil {
		t.Fatal("list_models tool is nil")
	}

	// Verify tool name
	if tool.Name() != "list_models" {
		t.Errorf("Expected tool name 'list_models', got %s", tool.Name())
	}

	// Verify tool description is not empty
	if tool.Description() == "" {
		t.Error("Tool description is empty")
	}
}

// TestModelInfoToolCreation tests that the model_info tool can be created
func TestModelInfoToolCreation(t *testing.T) {
	tool, err := NewModelInfoTool()
	if err != nil {
		t.Fatalf("Failed to create model_info tool: %v", err)
	}

	if tool == nil {
		t.Fatal("model_info tool is nil")
	}

	// Verify tool name
	if tool.Name() != "model_info" {
		t.Errorf("Expected tool name 'model_info', got %s", tool.Name())
	}

	// Verify tool description is not empty
	if tool.Description() == "" {
		t.Error("Tool description is empty")
	}
}

// TestDiscoveryToolsInSearchCategory verifies tools are in the right category
func TestDiscoveryToolsInSearchCategory(t *testing.T) {
	registry := common.GetRegistry()

	searchTools := registry.GetByCategory(common.CategorySearchDiscovery)

	toolNames := make(map[string]bool)
	for _, metadata := range searchTools {
		toolNames[metadata.Tool.Name()] = true
	}

	expectedTools := []string{
		"list_models",
		"model_info",
	}

	for _, expectedTool := range expectedTools {
		if !toolNames[expectedTool] {
			t.Errorf("Expected tool %s not found in Search & Discovery category", expectedTool)
		}
	}
}
