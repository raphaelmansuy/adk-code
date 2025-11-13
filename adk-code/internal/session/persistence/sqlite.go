// Package persistence provides session persistence implementations
package persistence

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	pkgerrors "adk-code/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
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
		return pkgerrors.InternalError(fmt.Sprintf("failed to unmarshal JSON value: %T", value))
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
		return pkgerrors.InternalError(fmt.Sprintf("failed to unmarshal JSON value: %T", value))
	}

	if !json.Valid(bytes) {
		return pkgerrors.InvalidInputError(fmt.Sprintf("invalid JSON received from database: %s", string(bytes)))
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
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "error updating session state from event", err)
	}
	s.events = append(s.events, event)
	s.updatedAt = event.Timestamp
	return nil
}

// SQLiteSessionService provides SQLite-backed session persistence
type SQLiteSessionService struct {
	db *gorm.DB
}

// NewSQLiteSessionService creates a new SQLite-backed session service
// dbPath is the file path to the SQLite database (e.g., ~/.code_agent/sessions.db)
func NewSQLiteSessionService(dbPath string) (*SQLiteSessionService, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodePermission, "failed to create database directory", err)
	}

	// Configure GORM logger to only show errors (not warnings like "record not found")
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel:                  logger.Error, // Only log actual errors
			IgnoreRecordNotFoundError: true,         // Don't log "record not found"
			Colorful:                  false,
		},
	)

	// Open SQLite connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to open SQLite database", err)
	}

	service := &SQLiteSessionService{db: db}

	// Run migrations to ensure schema exists
	if err := service.migrate(); err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to run migrations", err)
	}

	return service, nil
}

// migrate creates the database schema if it doesn't exist
func (s *SQLiteSessionService) migrate() error {
	return s.db.AutoMigrate(
		&storageSession{},
		&storageEvent{},
		&storageAppState{},
		&storageUserState{},
	)
}

// Create creates a new session
func (s *SQLiteSessionService) Create(ctx context.Context, req *session.CreateRequest) (*session.CreateResponse, error) {
	if req.AppName == "" || req.UserID == "" {
		return nil, pkgerrors.InvalidInputError("app_name and user_id are required")
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	// Check if session already exists
	var existing storageSession
	if err := s.db.WithContext(ctx).
		Where("app_name = ? AND user_id = ? AND id = ?", req.AppName, req.UserID, sessionID).
		First(&existing).Error; err == nil {
		return nil, pkgerrors.InvalidInputError("session already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to check if session exists", err)
	}

	// Prepare session state
	state := req.State
	if state == nil {
		state = make(map[string]any)
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to start transaction", tx.Error)
	}

	// Extract state deltas
	appDelta, userDelta, sessionState := extractStateDeltas(state)

	// Get or create app state
	var appState storageAppState
	if err := tx.Where("app_name = ?", req.AppName).First(&appState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new app state
			appState = storageAppState{AppName: req.AppName, State: make(stateMap), UpdateTime: time.Now()}
			if err := tx.Create(&appState).Error; err != nil {
				tx.Rollback()
				return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create app state", err)
			}
		} else {
			tx.Rollback()
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch app state", err)
		}
	}

	// Merge app state delta
	if appState.State == nil {
		appState.State = make(stateMap)
	}
	for k, v := range appDelta {
		appState.State[k] = v
	}
	appState.UpdateTime = time.Now()
	if err := tx.Save(&appState).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to save app state", err)
	}

	// Get or create user state
	var userState storageUserState
	if err := tx.Where("app_name = ? AND user_id = ?", req.AppName, req.UserID).First(&userState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new user state
			userState = storageUserState{AppName: req.AppName, UserID: req.UserID, State: make(stateMap), UpdateTime: time.Now()}
			if err := tx.Create(&userState).Error; err != nil {
				tx.Rollback()
				return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create user state", err)
			}
		} else {
			tx.Rollback()
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch user state", err)
		}
	}

	// Merge user state delta
	if userState.State == nil {
		userState.State = make(stateMap)
	}
	for k, v := range userDelta {
		userState.State[k] = v
	}
	userState.UpdateTime = time.Now()
	if err := tx.Save(&userState).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to save user state", err)
	}

	// Create session
	now := time.Now()
	storageSession := &storageSession{AppName: req.AppName, UserID: req.UserID, ID: sessionID, State: sessionState, CreateTime: now, UpdateTime: now}
	if err := tx.Create(storageSession).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create session", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to commit transaction", err)
	}

	// Create response session
	localSession := &localSession{appName: req.AppName, userID: req.UserID, sessionID: sessionID, state: mergeStates(appState.State, userState.State, sessionState), updatedAt: now, events: make([]*session.Event, 0)}

	return &session.CreateResponse{Session: localSession}, nil
}

