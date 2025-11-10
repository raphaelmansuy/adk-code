package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseV4APatch tests the V4A patch parser
func TestParseV4APatch(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		wantHunks   int
		wantMarkers int // Expected markers in first hunk
	}{
		{
			name: "simple function patch",
			input: `@@ func ProcessRequest
-    return nil
+    return processData(req)`,
			wantErr:     false,
			wantHunks:   1,
			wantMarkers: 1,
		},
		{
			name: "nested class method patch",
			input: `*** Update File: src/models/user.py
@@ class User
@@     def validate():
-          return True
+          if not self.email:
+              raise ValueError("Email required")
+          return True`,
			wantErr:     false,
			wantHunks:   1,
			wantMarkers: 2,
		},
		{
			name: "multiple hunks",
			input: `@@ func Init
-    setupA()
+    setupB()

@@ func Cleanup
-    cleanupA()
+    cleanupB()`,
			wantErr:   false,
			wantHunks: 2,
		},
		{
			name:    "empty patch",
			input:   "",
			wantErr: true,
		},
		{
			name: "no hunks",
			input: `*** Update File: test.go
// just comments`,
			wantErr: true,
		},
		{
			name: "context without changes",
			input: `@@ func Test
`,
			wantErr: true, // Should error - no changes
		},
		{
			name: "changes before context",
			input: `-    invalid
+    content
@@ func Test`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch, err := ParseV4APatch(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseV4APatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(patch.Hunks) != tt.wantHunks {
				t.Errorf("ParseV4APatch() got %d hunks, want %d", len(patch.Hunks), tt.wantHunks)
			}
			if tt.wantMarkers > 0 && len(patch.Hunks) > 0 {
				if len(patch.Hunks[0].ContextMarkers) != tt.wantMarkers {
					t.Errorf("First hunk has %d markers, want %d",
						len(patch.Hunks[0].ContextMarkers), tt.wantMarkers)
				}
			}
		})
	}
}

// TestApplyV4APatch tests applying V4A patches to files
func TestApplyV4APatch(t *testing.T) {
	tests := []struct {
		name         string
		fileContent  string
		patchContent string
		wantContent  string
		wantErr      bool
	}{
		{
			name: "simple function replacement",
			fileContent: `package main

func ProcessRequest() error {
    return nil
}`,
			patchContent: `@@ func ProcessRequest
-    return nil
+    return processData(req)`,
			wantContent: `package main

func ProcessRequest() error {
    return processData(req)
}`,
			wantErr: false,
		},
		{
			name: "nested method with indentation",
			fileContent: `class User:
    def validate(self):
        return True
`,
			patchContent: `@@ class User
@@     def validate
-        return True
+        if not self.email:
+            raise ValueError("Email required")
+        return True`,
			wantContent: `class User:
    def validate(self):
        if not self.email:
            raise ValueError("Email required")
        return True
`,
			wantErr: false,
		},
		{
			name: "context not found",
			fileContent: `package main

func OtherFunc() {}`,
			patchContent: `@@ func NonExistent
-    line
+    replacement`,
			wantErr: true,
		},
		{
			name: "removal mismatch",
			fileContent: `package main

func Test() {
    actualLine()
}`,
			patchContent: `@@ func Test
-    expectedLine()
+    replacement()`,
			wantErr: true,
		},
		{
			name: "insertion only (no removals)",
			fileContent: `package main

func Init() {
}`,
			patchContent: `@@ func Init
+    setupCode()`,
			wantContent: `package main

func Init() {
    setupCode()
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test_file.go")
			if err := os.WriteFile(testFile, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse patch
			patch, err := ParseV4APatch(tt.patchContent)
			if err != nil {
				t.Fatalf("ParseV4APatch() failed: %v", err)
			}

			// Apply patch
			_, err = ApplyV4APatch(testFile, patch, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyV4APatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify content
			gotContent, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read result: %v", err)
			}

			if string(gotContent) != tt.wantContent {
				t.Errorf("ApplyV4APatch() content mismatch\nGot:\n%s\nWant:\n%s",
					string(gotContent), tt.wantContent)
			}
		})
	}
}

// TestApplyV4APatchDryRun tests dry run mode
func TestApplyV4APatchDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	original := `package main

func Test() {
    oldCode()
}`

	if err := os.WriteFile(testFile, []byte(original), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	patchContent := `@@ func Test
-    oldCode()
+    newCode()`

	patch, err := ParseV4APatch(patchContent)
	if err != nil {
		t.Fatalf("ParseV4APatch() failed: %v", err)
	}

	// Apply in dry run mode
	result, err := ApplyV4APatch(testFile, patch, true)
	if err != nil {
		t.Fatalf("ApplyV4APatch(dryRun=true) failed: %v", err)
	}

	// Verify result contains preview info
	if !strings.Contains(result, "DRY RUN") {
		t.Errorf("Dry run result should contain 'DRY RUN', got: %s", result)
	}

	// Verify file was NOT modified
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != original {
		t.Errorf("Dry run should not modify file. Got:\n%s\nWant:\n%s",
			string(content), original)
	}
}

// TestApplyV4APatchTool tests the tool integration
func TestApplyV4APatchTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.go")
	original := `package main

func Greet() string {
    return "hello"
}`

	if err := os.WriteFile(testFile, []byte(original), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create tool
	tool, err := NewApplyV4APatchTool(tmpDir)
	if err != nil {
		t.Fatalf("NewApplyV4APatchTool() failed: %v", err)
	}

	_ = tool // Tool is created successfully

	// Note: Full tool execution would require tool.Context which is not easy to mock
	// The parser and applier are tested separately above
}

// TestFindContextLocation tests context marker searching
func TestFindContextLocation(t *testing.T) {
	lines := []string{
		"package main",
		"",
		"type User struct {",
		"    Name string",
		"}",
		"",
		"func (u *User) Validate() error {",
		"    if u.Name == \"\" {",
		"        return errors.New(\"name required\")",
		"    }",
		"    return nil",
		"}",
	}

	tests := []struct {
		name    string
		markers []string
		want    int
		wantErr bool
	}{
		{
			name:    "find struct",
			markers: []string{"type User struct"},
			want:    2,
			wantErr: false,
		},
		{
			name:    "find method",
			markers: []string{"func (u *User) Validate"},
			want:    6,
			wantErr: false,
		},
		{
			name:    "nested context",
			markers: []string{"type User struct", "Name string"},
			want:    3,
			wantErr: false,
		},
		{
			name:    "not found",
			markers: []string{"NonExistent"},
			want:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findContextLocation(lines, tt.markers)
			if (err != nil) != tt.wantErr {
				t.Errorf("findContextLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("findContextLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}
