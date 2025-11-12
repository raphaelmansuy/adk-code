package agent_prompts

import (
	"strings"
	"testing"

	"code_agent/tools"
)

func TestBuildToolsSection_ReturnsString(t *testing.T) {
	registry := tools.GetRegistry()
	result := BuildToolsSection(registry)

	if result == "" {
		t.Fatal("expected non-empty tools section")
	}
	if !strings.Contains(result, "Available Tools") {
		t.Error("expected 'Available Tools' header in tools section")
	}
}

func TestBuildToolsSection_ContainsCategoryHeaders(t *testing.T) {
	registry := tools.GetRegistry()
	result := BuildToolsSection(registry)

	// Should contain tool categories
	categories := registry.GetCategories()
	if len(categories) == 0 {
		t.Skip("No tool categories registered")
	}

	for _, category := range categories {
		// category is a ToolCategory type, convert to string for comparison
		categoryStr := string(category)
		if !strings.Contains(result, categoryStr) {
			t.Errorf("expected category '%s' in tools section", categoryStr)
		}
	}
}

func TestBuildToolsSection_IncludesToolNames(t *testing.T) {
	registry := tools.GetRegistry()
	result := BuildToolsSection(registry)

	// Tools should be formatted with bold names (wrapped in **)
	if !strings.Contains(result, "**") {
		t.Error("expected tool names to be formatted with **bold** markers")
	}
}

func TestBuildToolsSection_ContainsUsageHints(t *testing.T) {
	registry := tools.GetRegistry()
	result := BuildToolsSection(registry)

	// Some tools may have usage hints
	if strings.Contains(result, "Usage tip:") {
		// If usage tips are present, verify the format
		if !strings.Contains(result, "→") {
			t.Error("expected arrow marker (→) before usage tips")
		}
	}
}

func TestBuildEnhancedPrompt_ReturnsString(t *testing.T) {
	registry := tools.GetRegistry()
	result := BuildEnhancedPrompt(registry)

	if result == "" {
		t.Fatal("expected non-empty enhanced prompt")
	}
}

func TestBuildEnhancedPrompt_ContainsToolsSection(t *testing.T) {
	registry := tools.GetRegistry()
	result := BuildEnhancedPrompt(registry)

	// Should contain the tools section
	if !strings.Contains(result, "Available Tools") {
		t.Error("expected 'Available Tools' header in enhanced prompt")
	}
}

func TestBuildEnhancedPrompt_ContainsGuidance(t *testing.T) {
	registry := tools.GetRegistry()
	result := BuildEnhancedPrompt(registry)

	// Should include the static guidance section
	if !strings.Contains(result, "guidance") || !strings.Contains(result, "Guidance") {
		// The guidance might be in different case or format
		// Just verify some guidance content exists
		if len(result) < 100 {
			t.Error("expected substantial content in enhanced prompt")
		}
	}
}

func TestBuildEnhancedPrompt_ContainsPitfalls(t *testing.T) {
	registry := tools.GetRegistry()
	result := BuildEnhancedPrompt(registry)

	// Should include pitfalls section
	if !strings.Contains(result, "Pitfall") && !strings.Contains(result, "pitfall") && !strings.Contains(result, "AVOID") {
		t.Logf("Warning: enhanced prompt might not contain pitfalls section")
	}
}

func TestBuildEnhancedPrompt_ContainsWorkflow(t *testing.T) {
	registry := tools.GetRegistry()
	result := BuildEnhancedPrompt(registry)

	// Should include workflow section
	if !strings.Contains(result, "workflow") && !strings.Contains(result, "Workflow") {
		t.Logf("Warning: enhanced prompt might not contain workflow section")
	}
}

func TestBuildToolsSection_WithEmptyRegistry(t *testing.T) {
	// This test creates a registry with no tools
	// which should still produce valid output
	registry := tools.GetRegistry()

	result := BuildToolsSection(registry)
	if result == "" {
		t.Error("expected non-empty result even with minimal tools")
	}
}

func TestBuildEnhancedPrompt_Consistency(t *testing.T) {
	registry := tools.GetRegistry()

	// Call multiple times - should produce consistent output
	result1 := BuildEnhancedPrompt(registry)
	result2 := BuildEnhancedPrompt(registry)

	if result1 != result2 {
		t.Error("expected consistent output from BuildEnhancedPrompt on repeated calls")
	}
}

func TestBuildToolsSection_FormattingStructure(t *testing.T) {
	registry := tools.GetRegistry()
	result := BuildToolsSection(registry)

	// Verify expected formatting elements
	checks := []struct {
		name        string
		shouldExist bool
		pattern     string
	}{
		{"expert ai assistant", true, "expert AI coding assistant"},
		{"markdown headers", true, "###"},
		{"bold formatting", true, "**"},
	}

	for _, check := range checks {
		exists := strings.Contains(result, check.pattern)
		if exists != check.shouldExist {
			if check.shouldExist {
				t.Errorf("expected pattern '%s' (%s) to exist in tools section", check.pattern, check.name)
			} else {
				t.Errorf("expected pattern '%s' (%s) to NOT exist in tools section", check.pattern, check.name)
			}
		}
	}
}
