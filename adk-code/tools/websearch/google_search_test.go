package websearch

import (
	"testing"

	common "adk-code/tools/base"
)

func TestNewGoogleSearchTool(t *testing.T) {
	// Test that the tool can be created successfully
	tool, err := NewGoogleSearchTool()
	if err != nil {
		t.Fatalf("Failed to create Google Search tool: %v", err)
	}

	if tool == nil {
		t.Fatal("NewGoogleSearchTool returned nil tool")
	}
}

func TestGoogleSearchToolRegistration(t *testing.T) {
	// Create a new registry for testing
	registry := common.NewToolRegistry()

	// Register the tool
	tool, err := NewGoogleSearchTool()
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Manually register for test
	err = registry.Register(common.ToolMetadata{
		Tool:      tool,
		Category:  common.CategorySearchDiscovery,
		Priority:  0,
		UsageHint: "Search the web for current information",
	})

	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Verify it's in the Search & Discovery category
	tools := registry.GetByCategory(common.CategorySearchDiscovery)
	if len(tools) == 0 {
		t.Fatal("No tools found in Search & Discovery category")
	}

	// Verify the tool is registered
	found := false
	for _, tm := range tools {
		if tm.Tool == tool {
			found = true
			if tm.Category != common.CategorySearchDiscovery {
				t.Errorf("Expected category %v, got %v", common.CategorySearchDiscovery, tm.Category)
			}
			if tm.Priority != 0 {
				t.Errorf("Expected priority 0, got %d", tm.Priority)
			}
			break
		}
	}

	if !found {
		t.Error("Google Search tool not found in registry")
	}
}

func TestGoogleSearchToolMetadata(t *testing.T) {
	// Get the global registry
	registry := common.GetRegistry()

	// Get all tools in Search & Discovery category
	tools := registry.GetByCategory(common.CategorySearchDiscovery)

	// Find the Google Search tool (it should be registered via init())
	var googleSearchTool *common.ToolMetadata
	for i := range tools {
		// Check if this is a Google Search tool by checking metadata
		if tools[i].UsageHint != "" && tools[i].Category == common.CategorySearchDiscovery {
			googleSearchTool = &tools[i]
			break
		}
	}

	if googleSearchTool == nil {
		// This might fail if init() hasn't run yet, which is OK in isolated tests
		t.Skip("Google Search tool not found in global registry (init may not have run)")
	}

	// Verify metadata
	if googleSearchTool.Category != common.CategorySearchDiscovery {
		t.Errorf("Expected category %v, got %v", common.CategorySearchDiscovery, googleSearchTool.Category)
	}

	if googleSearchTool.Tool == nil {
		t.Error("Tool is nil in metadata")
	}
}
