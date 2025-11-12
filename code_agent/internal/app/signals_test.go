package app

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestSignalHandler_CtrlC_CancelsContext(t *testing.T) {
	ctx := context.Background()
	handler := NewSignalHandler(ctx)

	// Capture stdout
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	// Send SIGINT to simulate Ctrl+C
	handler.sigChan <- syscall.SIGINT

	// Wait for the context to be canceled
	select {
	case <-handler.Context().Done():
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("context was not canceled after SIGINT")
	}

	// Restore stdout and read output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = origStdout

	out := buf.String()
	if !strings.Contains(out, "Interrupted by user") {
		t.Fatalf("expected output to contain 'Interrupted by user', got: %q", out)
	}
	if handler.ctrlCCount != 1 {
		t.Fatalf("expected ctrlCCount == 1 got %d", handler.ctrlCCount)
	}

	// Cleanup
	handler.Cancel()
}
