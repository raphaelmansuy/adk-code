// Package edit provides code editing tools for the coding agent.
package edit

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"code_agent/tools/base"
	"code_agent/tools/file"
)

// SearchReplaceBlock represents a single SEARCH/REPLACE operation
type SearchReplaceBlock struct {
	SearchContent  string
	ReplaceContent string
	MatchIndex     int // Where in the file this block matched (-1 if not matched)
}

// SearchReplaceInput defines input for SEARCH/REPLACE block-based editing
type SearchReplaceInput struct {
	// Path to the file to modify
	Path string `json:"path" jsonschema:"Path to the file to modify (relative to working directory)"`
	// Diff containing one or more SEARCH/REPLACE blocks
	Diff string `json:"diff" jsonschema:"One or more SEARCH/REPLACE blocks in the specified format"`
	// Preview mode - show what would change without applying
	Preview *bool `json:"preview,omitempty" jsonschema:"Preview changes without applying (default: false)"`
}

// SearchReplaceOutput defines output of SEARCH/REPLACE operation
type SearchReplaceOutput struct {
	Success        bool     `json:"success"`
	BlocksApplied  int      `json:"blocks_applied"`
	TotalBlocks    int      `json:"total_blocks"`
	PreviewContent string   `json:"preview_content,omitempty"`
	Message        string   `json:"message,omitempty"`
	Error          string   `json:"error,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
}

// Block marker patterns (flexible regex, inspired by Cline)
var (
	searchBlockStartRegex  = regexp.MustCompile(`^[-]{3,} SEARCH>?\s*$`)
	searchBlockEndRegex    = regexp.MustCompile(`^[=]{3,}\s*$`)
	replaceBlockEndRegex   = regexp.MustCompile(`^[+]{3,} REPLACE>?\s*$`)
	legacySearchStartRegex = regexp.MustCompile(`^[<]{3,} SEARCH>?\s*$`)
	legacyReplaceEndRegex  = regexp.MustCompile(`^[>]{3,} REPLACE>?\s*$`)
)

// isSearchBlockStart checks if a line is a search block start marker
func isSearchBlockStart(line string) bool {
	return searchBlockStartRegex.MatchString(line) || legacySearchStartRegex.MatchString(line)
}

// isSearchBlockEnd checks if a line is a search block end marker
func isSearchBlockEnd(line string) bool {
	return searchBlockEndRegex.MatchString(line)
}

// isReplaceBlockEnd checks if a line is a replace block end marker
func isReplaceBlockEnd(line string) bool {
	return replaceBlockEndRegex.MatchString(line) || legacyReplaceEndRegex.MatchString(line)
}

// ParseSearchReplaceBlocks parses SEARCH/REPLACE blocks from diff string
func ParseSearchReplaceBlocks(diff string) ([]SearchReplaceBlock, error) {
	lines := strings.Split(diff, "\n")
	var blocks []SearchReplaceBlock
	var currentBlock *SearchReplaceBlock
	state := "idle" // idle, in_search, in_replace

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		switch state {
		case "idle":
			if isSearchBlockStart(trimmedLine) {
				currentBlock = &SearchReplaceBlock{MatchIndex: -1}
				state = "in_search"
			}

		case "in_search":
			if isSearchBlockEnd(trimmedLine) {
				state = "in_replace"
			} else {
				if currentBlock.SearchContent != "" {
					currentBlock.SearchContent += "\n"
				}
				currentBlock.SearchContent += line // Keep original line with whitespace
			}

		case "in_replace":
			if isReplaceBlockEnd(trimmedLine) {
				// Block complete
				if currentBlock.SearchContent == "" {
					return nil, fmt.Errorf("empty SEARCH block at line %d", i+1)
				}
				blocks = append(blocks, *currentBlock)
				currentBlock = nil
				state = "idle"
			} else {
				if currentBlock.ReplaceContent != "" {
					currentBlock.ReplaceContent += "\n"
				}
				currentBlock.ReplaceContent += line // Keep original line with whitespace
			}
		}
	}

	if state != "idle" {
		return nil, fmt.Errorf("incomplete SEARCH/REPLACE block (state: %s)", state)
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no valid SEARCH/REPLACE blocks found")
	}

	return blocks, nil
}

// findExactMatch finds exact string match in content starting from offset
func findExactMatch(content, search string, startOffset int) int {
	return strings.Index(content[startOffset:], search)
}

// lineTrimmedMatch finds match with line-trimmed whitespace tolerance
// Inspired by Cline's lineTrimmedFallbackMatch
func lineTrimmedMatch(content, search string, startOffset int) int {
	contentLines := strings.Split(content[startOffset:], "\n")
	searchLines := strings.Split(search, "\n")

	// Trim trailing empty line if exists
	if len(searchLines) > 0 && searchLines[len(searchLines)-1] == "" {
		searchLines = searchLines[:len(searchLines)-1]
	}

	// Try to match search lines at each position in content
	for i := 0; i <= len(contentLines)-len(searchLines); i++ {
		matches := true
		for j := 0; j < len(searchLines); j++ {
			contentTrimmed := strings.TrimSpace(contentLines[i+j])
			searchTrimmed := strings.TrimSpace(searchLines[j])
			if contentTrimmed != searchTrimmed {
				matches = false
				break
			}
		}

		if matches {
			// Calculate character offset
			charOffset := 0
			for k := 0; k < i; k++ {
				charOffset += len(contentLines[k]) + 1 // +1 for \n
			}
			return charOffset
		}
	}

	return -1
}

// ApplySearchReplaceBlocks applies SEARCH/REPLACE blocks to file content
func ApplySearchReplaceBlocks(content string, blocks []SearchReplaceBlock) (string, []SearchReplaceBlock, error) {
	result := content
	appliedBlocks := []SearchReplaceBlock{}
	currentOffset := 0

	for blockNum, block := range blocks {
		// Try exact match first
		matchIdx := findExactMatch(result, block.SearchContent, currentOffset)

		// Fall back to line-trimmed match if exact fails
		if matchIdx == -1 {
			matchIdx = lineTrimmedMatch(result, block.SearchContent, currentOffset)
		}

		if matchIdx == -1 {
			return "", appliedBlocks, fmt.Errorf(
				"block %d: SEARCH content not found after offset %d\n"+
					"SEARCH content:\n%s",
				blockNum+1, currentOffset, block.SearchContent,
			)
		}

		// Adjust for current offset
		absoluteMatchIdx := currentOffset + matchIdx

		// Apply replacement
		before := result[:absoluteMatchIdx]
		after := result[absoluteMatchIdx+len(block.SearchContent):]
		result = before + block.ReplaceContent + after

		// Update offset for next block
		// Move past the replaced content
		currentOffset = absoluteMatchIdx + len(block.ReplaceContent)

		// Track applied block
		block.MatchIndex = absoluteMatchIdx
		appliedBlocks = append(appliedBlocks, block)
	}

	return result, appliedBlocks, nil
}

// NewSearchReplaceTool creates a tool for SEARCH/REPLACE block-based editing
func NewSearchReplaceTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input SearchReplaceInput) SearchReplaceOutput {
		// Validate input
		if input.Path == "" {
			return SearchReplaceOutput{
				Success: false,
				Error:   "Path is required",
			}
		}

		if input.Diff == "" {
			return SearchReplaceOutput{
				Success: false,
				Error:   "Diff is required",
			}
		}

		// Parse SEARCH/REPLACE blocks
		blocks, err := ParseSearchReplaceBlocks(input.Diff)
		if err != nil {
			return SearchReplaceOutput{
				Success:     false,
				TotalBlocks: 0,
				Error:       fmt.Sprintf("Failed to parse SEARCH/REPLACE blocks: %v", err),
			}
		}

		// Read file content
		content, err := os.ReadFile(input.Path)
		if err != nil {
			return SearchReplaceOutput{
				Success:     false,
				TotalBlocks: len(blocks),
				Error:       fmt.Sprintf("Failed to read file: %v", err),
			}
		}

		originalContent := string(content)

		// Apply blocks
		newContent, appliedBlocks, err := ApplySearchReplaceBlocks(originalContent, blocks)
		if err != nil {
			return SearchReplaceOutput{
				Success:       false,
				TotalBlocks:   len(blocks),
				BlocksApplied: len(appliedBlocks),
				Error:         fmt.Sprintf("Failed to apply blocks: %v", err),
			}
		}

		// Preview mode - just return what would change
		preview := false
		if input.Preview != nil {
			preview = *input.Preview
		}

		if preview {
			// Generate preview showing changes
			previewLines := []string{
				fmt.Sprintf("Would apply %d SEARCH/REPLACE blocks to %s:", len(blocks), input.Path),
				"",
			}
			for i, block := range appliedBlocks {
				previewLines = append(previewLines,
					fmt.Sprintf("Block %d (match at offset %d):", i+1, block.MatchIndex),
					"------- SEARCH",
					block.SearchContent,
					"=======",
					block.ReplaceContent,
					"+++++++ REPLACE",
					"",
				)
			}

			return SearchReplaceOutput{
				Success:        true,
				TotalBlocks:    len(blocks),
				BlocksApplied:  len(appliedBlocks),
				PreviewContent: strings.Join(previewLines, "\n"),
				Message:        fmt.Sprintf("Preview: %d blocks would be applied", len(blocks)),
			}
		}

		// Actually write the file
		err = file.AtomicWrite(input.Path, []byte(newContent), 0644)
		if err != nil {
			return SearchReplaceOutput{
				Success:       false,
				TotalBlocks:   len(blocks),
				BlocksApplied: len(appliedBlocks),
				Error:         fmt.Sprintf("Failed to write file: %v", err),
			}
		}

		return SearchReplaceOutput{
			Success:       true,
			TotalBlocks:   len(blocks),
			BlocksApplied: len(appliedBlocks),
			Message: fmt.Sprintf(
				"Successfully applied %d SEARCH/REPLACE block(s) to %s",
				len(appliedBlocks), input.Path,
			),
		}
	}

	t, err := functiontool.New(functiontool.Config{
		Name: "search_replace",
		Description: `Request to replace sections of content in an existing file using SEARCH/REPLACE blocks. 
This is the PREFERRED tool for making targeted changes to specific parts of a file. 
Use this tool when you need to modify, add, or delete code in precise locations.

Format:
` + "```" + `
------- SEARCH
[exact content to find]
=======
[new content to replace with]
+++++++ REPLACE
` + "```" + `

Critical Rules:
1. SEARCH content must match EXACTLY (including whitespace, indentation)
2. Each SEARCH/REPLACE block replaces ONLY the first match
3. Use multiple blocks for multiple changes (list in file order)
4. Keep blocks concise - just the changing lines + a few context lines
5. To delete code: use empty REPLACE section
6. To move code: use two blocks (one to delete, one to insert)

Example (adding error handling):
` + "```" + `
------- SEARCH
func processData(data string) {
    result := transform(data)
    return result
}
=======
func processData(data string) error {
    if data == "" {
        return errors.New("empty data")
    }
    result := transform(data)
    return result
}
+++++++ REPLACE
` + "```" + `

The tool uses whitespace-tolerant matching as a fallback, so minor indentation differences are handled gracefully.`,
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategoryCodeEditing,
			Priority:  0,
			UsageHint: "PREFERRED for targeted edits, supports multiple blocks, whitespace-tolerant",
		})
	}

	return t, err
}
