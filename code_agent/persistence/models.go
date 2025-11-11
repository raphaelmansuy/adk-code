// Package persistence provides session persistence using SQLite
package persistence

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
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
func (s *localSession) ID() string {
	return s.sessionID
}

// AppName returns the app name
func (s *localSession) AppName() string {
	return s.appName
}

// UserID returns the user ID
func (s *localSession) UserID() string {
	return s.userID
}

// State returns the session state
func (s *localSession) State() session.State {
	return &localState{state: s.state}
}

// Events returns the events
func (s *localSession) Events() session.Events {
	return &localEvents{events: s.events}
}

// LastUpdateTime returns the last update time
func (s *localSession) LastUpdateTime() time.Time {
	return s.updatedAt
}

// localState implements session.State
type localState struct {
	state stateMap
}

// Get retrieves a value from state
func (s *localState) Get(key string) (any, error) {
	if val, ok := s.state[key]; ok {
		return val, nil
	}
	return nil, session.ErrStateKeyNotExist
}

// Set sets a value in state
func (s *localState) Set(key string, value any) error {
	s.state[key] = value
	return nil
}

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
type localEvents struct {
	events []*session.Event
}

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
func (e *localEvents) Len() int {
	return len(e.events)
}

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

	// Trim temporary keys from event before processing state
	processedEvent := trimTempDeltaState(event)

	// Update the session's state based on the event's state delta (temp keys already filtered)
	if err := updateSessionState(s, processedEvent); err != nil {
		return fmt.Errorf("error updating session state from event: %w", err)
	}

	s.events = append(s.events, event)
	s.updatedAt = event.Timestamp
	return nil
}

// Helper functions

func generateSessionID() string {
	return uuid.NewString()
}

func extractStateDeltas(state map[string]any) (map[string]any, map[string]any, map[string]any) {
	appDelta := make(map[string]any)
	userDelta := make(map[string]any)
	sessionState := make(map[string]any)

	if state == nil {
		return appDelta, userDelta, sessionState
	}

	const (
		keyPrefixApp  = "app:"
		keyPrefixUser = "user:"
		keyPrefixTemp = "temp:"
	)

	for key, value := range state {
		// Check for app-level state (remove prefix from key)
		if cleanKey, found := strings.CutPrefix(key, keyPrefixApp); found {
			appDelta[cleanKey] = value
		} else if cleanKey, found := strings.CutPrefix(key, keyPrefixUser); found {
			// Check for user-level state (remove prefix from key)
			userDelta[cleanKey] = value
		} else if !strings.HasPrefix(key, keyPrefixTemp) {
			// Only include session state if it's not a temporary key
			sessionState[key] = value
		}
	}

	return appDelta, userDelta, sessionState
}

func mergeStates(appState, userState, sessionState stateMap) stateMap {
	merged := make(stateMap)

	// Add session state first (no prefix)
	for k, v := range sessionState {
		merged[k] = v
	}

	// Add app state with re-added "app:" prefix
	for k, v := range appState {
		merged["app:"+k] = v
	}

	// Add user state with re-added "user:" prefix
	for k, v := range userState {
		merged["user:"+k] = v
	}

	return merged
}

func convertStorageEventToSessionEvent(se *storageEvent) (*session.Event, error) {
	var actions session.EventActions
	if len(se.Actions) > 0 {
		if err := json.Unmarshal(se.Actions, &actions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
		}
	}

	var content *genai.Content
	if len(se.Content) > 0 {
		if err := json.Unmarshal(se.Content, &content); err != nil {
			return nil, fmt.Errorf("failed to unmarshal content: %w", err)
		}
	}

	var groundingMetadata *genai.GroundingMetadata
	if len(se.GroundingMetadata) > 0 {
		if err := json.Unmarshal(se.GroundingMetadata, &groundingMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal grounding metadata: %w", err)
		}
	}

	var customMetadata map[string]any
	if len(se.CustomMetadata) > 0 {
		if err := json.Unmarshal(se.CustomMetadata, &customMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom metadata: %w", err)
		}
	}

	var usageMetadata *genai.GenerateContentResponseUsageMetadata
	if len(se.UsageMetadata) > 0 {
		if err := json.Unmarshal(se.UsageMetadata, &usageMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal usage metadata: %w", err)
		}
	}

	var citationMetadata *genai.CitationMetadata
	if len(se.CitationMetadata) > 0 {
		if err := json.Unmarshal(se.CitationMetadata, &citationMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal citation metadata: %w", err)
		}
	}

	var toolIDs []string
	if se.LongRunningToolIDsJSON != nil {
		if err := json.Unmarshal([]byte(se.LongRunningToolIDsJSON), &toolIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool IDs: %w", err)
		}
	}

	branch := ""
	if se.Branch != nil {
		branch = *se.Branch
	}
	errorCode := ""
	if se.ErrorCode != nil {
		errorCode = *se.ErrorCode
	}
	errorMessage := ""
	if se.ErrorMessage != nil {
		errorMessage = *se.ErrorMessage
	}
	partial := false
	if se.Partial != nil {
		partial = *se.Partial
	}
	turnComplete := false
	if se.TurnComplete != nil {
		turnComplete = *se.TurnComplete
	}
	interrupted := false
	if se.Interrupted != nil {
		interrupted = *se.Interrupted
	}

	event := &session.Event{
		ID:                 se.ID,
		InvocationID:       se.InvocationID,
		Author:             se.Author,
		Timestamp:          se.Timestamp,
		Actions:            actions,
		LongRunningToolIDs: toolIDs,
		Branch:             branch,
		LLMResponse: model.LLMResponse{
			Content:           content,
			GroundingMetadata: groundingMetadata,
			CustomMetadata:    customMetadata,
			UsageMetadata:     usageMetadata,
			CitationMetadata:  citationMetadata,
			ErrorCode:         errorCode,
			ErrorMessage:      errorMessage,
			Partial:           partial,
			TurnComplete:      turnComplete,
			Interrupted:       interrupted,
		},
	}

	return event, nil
}