// Get retrieves a session
func (s *SQLiteSessionService) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	if req.AppName == "" || req.UserID == "" || req.SessionID == "" {
		return nil, pkgerrors.InvalidInputError("app_name, user_id, and session_id are required")
	}
	var storageSession storageSession
	if err := s.db.WithContext(ctx).
		Where("app_name = ? AND user_id = ? AND id = ?", req.AppName, req.UserID, req.SessionID).
		First(&storageSession).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.InvalidInputError("session not found")
		}
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch session", err)
	}
	var appState storageAppState
	if err := s.db.WithContext(ctx).
		Where("app_name = ?", req.AppName).
		First(&appState).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch app state", err)
	}
	var userState storageUserState
	if err := s.db.WithContext(ctx).
		Where("app_name = ? AND user_id = ?", req.AppName, req.UserID).
		First(&userState).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch user state", err)
	}
	var events []storageEvent
	eventQuery := s.db.WithContext(ctx).
		Where("app_name = ? AND user_id = ? AND session_id = ?", req.AppName, req.UserID, req.SessionID)
	if !req.After.IsZero() {
		eventQuery = eventQuery.Where("timestamp >= ?", req.After)
	}
	if req.NumRecentEvents > 0 {
		eventQuery = eventQuery.Order("timestamp DESC").Limit(req.NumRecentEvents)
	} else {
		eventQuery = eventQuery.Order("timestamp ASC")
	}
	if err := eventQuery.Find(&events).Error; err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch events", err)
	}
	if req.NumRecentEvents > 0 {
		for i := len(events)/2 - 1; i >= 0; i-- {
			opp := len(events) - 1 - i
			events[i], events[opp] = events[opp], events[i]
		}
	}
	localSession := &localSession{appName: req.AppName, userID: req.UserID, sessionID: req.SessionID, state: mergeStates(appState.State, userState.State, storageSession.State), updatedAt: storageSession.UpdateTime, events: make([]*session.Event, len(events))}
	for i, e := range events {
		evt, err := convertStorageEventToSessionEvent(&e)
		if err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to convert event", err)
		}
		localSession.events[i] = evt
	}
	return &session.GetResponse{Session: localSession}, nil
}

// List lists sessions
func (s *SQLiteSessionService) List(ctx context.Context, req *session.ListRequest) (*session.ListResponse, error) {
	if req.AppName == "" {
		return nil, pkgerrors.InvalidInputError("app_name is required")
	}
	var sessions []storageSession
	query := s.db.WithContext(ctx).Where("app_name = ?", req.AppName)
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
	if err := query.Find(&sessions).Error; err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to list sessions", err)
	}
	var appState storageAppState
	if err := s.db.WithContext(ctx).Where("app_name = ?", req.AppName).First(&appState).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch app state", err)
	}
	response := &session.ListResponse{Sessions: make([]session.Session, len(sessions))}
	for i, sess := range sessions {
		var userState storageUserState
		if err := s.db.WithContext(ctx).Where("app_name = ? AND user_id = ?", req.AppName, sess.UserID).First(&userState).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch user state", err)
		}

		var events []storageEvent
		if err := s.db.WithContext(ctx).Where("app_name = ? AND user_id = ? AND session_id = ?", req.AppName, sess.UserID, sess.ID).Order("timestamp ASC").Find(&events).Error; err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch events", err)
		}
		sessionEvents := make([]*session.Event, len(events))
		for j, e := range events {
			evt, err := convertStorageEventToSessionEvent(&e)
			if err != nil {
				return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to convert event", err)
			}
			sessionEvents[j] = evt
		}
		localSession := &localSession{
			appName:   sess.AppName,
			userID:    sess.UserID,
			sessionID: sess.ID,
			state:     mergeStates(appState.State, userState.State, sess.State),
			updatedAt: sess.UpdateTime,
			events:    sessionEvents,
		}
		response.Sessions[i] = localSession
	}
	return response, nil
}

// Delete deletes a session
func (s *SQLiteSessionService) Delete(ctx context.Context, req *session.DeleteRequest) error {
	if req.AppName == "" || req.UserID == "" || req.SessionID == "" {
		return pkgerrors.InvalidInputError("app_name, user_id, and session_id are required")
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to start transaction", tx.Error)
	}
	if err := tx.Where("app_name = ? AND user_id = ? AND session_id = ?", req.AppName, req.UserID, req.SessionID).Delete(&storageEvent{}).Error; err != nil {
		tx.Rollback()
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to delete events", err)
	}
	if err := tx.Where("app_name = ? AND user_id = ? AND id = ?", req.AppName, req.UserID, req.SessionID).Delete(&storageSession{}).Error; err != nil {
		tx.Rollback()
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to delete session", err)
	}
	if err := tx.Commit().Error; err != nil {
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to commit transaction", err)
	}
	return nil
}

