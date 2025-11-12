package app

import (
	"code_agent/display"
	"code_agent/internal/orchestration"
	"code_agent/internal/session"
)

// SessionInitializer is a facade for backward compatibility
type SessionInitializer = orchestration.SessionInitializer

// NewSessionInitializer is a facade for backward compatibility
func NewSessionInitializer(manager *session.SessionManager, bannerRenderer *display.BannerRenderer) *SessionInitializer {
	return orchestration.NewSessionInitializer(manager, bannerRenderer)
}
