package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFileExistence tests file existence checking in resolver
func TestFileExistence(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create workspace directories
	ws1 := filepath.Join(tempDir, "workspace1")
	ws2 := filepath.Join(tempDir, "workspace2")

	if err := os.MkdirAll(ws1, 0755); err != nil {
		t.Fatalf("Failed to create workspace1: %v", err)
	}
	if err := os.MkdirAll(ws2, 0755); err != nil {
		t.Fatalf("Failed to create workspace2: %v", err)
	}

	// Create test files
	file1 := filepath.Join(ws1, "test.txt")
	file2 := filepath.Join(ws2, "test.txt")
	uniqueFile := filepath.Join(ws1, "unique.txt")

	if err := os.WriteFile(file1, []byte("workspace1 content"), 0644); err != nil {
		t.Fatalf("Failed to create test file in workspace1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("workspace2 content"), 0644); err != nil {
		t.Fatalf("Failed to create test file in workspace2: %v", err)
	}
	if err := os.WriteFile(uniqueFile, []byte("unique content"), 0644); err != nil {
		t.Fatalf("Failed to create unique file: %v", err)
	}

	// Create workspace manager
	roots := []WorkspaceRoot{
		{Path: ws1, Name: "workspace1", VCS: VCSTypeNone},
		{Path: ws2, Name: "workspace2", VCS: VCSTypeNone},
	}
	manager := NewManager(roots, 0)
	resolver := NewResolver(manager)

	// Test 1: DisambiguatePath with file existing in both workspaces
	matches := resolver.DisambiguatePath("test.txt")
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches for test.txt, got %d", len(matches))
	}

	// Test 2: DisambiguatePath with file existing in only one workspace
	matches = resolver.DisambiguatePath("unique.txt")
	if len(matches) != 1 {
		t.Errorf("Expected 1 match for unique.txt, got %d", len(matches))
	}
	if len(matches) > 0 && matches[0] != "workspace1" {
		t.Errorf("Expected match in workspace1, got %s", matches[0])
	}

	// Test 3: DisambiguatePath with non-existent file
	matches = resolver.DisambiguatePath("nonexistent.txt")
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for nonexistent.txt, got %d", len(matches))
	}

	// Test 4: ResolvePathWithDisambiguation for unique file
	resolved, err := resolver.ResolvePathWithDisambiguation("unique.txt")
	if err != nil {
		t.Errorf("Failed to resolve unique.txt: %v", err)
	}
	if resolved.Root.Name != "workspace1" {
		t.Errorf("Expected resolution to workspace1, got %s", resolved.Root.Name)
	}

	// Test 5: ResolvePathWithDisambiguation for ambiguous file (should prefer primary)
	resolved, err = resolver.ResolvePathWithDisambiguation("test.txt")
	if err != nil {
		t.Errorf("Failed to resolve test.txt: %v", err)
	}
	if resolved.Root.Name != "workspace1" {
		t.Errorf("Expected resolution to primary workspace (workspace1), got %s", resolved.Root.Name)
	}

	// Test 6: FileExists method
	if !resolver.FileExists("test.txt") {
		t.Error("FileExists returned false for existing file")
	}
	if resolver.FileExists("nonexistent.txt") {
		t.Error("FileExists returned true for non-existent file")
	}
}

