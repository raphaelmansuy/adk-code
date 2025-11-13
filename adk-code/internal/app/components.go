package app

import (
	// Re-export component types from orchestration for backward compatibility
	orchcomp "adk-code/internal/orchestration"
)

// DisplayComponents is a facade for backward compatibility
type DisplayComponents = orchcomp.DisplayComponents

// ModelComponents is a facade for backward compatibility
type ModelComponents = orchcomp.ModelComponents

// SessionComponents is a facade for backward compatibility
type SessionComponents = orchcomp.SessionComponents

// MCPComponents is a facade for backward compatibility
type MCPComponents = orchcomp.MCPComponents
