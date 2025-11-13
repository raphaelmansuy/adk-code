// Package file provides file operation tools for the coding agent.
package file

import (
	"strings"
)

// normalizeText normalizes whitespace in text for better matching.
// Converts escaped newlines, tabs, and carriage returns to actual characters.
func normalizeText(text string) string {
	// Convert escaped newlines to actual newlines
	text = strings.ReplaceAll(text, "\\n", "\n")
	text = strings.ReplaceAll(text, "\\t", "\t")
	text = strings.ReplaceAll(text, "\\r", "\r")
	return text
}
