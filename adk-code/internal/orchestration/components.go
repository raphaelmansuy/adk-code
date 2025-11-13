package orchestration

import (
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"

	"adk-code/internal/display"
	"adk-code/internal/mcp"
	"adk-code/internal/session"
	"adk-code/internal/tracking"
	"adk-code/pkg/models"
)

// DisplayComponents groups all display-related fields
type DisplayComponents struct {
	Renderer       *display.Renderer
	BannerRenderer *display.BannerRenderer
	Typewriter     *display.TypewriterPrinter
	StreamDisplay  *display.StreamingDisplay
}

// ModelComponents groups all model-related fields
type ModelComponents struct {
	Registry *models.Registry
	Selected models.Config
	LLM      model.LLM
}

// SessionComponents groups all session-related fields
type SessionComponents struct {
	Manager *session.SessionManager
	Runner  *runner.Runner
	Tokens  *tracking.SessionTokens
}

// MCPComponents groups MCP-related fields
type MCPComponents struct {
	Manager *mcp.Manager
	Enabled bool
}
