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
	"code_agent/session"
)

func TestInitializeSession_CreatesNewSessionIfMissing(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	manager, err := session.NewSessionManager("test_app", dbPath)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	defer manager.Close()

	renderer, err := display.NewRenderer(display.OutputFormatPlain)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}
	banner := display.NewBannerRenderer(renderer)
	sessionInit := NewSessionInitializer(manager, banner)

	// Capture stdout
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	if err := sessionInit.InitializeSession(context.Background(), "user1", "session1"); err != nil {
		t.Fatalf("InitializeSession returned error: %v", err)
	}

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = origStdout
	output := buf.String()
	if !strings.Contains(output, "Created new session:") {
		t.Fatalf("expected output to contain Created new session, got: %q", output)
	}
}

func TestInitializeSession_ResumesExistingSession(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	manager, err := session.NewSessionManager("test_app", dbPath)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()
	if _, err := manager.CreateSession(ctx, "user1", "session2"); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	renderer, err := display.NewRenderer(display.OutputFormatPlain)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}
	banner := display.NewBannerRenderer(renderer)
	sessionInit := NewSessionInitializer(manager, banner)

	// Capture stdout
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	if err := sessionInit.InitializeSession(ctx, "user1", "session2"); err != nil {
		t.Fatalf("InitializeSession returned error: %v", err)
	}

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = origStdout
	output := buf.String()
	if !strings.Contains(output, "Resumed session") {
		t.Fatalf("expected output to contain Resumed session, got: %q", output)
	}
}
