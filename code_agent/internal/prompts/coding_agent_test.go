package agent_prompts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Fields(t *testing.T) {
	tests := []struct {
		name                 string
		workingDirectory     string
		enableMultiWorkspace bool
		enableThinking       bool
		thinkingBudget       int32
	}{
		{
			name:                 "basic config",
			workingDirectory:     "/tmp",
			enableMultiWorkspace: false,
			enableThinking:       false,
			thinkingBudget:       0,
		},
		{
			name:                 "with working directory",
			workingDirectory:     "/home/user/project",
			enableMultiWorkspace: false,
			enableThinking:       false,
			thinkingBudget:       0,
		},
		{
			name:                 "with multi-workspace",
			workingDirectory:     "/project",
			enableMultiWorkspace: true,
			enableThinking:       false,
			thinkingBudget:       0,
		},
		{
			name:                 "with thinking enabled",
			workingDirectory:     "/project",
			enableMultiWorkspace: false,
			enableThinking:       true,
			thinkingBudget:       10000,
		},
		{
			name:                 "all features enabled",
			workingDirectory:     "/project",
			enableMultiWorkspace: true,
			enableThinking:       true,
			thinkingBudget:       50000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Model:                nil,
				WorkingDirectory:     tt.workingDirectory,
				EnableMultiWorkspace: tt.enableMultiWorkspace,
				EnableThinking:       tt.enableThinking,
				ThinkingBudget:       tt.thinkingBudget,
			}

			if cfg.Model != nil {
				t.Errorf("expected Model=nil, got %v", cfg.Model)
			}
			if cfg.WorkingDirectory != tt.workingDirectory {
				t.Errorf("expected WorkingDirectory=%s, got %s", tt.workingDirectory, cfg.WorkingDirectory)
			}
			if cfg.EnableMultiWorkspace != tt.enableMultiWorkspace {
				t.Errorf("expected EnableMultiWorkspace=%v, got %v", tt.enableMultiWorkspace, cfg.EnableMultiWorkspace)
			}
			if cfg.EnableThinking != tt.enableThinking {
				t.Errorf("expected EnableThinking=%v, got %v", tt.enableThinking, cfg.EnableThinking)
			}
			if cfg.ThinkingBudget != tt.thinkingBudget {
				t.Errorf("expected ThinkingBudget=%d, got %d", tt.thinkingBudget, cfg.ThinkingBudget)
			}
		})
	}
}

func TestConfig_Default(t *testing.T) {
	// Test that Config can be created with zero values
	cfg := Config{}

	if cfg.Model != nil {
		t.Errorf("expected Model=nil for empty config, got %v", cfg.Model)
	}
	if cfg.WorkingDirectory != "" {
		t.Errorf("expected WorkingDirectory='', got '%s'", cfg.WorkingDirectory)
	}
	if cfg.EnableMultiWorkspace {
		t.Error("expected EnableMultiWorkspace=false for empty config")
	}
	if cfg.EnableThinking {
		t.Error("expected EnableThinking=false for empty config")
	}
	if cfg.ThinkingBudget != 0 {
		t.Errorf("expected ThinkingBudget=0, got %d", cfg.ThinkingBudget)
	}
}

func TestConfig_WithThinkingBudget(t *testing.T) {
	// Test Config with thinking enabled
	cfg := Config{
		EnableThinking: true,
		ThinkingBudget: 50000,
	}

	if !cfg.EnableThinking {
		t.Error("expected EnableThinking=true")
	}
	if cfg.ThinkingBudget != 50000 {
		t.Errorf("expected ThinkingBudget=50000, got %d", cfg.ThinkingBudget)
	}
}

func TestPromptContext_Empty(t *testing.T) {
	// Test PromptContext with no workspace
	ctx := PromptContext{
		HasWorkspace: false,
	}

	if ctx.HasWorkspace {
		t.Error("expected HasWorkspace=false")
	}
	if ctx.WorkspaceRoot != "" {
		t.Errorf("expected WorkspaceRoot='', got '%s'", ctx.WorkspaceRoot)
	}
}

func TestPromptContext_WithWorkspace(t *testing.T) {
	// Test PromptContext with workspace info
	ctx := PromptContext{
		HasWorkspace:        true,
		WorkspaceRoot:       "/home/user/project",
		WorkspaceSummary:    "Project with Go files",
		EnvironmentMetadata: "Git: main branch, 10 commits",
	}

	if !ctx.HasWorkspace {
		t.Error("expected HasWorkspace=true")
	}
	if ctx.WorkspaceRoot != "/home/user/project" {
		t.Errorf("expected WorkspaceRoot=/home/user/project, got %s", ctx.WorkspaceRoot)
	}
	if ctx.WorkspaceSummary != "Project with Go files" {
		t.Errorf("expected proper WorkspaceSummary, got %s", ctx.WorkspaceSummary)
	}
}

func TestConfig_WorkingDirectoryCanBeEmpty(t *testing.T) {
	// Working directory should be able to be empty (will use current dir)
	cfg := Config{
		WorkingDirectory: "",
	}

	if cfg.WorkingDirectory != "" {
		t.Errorf("expected empty WorkingDirectory, got '%s'", cfg.WorkingDirectory)
	}
}

