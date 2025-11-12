package app

import (
	"code_agent/internal/runtime"
	"context"
)

// SignalHandler is a facade for runtime.SignalHandler
// Deprecated: Use code_agent/internal/runtime.SignalHandler instead
type SignalHandler = runtime.SignalHandler

// NewSignalHandler creates a new signal handler
// Deprecated: Use code_agent/internal/runtime.NewSignalHandler instead
func NewSignalHandler(ctx context.Context) *SignalHandler {
	return runtime.NewSignalHandler(ctx)
}
