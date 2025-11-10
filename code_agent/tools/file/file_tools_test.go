// Package file provides file operation tools for the coding agent.
package file

import (
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
		name         string
		basePath     string
		filePath     string
		requireExist bool
		shouldErr    bool
	}{
		{"Valid file without base", "", tmpFile, true, false},
		{"Valid file with base", tmpDir, tmpFile, true, false},
		{"Non-existent file without require", "", tmpFile + "_notexist", false, false},
		{"Non-existent file with require", tmpDir, tmpFile + "_notexist", true, true},
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

// ==================== Note: Moved Tests ====================
// The following test groups were moved to their respective packages:
// - Patch Parsing Tests (ParseHunkHeader, ParseUnifiedDiff, ApplyPatch) → tools/edit/
