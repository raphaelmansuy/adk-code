package components

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"code_agent/internal/display/core"

	"golang.org/x/term"
)

// Re-export types and functions from core
var (
	IsTTY             = core.IsTTY
	GetTerminalHeight = core.GetTerminalHeight
)

// Paginator handles displaying long content with pagination in the terminal,
// similar to the 'more' or 'less' command.
type Paginator struct {
	renderer core.StyleRenderer
}

// NewPaginator creates a new paginator instance.
func NewPaginator(renderer core.StyleRenderer) *Paginator {
	return &Paginator{renderer: renderer}
}

// DisplayPaged displays content with pagination if it exceeds terminal height.
// Lines is an array of strings representing content lines.
// Returns true if user completed viewing all pages, false if they quit early.
func (p *Paginator) DisplayPaged(lines []string) bool {
	if len(lines) == 0 {
		return true
	}

	// If not in a TTY or content fits in one screen, just print it all
	if !IsTTY() {
		for _, line := range lines {
			fmt.Println(line)
		}
		return true
	}

	termHeight := GetTerminalHeight()
	// Reserve space for the pagination prompt (2 lines)
	pageHeight := termHeight - 2
	if pageHeight < 5 {
		pageHeight = 5 // Minimum page height
	}

	// Display pages
	currentPage := 0
	totalPages := (len(lines) + pageHeight - 1) / pageHeight

	for currentPage < totalPages {
		// Calculate page boundaries
		startLine := currentPage * pageHeight
		endLine := startLine + pageHeight
		if endLine > len(lines) {
			endLine = len(lines)
		}

		// Print page content
		for _, line := range lines[startLine:endLine] {
			fmt.Println(line)
		}

		currentPage++

		// If there are more pages, show pagination prompt
		if currentPage < totalPages {
			if !p.showPaginationPrompt(currentPage, totalPages) {
				return false // User quit
			}
		}
	}

	return true
}

// DisplayPagedString displays content (as a single string) with pagination.
// Splits the string by newlines and displays it page by page.
func (p *Paginator) DisplayPagedString(content string) bool {
	lines := strings.Split(content, "\n")
	return p.DisplayPaged(lines)
}

// showPaginationPrompt shows the pagination prompt and waits for user input.
// Returns true to continue, false to quit.
func (p *Paginator) showPaginationPrompt(currentPage, totalPages int) bool {
	// Build prompt
	promptStr := fmt.Sprintf("[Page %d/%d] Press SPACE to continue, Q to quit: ", currentPage, totalPages)
	fmt.Print(p.renderer.Dim(promptStr))

	// Try to read from stdin in raw mode
	if !isStdinAvailable() {
		// Can't read from stdin (piped or redirected), just continue
		return true
	}

	// Save terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fall back to simple reader if we can't make raw
		return p.fallbackPrompt()
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Read one byte at a time
	buffer := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buffer)
		if err != nil || n == 0 {
			return true // Default to continue on error
		}

		char := buffer[0]
		switch char {
		case ' ', '\n', '\r': // Space, Enter
			clearPromptLine()
			return true
		case 'q', 'Q': // Quit
			clearPromptLine()
			return false
		case 3: // Ctrl-C
			clearPromptLine()
			return false
		default:
			// Ignore other keys and keep waiting for valid input
			continue
		}
	}
}

// fallbackPrompt is used when we can't set raw terminal mode
// Falls back to buffered input with less control
func (p *Paginator) fallbackPrompt() bool {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	clearPromptLine()
	if input == "q" || input == "quit" {
		return false
	}
	return true
}

// clearPromptLine clears the pagination prompt line
func clearPromptLine() {
	// Move cursor to start of line and clear it
	fmt.Print("\r")
	// Clear with spaces (80 chars should be enough for most terminals)
	fmt.Print(strings.Repeat(" ", 120))
	fmt.Print("\r")
}

// isStdinAvailable checks if stdin is available for reading
func isStdinAvailable() bool {
	// Check if stdin is a terminal or pipe
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return true
	}
	// stdin might be available but not a terminal (piped)
	// We can still try to read from it, so return true
	return os.Stdin != nil
}
