package instructions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewInstructionLoader(t *testing.T) {
	workdir, _ := os.Getwd()
	loader := NewInstructionLoader(workdir)

	if loader == nil {
		t.Fatal("Expected non-nil loader")
	}

	if loader.workingDir != workdir {
		t.Errorf("Expected workdir %s, got %s", workdir, loader.workingDir)
	}
}

func TestFindProjectRoot(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create a go.mod file to mark project root
	goModPath := filepath.Join(tmpDir, "go.mod")
	os.WriteFile(goModPath, []byte("module test"), 0644)

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	// Find project root from subdirectory
	root := findProjectRoot(subDir)

	if root != tmpDir {
		t.Errorf("Expected project root %s, got %s", tmpDir, root)
	}
}

func TestLoad_NoInstructions(t *testing.T) {
	// Use a temp directory with no AGENTS.md files
	tmpDir := t.TempDir()

	loader := NewInstructionLoader(tmpDir)
	result := loader.Load()

	if result.Global != "" {
		t.Errorf("Expected empty global instructions")
	}

	if result.ProjectRoot != "" {
		t.Errorf("Expected empty project root instructions")
	}

	if len(result.Nested) != 0 {
		t.Errorf("Expected no nested instructions")
	}

	if result.Merged != "" {
		t.Errorf("Expected empty merged instructions")
	}
}

func TestLoad_WithProjectInstructions(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create a go.mod to mark project root
	goModPath := filepath.Join(tmpDir, "go.mod")
	os.WriteFile(goModPath, []byte("module test"), 0644)

	// Create AGENTS.md at project root
	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	content := "Project-level instructions"
	os.WriteFile(agentsPath, []byte(content), 0644)

	loader := NewInstructionLoader(tmpDir)
	result := loader.Load()

	if result.ProjectRoot != content {
		t.Errorf("Expected project root instructions to be %q, got %q", content, result.ProjectRoot)
	}

	// Merged should contain the project root content
	if result.Merged == "" || len(result.Merged) == 0 {
		t.Errorf("Expected merged instructions to be non-empty")
	}
}

func TestMergeInstructions_Truncation(t *testing.T) {
	result := LoadedInstructions{
		MaxBytes: 100, // Very small limit
		Nested:   make(map[string]string),
	}

	// Create content that exceeds limit
	result.Global = "This is a very long instruction that will definitely exceed the 100 byte limit that we have set for testing purposes"

	loader := NewInstructionLoader(".")

	// Load to trigger merge with truncation check
	result.Global = "This is a very long instruction that will definitely exceed the 100 byte limit that we have set for testing purposes"
	result.Merged = loader.mergeInstructions(result)

	// Check if truncation occurred
	if len(result.Merged) > result.MaxBytes+50 { // Allow some buffer for marker
		t.Errorf("Expected content to be truncated, got length %d", len(result.Merged))
	}
}

func TestFindProjectRoot_NoMarkers(t *testing.T) {
	// Use a temp directory with no project markers
	tmpDir := t.TempDir()

	root := findProjectRoot(tmpDir)

	// Should return empty string when no markers found
	if root != "" {
		t.Errorf("Expected empty string for no project root, got %s", root)
	}
}
