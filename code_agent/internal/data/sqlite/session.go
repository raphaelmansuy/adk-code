// Package sqlite provides SQLite-backed implementations of data repositories
package sqlite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	pkgerrors "code_agent/pkg/errors"

	"google.golang.org/adk/session"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
