package app

import (
	"testing"

	"adk-code/internal/display"
	intrepl "adk-code/internal/repl"
)

// TestNewREPL_CreatesAndCloses tests REPL creation
func TestNewREPL_CreatesAndCloses(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	renderer, err := display.NewRenderer(display.OutputFormatPlain)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	cfg := intrepl.Config{
		Renderer:       renderer,
		BannerRenderer: display.NewBannerRenderer(renderer),
	}

	r, err := intrepl.New(cfg)
	if err != nil {
		t.Fatalf("repl.New error: %v", err)
	}

	if r == nil {
		t.Fatal("expected REPL instance, got nil")
	}

	err = r.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
}
