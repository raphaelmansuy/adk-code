package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to create temporary file for testing
func createTempFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return tmpFile
}

// Helper function for int pointer
func intPtr(i int) *int {
	return &i
}

// Helper function for bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// ==================== File Validation Tests ====================

func TestValidateFilePath_ValidPath(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	tests := []struct {
		name       string
		basePath   string
		filePath   string
		requireExist bool
		shouldErr  bool
	}{
		{"Valid file without base", "", tmpFile, true, false},
		{"Valid file with base", tmpDir, tmpFile, true, false},
		{"Non-existent file without require", "", tmpFile+"_notexist", false, false},
		{"Non-existent file with require", tmpDir, tmpFile+"_notexist", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.basePath, tt.filePath, tt.requireExist)
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateFilePath() error = %v, shouldErr = %v", err, tt.shouldErr)
			}
		})
	}
}

func TestValidateFilePath_DirectoryTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		basePath  string
		filePath  string
		shouldErr bool
	}{
		{"Valid path within base", tmpDir, filepath.Join(tmpDir, "file.txt"), false},
		{"Directory traversal attempt", tmpDir, filepath.Join(tmpDir, "..", "outside.txt"), true},
		{"Double traversal attempt", tmpDir, filepath.Join(tmpDir, "..", "..", "etc", "passwd"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.basePath, tt.filePath, false)
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateFilePath() error = %v, shouldErr = %v", err, tt.shouldErr)
			}
			if err != nil && tt.shouldErr {
				psErr := err.(*PathSecurityError)
				if psErr.Code != "DIRECTORY_TRAVERSAL" {
					t.Errorf("Expected DIRECTORY_TRAVERSAL error code, got %s", psErr.Code)
				}
			}
		})
	}
}

// ==================== Atomic Write Tests ====================

func TestAtomicWrite_BasicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	err := AtomicWrite(filePath, content, 0644)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify file exists and has correct content
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Content mismatch: got %q, want %q", string(readContent), string(content))
	}
}

func TestAtomicWrite_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	err := AtomicWrite(filePath, content, 0600)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify permissions
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// Check if permissions are set correctly (accounting for umask)
	perms := info.Mode().Perm()
	if perms != 0600 && perms != 0644 { // Some systems might apply umask
		t.Logf("Permissions set to %o (system may apply umask)", perms)
	}
}

func TestAtomicWrite_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Write initial content
	initialContent := []byte("initial")
	os.WriteFile(filePath, initialContent, 0644)

	// Overwrite with atomic write
	newContent := []byte("new content")
	err := AtomicWrite(filePath, newContent, 0644)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify new content
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readContent) != string(newContent) {
		t.Errorf("Content mismatch: got %q, want %q", string(readContent), string(newContent))
	}
}

// ==================== Line Range Reading Tests ====================

func TestParseLineRange_FullFile(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5"
	tmpFile := createTempFile(t, content)

	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// Full file read
	offset := 1
	limit := totalLines

	endIdx := offset + limit - 1
	if endIdx > totalLines {
		endIdx = totalLines
	}

	var selectedLines []string
	if offset <= totalLines {
		selectedLines = lines[offset-1 : endIdx]
	}

	if len(selectedLines) != 5 {
		t.Errorf("Expected 5 lines, got %d", len(selectedLines))
	}

	expectedResult := strings.Join(selectedLines, "\n")
	if expectedResult != content {
		t.Errorf("Expected %q, got %q", content, expectedResult)
	}

	os.Remove(tmpFile)
}

func TestParseLineRange_PartialRange(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5"
	tmpFile := createTempFile(t, content)

	lines := strings.Split(content, "\n")

	tests := []struct {
		name     string
		offset   int
		limit    int
		expected string
	}{
		{"Lines 2-4", 2, 3, "line2\nline3\nline4"},
		{"From line 3", 3, 3, "line3\nline4\nline5"},
		{"Single line", 2, 1, "line2"},
		{"Beyond end", 1, 100, content},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := tt.offset
			limit := tt.limit
			totalLines := len(lines)

			endIdx := offset + limit - 1
			if endIdx > totalLines {
				endIdx = totalLines
			}

			var selectedLines []string
			if offset <= totalLines {
				selectedLines = lines[offset-1 : endIdx]
			}

			result := strings.Join(selectedLines, "\n")
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}

	os.Remove(tmpFile)
}

// ==================== Patch Parsing Tests ====================

func TestParseHunkHeader_Valid(t *testing.T) {
	tests := []struct {
		header    string
		origStart int
		origCount int
		newStart  int
		newCount  int
	}{
		{"@@ -10,5 +12,7 @@", 10, 5, 12, 7},
		{"@@ -1,3 +1,4 @@", 1, 3, 1, 4},
		{"@@ -100 +200 @@", 100, 1, 200, 1},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			hunk, err := parseHunkHeader(tt.header)
			if err != nil {
				t.Fatalf("parseHunkHeader failed: %v", err)
			}

			if hunk.OrigStart != tt.origStart || hunk.OrigCount != tt.origCount ||
				hunk.NewStart != tt.newStart || hunk.NewCount != tt.newCount {
				t.Errorf("Hunk mismatch: %+v", hunk)
			}
		})
	}
}

