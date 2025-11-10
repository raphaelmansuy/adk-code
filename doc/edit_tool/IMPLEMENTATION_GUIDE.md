# Implementation Guide: Improving ADK Code Agent Tools

## Overview

This guide provides concrete, step-by-step implementation recommendations for enhancing the ADK Code Agent's file manipulation tools based on patterns and best practices from Cline.

---

## 1. Priority Implementation Plan

### Phase 1: Critical Enhancements (Week 1-2)

#### 1.1 Implement `apply_patch` Tool

**Why**: Patch-based editing is fundamentally more robust than string replacement.

**Input Structure**:
```go
type ApplyPatchInput struct {
    // Path to the file to patch
    FilePath string `json:"file_path" jsonschema:"Path to the file to patch"`
    
    // Patch in unified diff format (RFC 3881)
    Patch string `json:"patch" jsonschema:"Unified diff format patch"`
    
    // Preview mode: dry-run without making changes
    DryRun *bool `json:"dry_run,omitempty" jsonschema:"Preview mode - don't apply changes (default: false)"`
    
    // Strict mode: fail if patch doesn't apply cleanly
    Strict *bool `json:"strict,omitempty" jsonschema:"Require exact match (default: true)"`
}

type ApplyPatchOutput struct {
    Success     bool   `json:"success"`
    Message     string `json:"message,omitempty"`
    LinesAdded  int    `json:"lines_added"`
    LinesRemoved int   `json:"lines_removed"`
    Preview     string `json:"preview,omitempty"`  // Shown in dry-run mode
    Error       string `json:"error,omitempty"`
}
```

**Implementation**:
```go
// Go 1.21+ has no built-in unified diff application
// Use external library or implement custom patch parser

// Option 1: Use "github.com/go-patch/patch" or similar
// Option 2: Implement custom unified diff parser

// Key algorithm:
// 1. Parse unified diff hunks (hunk header: @@ -start,count +start,count @@)
// 2. Apply each hunk with context line matching
// 3. Handle fuzzy matching for line number offset
// 4. Generate reverse patch for rollback
```

**Benefits**:
- Resilient to code changes
- Multiple edits in single operation
- Reviewable/previewable
- Reversible

---

#### 1.2 Enhance `read_file` with Line Ranges

**Current**:
```go
type ReadFileInput struct {
    Path string
}
```

**Enhanced**:
```go
type ReadFileInput struct {
    Path   string `json:"path" jsonschema:"Path to the file to read"`
    
    // Start line (1-indexed, optional)
    Offset *int `json:"offset,omitempty" jsonschema:"Start line number (1-indexed, default: 1)"`
    
    // Number of lines to read (optional, 0 = all)
    Limit *int `json:"limit,omitempty" jsonschema:"Number of lines to read (default: all)"`
}

type ReadFileOutput struct {
    Content      string `json:"content"`
    Success      bool   `json:"success"`
    Error        string `json:"error,omitempty"`
    
    // NEW: Line information
    TotalLines   int `json:"total_lines"`
    ReturnedLines int `json:"returned_lines"`
    StartLine    int `json:"start_line"`
}
```

**Implementation**:
```go
func NewReadFileTool() (tool.Tool, error) {
    handler := func(ctx tool.Context, input ReadFileInput) ReadFileOutput {
        content, err := os.ReadFile(input.Path)
        if err != nil {
            return ReadFileOutput{Success: false, Error: fmt.Sprintf("Failed to read: %v", err)}
        }
        
        lines := strings.Split(string(content), "\n")
        totalLines := len(lines)
        
        // Handle offsets
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
            selectedLines = lines[offset-1:endIdx]
        }
        
        return ReadFileOutput{
            Content: strings.Join(selectedLines, "\n"),
            Success: true,
            TotalLines: totalLines,
            ReturnedLines: len(selectedLines),
            StartLine: offset,
        }
    }
    
    return functiontool.New(functiontool.Config{
        Name: "read_file",
        Description: "Reads file content with optional line range support",
    }, handler)
}
```

---

#### 1.3 Add Path Security Validation

