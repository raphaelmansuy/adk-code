// Package v4a provides V4A patch parsing and application tools for the coding agent.
package v4a

import (
	"fmt"
	"strings"
)

// ParseV4APatch parses a V4A format patch string into a structured V4APatch.
//
// V4A format structure:
//
//	*** Update File: <filepath>          (optional header)
//	@@ <context1>                         (context marker, e.g., class name)
//	@@     <context2>                     (nested context, indented)
//	-<line_to_remove>                     (removal, can be multiple)
//	+<line_to_add>                        (addition, can be multiple)
//
// Example:
//
//	*** Update File: src/handler.go
//	@@ func ProcessRequest
//	-    return nil
//	+    return processData(req)
//
// Returns error if the patch format is invalid.
func ParseV4APatch(patchText string) (*V4APatch, error) {
	if strings.TrimSpace(patchText) == "" {
		return nil, fmt.Errorf("empty patch text")
	}

	lines := strings.Split(patchText, "\n")
	patch := &V4APatch{}
	var currentHunk *V4AHunk

	for i, line := range lines {
		// Handle file path header
		if strings.HasPrefix(line, "*** Update File:") {
			filePath := strings.TrimSpace(strings.TrimPrefix(line, "*** Update File:"))
			if filePath == "" {
				return nil, fmt.Errorf("line %d: empty file path after '*** Update File:'", i+1)
			}
			patch.FilePath = filePath
			continue
		}

		// Handle context markers (@@)
		if strings.HasPrefix(line, "@@") {
			// Initialize hunk if needed
			if currentHunk == nil {
				currentHunk = &V4AHunk{
					ContextMarkers: []string{},
					Removals:       []string{},
					Additions:      []string{},
				}
			}

			// Extract context marker and indentation
			markerContent := strings.TrimPrefix(line, "@@")
			indentation := countLeadingSpaces(markerContent)
			marker := strings.TrimSpace(markerContent)

			if marker == "" {
				return nil, fmt.Errorf("line %d: empty context marker after '@@'", i+1)
			}

			currentHunk.ContextMarkers = append(currentHunk.ContextMarkers, marker)
			// Track the deepest (most indented) context marker's indentation
			if indentation > currentHunk.BaseIndentation {
				currentHunk.BaseIndentation = indentation
			}
			continue
		}

		// Handle removal lines (-)
		if strings.HasPrefix(line, "-") {
			if currentHunk == nil {
				return nil, fmt.Errorf("line %d: removal line before context marker", i+1)
			}
			content := strings.TrimPrefix(line, "-")
			currentHunk.Removals = append(currentHunk.Removals, content)
			continue
		}

		// Handle addition lines (+)
		if strings.HasPrefix(line, "+") {
			if currentHunk == nil {
				return nil, fmt.Errorf("line %d: addition line before context marker", i+1)
			}
			content := strings.TrimPrefix(line, "+")
			currentHunk.Additions = append(currentHunk.Additions, content)
			continue
		}

		// Handle blank lines (end of hunk)
		if strings.TrimSpace(line) == "" {
			if currentHunk != nil && len(currentHunk.ContextMarkers) > 0 {
				// Validate hunk has at least removals or additions
				if len(currentHunk.Removals) == 0 && len(currentHunk.Additions) == 0 {
					return nil, fmt.Errorf("hunk with context %v has no changes", currentHunk.ContextMarkers)
				}
				patch.Hunks = append(patch.Hunks, *currentHunk)
				currentHunk = nil
			}
			continue
		}

		// Ignore other lines (could be comments or context lines)
	}

	// Add the last hunk if present
	if currentHunk != nil && len(currentHunk.ContextMarkers) > 0 {
		if len(currentHunk.Removals) == 0 && len(currentHunk.Additions) == 0 {
			return nil, fmt.Errorf("hunk with context %v has no changes", currentHunk.ContextMarkers)
		}
		patch.Hunks = append(patch.Hunks, *currentHunk)
	}

	// Validate we have at least one hunk
	if len(patch.Hunks) == 0 {
		return nil, fmt.Errorf("no valid hunks found in patch")
	}

	return patch, nil
}

// countLeadingSpaces counts the number of leading spaces in a string.
func countLeadingSpaces(s string) int {
	count := 0
	for _, ch := range s {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4 // Treat tab as 4 spaces
		} else {
			break
		}
	}
	return count
}