// TestConfigPersistence tests saving and loading workspace configuration
func TestConfigPersistence(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create workspace roots
	ws1 := filepath.Join(tempDir, "workspace1")
	ws2 := filepath.Join(tempDir, "workspace2")

	if err := os.MkdirAll(ws1, 0755); err != nil {
		t.Fatalf("Failed to create workspace1: %v", err)
	}
	if err := os.MkdirAll(ws2, 0755); err != nil {
		t.Fatalf("Failed to create workspace2: %v", err)
	}

	// Create manager
	roots := []WorkspaceRoot{
		{Path: ws1, Name: "workspace1", VCS: VCSTypeGit},
		{Path: ws2, Name: "workspace2", VCS: VCSTypeNone},
	}
	manager := NewManager(roots, 1) // Set workspace2 as primary

	// Save config
	prefs := DefaultPreferences()
	prefs.MaxWorkspaces = 5
	err := SaveManagerToDirectory(tempDir, manager, &prefs)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Check if config file exists
	if !ConfigExists(tempDir) {
		t.Error("Config file was not created")
	}

	// Load config
	loadedManager, loadedPrefs, err := LoadManagerFromDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded data
	if len(loadedManager.GetRoots()) != 2 {
		t.Errorf("Expected 2 roots, got %d", len(loadedManager.GetRoots()))
	}

	if loadedManager.GetPrimaryIndex() != 1 {
		t.Errorf("Expected primary index 1, got %d", loadedManager.GetPrimaryIndex())
	}

	if loadedPrefs.MaxWorkspaces != 5 {
		t.Errorf("Expected MaxWorkspaces 5, got %d", loadedPrefs.MaxWorkspaces)
	}

	// Test validation
	config := &Config{
		Version:      1,
		Roots:        roots,
		PrimaryIndex: 1,
		Preferences:  prefs,
	}

	if err := ValidateConfig(config); err != nil {
		t.Errorf("Valid config failed validation: %v", err)
	}

	// Test invalid config
	invalidConfig := &Config{
		Version:      1,
		Roots:        []WorkspaceRoot{},
		PrimaryIndex: 0,
	}

	if err := ValidateConfig(invalidConfig); err == nil {
		t.Error("Empty roots config should fail validation")
	}
}

// TestMultiWorkspaceDetection tests automatic workspace detection
func TestMultiWorkspaceDetection(t *testing.T) {
	// Create temporary directory with multiple projects
	tempDir := t.TempDir()

	// Create project structures
	frontend := filepath.Join(tempDir, "frontend")
	backend := filepath.Join(tempDir, "backend")
	nested := filepath.Join(tempDir, "nested", "project")

	for _, dir := range []string{frontend, backend, nested} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Add workspace markers
	os.WriteFile(filepath.Join(frontend, "package.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(backend, "go.mod"), []byte("module test\n"), 0644)
	os.WriteFile(filepath.Join(nested, "Cargo.toml"), []byte("[package]\n"), 0644)

	// Test detection with default options
	options := DefaultDetectionOptions()
	roots, err := DetectWorkspaces(tempDir, options)
	if err != nil {
		t.Fatalf("Failed to detect workspaces: %v", err)
	}

	if len(roots) < 3 {
		t.Errorf("Expected at least 3 workspaces, found %d", len(roots))
	}

	// Verify workspace names
	names := make(map[string]bool)
	for _, root := range roots {
		names[root.Name] = true
	}

	expectedNames := []string{"frontend", "backend", "project"}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("Expected to find workspace %s", name)
		}
	}

	// Test detection with max workspace limit
	options.MaxWorkspaces = 2
	roots, err = DetectWorkspaces(tempDir, options)
	if err != nil {
		t.Fatalf("Failed to detect workspaces with limit: %v", err)
	}

	if len(roots) > 2 {
		t.Errorf("Expected at most 2 workspaces, found %d", len(roots))
	}

	// Test detection with max depth limit
	options = DefaultDetectionOptions()
	options.MaxDepth = 1
	roots, err = DetectWorkspaces(tempDir, options)
	if err != nil {
		t.Fatalf("Failed to detect workspaces with depth limit: %v", err)
	}

	// Should not find the nested project
	for _, root := range roots {
		if root.Name == "project" {
			t.Error("Should not have found nested project with MaxDepth=1")
		}
	}
}