func TestConfig_MultiWorkspaceIndependent(t *testing.T) {
	// MultiWorkspace and thinking should be independent flags
	tests := []struct {
		name                 string
		enableMultiWorkspace bool
		enableThinking       bool
	}{
		{"both enabled", true, true},
		{"only multi-workspace", true, false},
		{"only thinking", false, true},
		{"both disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				EnableMultiWorkspace: tt.enableMultiWorkspace,
				EnableThinking:       tt.enableThinking,
			}

			if cfg.EnableMultiWorkspace != tt.enableMultiWorkspace {
				t.Errorf("expected EnableMultiWorkspace=%v, got %v", tt.enableMultiWorkspace, cfg.EnableMultiWorkspace)
			}
			if cfg.EnableThinking != tt.enableThinking {
				t.Errorf("expected EnableThinking=%v, got %v", tt.enableThinking, cfg.EnableThinking)
			}
		})
	}
}

func TestGetProjectRoot_FindsGoMod(t *testing.T) {
	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// GetProjectRoot should find the go.mod in a parent directory
	projectRoot, err := GetProjectRoot(workDir)
	if err != nil {
		t.Fatalf("GetProjectRoot failed: %v", err)
	}

	if projectRoot == "" {
		t.Fatal("expected non-empty project root")
	}

	// Verify go.mod exists at project root
	goModPath := filepath.Join(projectRoot, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		t.Fatalf("expected go.mod at %s, but got error: %v", goModPath, err)
	}
}

func TestGetProjectRoot_ValidPath(t *testing.T) {
	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Find project root
	projectRoot, err := GetProjectRoot(workDir)
	if err != nil {
		t.Fatalf("GetProjectRoot failed: %v", err)
	}

	// Path should be an absolute path
	if !filepath.IsAbs(projectRoot) {
		t.Errorf("expected absolute path, got %s", projectRoot)
	}

	// Directory should exist
	stat, err := os.Stat(projectRoot)
	if err != nil {
		t.Fatalf("expected project root to exist: %v", err)
	}
	if !stat.IsDir() {
		t.Errorf("expected project root to be a directory, but it's a file")
	}
}

func TestGetProjectRoot_Deprecated(t *testing.T) {
	// This tests that the deprecated function still works
	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// The deprecated GetProjectRoot should delegate to workspace.GetProjectRoot
	result, err := GetProjectRoot(workDir)
	if err != nil {
		t.Fatalf("GetProjectRoot failed: %v", err)
	}

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	// Should have go.mod at result
	_, err = os.Stat(filepath.Join(result, "go.mod"))
	if err != nil {
		t.Fatalf("expected go.mod at project root: %v", err)
	}
}

func TestPromptContext_Fields(t *testing.T) {
	tests := []struct {
		name                 string
		hasWorkspace         bool
		workspaceRoot        string
		workspaceSummary     string
		environmentMetadata  string
		enableMultiWorkspace bool
	}{
		{
			name:                 "no workspace",
			hasWorkspace:         false,
			workspaceRoot:        "",
			workspaceSummary:     "",
			environmentMetadata:  "",
			enableMultiWorkspace: false,
		},
		{
			name:                 "with workspace",
			hasWorkspace:         true,
			workspaceRoot:        "/home/user/project",
			workspaceSummary:     "Project with main.go and test files",
			environmentMetadata:  "Git: main branch, 5 commits",
			enableMultiWorkspace: false,
		},
		{
			name:                 "multi-workspace enabled",
			hasWorkspace:         true,
			workspaceRoot:        "/monorepo",
			workspaceSummary:     "Monorepo with 3 packages",
			environmentMetadata:  "Git: develop branch",
			enableMultiWorkspace: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := PromptContext{
				HasWorkspace:         tt.hasWorkspace,
				WorkspaceRoot:        tt.workspaceRoot,
				WorkspaceSummary:     tt.workspaceSummary,
				EnvironmentMetadata:  tt.environmentMetadata,
				EnableMultiWorkspace: tt.enableMultiWorkspace,
			}

			if ctx.HasWorkspace != tt.hasWorkspace {
				t.Errorf("expected HasWorkspace=%v, got %v", tt.hasWorkspace, ctx.HasWorkspace)
			}
			if ctx.WorkspaceRoot != tt.workspaceRoot {
				t.Errorf("expected WorkspaceRoot=%s, got %s", tt.workspaceRoot, ctx.WorkspaceRoot)
			}
			if ctx.WorkspaceSummary != tt.workspaceSummary {
				t.Errorf("expected WorkspaceSummary=%s, got %s", tt.workspaceSummary, ctx.WorkspaceSummary)
			}
			if ctx.EnvironmentMetadata != tt.environmentMetadata {
				t.Errorf("expected EnvironmentMetadata=%s, got %s", tt.environmentMetadata, ctx.EnvironmentMetadata)
			}
			if ctx.EnableMultiWorkspace != tt.enableMultiWorkspace {
				t.Errorf("expected EnableMultiWorkspace=%v, got %v", tt.enableMultiWorkspace, ctx.EnableMultiWorkspace)
			}
		})
	}
}
