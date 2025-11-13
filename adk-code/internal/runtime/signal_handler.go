package runtime

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SignalHandler manages graceful shutdown on Ctrl+C
type SignalHandler struct {
	sigChan    chan os.Signal
	ctx        context.Context
	cancel     context.CancelFunc
	ctrlCCount int
}

// NewSignalHandler creates a new signal handler
func NewSignalHandler(ctx context.Context) *SignalHandler {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)

	handler := &SignalHandler{
		sigChan: sigChan,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Start signal handling goroutine
	go handler.handleSignals()

	return handler
}

// handleSignals processes OS signals in a goroutine
func (h *SignalHandler) handleSignals() {
	for sig := range h.sigChan {
		h.ctrlCCount++
		if sig == syscall.SIGINT {
			if h.ctrlCCount == 1 {
				fmt.Println("\n\n⚠️  Interrupted by user (Ctrl+C)")
				fmt.Println("Cancelling current operation...")
			} else {
				fmt.Println("\n\n⚠️  Ctrl+C pressed again - forcing exit")
				os.Exit(130) // Standard exit code for SIGINT
			}
		}
		h.cancel()
	}
}

// Context returns the cancellable context
func (h *SignalHandler) Context() context.Context {
	return h.ctx
}

// Cancel cancels the context
func (h *SignalHandler) Cancel() {
	h.cancel()
}
