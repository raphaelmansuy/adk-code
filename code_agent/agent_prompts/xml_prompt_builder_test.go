// Tests for XML prompt builder
package agent_prompts

import (
	"strings"
	"testing"

	pkgerrors "code_agent/pkg/errors"
	"code_agent/tools"
)

func TestPromptBuilderBasic(t *testing.T) {
	registry := tools.GetRegistry()
	builder := NewPromptBuilder(registry)

	ctx := PromptContext{
		HasWorkspace: false,
	}

	prompt := builder.BuildXMLPrompt(ctx)

	// Verify basic structure
	if !strings.Contains(prompt, "<agent_system_prompt>") {
		t.Error("Expected <agent_system_prompt> tag")
	}
	if !strings.Contains(prompt, "</agent_system_prompt>") {
		t.Error("Expected </agent_system_prompt> closing tag")
	}
	if !strings.Contains(prompt, "<agent_identity>") {
		t.Error("Expected <agent_identity> tag")
	}
	if !strings.Contains(prompt, "<tools>") {
		t.Error("Expected <tools> tag")
	}
	if !strings.Contains(prompt, "<guidance>") {
		t.Error("Expected <guidance> tag")
	}
	if !strings.Contains(prompt, "<critical_rules") {
		t.Error("Expected <critical_rules> tag")
	}
}

func TestPromptBuilderWithWorkspace(t *testing.T) {
	registry := tools.GetRegistry()
	builder := NewPromptBuilder(registry)

	ctx := PromptContext{
		HasWorkspace:     true,
		WorkspaceRoot:    "/path/to/workspace",
		WorkspaceSummary: "Test workspace summary",
	}

	prompt := builder.BuildXMLPrompt(ctx)

	// Verify workspace context is included
	if !strings.Contains(prompt, "<workspace_context>") {
		t.Error("Expected <workspace_context> tag when HasWorkspace is true")
	}
	if !strings.Contains(prompt, "</workspace_context>") {
		t.Error("Expected </workspace_context> closing tag")
	}
	if !strings.Contains(prompt, "<file_system>") {
		t.Error("Expected <file_system> tag")
	}
	if !strings.Contains(prompt, ctx.WorkspaceRoot) {
		t.Error("Expected workspace root in prompt")
	}
	if !strings.Contains(prompt, "<path_usage>") {
		t.Error("Expected <path_usage> tag")
	}
}

func TestPromptBuilderWithEnvironment(t *testing.T) {
	registry := tools.GetRegistry()
	builder := NewPromptBuilder(registry)

	ctx := PromptContext{
		HasWorkspace:        true,
		WorkspaceRoot:       "/path/to/workspace",
		WorkspaceSummary:    "Test workspace",
		EnvironmentMetadata: "Git branch: main\nCommit: abc123",
	}

	prompt := builder.BuildXMLPrompt(ctx)

	// Verify environment section is included
	if !strings.Contains(prompt, "<environment>") {
		t.Error("Expected <environment> tag when metadata is provided")
	}
	if !strings.Contains(prompt, "Git branch: main") {
		t.Error("Expected environment metadata in prompt")
	}
}

func TestPromptBuilderMultiWorkspace(t *testing.T) {
	registry := tools.GetRegistry()
	builder := NewPromptBuilder(registry)

	ctx := PromptContext{
		HasWorkspace:         true,
		WorkspaceRoot:        "/path/to/workspace",
		WorkspaceSummary:     "Multi-workspace",
		EnableMultiWorkspace: true,
	}

	prompt := builder.BuildXMLPrompt(ctx)

	// Verify multi-workspace instructions
	if !strings.Contains(prompt, "@workspace:path") {
		t.Error("Expected multi-workspace syntax instructions")
	}
	if !strings.Contains(prompt, "@frontend:") {
		t.Error("Expected multi-workspace example")
	}
}

