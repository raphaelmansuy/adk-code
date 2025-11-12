package app

import (
	intrepl "code_agent/internal/repl"
)

// REPL is a facade for internal/repl.REPL
// Deprecated: Use code_agent/internal/repl.REPL instead
type REPL = intrepl.REPL

// REPLConfig is a facade for internal/repl.Config
// Deprecated: Use code_agent/internal/repl.Config instead
type REPLConfig = intrepl.Config

// NewREPL creates a new REPL instance
// Deprecated: Use code_agent/internal/repl.New instead
func NewREPL(config REPLConfig) (*REPL, error) {
	return intrepl.New(config)
}