**Create new utility**:
```go
// file_validation.go
package tools

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// PathSecurityError represents a path validation error
type PathSecurityError struct {
    Code    string // e.g., "DIRECTORY_TRAVERSAL", "OUTSIDE_BASE"
    Path    string
    Message string
}

// ValidateFilePath validates a file path for security and existence
func ValidateFilePath(basePath, requestedPath string, requireExist bool) error {
    // 1. Prevent directory traversal
    absRequested, err := filepath.Abs(requestedPath)
    if err != nil {
        return &PathSecurityError{
            Code: "INVALID_PATH",
            Path: requestedPath,
            Message: fmt.Sprintf("Invalid path: %v", err),
        }
    }
    
    absBase, err := filepath.Abs(basePath)
    if err != nil {
        return &PathSecurityError{
            Code: "INVALID_BASE",
            Path: basePath,
            Message: fmt.Sprintf("Invalid base path: %v", err),
        }
    }
    
    // 2. Check for directory traversal
    if !strings.HasPrefix(absRequested, absBase) {
        return &PathSecurityError{
            Code: "DIRECTORY_TRAVERSAL",
            Path: requestedPath,
            Message: fmt.Sprintf("Path traversal detected: %s is outside %s", absRequested, absBase),
        }
    }
    
    // 3. Resolve symlinks and verify again
    realPath, err := filepath.EvalSymlinks(absRequested)
    if err == nil && !strings.HasPrefix(realPath, absBase) {
        return &PathSecurityError{
            Code: "SYMLINK_ESCAPE",
            Path: requestedPath,
            Message: fmt.Sprintf("Symlink points outside base directory: %s", realPath),
        }
    }
    
    // 4. Check if file exists (if required)
    if requireExist {
        if _, err := os.Stat(absRequested); os.IsNotExist(err) {
            return &PathSecurityError{
                Code: "FILE_NOT_FOUND",
                Path: requestedPath,
                Message: fmt.Sprintf("File not found: %s", requestedPath),
            }
        }
    }
    
    return nil
}
```

**Integrate into tools**:
```go
// Updated ReadFileInput with base path
type ReadFileInput struct {
    Path     string `json:"path"`
    BasePath *string `json:"base_path,omitempty"` // Optional security boundary
}

// In handler:
if input.BasePath != nil {
    if err := ValidateFilePath(*input.BasePath, input.Path, true); err != nil {
        return ReadFileOutput{
            Success: false,
            Error: err.(*PathSecurityError).Message,
        }
    }
}
```

---

### Phase 2: Important Enhancements (Week 3-4)

#### 2.1 Atomic File Operations

```go
// atomic_write.go
package tools

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
)

// AtomicWrite performs a safe, atomic file write
func AtomicWrite(path string, content []byte, perm os.FileMode) error {
    // 1. Create temp file in same directory
    dir := filepath.Dir(path)
    tmpFile, err := ioutil.TempFile(dir, ".tmp-")
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    defer os.Remove(tmpFile.Name())
    
    // 2. Write content
    if _, err := tmpFile.Write(content); err != nil {
        tmpFile.Close()
        return fmt.Errorf("failed to write temp file: %w", err)
    }
    
    // 3. Set permissions
    if err := os.Chmod(tmpFile.Name(), perm); err != nil {
        tmpFile.Close()
        return fmt.Errorf("failed to set permissions: %w", err)
    }
    
    // 4. Sync to disk
    if err := tmpFile.Sync(); err != nil {
        tmpFile.Close()
        return fmt.Errorf("failed to sync: %w", err)
    }
    tmpFile.Close()
    
    // 5. Atomic rename
    if err := os.Rename(tmpFile.Name(), path); err != nil {
        return fmt.Errorf("failed to rename: %w", err)
    }
    
    return nil
}
```

**Enhance WriteFileTool**:
```go
type WriteFileInput struct {
    Path       string `json:"path"`
    Content    string `json:"content"`
    CreateDirs *bool  `json:"create_dirs,omitempty"`
    Mode       *int   `json:"mode,omitempty"`        // Unix permissions (e.g., 0644)
    Atomic     *bool  `json:"atomic,omitempty"`      // Use atomic write (default: true)
    Overwrite  *bool  `json:"overwrite,omitempty"`   // Require file not to exist (default: true)
}

// In handler:
if input.Atomic == nil || *input.Atomic {
    err = AtomicWrite(input.Path, []byte(input.Content), perm)
} else {
    err = os.WriteFile(input.Path, []byte(input.Content), perm)
}
```

---

#### 2.2 Enhanced Error Handling

