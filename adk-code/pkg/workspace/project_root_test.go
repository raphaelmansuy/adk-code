package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetProjectRoot_FindsGoModInCurrentPath(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a go.mod file in the temp directory
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatalf("failed to create test go.mod: %v", err)
	}

	root, err := GetProjectRoot(tmpDir)
	if err != nil {
		t.Fatalf("GetProjectRoot failed: %v", err)
	}
	if root != tmpDir {
		t.Fatalf("expected root %s, got %s", tmpDir, root)
	}
}

func TestGetProjectRoot_FindsGoModInSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a subdirectory with go.mod
	subdir := filepath.Join(tmpDir, "code_agent")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	goModPath := filepath.Join(subdir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatalf("failed to create test go.mod: %v", err)
	}

	root, err := GetProjectRoot(tmpDir)
	if err != nil {
		t.Fatalf("GetProjectRoot failed: %v", err)
	}
	if root != subdir {
		t.Fatalf("expected root %s, got %s", subdir, root)
	}
}

func TestGetProjectRoot_FindsGoModInParentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Create go.mod in the temp directory
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatalf("failed to create test go.mod: %v", err)
	}

	// Create a subdirectory and test from there
	subdir := filepath.Join(tmpDir, "subdir", "nested")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create nested directories: %v", err)
	}

	root, err := GetProjectRoot(subdir)
	if err != nil {
		t.Fatalf("GetProjectRoot failed: %v", err)
	}
	if root != tmpDir {
		t.Fatalf("expected root %s, got %s", tmpDir, root)
	}
}

func TestGetProjectRoot_NoGoModReturnsFallback(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create any go.mod file
	root, err := GetProjectRoot(tmpDir)
	if err != nil {
		t.Fatalf("GetProjectRoot should not fail when go.mod not found, got error: %v", err)
	}
	if root != tmpDir {
		t.Fatalf("expected fallback to start path %s, got %s", tmpDir, root)
	}
}

// TestGetProjectRoot_NonGoProjectUsage tests the scenario where adk-code is used
// as a global CLI tool in non-Go projects (e.g., Python, JavaScript, or any directory)
func TestGetProjectRoot_NonGoProjectUsage(t *testing.T) {
	// Create a directory structure mimicking a non-Go project
	tmpDir := t.TempDir()

	// Create some files that don't include go.mod
	pythonFile := filepath.Join(tmpDir, "main.py")
	if err := os.WriteFile(pythonFile, []byte("print('hello')"), 0644); err != nil {
		t.Fatalf("failed to create python file: %v", err)
	}

	packageJSON := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	// GetProjectRoot should return tmpDir without error
	root, err := GetProjectRoot(tmpDir)
	if err != nil {
		t.Fatalf("GetProjectRoot should work in non-Go projects, got error: %v", err)
	}

	if root != tmpDir {
		t.Fatalf("expected root %s for non-Go project, got %s", tmpDir, root)
	}
}
