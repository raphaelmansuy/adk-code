// Package display provides rich terminal display functionality for the code agent.
package display

import (
	"os"
	"strconv"

	"golang.org/x/term"
)

// IsTTY returns true if stdout is a terminal (not piped or redirected).
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// GetTerminalWidth returns the current terminal width in columns.
// Returns the fallback width if unable to determine the actual width.
func GetTerminalWidth() int {
	return getTerminalWidthOr(80)
}

// GetTerminalHeight returns the current terminal height in rows.
// Returns the fallback height if unable to determine the actual height.
func GetTerminalHeight() int {
	return getTerminalHeightOr(24)
}

// getTerminalWidthOr returns the terminal width or the provided fallback.
// It first tries term.GetSize, then falls back to $COLUMNS if set.
func getTerminalWidthOr(fallback int) int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

// getTerminalHeightOr returns the terminal height or the provided fallback.
// It first tries term.GetSize, then falls back to $LINES if set.
func getTerminalHeightOr(fallback int) int {
	if _, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil && h > 0 {
		return h
	}
	if lines := os.Getenv("LINES"); lines != "" {
		if n, err := strconv.Atoi(lines); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

// ClearLine clears the current line in the terminal.
func ClearLine() {
	if !IsTTY() {
		return
	}
	os.Stdout.WriteString("\r\033[K")
}

// ClearToEnd clears from cursor to end of screen.
func ClearToEnd() {
	if !IsTTY() {
		return
	}
	os.Stdout.WriteString("\033[J")
}

// MoveCursorUp moves the cursor up by n lines.
func MoveCursorUp(n int) {
	if !IsTTY() {
		return
	}
	os.Stdout.WriteString("\033[" + strconv.Itoa(n) + "A")
}
