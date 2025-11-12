// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

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