func TestValidatePromptStructure(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		wantErr bool
	}{
		{
			name: "valid simple structure",
			prompt: `<agent>
<tools>
</tools>
</agent>`,
			wantErr: false,
		},
		{
			name: "valid nested structure",
			prompt: `<agent>
<tools>
  <tool name="test">
  </tool>
</tools>
</agent>`,
			wantErr: false,
		},
		{
			name: "unclosed tag",
			prompt: `<agent>
<tools>
</agent>`,
			wantErr: true,
		},
		{
			name: "mismatched tags",
			prompt: `<agent>
<tools>
</tool>
</agent>`,
			wantErr: true,
		},
		{
			name: "unexpected closing tag",
			prompt: `</agent>
<tools>
</tools>`,
			wantErr: true,
		},
		{
			name:    "valid empty",
			prompt:  "",
			wantErr: false,
		},
		{
			name: "mixed content with tags",
			prompt: `<agent>
Some text content
<tools>
More content
</tools>
Final content
</agent>`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePromptStructure(tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePromptStructure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePromptStructure_CDATAUnclosed(t *testing.T) {
	invalid := "<agent_system_prompt><![CDATA[Unclosed content"
	err := ValidatePromptStructure(invalid)
	if err == nil {
		t.Fatalf("expected error for unclosed CDATA section")
	}
	if !pkgerrors.Is(err, pkgerrors.CodeInvalidInput) {
		t.Fatalf("expected invalid input code; got: %v", err)
	}
}

func TestBuildEnhancedPromptV2BackwardCompatibility(t *testing.T) {
	registry := tools.GetRegistry()

	// Test backward compatibility function
	prompt := BuildEnhancedPromptV2(registry)

	// Should produce XML-tagged output
	if !strings.Contains(prompt, "<agent_system_prompt>") {
		t.Error("BuildEnhancedPromptV2 should produce XML-tagged output")
	}

	// Should validate correctly
	err := ValidatePromptStructure(prompt)
	if err != nil {
		t.Errorf("BuildEnhancedPromptV2 produced invalid XML structure: %v", err)
	}
}

func TestBuildEnhancedPromptWithContext(t *testing.T) {
	registry := tools.GetRegistry()

	ctx := PromptContext{
		HasWorkspace:     true,
		WorkspaceRoot:    "/test/workspace",
		WorkspaceSummary: "Test summary",
	}

	prompt := BuildEnhancedPromptWithContext(registry, ctx)

	// Should include workspace context
	if !strings.Contains(prompt, "<workspace_context>") {
		t.Error("Expected workspace context in prompt")
	}

	// Should validate correctly
	err := ValidatePromptStructure(prompt)
	if err != nil {
		t.Errorf("BuildEnhancedPromptWithContext produced invalid XML structure: %v", err)
	}
}

func TestRenderToolsXML(t *testing.T) {
	registry := tools.GetRegistry()

	// Register a test tool if the registry is empty
	// (In real code, tools auto-register via init())

	builder := NewPromptBuilder(registry)
	toolsXML := builder.renderToolsXML()

	// Should contain tool_category tags if tools are registered
	// If registry is empty, the output will be empty (which is valid)
	// Just verify it doesn't crash
	if registry.Count() > 0 && !strings.Contains(toolsXML, "<tool_category") {
		t.Error("Expected <tool_category> tags in tools XML when tools are registered")
	}
}

func TestPromptValidationWithRealOutput(t *testing.T) {
	registry := tools.GetRegistry()
	builder := NewPromptBuilder(registry)

	ctx := PromptContext{
		HasWorkspace:         true,
		WorkspaceRoot:        "/test",
		WorkspaceSummary:     "Test",
		EnvironmentMetadata:  "Test env",
		EnableMultiWorkspace: true,
	}

	prompt := builder.BuildXMLPrompt(ctx)

	// The generated prompt should pass validation
	err := ValidatePromptStructure(prompt)
	if err != nil {
		t.Errorf("Generated prompt failed validation: %v", err)
		t.Logf("Prompt preview:\n%s", prompt[:min(500, len(prompt))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
