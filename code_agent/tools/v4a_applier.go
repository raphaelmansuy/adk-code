package tools

import (
	"fmt"
	"os"
	"strings"
)

// ApplyV4APatch applies a parsed V4A patch to a file.
//
// The algorithm:
//  1. Read the target file
//  2. For each hunk:
//     a. Find location using context markers (search for class/function names)
//     b. Find and match the removal lines
//     c. Replace removals with additions
//  3. Write back atomically
//
// Returns error if context is not found, removals don't match, or file I/O fails.
func ApplyV4APatch(filePath string, patch *V4APatch, dryRun bool) (string, error) {
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	originalLineCount := len(lines)

	// Apply each hunk
	for hunkIdx, hunk := range patch.Hunks {
		// Find the context location
		contextLine, err := findContextLocation(lines, hunk.ContextMarkers)
		if err != nil {
			return "", fmt.Errorf("hunk %d: %w", hunkIdx+1, err)
		}

		// Find the exact removal range
		removeStart, removeEnd, err := findRemovalRange(lines, contextLine, hunk.Removals)
		if err != nil {
			return "", fmt.Errorf("hunk %d: %w", hunkIdx+1, err)
		}

		// Build new lines with replacements
		newLines := make([]string, 0, len(lines)-len(hunk.Removals)+len(hunk.Additions))
		newLines = append(newLines, lines[:removeStart]...)
		newLines = append(newLines, hunk.Additions...)
		newLines = append(newLines, lines[removeEnd+1:]...)

		lines = newLines
	}

	newContent := strings.Join(lines, "\n")

	// Dry run: return preview
	if dryRun {
		preview := "=== DRY RUN ===\n"
		preview += fmt.Sprintf("File: %s\n", filePath)
		preview += fmt.Sprintf("Original lines: %d\n", originalLineCount)
		preview += fmt.Sprintf("Modified lines: %d\n", len(lines))
		preview += fmt.Sprintf("Hunks applied: %d\n", len(patch.Hunks))
		preview += fmt.Sprintf("\n=== NEW CONTENT ===\n%s\n", newContent)
		return preview, nil
	}

	// Write back atomically
	if err := AtomicWrite(filePath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	result := fmt.Sprintf("Successfully applied %d hunk(s) to %s", len(patch.Hunks), filePath)
	return result, nil
}

// findContextLocation searches for context markers in sequence and returns the line number
// of the deepest (last) context marker found.
//
// Example: markers = ["class User", "def validate"]
// Finds "class User" first, then searches for "def validate" after that line.
func findContextLocation(lines []string, markers []string) (int, error) {
	if len(markers) == 0 {
		return -1, fmt.Errorf("no context markers provided")
	}

	currentLine := 0

	for _, marker := range markers {
		found := false
		for i := currentLine; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			// Check if line contains the marker
			if strings.Contains(line, marker) {
				currentLine = i
				found = true
				break
			}
		}
		if !found {
			return -1, fmt.Errorf("context marker not found: %q", marker)
		}
	}

	return currentLine, nil
}

// findRemovalRange finds the exact lines to remove starting from contextLine.
// Returns the start and end line indices (inclusive) of the removal block.
func findRemovalRange(lines []string, contextLine int, removals []string) (int, int, error) {
	if len(removals) == 0 {
		// No removals, this is an insertion-only hunk
		// Insert after the context line
		return contextLine + 1, contextLine, nil
	}

	// Search for the removal block starting from context line
	// Allow some flexibility in whitespace
	searchStart := contextLine
	if searchStart >= len(lines) {
		return -1, -1, fmt.Errorf("context line %d is beyond file length %d", searchStart, len(lines))
	}

	// Try to find the removal block
	for startLine := searchStart; startLine < len(lines) && startLine < searchStart+50; startLine++ {
		// Check if removals match starting at this line
		if matchesRemovalBlock(lines, startLine, removals) {
			endLine := startLine + len(removals) - 1
			return startLine, endLine, nil
		}
	}

	return -1, -1, fmt.Errorf("removal lines not found near context (searched lines %d-%d)", searchStart, min(searchStart+50, len(lines)))
}

// matchesRemovalBlock checks if the removal lines match at the given start position.
// Uses whitespace-tolerant matching (trims both sides).
func matchesRemovalBlock(lines []string, startLine int, removals []string) bool {
	if startLine+len(removals) > len(lines) {
		return false
	}

	for i, removal := range removals {
		fileLine := lines[startLine+i]
		// Whitespace-tolerant comparison
		if strings.TrimSpace(fileLine) != strings.TrimSpace(removal) {
			return false
		}
	}

	return true
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
