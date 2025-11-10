// Package tools provides file operation tools for the coding agent.
package tools

import (
	"fmt"
	"os"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// ApplyPatchInput defines the input parameters for applying a patch.
type ApplyPatchInput struct {
	// FilePath is the path to the file to patch
	FilePath string `json:"file_path" jsonschema:"Path to the file to patch"`
	// Patch is the patch in unified diff format (RFC 3881)
	Patch string `json:"patch" jsonschema:"Unified diff format patch"`
	// DryRun indicates whether to preview changes without applying them
	DryRun *bool `json:"dry_run,omitempty" jsonschema:"Preview mode - don't apply changes (default: false)"`
	// Strict indicates whether to require exact match (fail if patch doesn't apply cleanly)
	Strict *bool `json:"strict,omitempty" jsonschema:"Require exact match (default: true)"`
}

// ApplyPatchOutput defines the output of applying a patch.
type ApplyPatchOutput struct {
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`
	// Message is a description of the result
	Message string `json:"message,omitempty"`
	// LinesAdded is the number of lines added
	LinesAdded int `json:"lines_added"`
	// LinesRemoved is the number of lines removed
	LinesRemoved int `json:"lines_removed"`
	// Preview is the patched content (only in dry-run mode)
	Preview string `json:"preview,omitempty"`
	// Error is the error message if the operation failed
	Error string `json:"error,omitempty"`
}

// PatchHunk represents a single hunk in a unified diff
type PatchHunk struct {
	OrigStart     int      // Starting line in original file
	OrigCount     int      // Number of lines in original
	NewStart      int      // Starting line in new file
	NewCount      int      // Number of lines in new
	Lines         []string // Hunk lines (with +/- prefixes)
	ContextBefore []string // Context lines before the hunk
	ContextAfter  []string // Context lines after the hunk
}

// ParseUnifiedDiff parses a unified diff format patch
func ParseUnifiedDiff(patch string) ([]PatchHunk, error) {
	var hunks []PatchHunk
	lines := strings.Split(patch, "\n")

	var currentHunk *PatchHunk

	for _, line := range lines {
		// Skip file headers (---/+++)
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			continue
		}

		// Check for hunk header (@@ -X,Y +A,B @@)
		if strings.HasPrefix(line, "@@") {
			// Save previous hunk if exists
			if currentHunk != nil {
				hunks = append(hunks, *currentHunk)
			}

			// Parse hunk header
			hunk, err := parseHunkHeader(line)
			if err != nil {
				return nil, err
			}
			currentHunk = hunk
			continue
		}

		// Add lines to current hunk
		if currentHunk != nil {
			// Trim the prefix character (+ or -)
			if len(line) > 0 {
				currentHunk.Lines = append(currentHunk.Lines, line)
			}
		}
	}

	// Add last hunk
	if currentHunk != nil {
		hunks = append(hunks, *currentHunk)
	}

	return hunks, nil
}

// parseHunkHeader parses a hunk header line like "@@ -10,5 +12,7 @@"
func parseHunkHeader(header string) (*PatchHunk, error) {
	// Format: @@ -origStart,origCount +newStart,newCount @@
	var origStart, origCount, newStart, newCount int

	_, err := fmt.Sscanf(header, "@@ -%d,%d +%d,%d @@", &origStart, &origCount, &newStart, &newCount)
	if err != nil {
		// Try format without counts (single line)
		_, err = fmt.Sscanf(header, "@@ -%d +%d @@", &origStart, &newStart)
		if err != nil {
			return nil, fmt.Errorf("invalid hunk header: %s", header)
		}
		origCount = 1
		newCount = 1
	}

	return &PatchHunk{
		OrigStart: origStart,
		OrigCount: origCount,
		NewStart:  newStart,
		NewCount:  newCount,
		Lines:     make([]string, 0),
	}, nil
}

// ApplyPatch applies a unified diff patch to file content
func ApplyPatch(originalContent string, patch string, strict bool) (string, int, int, error) {
	hunks, err := ParseUnifiedDiff(patch)
	if err != nil {
		return "", 0, 0, PatchFailedError(err.Error())
	}

	if len(hunks) == 0 {
		return originalContent, 0, 0, nil
	}

	lines := strings.Split(originalContent, "\n")
	var result []string
	var totalAdded, totalRemoved int
	lineOffset := 0 // Track offset due to previous hunks

	for _, hunk := range hunks {
		// Calculate actual line numbers accounting for offset
		origStart := hunk.OrigStart - 1 + lineOffset // Convert to 0-indexed
		if origStart < 0 {
			origStart = 0
		}

		// Verify context matches (if in strict mode)
		if strict {
			// Check lines before the change
			contextIdx := origStart
			for _, contextLine := range hunk.ContextBefore {
				if contextIdx < len(lines) {
					if !lineMatches(contextLine, lines[contextIdx]) {
						return "", 0, 0, PatchFailedError("context mismatch before hunk")
					}
				}
				contextIdx++
			}
		}

		// Apply the hunk
		var hunkAdded, hunkRemoved int
		newLines, added, removed, err := applyHunk(lines, hunk, origStart)
		if err != nil {
			return "", 0, 0, PatchFailedError(err.Error())
		}

		lines = newLines
		hunkAdded = added
		hunkRemoved = removed
		lineOffset += (hunkAdded - hunkRemoved)
		totalAdded += hunkAdded
		totalRemoved += hunkRemoved
	}

	result = lines
	return strings.Join(result, "\n"), totalAdded, totalRemoved, nil
}

// applyHunk applies a single hunk to the file content
func applyHunk(lines []string, hunk PatchHunk, startIdx int) ([]string, int, int, error) {
	var result []string
	var added, removed int

	// Copy lines before the hunk
	result = append(result, lines[:startIdx]...)

	// Process hunk lines
	origIdx := startIdx
	for _, line := range hunk.Lines {
		if len(line) == 0 {
			continue
		}

		prefix := line[0]
		content := line[1:]

		switch prefix {
		case ' ': // Context line
			if origIdx < len(lines) {
				result = append(result, content)
				origIdx++
			}
		case '-': // Remove line
			if origIdx < len(lines) {
				origIdx++
				removed++
			}
		case '+': // Add line
			result = append(result, content)
			added++
		case '\\': // No newline marker, skip
			continue
		}
	}

	// Copy remaining lines
	if origIdx < len(lines) {
		result = append(result, lines[origIdx:]...)
	}

	return result, added, removed, nil
}

// lineMatches checks if a context line matches, ignoring prefix
func lineMatches(contextLine string, fileLine string) bool {
	// Remove the context marker if present
	expectedLine := contextLine
	if len(contextLine) > 0 && contextLine[0] == ' ' {
		expectedLine = contextLine[1:]
	}
	return expectedLine == fileLine
}

// NewApplyPatchTool creates a tool for applying unified diff patches.
func NewApplyPatchTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input ApplyPatchInput) ApplyPatchOutput {
		// Read the file
		content, err := os.ReadFile(input.FilePath)
		if err != nil {
			return ApplyPatchOutput{
				Success: false,
				Error:   fmt.Sprintf("Failed to read file: %v", err),
			}
		}

		originalContent := string(content)

		// Handle dry-run and strict options
		dryRun := false
		if input.DryRun != nil {
			dryRun = *input.DryRun
		}

		strict := true
		if input.Strict != nil {
			strict = *input.Strict
		}

		// Apply patch
		patchedContent, linesAdded, linesRemoved, err := ApplyPatch(originalContent, input.Patch, strict)
		if err != nil {
			toolErr := err.(*ToolError)
			return ApplyPatchOutput{
				Success: false,
				Error:   toolErr.Message,
			}
		}

		// In dry-run mode, return preview without writing
		if dryRun {
			return ApplyPatchOutput{
				Success:      true,
				Message:      "Patch preview (not applied)",
				LinesAdded:   linesAdded,
				LinesRemoved: linesRemoved,
				Preview:      patchedContent,
			}
		}

		// Write the patched content back to file
		if err := AtomicWrite(input.FilePath, []byte(patchedContent), 0644); err != nil {
			return ApplyPatchOutput{
				Success: false,
				Error:   fmt.Sprintf("Failed to write patched file: %v", err),
			}
		}

		return ApplyPatchOutput{
			Success:      true,
			Message:      fmt.Sprintf("Successfully applied patch: %d lines added, %d lines removed", linesAdded, linesRemoved),
			LinesAdded:   linesAdded,
			LinesRemoved: linesRemoved,
		}
	}

	return functiontool.New(functiontool.Config{
		Name:        "apply_patch",
		Description: "Applies a unified diff format patch to a file. Supports dry-run preview and strict mode. More robust than string replacement for targeted edits.",
	}, handler)
}
