// Package display provides rich terminal display functionality for the code agent.
package display

import t "code_agent/display/terminal"

// IsTTY returns true if stdout is a terminal (not piped or redirected).
func IsTTY() bool {
	return t.IsTTY()
}

// GetTerminalWidth returns the current terminal width in columns.
// Returns the fallback width if unable to determine the actual width.
func GetTerminalWidth() int {
	return t.GetTerminalWidth()
}

// GetTerminalHeight returns the current terminal height in rows.
// Returns the fallback height if unable to determine the actual height.
func GetTerminalHeight() int {
	return t.GetTerminalHeight()
}

// ClearLine clears the current line in the terminal.
func ClearLine() {
	t.ClearLine()
}

// ClearToEnd clears from cursor to end of screen.
func ClearToEnd() {
	t.ClearToEnd()
}

// MoveCursorUp moves the cursor up by n lines.
func MoveCursorUp(n int) {
	t.MoveCursorUp(n)
}