```go
// error_types.go
package tools

// ErrorCode represents a structured error type
type ErrorCode string

const (
    ErrorCodeFileNotFound      ErrorCode = "FILE_NOT_FOUND"
    ErrorCodePermissionDenied  ErrorCode = "PERMISSION_DENIED"
    ErrorCodePathTraversal     ErrorCode = "PATH_TRAVERSAL"
    ErrorCodeInvalidInput      ErrorCode = "INVALID_INPUT"
    ErrorCodeOperationFailed   ErrorCode = "OPERATION_FAILED"
    ErrorCodePatchFailed       ErrorCode = "PATCH_FAILED"
)

// ToolError represents a structured error with suggestions
type ToolError struct {
    Code       ErrorCode `json:"code"`
    Message    string    `json:"message"`
    Suggestion string    `json:"suggestion,omitempty"`
    Details    map[string]interface{} `json:"details,omitempty"`
}

// Error implements error interface
func (e *ToolError) Error() string {
    return e.Message
}

// Example usage
func readFileWithError(path string) (string, *ToolError) {
    content, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return "", &ToolError{
                Code:    ErrorCodeFileNotFound,
                Message: fmt.Sprintf("File not found: %s", path),
                Suggestion: fmt.Sprintf("Check the path is correct. Current: %s", path),
            }
        }
        if os.IsPermission(err) {
            return "", &ToolError{
                Code:    ErrorCodePermissionDenied,
                Message: fmt.Sprintf("Permission denied: %s", path),
                Suggestion: "Check file permissions with 'ls -la'",
            }
        }
    }
    return string(content), nil
}
```

---

#### 2.3 Diff Generation and Preview

```go
// diff_tools.go
package tools

import (
    "fmt"
    "strings"
)

// DiffLine represents a single line in a diff
type DiffLine struct {
    Type    string // "+", "-", " " (context)
    LineNum int    // Line number in result
    Content string
}

// GenerateDiff creates a unified diff between two strings
func GenerateDiff(original, modified string, contextLines int) string {
    origLines := strings.Split(original, "\n")
    modLines := strings.Split(modified, "\n")
    
    // Simple implementation - for production, use a proper diff library
    // e.g., "github.com/sergi/go-diff/diffmatchpatch"
    
    var diff strings.Builder
    diff.WriteString("--- original\n")
    diff.WriteString("+++ modified\n")
    
    // TODO: Implement proper LCS-based diff
    
    return diff.String()
}

// PreviewReplaceInput for preview tool
type PreviewReplaceInput struct {
    FilePath  string `json:"file_path"`
    OldText   string `json:"old_text"`
    NewText   string `json:"new_text"`
    Context   *int   `json:"context,omitempty"`  // Lines of context (default: 3)
}

type PreviewReplaceOutput struct {
    Success bool   `json:"success"`
    Diff    string `json:"diff"`           // Unified diff preview
    Changes int    `json:"changes"`        // Number of changes
    Error   string `json:"error,omitempty"`
}

// NewPreviewReplaceTool creates a tool to preview changes
func NewPreviewReplaceTool() (tool.Tool, error) {
    handler := func(ctx tool.Context, input PreviewReplaceInput) PreviewReplaceOutput {
        content, err := os.ReadFile(input.FilePath)
        if err != nil {
            return PreviewReplaceOutput{
                Success: false,
                Error: fmt.Sprintf("Failed to read file: %v", err),
            }
        }
        
        original := string(content)
        modified := strings.ReplaceAll(original, input.OldText, input.NewText)
        changes := strings.Count(original, input.OldText)
        
        if changes == 0 {
            return PreviewReplaceOutput{
                Success: false,
                Error: "No matches found",
            }
        }
        
        contextLines := 3
        if input.Context != nil {
            contextLines = *input.Context
        }
        
        diff := GenerateDiff(original, modified, contextLines)
        
        return PreviewReplaceOutput{
            Success: true,
            Diff: diff,
            Changes: changes,
        }
    }
    
    return functiontool.New(functiontool.Config{
        Name: "preview_replace_in_file",
        Description: "Preview changes before applying replace operation",
    }, handler)
}
```

---

### Phase 3: Advanced Features (Week 5+)

#### 3.1 Hook System for Tool Execution

