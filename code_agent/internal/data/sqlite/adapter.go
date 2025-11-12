// Package sqlite provides SQLite implementations of data repositories
package sqlite

import (
	"context"

	sessionsvc "code_agent/session"
	"google.golang.org/adk/session"
)

// SessionRepositoryAdapter wraps SQLiteSessionService to implement SessionRepository interface
type SessionRepositoryAdapter struct {
	service *sessionsvc.SQLiteSessionService
}

// NewSessionRepositoryAdapter creates a new adapter wrapping an SQLiteSessionService
func NewSessionRepositoryAdapter(service *sessionsvc.SQLiteSessionService) *SessionRepositoryAdapter {
	return &SessionRepositoryAdapter{service: service}
}

// Create creates a new session
func (a *SessionRepositoryAdapter) Create(ctx context.Context, req *session.CreateRequest) (*session.CreateResponse, error) {
	return a.service.Create(ctx, req)
}

// Get retrieves a session
func (a *SessionRepositoryAdapter) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	return a.service.Get(ctx, req)
}

// List lists sessions
func (a *SessionRepositoryAdapter) List(ctx context.Context, req *session.ListRequest) (*session.ListResponse, error) {
	return a.service.List(ctx, req)
}

// Delete deletes a session
func (a *SessionRepositoryAdapter) Delete(ctx context.Context, req *session.DeleteRequest) error {
	return a.service.Delete(ctx, req)
}

// AppendEvent appends an event to a session
func (a *SessionRepositoryAdapter) AppendEvent(ctx context.Context, sess session.Session, event *session.Event) error {
	return a.service.AppendEvent(ctx, sess, event)
}

// Close closes the repository
func (a *SessionRepositoryAdapter) Close() error {
	return a.service.Close()
}
