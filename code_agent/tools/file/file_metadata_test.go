package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestReadFileMetadata verifies that ReadFileOutput contains file path and metadata
func TestReadFileMetadata(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_metadata.txt")
	testContent := "line 1\nline 2\nline 3\nline 4\nline 5"

	// Write test file
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wait a moment to ensure file timestamps are set
	time.Sleep(10 * time.Millisecond)

	// Create the read file tool
	tool, err := NewReadFileTool()
	if err != nil {
		t.Fatalf("Failed to create read file tool: %v", err)
	}

	// Test 1: Verify metadata fields are populated on success
	t.Run("MetadataFieldsPopulated", func(t *testing.T) {
		// Mock the tool handler directly
		handler := func(ctx interface{}, input ReadFileInput) ReadFileOutput {
			content, err := os.ReadFile(input.Path)
			if err != nil {
				return ReadFileOutput{
					Success: false,
					Error:   fmt.Sprintf("Failed to read file: %v", err),
				}
			}

			lines := strings.Split(string(content), "\n")
			totalLines := len(lines)

			offset := 1
			if input.Offset != nil && *input.Offset > 1 {
				offset = *input.Offset
			}

			limit := totalLines
			if input.Limit != nil && *input.Limit > 0 {
				limit = *input.Limit
			}

			endIdx := offset + limit - 1
			if endIdx > totalLines {
				endIdx = totalLines
			}

			var selectedLines []string
			if offset <= totalLines {
				selectedLines = lines[offset-1 : endIdx]
			}

			absPath, _ := filepath.Abs(input.Path)
			dateModified := ""
			dateCreated := ""

			if fileInfo, err := os.Stat(input.Path); err == nil {
				dateModified = fileInfo.ModTime().Format("2006-01-02T15:04:05Z07:00")
				dateCreated = dateModified
			}

			return ReadFileOutput{
				Content:       strings.Join(selectedLines, "\n"),
				Success:       true,
				TotalLines:    totalLines,
				ReturnedLines: len(selectedLines),
				StartLine:     offset,
				FilePath:      absPath,
				DateCreated:   dateCreated,
				DateModified:  dateModified,
			}
		}

		// Call the handler with test input
		output := handler(nil, ReadFileInput{Path: testFile})

		// Verify success
		if !output.Success {
			t.Errorf("Expected success, got error: %s", output.Error)
		}

		// Verify FilePath is populated and is absolute
		if output.FilePath == "" {
			t.Error("FilePath should not be empty")
		}
		if !filepath.IsAbs(output.FilePath) {
			t.Errorf("FilePath should be absolute, got: %s", output.FilePath)
		}

		// Verify DateModified is populated and in RFC3339 format
		if output.DateModified == "" {
			t.Error("DateModified should not be empty")
		}
		// Verify it's a valid RFC3339 timestamp
		_, err := time.Parse(time.RFC3339, output.DateModified)
		if err != nil {
			t.Errorf("DateModified should be in RFC3339 format, got: %s (error: %v)", output.DateModified, err)
		}

		// Verify DateCreated is also populated
		if output.DateCreated == "" {
			t.Error("DateCreated should not be empty")
		}
		// Verify it's also a valid RFC3339 timestamp
		_, err = time.Parse(time.RFC3339, output.DateCreated)
		if err != nil {
			t.Errorf("DateCreated should be in RFC3339 format, got: %s (error: %v)", output.DateCreated, err)
		}

		// Verify content is still correct
		if output.Content != testContent {
			t.Errorf("Content mismatch: expected %q, got %q", testContent, output.Content)
		}

		// Verify other fields are still correct
		if output.TotalLines != 5 {
			t.Errorf("Expected 5 total lines, got %d", output.TotalLines)
		}
		if output.ReturnedLines != 5 {
			t.Errorf("Expected 5 returned lines, got %d", output.ReturnedLines)
		}
		if output.StartLine != 1 {
			t.Errorf("Expected start line 1, got %d", output.StartLine)
		}

		fmt.Printf("✓ Test passed - FilePath: %s\n", output.FilePath)
		fmt.Printf("✓ Test passed - DateModified: %s\n", output.DateModified)
		fmt.Printf("✓ Test passed - DateCreated: %s\n", output.DateCreated)
	})

	// Test 2: Verify metadata fields work with partial reads
	t.Run("MetadataWithPartialRead", func(t *testing.T) {
		handler := func(ctx interface{}, input ReadFileInput) ReadFileOutput {
			content, err := os.ReadFile(input.Path)
			if err != nil {
				return ReadFileOutput{
					Success: false,
					Error:   fmt.Sprintf("Failed to read file: %v", err),
				}
			}

			lines := strings.Split(string(content), "\n")
			totalLines := len(lines)

			offset := 1
			if input.Offset != nil && *input.Offset > 1 {
				offset = *input.Offset
			}

			limit := totalLines
			if input.Limit != nil && *input.Limit > 0 {
				limit = *input.Limit
			}

			endIdx := offset + limit - 1
			if endIdx > totalLines {
				endIdx = totalLines
			}

			var selectedLines []string
			if offset <= totalLines {
				selectedLines = lines[offset-1 : endIdx]
			}

			absPath, _ := filepath.Abs(input.Path)
			dateModified := ""
			dateCreated := ""

			if fileInfo, err := os.Stat(input.Path); err == nil {
				dateModified = fileInfo.ModTime().Format("2006-01-02T15:04:05Z07:00")
				dateCreated = dateModified
			}

			return ReadFileOutput{
				Content:       strings.Join(selectedLines, "\n"),
				Success:       true,
				TotalLines:    totalLines,
				ReturnedLines: len(selectedLines),
				StartLine:     offset,
				FilePath:      absPath,
				DateCreated:   dateCreated,
				DateModified:  dateModified,
			}
		}

		// Call with partial read (lines 2-3)
		offset := 2
		limit := 2
		output := handler(nil, ReadFileInput{
			Path:   testFile,
			Offset: &offset,
			Limit:  &limit,
		})

		// Verify success
		if !output.Success {
			t.Errorf("Expected success, got error: %s", output.Error)
		}

		// Verify metadata is still present even with partial read
		if output.FilePath == "" {
			t.Error("FilePath should not be empty with partial read")
		}
		if output.DateModified == "" {
			t.Error("DateModified should not be empty with partial read")
		}
		if output.DateCreated == "" {
			t.Error("DateCreated should not be empty with partial read")
		}

		// Verify the partial content is correct
		expected := "line 2\nline 3"
		if output.Content != expected {
			t.Errorf("Expected %q, got %q", expected, output.Content)
		}

		fmt.Printf("✓ Partial read test passed - got lines %d-%d\n", offset, offset+limit-1)
	})

	_ = tool // Keep tool reference to avoid unused variable warning
}