```go
// hooks.go
package tools

import (
    "context"
)

// ToolHook defines callbacks for tool execution lifecycle
type ToolHook interface {
    // BeforeExecute is called before a tool executes
    BeforeExecute(ctx context.Context, toolName string, input interface{}) error
    
    // AfterExecute is called after a tool executes successfully
    AfterExecute(ctx context.Context, toolName string, output interface{}) error
    
    // OnError is called when a tool execution fails
    OnError(ctx context.Context, toolName string, err error) error
}

// ToolExecutor wraps tool execution with hooks
type ToolExecutor struct {
    hooks []ToolHook
}

// Execute runs a tool with all registered hooks
func (e *ToolExecutor) Execute(ctx context.Context, toolName string, input interface{}, 
    handler func() (interface{}, error)) (interface{}, error) {
    
    // Before hooks
    for _, hook := range e.hooks {
        if err := hook.BeforeExecute(ctx, toolName, input); err != nil {
            return nil, err
        }
    }
    
    // Execute
    output, err := handler()
    if err != nil {
        // Error hooks
        for _, hook := range e.hooks {
            hook.OnError(ctx, toolName, err)
        }
        return nil, err
    }
    
    // After hooks
    for _, hook := range e.hooks {
        if err := hook.AfterExecute(ctx, toolName, output); err != nil {
            return nil, err
        }
    }
    
    return output, nil
}
```

---

## 2. Testing Strategy

### 2.1 Unit Tests for Each Tool

```go
// file_tools_test.go
package tools

import (
    "os"
    "path/filepath"
    "testing"
)

func TestReadFileWithLineRange(t *testing.T) {
    // Create temp file
    tmpFile := filepath.Join(t.TempDir(), "test.txt")
    content := "line1\nline2\nline3\nline4\nline5"
    os.WriteFile(tmpFile, []byte(content), 0644)
    
    tests := []struct {
        name   string
        offset *int
        limit  *int
        want   string
    }{
        {"All lines", nil, nil, content},
        {"Lines 2-4", intPtr(2), intPtr(3), "line2\nline3\nline4"},
        {"From line 3", intPtr(3), nil, "line3\nline4\nline5"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            input := ReadFileInput{
                Path: tmpFile,
                Offset: tt.offset,
                Limit: tt.limit,
            }
            // Test implementation
        })
    }
}

func intPtr(i int) *int { return &i }
```

---

### 2.2 Integration Tests

Test complete workflows:
- Create file → Write content → Read with range → Replace → Verify
- Test patch application with complex scenarios
- Test atomicity and rollback

---

## 3. Migration Path

### 3.1 Backward Compatibility

Keep existing tool signatures, add new tools:
- Existing: `read_file`, `write_file`, `replace_in_file`
- New: `read_file_with_range`, `apply_patch`, `preview_replace_in_file`

### 3.2 Gradual Deprecation

```go
const DeprecationWarning = "replace_in_file is deprecated. Use apply_patch for more robust edits."

type ReplaceInFileOutput struct {
    // ... existing fields ...
    Deprecated string `json:"deprecated,omitempty"`  // Add deprecation notice
}
```

---

## 4. Performance Considerations

### 4.1 Large File Handling

For files > 100MB:
- Stream reading/writing
- Chunk-based processing
- Memory-mapped files

```go
type ReadFileInput struct {
    Path      string `json:"path"`
    StreamMode *bool `json:"stream_mode,omitempty"`  // For large files
    ChunkSize *int   `json:"chunk_size,omitempty"`   // Bytes per chunk
}
```

### 4.2 Search Optimization

- Use parallel searching for large directories
- Index frequently accessed paths
- Cache directory listings

---

## 5. Documentation Templates

### Tool Documentation Format

```markdown
## Tool Name: `apply_patch`

### Purpose
Apply unified diff patches to files robustly

### Parameters
| Name | Type | Required | Description |
|------|------|----------|-------------|
| file_path | string | Yes | Path to file to patch |
| patch | string | Yes | Unified diff format |
| dry_run | bool | No | Preview without applying |

### Returns
| Field | Type | Description |
|-------|------|-------------|
| success | bool | Operation success |
| lines_added | int | Number of lines added |
| lines_removed | int | Number of lines removed |
| error | string | Error message if failed |

### Examples
[Concrete examples...]

### Error Codes
| Code | Description |
|------|-------------|
| PATCH_FAILED | Patch didn't apply cleanly |
| FILE_NOT_FOUND | Target file not found |
```

---

## 6. Success Metrics

### KPI's for Implementation

1. **Robustness**: Reduce failed edits from 5% to <0.5%
2. **Performance**: Read/write speed maintains <100ms for typical files
3. **Usability**: Support >95% of Cline tool use cases
4. **Coverage**: 100% unit test coverage for core logic

---

## Conclusion

This implementation guide provides a practical roadmap for enhancing the ADK Code Agent tools. Prioritize Phase 1 for maximum impact, then incrementally add Phase 2 and 3 features based on actual usage patterns and feedback.

