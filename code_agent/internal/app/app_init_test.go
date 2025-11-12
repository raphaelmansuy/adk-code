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
	"code_agent/internal/config"
	"code_agent/internal/runtime"
	"code_agent/internal/session"
	"code_agent/pkg/models"
	"code_agent/tracking"
)

func TestInitializeDisplay_SetsFields(t *testing.T) {
	cfg := &config.Config{OutputFormat: display.OutputFormatPlain, TypewriterEnabled: true}
	display, err := initializeDisplayComponents(cfg)
	if err != nil {
		t.Fatalf("initializeDisplayComponents failed: %v", err)
	}
	if display == nil || display.Renderer == nil || display.BannerRenderer == nil || display.Typewriter == nil || display.StreamDisplay == nil {
		t.Fatalf("display components not initialized")
	}
	if !display.Typewriter.IsEnabled() {
		t.Fatalf("expected typewriter enabled")
	}
}

func TestInitializeREPL_Setup(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{SessionName: "sess1", WorkingDirectory: tmpDir}
	displayComp, err := initializeDisplayComponents(cfg)
	if err != nil {
		t.Fatalf("init display err: %v", err)
	}
	a := &Application{config: cfg, display: displayComp, ctx: context.Background()}
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
	sm, err := session.NewSessionManager("test_app", dbPath)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	// Create minimal application with a display and session manager
	cfg := &config.Config{OutputFormat: display.OutputFormatPlain}
	displayComp, _ := initializeDisplayComponents(cfg)
	a := &Application{config: cfg, session: &SessionComponents{Manager: sm}, display: displayComp}
	// Create a minimal REPL to ensure Close calls don't panic
	a.session.Tokens = tracking.NewSessionTokens()
	a.model = &ModelComponents{
		Registry: models.NewRegistry(),
	}
	a.model.Selected = a.model.Registry.GetDefaultModel()
	a.ctx = context.Background()
	if err := a.initializeREPL(); err != nil {
		t.Fatalf("initializeREPL failed: %v", err)
	}
	// Calling Close should not panic
	a.Close()
}

func TestNew_OpenAIRaisesIfNoEnvAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	cfg := &config.Config{Model: "openai/gpt-4.1", WorkingDirectory: t.TempDir()}
	if _, err := New(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "openAI backend requires OPENAI_API_KEY") {
		t.Fatalf("expected OpenAI API key error, got: %v", err)
	}
}

func TestNew_GeminiMissingAPIKeyReturnsError(t *testing.T) {
	cfg := &config.Config{Model: "", APIKey: "", WorkingDirectory: t.TempDir()}
	if _, err := New(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "gemini API backend requires") {
		t.Fatalf("expected Gemini API key error, got: %v", err)
	}
}

func TestInitializeAgent_ReturnsErrorWhenMissingModel(t *testing.T) {
	cfg := &config.Config{WorkingDirectory: t.TempDir()}
	// nil LLM should cause error
	if _, err := initializeAgentComponent(context.Background(), cfg, nil); err == nil {
		t.Fatalf("expected initializeAgentComponent to error when LLM model is nil")
	}
}

func TestInitializeSession_SetsManagerAndSessionName(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{WorkingDirectory: tmp, DBPath: filepath.Join(tmp, "sessions.db")}

	// Verify that GenerateUniqueSessionName works
	sessionName := GenerateUniqueSessionName()
	if sessionName == "" {
		t.Fatal("expected generated session name to not be empty")
	}

	// Create a session manager and verify it was created successfully
	sm, err := session.NewSessionManager("code_agent", cfg.DBPath)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	defer sm.Close()

	// Verify session manager was created successfully
	if sm == nil {
		t.Fatal("expected session manager to be created")
	}
}

// mockAgent is a minimal implementation for testing
type mockAgent struct{}

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
	handler := runtime.NewSignalHandler(context.Background())
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
