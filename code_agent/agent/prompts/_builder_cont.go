// (continuation for builder.go)
package prompts

import (
	"fmt"
	"strings"

	pkgerrors "code_agent/pkg/errors"
	"code_agent/tools"
)

// (This continuation finishes ValidatePromptStructure and additional helpers.)

func ValidatePromptStructure(prompt string) error {
	stack := []string{}
	pos := 0

	for pos < len(prompt) {
		// Find next '<'
		tagStart := strings.IndexByte(prompt[pos:], '<')
		if tagStart == -1 {
			break // No more tags
		}
		tagStart += pos

		// Check for CDATA section
		if strings.HasPrefix(prompt[tagStart:], "<![CDATA[") {
			// Find end of CDATA
			cdataEnd := strings.Index(prompt[tagStart:], "]]>")
			if cdataEnd == -1 {
				return pkgerrors.InvalidInputError(fmt.Sprintf("unclosed CDATA section starting at position %d", tagStart))
			}
			pos = tagStart + cdataEnd + 3
			continue
		}

		// Find matching '>'
		tagEnd := strings.IndexByte(prompt[tagStart:], '>')
		if tagEnd == -1 {
			return pkgerrors.InvalidInputError(fmt.Sprintf("unclosed tag starting at position %d", tagStart))
		}
		tagEnd += tagStart

		// Extract tag content (between < and >)
		tagContent := prompt[tagStart+1 : tagEnd]
		tagContent = strings.TrimSpace(tagContent)

		// Skip empty tags, XML declarations, or comments
		if tagContent == "" || strings.HasPrefix(tagContent, "?") || strings.HasPrefix(tagContent, "!") {
			pos = tagEnd + 1
			continue
		}

		// Check if it's a closing tag
		if strings.HasPrefix(tagContent, "/") {
			// Closing tag
			tagName := strings.TrimSpace(tagContent[1:])
			// Extract just the tag name (before any space)
			if idx := strings.IndexAny(tagName, " \t\n"); idx != -1 {
				tagName = tagName[:idx]
			}

			if len(stack) == 0 {
				return pkgerrors.InvalidInputError(fmt.Sprintf("unexpected closing tag </%s> with no matching opening tag at position %d", tagName, tagStart))
			}
			lastTag := stack[len(stack)-1]
			if lastTag != tagName {
				return pkgerrors.InvalidInputError(fmt.Sprintf("mismatched tag at position %d - expected </%s> but got </%s>", tagStart, lastTag, tagName))
			}
			stack = stack[:len(stack)-1]
		} else if !strings.HasSuffix(tagContent, "/") {
			// Opening tag (ignore self-closing tags ending with /)
			// Extract tag name (before space or special chars)
			tagName := tagContent
			if idx := strings.IndexAny(tagName, " \t\n"); idx != -1 {
				tagName = tagName[:idx]
			}
			stack = append(stack, tagName)
		}

		pos = tagEnd + 1
	}

	if len(stack) > 0 {
		return pkgerrors.InvalidInputError(fmt.Sprintf("unclosed tags: %v", stack))
	}

	return nil
}

// Backward-compatible wrappers
func BuildEnhancedPromptV2(registry *tools.ToolRegistry) string {
	builder := NewPromptBuilder(registry)
	ctx := PromptContext{HasWorkspace: false}
	return builder.BuildXMLPrompt(ctx)
}

func BuildEnhancedPromptWithContext(registry *tools.ToolRegistry, ctx PromptContext) string {
	builder := NewPromptBuilder(registry)
	return builder.BuildXMLPrompt(ctx)
}
