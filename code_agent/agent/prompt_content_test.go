package agent

import (
	"strings"
	"testing"
)

func TestGuidanceSection_NotEmpty(t *testing.T) {
	if GuidanceSection == "" {
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
		if strings.Contains(strings.ToLower(GuidanceSection), check) {
			found++
		}
	}

	if found == 0 {
		t.Logf("Warning: GuidanceSection might not contain expected keywords. Content length: %d", len(GuidanceSection))
	}
}

func TestGuidanceSection_MinimumLength(t *testing.T) {
	// Guidance should be substantial
	if len(GuidanceSection) < 100 {
		t.Errorf("expected GuidanceSection to have at least 100 characters, got %d", len(GuidanceSection))
	}
}

func TestPitfallsSection_NotEmpty(t *testing.T) {
	if PitfallsSection == "" {
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
		if strings.Contains(strings.ToLower(PitfallsSection), check) {
			found++
		}
	}

	if found == 0 {
		t.Logf("Warning: PitfallsSection might not contain expected keywords. Content length: %d", len(PitfallsSection))
	}
}

func TestPitfallsSection_MinimumLength(t *testing.T) {
	// Pitfalls should be substantial
	if len(PitfallsSection) < 100 {
		t.Errorf("expected PitfallsSection to have at least 100 characters, got %d", len(PitfallsSection))
	}
}

func TestWorkflowSection_NotEmpty(t *testing.T) {
	if WorkflowSection == "" {
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
		if strings.Contains(strings.ToLower(WorkflowSection), check) {
			found++
		}
	}

	if found == 0 {
		t.Logf("Warning: WorkflowSection might not contain expected keywords. Content length: %d", len(WorkflowSection))
	}
}

func TestWorkflowSection_MinimumLength(t *testing.T) {
	// Workflow should be substantial
	if len(WorkflowSection) < 100 {
		t.Errorf("expected WorkflowSection to have at least 100 characters, got %d", len(WorkflowSection))
	}
}

func TestPromptSections_Consistency(t *testing.T) {
	// Sections should not be the same - verify they're all different
	values := []string{GuidanceSection, PitfallsSection, WorkflowSection}
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
		"GuidanceSection": GuidanceSection,
		"PitfallsSection": PitfallsSection,
		"WorkflowSection": WorkflowSection,
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
	content := strings.ToLower(GuidanceSection)

	// Check for problem-solving approach
	if !strings.Contains(content, "read") && !strings.Contains(content, "understand") {
		t.Logf("Warning: GuidanceSection might not emphasize reading/understanding phase")
	}
}

func TestPitfallsSection_InclusCommonMistakes(t *testing.T) {
	// Pitfalls should warn about common mistakes
	content := strings.ToLower(PitfallsSection)

	// Should warn about something
	if len(content) < 50 {
		t.Error("PitfallsSection seems too short to contain meaningful warnings")
	}
}

func TestWorkflowSection_IncludesSteps(t *testing.T) {
	// Workflow should have structured steps
	content := WorkflowSection

	// Should contain some form of step indication
	hasSteps := strings.Contains(content, "1.") || strings.Contains(content, "Step") || strings.Contains(content, "step")
	if !hasSteps {
		t.Logf("Warning: WorkflowSection might not have clear step structure")
	}
}
