// Package data - Data access layer with repository pattern
package data

import (
	"context"

	"google.golang.org/adk/session"
)

// SessionRepository defines the interface for session persistence
// Implementations can use SQLite, in-memory, or other backends
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, req *session.CreateRequest) (*session.CreateResponse, error)

	// Get retrieves a session by ID
	Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error)

	// List retrieves all sessions for a user
	List(ctx context.Context, req *session.ListRequest) (*session.ListResponse, error)

	// Delete deletes a session
	Delete(ctx context.Context, req *session.DeleteRequest) error

	// AppendEvent appends an event to a session and updates session/app/user state
	AppendEvent(ctx context.Context, sess session.Session, event *session.Event) error

	// Close closes the repository connection
	Close() error
}

// ModelRegistry defines the interface for model registry persistence
// Implementations can use in-memory, SQLite, or other backends
type ModelRegistry interface {
	// GetModel retrieves a model by ID
	GetModel(id string) (any, error)

	// GetModelByName retrieves a model by display name (case-insensitive)
	GetModelByName(name string) (any, error)

	// GetDefaultModel returns the default model
	GetDefaultModel() any

	// ListModels returns all available models
	ListModels() []any

	// ListModelsByBackend returns models filtered by backend provider
	ListModelsByBackend(backend string) []any
}

// RepositoryFactory creates repository instances
type RepositoryFactory interface {
	// CreateSessionRepository creates a session repository
	CreateSessionRepository(ctx context.Context, dbPath string, appName string) (SessionRepository, error)

	// CreateModelRegistry creates a model registry
	CreateModelRegistry() ModelRegistry
}
