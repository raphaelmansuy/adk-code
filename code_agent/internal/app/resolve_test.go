package app

import (
	"os"
	"path/filepath"
	"testing"

	"code_agent/internal/config"
)

func TestResolveWorkingDirectory_Default(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{WorkingDirectory: ""}
	a := &Application{config: cfg}
	got := a.resolveWorkingDirectory()
	if got != wd {
		t.Fatalf("expected %q got %q", wd, got)
	}
}

func TestResolveWorkingDirectory_TildeExpand(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg := &config.Config{WorkingDirectory: "~/myproj"}
	a := &Application{config: cfg}
	got := a.resolveWorkingDirectory()
	want := filepath.Join(tmpHome, "myproj")
	if got != want {
		t.Fatalf("expected %s got %s", want, got)
	}
}

func TestResolveWorkingDirectory_Absolute(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{WorkingDirectory: tmpDir}
	a := &Application{config: cfg}
	got := a.resolveWorkingDirectory()
	if got != tmpDir {
		t.Fatalf("expected %s got %s", tmpDir, got)
	}
}
