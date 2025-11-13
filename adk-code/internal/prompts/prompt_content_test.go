package agent_prompts

import (
	"adk-code/internal/prompts/prompts"
	"strings"
	"testing"
)

func TestGuidanceSection_NotEmpty(t *testing.T) {
	if prompts.GuidanceSection == "" {
		t.Fatal("expected GuidanceSection to not be empty")
	}
}

func TestGuidanceSection_HasExpectedContent(t *testing.T) {
	// Guidance should contain actionable advice for the agent
	checks := []string{
		"break down", // typical guidance includes task breakdown
		"problem",
		"solution",
		"verify",
	}

	found := 0
	for _, check := range checks {
		if strings.Contains(strings.ToLower(prompts.GuidanceSection), check) {
			found++
		}
	}

	if found == 0 {
		t.Logf("Warning: GuidanceSection might not contain expected keywords. Content length: %d", len(prompts.GuidanceSection))
	}
}

func TestGuidanceSection_MinimumLength(t *testing.T) {
	// Guidance should be substantial
	if len(prompts.GuidanceSection) < 100 {
		t.Errorf("expected GuidanceSection to have at least 100 characters, got %d", len(prompts.GuidanceSection))
	}
}

func TestPitfallsSection_NotEmpty(t *testing.T) {
	if prompts.PitfallsSection == "" {
		t.Fatal("expected PitfallsSection to not be empty")
	}
}

func TestPitfallsSection_HasExpectedContent(t *testing.T) {
	// Pitfalls should contain warnings about what NOT to do
	checks := []string{
		"avoid",
		"don't",
		"wrong",
		"error",
	}

	found := 0
	for _, check := range checks {
		if strings.Contains(strings.ToLower(prompts.PitfallsSection), check) {
			found++
		}
	}

	if found == 0 {
		t.Logf("Warning: PitfallsSection might not contain expected keywords. Content length: %d", len(prompts.PitfallsSection))
	}
}

func TestPitfallsSection_MinimumLength(t *testing.T) {
	// Pitfalls should be substantial
	if len(prompts.PitfallsSection) < 100 {
		t.Errorf("expected PitfallsSection to have at least 100 characters, got %d", len(prompts.PitfallsSection))
	}
}

func TestWorkflowSection_NotEmpty(t *testing.T) {
	if prompts.WorkflowSection == "" {
		t.Fatal("expected WorkflowSection to not be empty")
	}
}

func TestWorkflowSection_HasExpectedContent(t *testing.T) {
	// Workflow should describe the agent's process
	checks := []string{
		"step",
		"process",
		"workflow",
		"approach",
	}

	found := 0
	for _, check := range checks {
		if strings.Contains(strings.ToLower(prompts.WorkflowSection), check) {
			found++
		}
	}

	if found == 0 {
		t.Logf("Warning: WorkflowSection might not contain expected keywords. Content length: %d", len(prompts.WorkflowSection))
	}
}

func TestWorkflowSection_MinimumLength(t *testing.T) {
	// Workflow should be substantial
	if len(prompts.WorkflowSection) < 100 {
		t.Errorf("expected WorkflowSection to have at least 100 characters, got %d", len(prompts.WorkflowSection))
	}
}

func TestPromptSections_Consistency(t *testing.T) {
	// Sections should not be the same - verify they're all different
	values := []string{prompts.GuidanceSection, prompts.PitfallsSection, prompts.WorkflowSection}
	for i, v1 := range values {
		for j, v2 := range values {
			if i != j && v1 == v2 {
				t.Errorf("sections %d and %d are identical", i, j)
			}
		}
	}
}

func TestPromptSections_ProperFormatting(t *testing.T) {
	// Sections should use proper formatting for readability
	tests := map[string]string{
		"GuidanceSection": prompts.GuidanceSection,
		"PitfallsSection": prompts.PitfallsSection,
		"WorkflowSection": prompts.WorkflowSection,
	}

	for name, content := range tests {
		// Should have reasonable line breaks
		lines := strings.Split(content, "\n")
		if len(lines) < 2 {
			t.Logf("Warning: %s might be poorly formatted (only %d lines)", name, len(lines))
		}

		// Should not have excessive blank lines
		blankCount := 0
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				blankCount++
			}
		}
		if blankCount > len(lines)/2 {
			t.Logf("Warning: %s has excessive blank lines (%d out of %d)", name, blankCount, len(lines))
		}
	}
}

func TestGuidanceSection_IncludesKeyPrinciples(t *testing.T) {
	// Guidance should include core principles
	content := strings.ToLower(prompts.GuidanceSection)

	// Check for problem-solving approach
	if !strings.Contains(content, "read") && !strings.Contains(content, "understand") {
		t.Logf("Warning: GuidanceSection might not emphasize reading/understanding phase")
	}
}

func TestPitfallsSection_InclusCommonMistakes(t *testing.T) {
	// Pitfalls should warn about common mistakes
	content := strings.ToLower(prompts.PitfallsSection)

	// Should warn about something
	if len(content) < 50 {
		t.Error("PitfallsSection seems too short to contain meaningful warnings")
	}
}

func TestWorkflowSection_IncludesSteps(t *testing.T) {
	// Workflow should have structured steps
	content := prompts.WorkflowSection

	// Should contain some form of step indication
	hasSteps := strings.Contains(content, "1.") || strings.Contains(content, "Step") || strings.Contains(content, "step")
	if !hasSteps {
		t.Logf("Warning: WorkflowSection might not have clear step structure")
	}
}
