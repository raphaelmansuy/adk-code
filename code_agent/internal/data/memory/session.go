// Package memory provides in-memory implementation of SessionRepository for testing
package memory

import (
	"context"
	"iter"
	"sync"
	"time"

	"code_agent/internal/data"
	pkgerrors "code_agent/pkg/errors"
	"google.golang.org/adk/session"
)

// InMemorySession represents a session stored in memory
type InMemorySession struct {
	AppName   string
	UserID    string
	ID        string
	State     map[string]any
	Events    []*session.Event
	CreatedAt time.Time
	UpdatedAt time.Time
}

// InMemorySessionRepository implements SessionRepository interface using in-memory storage
type InMemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*InMemorySession // key: "appName:userID:sessionID"
}

// NewInMemorySessionRepository creates a new in-memory session repository
func NewInMemorySessionRepository() data.SessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[string]*InMemorySession),
	}
}

// Create creates a new session in memory
func (r *InMemorySessionRepository) Create(ctx context.Context, req *session.CreateRequest) (*session.CreateResponse, error) {
	if req.AppName == "" || req.UserID == "" {
		return nil, pkgerrors.InvalidInputError("app_name and user_id are required")
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(req.AppName, req.UserID, sessionID)
	if _, exists := r.sessions[key]; exists {
		return nil, pkgerrors.InvalidInputError("session already exists")
	}

	state := req.State
	if state == nil {
		state = make(map[string]any)
	}

	now := time.Now()
	inMemSession := &InMemorySession{
		AppName:   req.AppName,
		UserID:    req.UserID,
		ID:        sessionID,
		State:     state,
		Events:    make([]*session.Event, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}

	r.sessions[key] = inMemSession

	// Create session wrapper
	sessionWrapper := &sessionWrapper{
		appName:   req.AppName,
		userID:    req.UserID,
		sessionID: sessionID,
		state:     inMemSession.State,
		updatedAt: now,
		events:    inMemSession.Events,
	}

	return &session.CreateResponse{Session: sessionWrapper}, nil
}

// Get retrieves a session from memory
func (r *InMemorySessionRepository) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	if req.AppName == "" || req.UserID == "" || req.SessionID == "" {
		return nil, pkgerrors.InvalidInputError("app_name, user_id, and session_id are required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	key := r.makeKey(req.AppName, req.UserID, req.SessionID)
	inMemSession, exists := r.sessions[key]
	if !exists {
		return nil, pkgerrors.InvalidInputError("session not found")
	}

	// Build events list based on filters
	var events []*session.Event
	if req.NumRecentEvents > 0 {
		// Return most recent N events, reversed in order
		start := len(inMemSession.Events) - int(req.NumRecentEvents)
		if start < 0 {
			start = 0
		}
		events = inMemSession.Events[start:]
		// Reverse the order for recent events
		reversedEvents := make([]*session.Event, len(events))
		for i, e := range events {
			reversedEvents[len(events)-1-i] = e
		}
		events = reversedEvents
	} else if !req.After.IsZero() {
		// Filter events after a timestamp
		for _, e := range inMemSession.Events {
			if e.Timestamp.After(req.After) || e.Timestamp.Equal(req.After) {
				events = append(events, e)
			}
		}
	} else {
		// Return all events
		events = inMemSession.Events
	}

	sessionWrapper := &sessionWrapper{
		appName:   inMemSession.AppName,
		userID:    inMemSession.UserID,
		sessionID: inMemSession.ID,
		state:     inMemSession.State,
		updatedAt: inMemSession.UpdatedAt,
		events:    events,
	}

	return &session.GetResponse{Session: sessionWrapper}, nil
}

// List lists all sessions for a user
func (r *InMemorySessionRepository) List(ctx context.Context, req *session.ListRequest) (*session.ListResponse, error) {
	if req.AppName == "" {
		return nil, pkgerrors.InvalidInputError("app_name is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []session.Session
	for _, inMemSession := range r.sessions {
		if inMemSession.AppName != req.AppName {
			continue
		}
		if req.UserID != "" && inMemSession.UserID != req.UserID {
			continue
		}

		sessionWrapper := &sessionWrapper{
			appName:   inMemSession.AppName,
			userID:    inMemSession.UserID,
			sessionID: inMemSession.ID,
			state:     inMemSession.State,
			updatedAt: inMemSession.UpdatedAt,
			events:    inMemSession.Events,
		}
		result = append(result, sessionWrapper)
	}

	return &session.ListResponse{Sessions: result}, nil
}

// Delete deletes a session from memory
func (r *InMemorySessionRepository) Delete(ctx context.Context, req *session.DeleteRequest) error {
	if req.AppName == "" || req.UserID == "" || req.SessionID == "" {
		return pkgerrors.InvalidInputError("app_name, user_id, and session_id are required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(req.AppName, req.UserID, req.SessionID)
	if _, exists := r.sessions[key]; !exists {
		return pkgerrors.InvalidInputError("session not found")
	}

	delete(r.sessions, key)
	return nil
}

// AppendEvent appends an event to a session in memory
func (r *InMemorySessionRepository) AppendEvent(ctx context.Context, sess session.Session, event *session.Event) error {
	if sess == nil {
		return pkgerrors.InvalidInputError("session is nil")
	}
	if event == nil {
		return pkgerrors.InvalidInputError("event is nil")
	}
	if event.Partial {
		return nil
	}

	wrapper, ok := sess.(*sessionWrapper)
	if !ok {
		return pkgerrors.InternalError("unexpected session type")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(wrapper.appName, wrapper.userID, wrapper.sessionID)
	inMemSession, exists := r.sessions[key]
	if !exists {
		return pkgerrors.InvalidInputError("session not found")
	}

	// Add event to session
	inMemSession.Events = append(inMemSession.Events, event)
	inMemSession.UpdatedAt = time.Now()

	// Update session state if event has state delta
	if len(event.Actions.StateDelta) > 0 {
		for k, v := range event.Actions.StateDelta {
			inMemSession.State[k] = v
		}
	}

	return nil
}

// Close closes the in-memory repository (no-op for memory-based storage)
func (r *InMemorySessionRepository) Close() error {
	return nil
}

// Helper methods

func (r *InMemorySessionRepository) makeKey(appName, userID, sessionID string) string {
	return appName + ":" + userID + ":" + sessionID
}

// sessionWrapper wraps an in-memory session to implement session.Session interface
type sessionWrapper struct {
	appName   string
	userID    string
	sessionID string
	state     map[string]any
	updatedAt time.Time
	events    []*session.Event
}

func (s *sessionWrapper) AppName() string           { return s.appName }
func (s *sessionWrapper) ID() string                { return s.sessionID }
func (s *sessionWrapper) UserID() string            { return s.userID }
func (s *sessionWrapper) SessionID() string         { return s.sessionID }
func (s *sessionWrapper) State() session.State      { return &sessionState{state: s.state} }
func (s *sessionWrapper) Events() session.Events    { return &sessionEvents{events: s.events} }
func (s *sessionWrapper) LastUpdateTime() time.Time { return s.updatedAt }

// sessionState implements session.State
type sessionState struct{ state map[string]any }

func (s *sessionState) Get(key string) (any, error) {
	if val, ok := s.state[key]; ok {
		return val, nil
	}
	return nil, session.ErrStateKeyNotExist
}

func (s *sessionState) Set(key string, value any) error {
	s.state[key] = value
	return nil
}

func (s *sessionState) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, v := range s.state {
			if !yield(k, v) {
				return
			}
		}
	}
}

// sessionEvents implements session.Events
type sessionEvents struct{ events []*session.Event }

func (e *sessionEvents) All() iter.Seq[*session.Event] {
	return func(yield func(*session.Event) bool) {
		for _, event := range e.events {
			if !yield(event) {
				return
			}
		}
	}
}

func (e *sessionEvents) Len() int { return len(e.events) }

func (e *sessionEvents) At(i int) *session.Event {
	if i >= 0 && i < len(e.events) {
		return e.events[i]
	}
	return nil
}

// generateSessionID creates a unique session ID
func generateSessionID() string {
	return "session-" + time.Now().Format("20060102150405")
}