// AppendEvent appends an event to a session and updates session/app/user state
func (s *SQLiteSessionService) AppendEvent(ctx context.Context, sess session.Session, event *session.Event) error {
	if sess == nil {
		return pkgerrors.InvalidInputError("session is nil")
	}
	if event == nil {
		return pkgerrors.InvalidInputError("event is nil")
	}
	if event.Partial {
		return nil
	}
	trimTempDeltaState(event)
	localSession, ok := sess.(*localSession)
	if !ok {
		return pkgerrors.InternalError(fmt.Sprintf("unexpected session type: %T", sess))
	}
	if err := localSession.appendEvent(event); err != nil {
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to append event to in-memory session", err)
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to start transaction", tx.Error)
	}
	storageEvent, err := convertSessionEventToStorageEvent(localSession, event)
	if err != nil {
		tx.Rollback()
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to convert event", err)
	}
	if err := tx.Create(storageEvent).Error; err != nil {
		tx.Rollback()
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to append event", err)
	}
	if len(event.Actions.StateDelta) > 0 {
		appDelta, userDelta, sessionDelta := extractStateDeltas(event.Actions.StateDelta)
		if len(appDelta) > 0 {
			var appState storageAppState
			if err := tx.Where("app_name = ?", localSession.appName).First(&appState).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					appState = storageAppState{AppName: localSession.appName, State: make(stateMap), UpdateTime: time.Now()}
					if err := tx.Create(&appState).Error; err != nil {
						tx.Rollback()
						return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create app state", err)
					}
				} else {
					tx.Rollback()
					return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch app state", err)
				}
			}
			if appState.State == nil {
				appState.State = make(stateMap)
			}
			for k, v := range appDelta {
				appState.State[k] = v
			}
			appState.UpdateTime = time.Now()
			if err := tx.Save(&appState).Error; err != nil {
				tx.Rollback()
				return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to update app state", err)
			}
		}
		if len(userDelta) > 0 {
			var userState storageUserState
			if err := tx.Where("app_name = ? AND user_id = ?", localSession.appName, localSession.userID).First(&userState).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					userState = storageUserState{AppName: localSession.appName, UserID: localSession.userID, State: make(stateMap), UpdateTime: time.Now()}
					if err := tx.Create(&userState).Error; err != nil {
						tx.Rollback()
						return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create user state", err)
					}
				} else {
					tx.Rollback()
					return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch user state", err)
				}
			}
			if userState.State == nil {
				userState.State = make(stateMap)
			}
			for k, v := range userDelta {
				userState.State[k] = v
			}
			userState.UpdateTime = time.Now()
			if err := tx.Save(&userState).Error; err != nil {
				tx.Rollback()
				return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to update user state", err)
			}
		}
		if len(sessionDelta) > 0 {
			var currentSession storageSession
			if err := tx.Where("app_name = ? AND user_id = ? AND id = ?", localSession.appName, localSession.userID, localSession.sessionID).First(&currentSession).Error; err != nil {
				tx.Rollback()
				return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to fetch session for state update", err)
			}
			if currentSession.State == nil {
				currentSession.State = make(stateMap)
			}
			for k, v := range sessionDelta {
				currentSession.State[k] = v
			}
			if err := tx.Model(&storageSession{}).Where("app_name = ? AND user_id = ? AND id = ?", localSession.appName, localSession.userID, localSession.sessionID).Update("state", currentSession.State).Error; err != nil {
				tx.Rollback()
				return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to update session state", err)
			}
		}
	}
	if err := tx.Model(&storageSession{}).Where("app_name = ? AND user_id = ? AND id = ?", localSession.appName, localSession.userID, localSession.sessionID).Update("update_time", time.Now()).Error; err != nil {
		tx.Rollback()
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to update session timestamp", err)
	}
	if err := tx.Commit().Error; err != nil {
		return pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to commit transaction", err)
	}
	return nil
}

// Close closes the database connection
func (s *SQLiteSessionService) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
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
		if cleanKey, found := strings.CutPrefix(key, keyPrefixApp); found {
			appDelta[cleanKey] = value
		} else if cleanKey, found := strings.CutPrefix(key, keyPrefixUser); found {
			userDelta[cleanKey] = value
		} else if !strings.HasPrefix(key, keyPrefixTemp) {
			sessionState[key] = value
		}
	}
	return appDelta, userDelta, sessionState
}

