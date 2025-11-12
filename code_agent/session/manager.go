package session

import (
	internalsession "code_agent/internal/session"
)

// SessionManager is a re-export for backward compatibility
type SessionManager = internalsession.SessionManager

// NewSessionManager creates a new session manager
func NewSessionManager(appName, dbPath string) (*SessionManager, error) {
	return internalsession.NewSessionManager(appName, dbPath)
}