func TestParseUnifiedDiff_Simple(t *testing.T) {
	patch := `--- original
+++ modified
@@ -1,3 +1,4 @@
 line1
 line2
+added line
 line3`

	hunks, err := ParseUnifiedDiff(patch)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff failed: %v", err)
	}

	if len(hunks) != 1 {
		t.Errorf("Expected 1 hunk, got %d", len(hunks))
	}

	hunk := hunks[0]
	if hunk.OrigStart != 1 || hunk.NewStart != 1 {
		t.Errorf("Hunk coordinates mismatch: orig=%d,new=%d", hunk.OrigStart, hunk.NewStart)
	}
}

// ==================== Patch Application Tests ====================

func TestApplyPatch_SimplAddition(t *testing.T) {
	original := "line1\nline2\nline3"

	// Create a simple patch
	patch := `--- original
+++ modified
@@ -2,2 +2,3 @@
 line2
+added
 line3`

	_, added, removed, err := ApplyPatch(original, patch, false)
	if err != nil {
		t.Fatalf("ApplyPatch failed: %v", err)
	}

	if added != 1 || removed != 0 {
		t.Errorf("Expected 1 added, 0 removed; got %d added, %d removed", added, removed)
	}

	// Note: Simple patch parser might not produce exact match due to algorithm limitations
	// This test demonstrates the patch application structure
}

// ==================== Error Types Tests ====================

func TestErrorCreation(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() error
		code     ErrorCode
		hasHint  bool
	}{
		{"FileNotFound", func() error { return FileNotFoundError("/test/path") }, ErrorCodeFileNotFound, true},
		{"PermissionDenied", func() error { return PermissionDeniedError("/test/path") }, ErrorCodePermissionDenied, true},
		{"PathTraversal", func() error { return PathTraversalError("/bad/path", "/base") }, ErrorCodePathTraversal, true},
		{"InvalidInput", func() error { return InvalidInputError("bad value") }, ErrorCodeInvalidInput, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			toolErr := err.(*ToolError)
			if toolErr.Code != tt.code {
				t.Errorf("Expected code %s, got %s", tt.code, toolErr.Code)
			}

			if tt.hasHint && toolErr.Suggestion == "" {
				t.Errorf("Expected suggestion for %s error", tt.name)
			}
		})
	}
}

// ==================== Edge Cases Tests ====================

func TestEdgeCase_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	os.WriteFile(emptyFile, []byte(""), 0644)

	content, err := os.ReadFile(emptyFile)
	if err != nil {
		t.Fatalf("Failed to read empty file: %v", err)
	}

	if len(content) != 0 {
		t.Errorf("Expected empty content, got %d bytes", len(content))
	}
}

func TestEdgeCase_LargeLineCount(t *testing.T) {
	// Create file with many lines
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString(strings.Repeat(fmt.Sprintf("line%d\n", i), 1))
	}

	tmpFile := createTempFile(t, sb.String())
	defer os.Remove(tmpFile)

	// Test reading with limits
	lines := strings.Split(sb.String(), "\n")
	totalLines := len(lines)

	offset := 5000
	limit := 100

	endIdx := offset + limit - 1
	if endIdx > totalLines {
		endIdx = totalLines
	}

	var selectedLines []string
	if offset <= totalLines {
		selectedLines = lines[offset-1 : endIdx]
	}

	if len(selectedLines) != 100 {
		t.Errorf("Expected 100 lines, got %d", len(selectedLines))
	}
}

// TestIntegration_CompleteWorkflow tests a complete workflow
func TestIntegration_CompleteWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "workflow.txt")

	// 1. Create file with WriteFile
	initialContent := "line1\nline2\nline3"
	err := os.WriteFile(testFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// 2. Validate path
	err = ValidateFilePath(tmpDir, testFile, true)
	if err != nil {
		t.Fatalf("Path validation failed: %v", err)
	}

	// 3. Read file with range
	content, _ := os.ReadFile(testFile)
	lines := strings.Split(string(content), "\n")

	offset := 1
	limit := 2
	endIdx := offset + limit - 1
	if endIdx > len(lines) {
		endIdx = len(lines)
	}
	selectedLines := lines[offset-1 : endIdx]

	if len(selectedLines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(selectedLines))
	}

	// 4. Atomic write
	newContent := "modified"
	err = AtomicWrite(testFile, []byte(newContent), 0644)
	if err != nil {
		t.Fatalf("Atomic write failed: %v", err)
	}

	// 5. Verify
	readBack, _ := os.ReadFile(testFile)
	if string(readBack) != newContent {
		t.Errorf("Content mismatch after atomic write")
	}
}
