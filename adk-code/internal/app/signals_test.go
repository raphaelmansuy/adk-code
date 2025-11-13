package app

import (
	"context"
	"testing"
)

// TestSignalHandler_CtrlC_CancelsContext tests that the facade works
// The actual implementation is tested in internal/runtime/signal_handler_test.go
func TestSignalHandler_CtrlC_CancelsContext(t *testing.T) {
	ctx := context.Background()
	handler := NewSignalHandler(ctx)

	// Test that the facade works
	if handler == nil {
		t.Fatal("NewSignalHandler returned nil")
	}

	handlerCtx := handler.Context()
	if handlerCtx == nil {
		t.Fatal("Context returned nil")
	}

	// Test cancellation
	handler.Cancel()
	select {
	case <-handlerCtx.Done():
		// Context was canceled as expected
	default:
		t.Fatal("Context was not canceled")
	}
}