func convertSessionEventToStorageEvent(sess *localSession, event *session.Event) (*storageEvent, error) {
	storageEv := &storageEvent{
		ID:           event.ID,
		InvocationID: event.InvocationID,
		Author:       event.Author,
		SessionID:    sess.sessionID,
		AppName:      sess.appName,
		UserID:       sess.userID,
		Timestamp:    event.Timestamp,
	}

	// Serialize actions
	actionsJSON, err := json.Marshal(event.Actions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal actions: %w", err)
	}
	storageEv.Actions = actionsJSON

	// Serialize tool IDs
	if len(event.LongRunningToolIDs) > 0 {
		toolIDsJSON, err := json.Marshal(event.LongRunningToolIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool IDs: %w", err)
		}
		storageEv.LongRunningToolIDsJSON = toolIDsJSON
	}

	// Handle optional fields
	if event.Branch != "" {
		storageEv.Branch = &event.Branch
	}
	if event.ErrorCode != "" {
		storageEv.ErrorCode = &event.ErrorCode
	}
	if event.ErrorMessage != "" {
		storageEv.ErrorMessage = &event.ErrorMessage
	}

	storageEv.Partial = &event.Partial
	storageEv.TurnComplete = &event.TurnComplete
	storageEv.Interrupted = &event.Interrupted

	// Serialize JSON fields
	if event.Content != nil {
		contentJSON, err := json.Marshal(event.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal content: %w", err)
		}
		storageEv.Content = contentJSON
	}
	if event.GroundingMetadata != nil {
		groundingJSON, err := json.Marshal(event.GroundingMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal grounding metadata: %w", err)
		}
		storageEv.GroundingMetadata = groundingJSON
	}
	if len(event.CustomMetadata) > 0 {
		customJSON, err := json.Marshal(event.CustomMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal custom metadata: %w", err)
		}
		storageEv.CustomMetadata = customJSON
	}
	if event.UsageMetadata != nil {
		usageJSON, err := json.Marshal(event.UsageMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal usage metadata: %w", err)
		}
		storageEv.UsageMetadata = usageJSON
	}
	if event.CitationMetadata != nil {
		citationJSON, err := json.Marshal(event.CitationMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal citation metadata: %w", err)
		}
		storageEv.CitationMetadata = citationJSON
	}

	return storageEv, nil
}

// trimTempDeltaState removes temporary (invocation-scoped) keys from an event's state delta.
// It mutates the event and returns it for convenience.
func trimTempDeltaState(event *session.Event) *session.Event {
	if event == nil || len(event.Actions.StateDelta) == 0 {
		return event
	}

	filtered := make(map[string]any)
	for key, value := range event.Actions.StateDelta {
		if !strings.HasPrefix(key, session.KeyPrefixTemp) {
			filtered[key] = value
		}
	}

	event.Actions.StateDelta = filtered
	return event
}

// updateSessionState updates the session state based on the event state delta.
// It applies all non-temporary state changes to the session's state map.
func updateSessionState(sess *localSession, event *session.Event) error {
	if event == nil || event.Actions.StateDelta == nil {
		return nil // Nothing to do
	}

	// Ensure the session state map is initialized
	if sess.state == nil {
		sess.state = make(stateMap)
	}

	// Apply each state delta entry to the session state
	for key, value := range event.Actions.StateDelta {
		// Skip temporary keys (should already be filtered, but be defensive)
		if strings.HasPrefix(key, session.KeyPrefixTemp) {
			continue
		}
		sess.state[key] = value
	}

	return nil
}
