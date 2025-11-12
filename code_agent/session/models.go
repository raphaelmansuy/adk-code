// Package session provides session persistence data models and helpers
package session

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	"google.golang.org/adk/session"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// stateMap is a custom type for map[string]any that handles JSON serialization
type stateMap map[string]any

// GormDataType defines the generic fallback data type
func (stateMap) GormDataType() string {
	return "text"
}

// GormDBDataType defines database specific data types
func (stateMap) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "TEXT"
	case "postgres":
		return "JSONB"
	case "mysql":
		return "LONGTEXT"
	default:
		return ""
	}
}

// Value implements the driver.Valuer interface
func (sm stateMap) Value() (driver.Value, error) {
	if sm == nil {
		sm = make(map[string]any)
	}
	b, err := json.Marshal(sm)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements the sql.Scanner interface
func (sm *stateMap) Scan(value any) error {
	if value == nil {
		*sm = make(map[string]any)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSON value: %T", value)
	}

	if len(bytes) == 0 {
		*sm = make(map[string]any)
		return nil
	}

	return json.Unmarshal(bytes, sm)
}

func (sm stateMap) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	data, _ := json.Marshal(sm)
	return gorm.Expr("?", string(data))
}

// dynamicJSON is a custom JSON type that handles serialization
type dynamicJSON json.RawMessage

// Value implements the driver.Valuer interface
func (j dynamicJSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// Scan implements the sql.Scanner interface
func (j *dynamicJSON) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*j = nil
			return nil
		}
		bytes = make([]byte, len(v))
		copy(bytes, v)
	case string:
		if v == "" {
			*j = nil
			return nil
		}
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSON value: %T", value)
	}

	if !json.Valid(bytes) {
		return fmt.Errorf("invalid JSON received from database: %s", string(bytes))
	}
	*j = dynamicJSON(bytes)
	return nil
}

func (j dynamicJSON) String() string {
	return string(j)
}

// GormDataType defines the generic fallback data type
func (dynamicJSON) GormDataType() string {
	return "text"
}

// GormDBDataType defines database specific data types
func (dynamicJSON) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "TEXT"
	case "mysql":
		return "LONGTEXT"
	case "postgres":
		return "JSONB"
	default:
		return ""
	}
}

func (js dynamicJSON) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if len(js) == 0 {
		return gorm.Expr("NULL")
	}
	return gorm.Expr("?", string(js))
}

// storageSession represents a session in the database
type storageSession struct {
	AppName    string   `gorm:"primaryKey;"`
	UserID     string   `gorm:"primaryKey;"`
	ID         string   `gorm:"primaryKey;"`
	State      stateMap `gorm:"type:text;serializer:json"`
	CreateTime time.Time
	UpdateTime time.Time
}

// TableName sets the table name
func (storageSession) TableName() string {
	return "sessions"
}

// storageEvent represents an event in the database
type storageEvent struct {
	ID        string `gorm:"primaryKey;"`
	AppName   string `gorm:"primaryKey;"`
	UserID    string `gorm:"primaryKey;"`
	SessionID string `gorm:"primaryKey;"`
	Timestamp time.Time

	InvocationID           string
	Author                 string
	Actions                []byte
	LongRunningToolIDsJSON dynamicJSON
	Branch                 *string
	Content                dynamicJSON `gorm:"type:text"`
	GroundingMetadata      dynamicJSON `gorm:"type:text"`
	CustomMetadata         dynamicJSON `gorm:"type:text"`
	UsageMetadata          dynamicJSON `gorm:"type:text"`
	CitationMetadata       dynamicJSON `gorm:"type:text"`
	Partial                *bool
	TurnComplete           *bool
	ErrorCode              *string
	ErrorMessage           *string
	Interrupted            *bool
}

// TableName sets the table name
func (storageEvent) TableName() string {
	return "events"
}

// storageAppState represents application state
type storageAppState struct {
	AppName    string   `gorm:"primaryKey;"`
	State      stateMap `gorm:"type:text;serializer:json"`
	UpdateTime time.Time
}

// TableName sets the table name
func (storageAppState) TableName() string {
	return "app_states"
}

// storageUserState represents user state
type storageUserState struct {
	AppName    string   `gorm:"primaryKey;"`
	UserID     string   `gorm:"primaryKey;"`
	State      stateMap `gorm:"type:text;serializer:json"`
	UpdateTime time.Time
}

// TableName sets the table name
func (storageUserState) TableName() string {
	return "user_states"
}

// localSession is the in-memory representation of a session
type localSession struct {
	appName   string
	userID    string
	sessionID string
	state     stateMap
	updatedAt time.Time
	events    []*session.Event
}

// ID returns the session ID
func (s *localSession) ID() string { return s.sessionID }

// AppName returns the app name
func (s *localSession) AppName() string { return s.appName }

// UserID returns the user ID
func (s *localSession) UserID() string { return s.userID }

// State returns the session state
func (s *localSession) State() session.State { return &localState{state: s.state} }

// Events returns the events
func (s *localSession) Events() session.Events { return &localEvents{events: s.events} }

// LastUpdateTime returns the last update time
func (s *localSession) LastUpdateTime() time.Time { return s.updatedAt }

// localState implements session.State
type localState struct{ state stateMap }

// Get retrieves a value from state
func (s *localState) Get(key string) (any, error) {
	if val, ok := s.state[key]; ok {
		return val, nil
	}
	return nil, session.ErrStateKeyNotExist
}

// Set sets a value in state
func (s *localState) Set(key string, value any) error { s.state[key] = value; return nil }

// All returns an iterator over all state entries
func (s *localState) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, v := range s.state {
			if !yield(k, v) {
				return
			}
		}
	}
}

// localEvents implements session.Events
type localEvents struct{ events []*session.Event }

// All returns an iterator over all events
func (e *localEvents) All() iter.Seq[*session.Event] {
	return func(yield func(*session.Event) bool) {
		for _, event := range e.events {
			if !yield(event) {
				return
			}
		}
	}
}

// Len returns the number of events
func (e *localEvents) Len() int { return len(e.events) }

// At returns the event at the given index
func (e *localEvents) At(i int) *session.Event {
	if i >= 0 && i < len(e.events) {
		return e.events[i]
	}
	return nil
}

// appendEvent adds an event to the session and updates session state from event deltas
func (s *localSession) appendEvent(event *session.Event) error {
	if event.Partial {
		return nil
	}
	processedEvent := trimTempDeltaState(event)
	if err := updateSessionState(s, processedEvent); err != nil {
		return fmt.Errorf("error updating session state from event: %w", err)
	}
	s.events = append(s.events, event)
	s.updatedAt = event.Timestamp
	return nil
}
