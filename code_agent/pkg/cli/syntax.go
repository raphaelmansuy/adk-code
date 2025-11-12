// Package cli - Syntax parsing utilities
package cli

import (
	"fmt"
	"strings"
)

// ParseProviderModelSyntax parses a "provider/model" string into components
// Examples:
//
//	"gemini/2.5-flash" → ("gemini", "2.5-flash", nil)
//	"gemini/flash" → ("gemini", "flash", nil)
//	"flash" → ("", "flash", nil)
//	"/flash" → ("", "", error)
//	"a/b/c" → ("", "", error)
func ParseProviderModelSyntax(input string) (string, string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", fmt.Errorf("model syntax cannot be empty")
	}

	parts := strings.Split(input, "/")
	switch len(parts) {
	case 1:
		// Shorthand without provider: "flash" → ("", "flash")
		return "", parts[0], nil
	case 2:
		// Full syntax: "provider/model" → ("provider", "model")
		if parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("invalid model syntax: %q (use provider/model)", input)
		}
		return parts[0], parts[1], nil
	default:
		return "", "", fmt.Errorf("invalid model syntax: %q (use provider/model)", input)
	}
}
