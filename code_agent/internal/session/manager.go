package session

import (
	"context"
	"os"
	"path/filepath"

	"code_agent/internal/session/persistence"
	pkgerrors "code_agent/pkg/errors"

	"google.golang.org/adk/session"
)

// SessionManager provides utilities for managing sessions
type SessionManager struct {
	sessionService session.Service
	dbPath         string
	appName        string
}

// NewSessionManager creates a new session manager
func NewSessionManager(appName, dbPath string) (*SessionManager, error) {
	// Ensure dbPath is provided, use default if not
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to get home directory", err)
		}
		dbPath = filepath.Join(home, ".code_agent", "sessions.db")
	}

	// Create persistence service
	persistenceSvc, err := persistence.NewSQLiteSessionService(dbPath)
	if err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create session service", err)
	}

	return &SessionManager{
		sessionService: persistenceSvc,
		dbPath:         dbPath,
		appName:        appName,
	}, nil
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(ctx context.Context, userID, sessionName string) (session.Session, error) {
	req := &session.CreateRequest{
		AppName:   sm.appName,
		UserID:    userID,
		SessionID: sessionName,
		State:     make(map[string]any),
	}

	resp, err := sm.sessionService.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Session, nil
}

// GetSession retrieves a session
func (sm *SessionManager) GetSession(ctx context.Context, userID, sessionID string) (session.Session, error) {
	req := &session.GetRequest{
		AppName:   sm.appName,
		UserID:    userID,
		SessionID: sessionID,
	}

	resp, err := sm.sessionService.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Session, nil
}

// ListSessions lists all sessions for a user
func (sm *SessionManager) ListSessions(ctx context.Context, userID string) ([]session.Session, error) {
	req := &session.ListRequest{
		AppName: sm.appName,
		UserID:  userID,
	}

	resp, err := sm.sessionService.List(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Sessions, nil
}

// DeleteSession deletes a session
func (sm *SessionManager) DeleteSession(ctx context.Context, userID, sessionID string) error {
	req := &session.DeleteRequest{
		AppName:   sm.appName,
		UserID:    userID,
		SessionID: sessionID,
	}

	return sm.sessionService.Delete(ctx, req)
}

// GetService returns the underlying session service
func (sm *SessionManager) GetService() session.Service {
	return sm.sessionService
}

// GetDBPath returns the database path
func (sm *SessionManager) GetDBPath() string {
	return sm.dbPath
}

// Close closes the session service
func (sm *SessionManager) Close() error {
	if sqlite, ok := sm.sessionService.(*persistence.SQLiteSessionService); ok {
		return sqlite.Close()
	}
	return nil
}
