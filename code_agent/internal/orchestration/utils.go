package orchestration

import (
	"context"
	"fmt"
	"time"

	"code_agent/display"
	"code_agent/internal/session"
)

// GenerateUniqueSessionName creates a unique session name based on timestamp
// Format: session-YYYYMMDD-HHMMSS (e.g., session-20251110-221530)
func GenerateUniqueSessionName() string {
	now := time.Now()
	return fmt.Sprintf("session-%d%02d%02d-%02d%02d%02d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second())
}

// SessionInitializer handles session creation and retrieval
type SessionInitializer struct {
	manager        *session.SessionManager
	bannerRenderer *display.BannerRenderer
}

// NewSessionInitializer creates a new session initializer
func NewSessionInitializer(manager *session.SessionManager, bannerRenderer *display.BannerRenderer) *SessionInitializer {
	return &SessionInitializer{
		manager:        manager,
		bannerRenderer: bannerRenderer,
	}
}

// InitializeSession gets or creates a session
func (s *SessionInitializer) InitializeSession(ctx context.Context, userID, sessionName string) error {
	sess, err := s.manager.GetSession(ctx, userID, sessionName)
	if err != nil {
		// Session doesn't exist, create it
		_, err = s.manager.CreateSession(ctx, userID, sessionName)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		fmt.Printf("✨ Created new session: %s\n\n", sessionName)
	} else {
		// Use enhanced session resume header with event count and tokens
		resumeInfo := s.bannerRenderer.RenderSessionResumeInfo(sessionName, sess.Events().Len(), 0)
		fmt.Print(resumeInfo)
	}
	return nil
}
