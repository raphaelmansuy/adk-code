package repl

import (
	"path/filepath"
	"testing"

	"code_agent/internal/display"
)

func TestNew_CreatesAndCloses(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	renderer, err := display.NewRenderer(display.OutputFormatPlain)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	cfg := Config{
		Renderer:       renderer,
		BannerRenderer: display.NewBannerRenderer(renderer),
	}

	r, err := New(cfg)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer r.Close()

	expected := filepath.Join(tmpHome, ".code_agent_history")
	if r.historyFile != expected {
		t.Fatalf("expected history file %s, got %s", expected, r.historyFile)
	}
}
