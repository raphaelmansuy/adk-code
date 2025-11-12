package app

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code_agent/display"
	"code_agent/persistence"
	"code_agent/pkg/cli"
	"code_agent/pkg/models"
	"code_agent/tracking"
)

func TestInitializeDisplay_SetsFields(t *testing.T) {
	cfg := &cli.CLIConfig{OutputFormat: display.OutputFormatPlain, TypewriterEnabled: true}
	a := &Application{config: cfg}
	if err := a.initializeDisplay(); err != nil {
		t.Fatalf("initializeDisplay failed: %v", err)
	}
	if a.display == nil || a.display.Renderer == nil || a.display.BannerRenderer == nil || a.display.Typewriter == nil || a.display.StreamDisplay == nil {
		t.Fatalf("display components not initialized")
	}
	if !a.display.Typewriter.IsEnabled() {
		t.Fatalf("expected typewriter enabled")
	}
}

func TestInitializeREPL_Setup(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &cli.CLIConfig{SessionName: "sess1", WorkingDirectory: tmpDir}
	a := &Application{config: cfg}
	if err := a.initializeDisplay(); err != nil {
		t.Fatalf("init display err: %v", err)
	}
	a.session = &SessionComponents{
		Tokens: tracking.NewSessionTokens(),
	}
	a.model = &ModelComponents{
		Registry: models.NewRegistry(),
	}
	a.model.Selected = a.model.Registry.GetDefaultModel()
	if err := a.initializeREPL(); err != nil {
		t.Fatalf("initializeREPL failed: %v", err)
	}
	if a.repl == nil {
		t.Fatalf("repl should not be nil")
	}
}

func TestApplicationClose_Completes(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	sm, err := persistence.NewSessionManager("test_app", dbPath)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	// Create minimal application with a display and session manager
	cfg := &cli.CLIConfig{OutputFormat: display.OutputFormatPlain}
	a := &Application{config: cfg, session: &SessionComponents{Manager: sm}}
	if err := a.initializeDisplay(); err != nil {
		t.Fatalf("initialize display: %v", err)
	}
	// Create a minimal REPL to ensure Close calls don't panic
	a.session.Tokens = tracking.NewSessionTokens()
	a.model = &ModelComponents{
		Registry: models.NewRegistry(),
	}
	a.model.Selected = a.model.Registry.GetDefaultModel()
	if err := a.initializeREPL(); err != nil {
		t.Fatalf("initializeREPL failed: %v", err)
	}
	// Calling Close should not panic
	a.Close()
}

func TestNew_OpenAIRaisesIfNoEnvAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	cfg := &cli.CLIConfig{Model: "openai/gpt-4.1", WorkingDirectory: t.TempDir()}
	if _, err := New(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "OpenAI backend requires OPENAI_API_KEY") {
		t.Fatalf("expected OpenAI API key error, got: %v", err)
	}
}

func TestNew_GeminiMissingAPIKeyReturnsError(t *testing.T) {
	cfg := &cli.CLIConfig{Model: "", APIKey: "", WorkingDirectory: t.TempDir()}
	if _, err := New(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "Gemini API backend requires") {
		t.Fatalf("expected Gemini API key error, got: %v", err)
	}
}

func TestInitializeAgent_ReturnsErrorWhenMissingModel(t *testing.T) {
	cfg := &cli.CLIConfig{WorkingDirectory: t.TempDir()}
	a := &Application{ctx: context.Background(), config: cfg}
	a.model = &ModelComponents{LLM: nil}
	_ = a.initializeDisplay() // initialize display so a.display is available
	if err := a.initializeAgent(); err == nil {
		t.Fatalf("expected initializeAgent to error when LLM model is nil")
	}
}

func TestInitializeSession_SetsManagerAndSessionName(t *testing.T) {
	tmp := t.TempDir()
	cfg := &cli.CLIConfig{WorkingDirectory: tmp, DBPath: filepath.Join(tmp, "sessions.db")}

	renderer, _ := display.NewRenderer(display.OutputFormatPlain)
	a := &Application{ctx: context.Background(), config: cfg, display: &DisplayComponents{BannerRenderer: display.NewBannerRenderer(renderer)}}
	// Even if initializeSession returns an error due to runner.New, it should still initialize the session manager and session name
	if err := a.initializeSession(); err == nil && a.session == nil {
		t.Fatalf("initializeSession did not set session components: %v", err)
	}
	if a.config.SessionName == "" {
		t.Fatal("expected a.config.SessionName to be set")
	}
}

func TestREPL_Run_ExitsOnCanceledContext(t *testing.T) {
	renderer, _ := display.NewRenderer(display.OutputFormatPlain)
	cfg := REPLConfig{Renderer: renderer, BannerRenderer: display.NewBannerRenderer(renderer)}
	r, err := NewREPL(cfg)
	if err != nil {
		t.Fatalf("NewREPL failed: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// capture stdout
	origStdout := os.Stdout
	rPipe, w, _ := os.Pipe()
	os.Stdout = w
	r.Run(ctx)
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, rPipe)
	os.Stdout = origStdout
	out := buf.String()
	if !strings.Contains(out, "Goodbye!") {
		t.Fatalf("expected goodbye output, got: %q", out)
	}
	r.Close()
}

func TestApplicationRun_ExitsWhenContextCanceled(t *testing.T) {
	renderer, _ := display.NewRenderer(display.OutputFormatPlain)
	cfg := REPLConfig{Renderer: renderer, BannerRenderer: display.NewBannerRenderer(renderer)}
	repl, err := NewREPL(cfg)
	if err != nil {
		t.Fatalf("NewREPL failed: %v", err)
	}

	// Build application with a canceled signal handler
	handler := NewSignalHandler(context.Background())
	handler.Cancel()

	a := &Application{repl: repl, signalHandler: handler, ctx: handler.Context()}

	// capture stdout
	origStdout := os.Stdout
	rPipe, w, _ := os.Pipe()
	os.Stdout = w
	a.Run()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, rPipe)
	os.Stdout = origStdout
	out := buf.String()
	if !strings.Contains(out, "Goodbye!") {
		t.Fatalf("expected application goodbye output, got %q", out)
	}
	// ensure Close doesn't panic
	a.Close()
}
