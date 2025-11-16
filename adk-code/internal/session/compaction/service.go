package compaction

import (
	"context"

	"google.golang.org/adk/session"
)

// CompactionSessionService wraps the underlying session service
// to provide transparent compaction filtering when sessions are retrieved
type CompactionSessionService struct {
	underlying session.Service
	config     *Config
}

// NewCompactionService creates a wrapper around the session service
func NewCompactionService(underlying session.Service, config *Config) *CompactionSessionService {
	return &CompactionSessionService{
		underlying: underlying,
		config:     config,
	}
}

// Create creates a new session (pass-through to underlying service)
func (c *CompactionSessionService) Create(ctx context.Context, req *session.CreateRequest) (*session.CreateResponse, error) {
	return c.underlying.Create(ctx, req)
}

// Get wraps the underlying Get to return a filtered session
func (c *CompactionSessionService) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	resp, err := c.underlying.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	// Wrap the session with filtering layer
	filteredSession := NewFilteredSession(resp.Session)

	return &session.GetResponse{
		Session: filteredSession,
	}, nil
}

// List lists all sessions (pass-through to underlying service)
func (c *CompactionSessionService) List(ctx context.Context, req *session.ListRequest) (*session.ListResponse, error) {
	return c.underlying.List(ctx, req)
}

// Delete deletes a session (pass-through to underlying service)
func (c *CompactionSessionService) Delete(ctx context.Context, req *session.DeleteRequest) error {
	return c.underlying.Delete(ctx, req)
}

// AppendEvent appends an event to a session (pass-through to underlying service)
func (c *CompactionSessionService) AppendEvent(ctx context.Context, sess session.Session, event *session.Event) error {
	// If the session is a FilteredSession, unwrap it to get the underlying session
	if filtered, ok := sess.(*FilteredSession); ok {
		sess = filtered.Underlying
	}
	return c.underlying.AppendEvent(ctx, sess, event)
}