// TestWorkspaceSwitching tests switching between workspaces
func TestWorkspaceSwitching(t *testing.T) {
	// Create temporary workspaces
	tempDir := t.TempDir()

	ws1 := filepath.Join(tempDir, "workspace1")
	ws2 := filepath.Join(tempDir, "workspace2")

	if err := os.MkdirAll(ws1, 0755); err != nil {
		t.Fatalf("Failed to create workspace1: %v", err)
	}
	if err := os.MkdirAll(ws2, 0755); err != nil {
		t.Fatalf("Failed to create workspace2: %v", err)
	}

	// Create manager
	roots := []WorkspaceRoot{
		{Path: ws1, Name: "workspace1", VCS: VCSTypeNone},
		{Path: ws2, Name: "workspace2", VCS: VCSTypeNone},
	}
	manager := NewManager(roots, 0)

	// Test 1: Initial primary workspace
	primary := manager.GetPrimaryRoot()
	if primary.Name != "workspace1" {
		t.Errorf("Expected primary to be workspace1, got %s", primary.Name)
	}

	// Test 2: Switch by name
	err := manager.SetPrimaryByName("workspace2")
	if err != nil {
		t.Errorf("Failed to switch workspace by name: %v", err)
	}

	primary = manager.GetPrimaryRoot()
	if primary.Name != "workspace2" {
		t.Errorf("Expected primary to be workspace2 after switch, got %s", primary.Name)
	}

	// Test 3: Switch by path
	err = manager.SetPrimaryByPath(ws1)
	if err != nil {
		t.Errorf("Failed to switch workspace by path: %v", err)
	}

	primary = manager.GetPrimaryRoot()
	if primary.Name != "workspace1" {
		t.Errorf("Expected primary to be workspace1 after switch, got %s", primary.Name)
	}

	// Test 4: Switch by index
	err = manager.SetPrimaryIndex(1)
	if err != nil {
		t.Errorf("Failed to switch workspace by index: %v", err)
	}

	primary = manager.GetPrimaryRoot()
	if primary.Name != "workspace2" {
		t.Errorf("Expected primary to be workspace2 after index switch, got %s", primary.Name)
	}

	// Test 5: SwitchWorkspace method with name
	newPrimary, err := manager.SwitchWorkspace("workspace1")
	if err != nil {
		t.Errorf("SwitchWorkspace failed: %v", err)
	}
	if newPrimary.Name != "workspace1" {
		t.Errorf("Expected workspace1, got %s", newPrimary.Name)
	}

	// Test 6: SwitchWorkspace method with path
	newPrimary, err = manager.SwitchWorkspace(ws2)
	if err != nil {
		t.Errorf("SwitchWorkspace with path failed: %v", err)
	}
	if newPrimary.Name != "workspace2" {
		t.Errorf("Expected workspace2, got %s", newPrimary.Name)
	}

	// Test 7: Error cases
	err = manager.SetPrimaryByName("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent workspace name")
	}

	err = manager.SetPrimaryByPath("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for nonexistent workspace path")
	}

	err = manager.SetPrimaryIndex(10)
	if err == nil {
		t.Error("Expected error for invalid workspace index")
	}
}

// TestSmartInitialization tests the smart workspace initialization
func TestSmartInitialization(t *testing.T) {
	// Test 1: With existing config
	tempDir := t.TempDir()

	ws1 := filepath.Join(tempDir, "workspace1")
	if err := os.MkdirAll(ws1, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Create initial manager and save config
	roots := []WorkspaceRoot{
		{Path: ws1, Name: "workspace1", VCS: VCSTypeNone},
	}
	initialManager := NewManager(roots, 0)
	SaveManagerToDirectory(tempDir, initialManager, nil)

	// Test smart initialization loads the config
	manager, err := SmartWorkspaceInitialization(tempDir)
	if err != nil {
		t.Fatalf("Smart initialization failed: %v", err)
	}

	if len(manager.GetRoots()) != 1 {
		t.Errorf("Expected 1 root from config, got %d", len(manager.GetRoots()))
	}

	// Test 2: Without config but with detectable workspaces
	tempDir2 := t.TempDir()

	project := filepath.Join(tempDir2, "project")
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	os.WriteFile(filepath.Join(project, "go.mod"), []byte("module test\n"), 0644)

	manager, err = SmartWorkspaceInitialization(tempDir2)
	if err != nil {
		t.Fatalf("Smart initialization failed for auto-detection: %v", err)
	}

	// Should have detected the project
	if len(manager.GetRoots()) == 0 {
		t.Error("Expected at least one workspace to be detected")
	}

	// Test 3: Fall back to single directory
	tempDir3 := t.TempDir()

	manager, err = SmartWorkspaceInitialization(tempDir3)
	if err != nil {
		t.Fatalf("Smart initialization failed for fallback: %v", err)
	}

	if !manager.IsSingleRoot() {
		t.Error("Expected single root for empty directory")
	}
}
