package renderer

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

// MarkdownRenderer handles rendering markdown text to terminal format.
type MarkdownRenderer struct {
	renderer *glamour.TermRenderer
	width    int
}

// NewMarkdownRenderer creates a new markdown renderer with auto-detected theme.
func NewMarkdownRenderer() (*MarkdownRenderer, error) {
	// Use terminal word wrap by setting width to 0
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0), // Let terminal handle wrapping
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		return nil, err
	}

	return &MarkdownRenderer{
		renderer: r,
		width:    0, // Unlimited width
	}, nil
}

// NewMarkdownRendererWithWidth creates a markdown renderer with a specific width.
func NewMarkdownRendererWithWidth(width int) (*MarkdownRenderer, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		return nil, err
	}

	return &MarkdownRenderer{
		renderer: r,
		width:    width,
	}, nil
}

// Render renders markdown text to ANSI-formatted terminal output.
func (mr *MarkdownRenderer) Render(markdown string) (string, error) {
	rendered, err := mr.renderer.Render(markdown)
	if err != nil {
		return "", err
	}
	// Trim leading and trailing newlines
	return strings.TrimLeft(strings.TrimRight(rendered, "\n"), "\n"), nil
}
