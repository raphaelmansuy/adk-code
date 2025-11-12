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

func TestGetProjectRoot_NoGoModReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create any go.mod file
	_, err := GetProjectRoot(tmpDir)
	if err == nil {
		t.Fatalf("expected error when go.mod not found, got nil")
	}
	if !os.IsNotExist(err) && err.Error() == "" {
		t.Fatalf("expected proper error message, got: %v", err)
	}
}
