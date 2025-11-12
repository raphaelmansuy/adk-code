package orchestration

import (
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"

	"code_agent/internal/display"
	"code_agent/internal/session"
	"code_agent/internal/tracking"
	"code_agent/pkg/models"
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
