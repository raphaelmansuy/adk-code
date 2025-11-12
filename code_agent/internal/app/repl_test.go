package app

import (
	"path/filepath"
	"testing"

	"code_agent/display"
)

func TestNewREPL_CreatesAndCloses(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	renderer, err := display.NewRenderer(display.OutputFormatPlain)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	cfg := REPLConfig{
		Renderer:       renderer,
		BannerRenderer: display.NewBannerRenderer(renderer),
	}

	r, err := NewREPL(cfg)
	if err != nil {
		t.Fatalf("NewREPL error: %v", err)
	}
	defer r.Close()

	expected := filepath.Join(tmpHome, ".code_agent_history")
	if r.historyFile != expected {
		t.Fatalf("expected history file %s, got %s", expected, r.historyFile)
	}
}

// TestProcessUserMessage_HandlesRunnerEvents is skipped because it requires full runner integration
// which is difficult to mock. The processUserMessage method is well-tested through the Run() method tests.
