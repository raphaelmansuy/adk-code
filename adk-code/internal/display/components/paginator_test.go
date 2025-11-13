package components

import (
	"testing"

	"adk-code/internal/display/renderer"
)

// Re-export for tests
var NewRenderer = renderer.NewRenderer

// TestNewPaginator creates and validates a paginator
func TestNewPaginator(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}

	p := NewPaginator(renderer)

	if p == nil {
		t.Fatal("NewPaginator returned nil")
	}
}

// TestPaginator_DisplayPagedEmptyLines handles empty input
func TestPaginator_DisplayPagedEmptyLines(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}

	p := NewPaginator(renderer)
	result := p.DisplayPaged([]string{})

	if !result {
		t.Errorf("DisplayPaged returned false for empty lines, expected true")
	}
}

// TestPaginator_DisplayPagedSingleLine handles single line
func TestPaginator_DisplayPagedSingleLine(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}

	p := NewPaginator(renderer)
	result := p.DisplayPaged([]string{"single line"})

	if !result {
		t.Errorf("DisplayPaged returned false for single line, expected true")
	}
}

// TestPaginator_DisplayPagedString converts string to lines and displays
func TestPaginator_DisplayPagedString(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}

	p := NewPaginator(renderer)
	content := "line1\nline2\nline3"
	result := p.DisplayPagedString(content)

	if !result {
		t.Errorf("DisplayPagedString returned false, expected true")
	}
}

// TestPaginator_DisplayPagedEmptyString handles empty string
func TestPaginator_DisplayPagedEmptyString(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}

	p := NewPaginator(renderer)
	result := p.DisplayPagedString("")

	if !result {
		t.Errorf("DisplayPagedString returned false for empty string, expected true")
	}
}