func mergeStates(appState, userState, sessionState stateMap) stateMap {
	merged := make(stateMap)
	for k, v := range sessionState {
		merged[k] = v
	}
	for k, v := range appState {
		merged["app:"+k] = v
	}
	for k, v := range userState {
		merged["user:"+k] = v
	}
	return merged
}

func convertStorageEventToSessionEvent(se *storageEvent) (*session.Event, error) {
	var actions session.EventActions
	if len(se.Actions) > 0 {
		if err := json.Unmarshal(se.Actions, &actions); err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to unmarshal actions", err)
		}
	}
	var content *genai.Content
	if len(se.Content) > 0 {
		if err := json.Unmarshal(se.Content, &content); err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to unmarshal content", err)
		}
	}
	var groundingMetadata *genai.GroundingMetadata
	if len(se.GroundingMetadata) > 0 {
		if err := json.Unmarshal(se.GroundingMetadata, &groundingMetadata); err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to unmarshal grounding metadata", err)
		}
	}
	var customMetadata map[string]any
	if len(se.CustomMetadata) > 0 {
		if err := json.Unmarshal(se.CustomMetadata, &customMetadata); err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to unmarshal custom metadata", err)
		}
	}
	var usageMetadata *genai.GenerateContentResponseUsageMetadata
	if len(se.UsageMetadata) > 0 {
		if err := json.Unmarshal(se.UsageMetadata, &usageMetadata); err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to unmarshal usage metadata", err)
		}
	}
	var citationMetadata *genai.CitationMetadata
	if len(se.CitationMetadata) > 0 {
		if err := json.Unmarshal(se.CitationMetadata, &citationMetadata); err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to unmarshal citation metadata", err)
		}
	}
	var toolIDs []string
	if se.LongRunningToolIDsJSON != nil {
		if err := json.Unmarshal([]byte(se.LongRunningToolIDsJSON), &toolIDs); err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to unmarshal tool IDs", err)
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
	event := &session.Event{ID: se.ID, InvocationID: se.InvocationID, Author: se.Author, Timestamp: se.Timestamp, Actions: actions, LongRunningToolIDs: toolIDs, Branch: branch, LLMResponse: model.LLMResponse{Content: content, GroundingMetadata: groundingMetadata, CustomMetadata: customMetadata, UsageMetadata: usageMetadata, CitationMetadata: citationMetadata, ErrorCode: errorCode, ErrorMessage: errorMessage, Partial: partial, TurnComplete: turnComplete, Interrupted: interrupted}}
	return event, nil
}

func convertSessionEventToStorageEvent(sess *localSession, event *session.Event) (*storageEvent, error) {
	storageEv := &storageEvent{ID: event.ID, InvocationID: event.InvocationID, Author: event.Author, SessionID: sess.sessionID, AppName: sess.appName, UserID: sess.userID, Timestamp: event.Timestamp}
	actionsJSON, err := json.Marshal(event.Actions)
	if err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to marshal actions", err)
	}
	storageEv.Actions = actionsJSON
	if len(event.LongRunningToolIDs) > 0 {
		toolIDsJSON, err := json.Marshal(event.LongRunningToolIDs)
		if err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to marshal tool IDs", err)
		}
		storageEv.LongRunningToolIDsJSON = toolIDsJSON
	}
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
	if event.Content != nil {
		contentJSON, err := json.Marshal(event.Content)
		if err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to marshal content", err)
		}
		storageEv.Content = contentJSON
	}
	if event.GroundingMetadata != nil {
		groundingJSON, err := json.Marshal(event.GroundingMetadata)
		if err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to marshal grounding metadata", err)
		}
		storageEv.GroundingMetadata = groundingJSON
	}
	if len(event.CustomMetadata) > 0 {
		customJSON, err := json.Marshal(event.CustomMetadata)
		if err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to marshal custom metadata", err)
		}
		storageEv.CustomMetadata = customJSON
	}
	if event.UsageMetadata != nil {
		usageJSON, err := json.Marshal(event.UsageMetadata)
		if err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to marshal usage metadata", err)
		}
		storageEv.UsageMetadata = usageJSON
	}
	if event.CitationMetadata != nil {
		citationJSON, err := json.Marshal(event.CitationMetadata)
		if err != nil {
			return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to marshal citation metadata", err)
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
func updateSessionState(sess *localSession, event *session.Event) error {
	if event == nil || event.Actions.StateDelta == nil {
		return nil
	}
	if sess.state == nil {
		sess.state = make(stateMap)
	}
	for key, value := range event.Actions.StateDelta {
		if strings.HasPrefix(key, session.KeyPrefixTemp) {
			continue
		}
		sess.state[key] = value
	}
	return nil
}
