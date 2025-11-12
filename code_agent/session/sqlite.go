// Package session provides session persistence using SQLite
package session

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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
		return nil, fmt.Errorf("failed to create database directory: %w", err)
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
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	service := &SQLiteSessionService{db: db}

	// Run migrations to ensure schema exists
	if err := service.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
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
		return nil, fmt.Errorf("app_name and user_id are required")
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
		return nil, fmt.Errorf("session already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check if session exists: %w", err)
	}

	// Prepare session state
	state := req.State
	if state == nil {
		state = make(map[string]any)
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
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
				return nil, fmt.Errorf("failed to create app state: %w", err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("failed to fetch app state: %w", err)
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
		return nil, fmt.Errorf("failed to save app state: %w", err)
	}

	// Get or create user state
	var userState storageUserState
	if err := tx.Where("app_name = ? AND user_id = ?", req.AppName, req.UserID).First(&userState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new user state
			userState = storageUserState{AppName: req.AppName, UserID: req.UserID, State: make(stateMap), UpdateTime: time.Now()}
			if err := tx.Create(&userState).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create user state: %w", err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("failed to fetch user state: %w", err)
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
		return nil, fmt.Errorf("failed to save user state: %w", err)
	}

	// Create session
	now := time.Now()
	storageSession := &storageSession{AppName: req.AppName, UserID: req.UserID, ID: sessionID, State: sessionState, CreateTime: now, UpdateTime: now}
	if err := tx.Create(storageSession).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Create response session
	localSession := &localSession{appName: req.AppName, userID: req.UserID, sessionID: sessionID, state: mergeStates(appState.State, userState.State, sessionState), updatedAt: now, events: make([]*session.Event, 0)}

	return &session.CreateResponse{Session: localSession}, nil
}

// Get retrieves a session
func (s *SQLiteSessionService) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	if req.AppName == "" || req.UserID == "" || req.SessionID == "" {
		return nil, fmt.Errorf("app_name, user_id, and session_id are required")
	}
	var storageSession storageSession
	if err := s.db.WithContext(ctx).
		Where("app_name = ? AND user_id = ? AND id = ?", req.AppName, req.UserID, req.SessionID).
		First(&storageSession).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to fetch session: %w", err)
	}
	var appState storageAppState
	if err := s.db.WithContext(ctx).
		Where("app_name = ?", req.AppName).
		First(&appState).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to fetch app state: %w", err)
	}
	var userState storageUserState
	if err := s.db.WithContext(ctx).
		Where("app_name = ? AND user_id = ?", req.AppName, req.UserID).
		First(&userState).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to fetch user state: %w", err)
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
		return nil, fmt.Errorf("failed to fetch events: %w", err)
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
			return nil, fmt.Errorf("failed to convert event: %w", err)
		}
		localSession.events[i] = evt
	}
	return &session.GetResponse{Session: localSession}, nil
}

// List lists sessions
func (s *SQLiteSessionService) List(ctx context.Context, req *session.ListRequest) (*session.ListResponse, error) {
	if req.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}
	var sessions []storageSession
	query := s.db.WithContext(ctx).Where("app_name = ?", req.AppName)
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
	if err := query.Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	var appState storageAppState
	if err := s.db.WithContext(ctx).Where("app_name = ?", req.AppName).First(&appState).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to fetch app state: %w", err)
	}
	response := &session.ListResponse{Sessions: make([]session.Session, len(sessions))}
	for i, sess := range sessions {
		var userState storageUserState
		if err := s.db.WithContext(ctx).Where("app_name = ? AND user_id = ?", req.AppName, sess.UserID).First(&userState).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to fetch user state: %w", err)
		}

		var events []storageEvent
		if err := s.db.WithContext(ctx).Where("app_name = ? AND user_id = ? AND session_id = ?", req.AppName, sess.UserID, sess.ID).Order("timestamp ASC").Find(&events).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch events: %w", err)
		}
		sessionEvents := make([]*session.Event, len(events))
		for j, e := range events {
			evt, err := convertStorageEventToSessionEvent(&e)
			if err != nil {
				return nil, fmt.Errorf("failed to convert event: %w", err)
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
		return fmt.Errorf("app_name, user_id, and session_id are required")
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	if err := tx.Where("app_name = ? AND user_id = ? AND session_id = ?", req.AppName, req.UserID, req.SessionID).Delete(&storageEvent{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete events: %w", err)
	}
	if err := tx.Where("app_name = ? AND user_id = ? AND id = ?", req.AppName, req.UserID, req.SessionID).Delete(&storageSession{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete session: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// AppendEvent appends an event to a session and updates session/app/user state
func (s *SQLiteSessionService) AppendEvent(ctx context.Context, sess session.Session, event *session.Event) error {
	if sess == nil {
		return fmt.Errorf("session is nil")
	}
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	if event.Partial {
		return nil
	}
	trimTempDeltaState(event)
	localSession, ok := sess.(*localSession)
	if !ok {
		return fmt.Errorf("unexpected session type: %T", sess)
	}
	if err := localSession.appendEvent(event); err != nil {
		return fmt.Errorf("failed to append event to in-memory session: %w", err)
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	storageEvent, err := convertSessionEventToStorageEvent(localSession, event)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to convert event: %w", err)
	}
	if err := tx.Create(storageEvent).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to append event: %w", err)
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
						return fmt.Errorf("failed to create app state: %w", err)
					}
				} else {
					tx.Rollback()
					return fmt.Errorf("failed to fetch app state: %w", err)
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
				return fmt.Errorf("failed to update app state: %w", err)
			}
		}
		if len(userDelta) > 0 {
			var userState storageUserState
			if err := tx.Where("app_name = ? AND user_id = ?", localSession.appName, localSession.userID).First(&userState).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					userState = storageUserState{AppName: localSession.appName, UserID: localSession.userID, State: make(stateMap), UpdateTime: time.Now()}
					if err := tx.Create(&userState).Error; err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to create user state: %w", err)
					}
				} else {
					tx.Rollback()
					return fmt.Errorf("failed to fetch user state: %w", err)
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
				return fmt.Errorf("failed to update user state: %w", err)
			}
		}
		if len(sessionDelta) > 0 {
			var currentSession storageSession
			if err := tx.Where("app_name = ? AND user_id = ? AND id = ?", localSession.appName, localSession.userID, localSession.sessionID).First(&currentSession).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to fetch session for state update: %w", err)
			}
			if currentSession.State == nil {
				currentSession.State = make(stateMap)
			}
			for k, v := range sessionDelta {
				currentSession.State[k] = v
			}
			if err := tx.Model(&storageSession{}).Where("app_name = ? AND user_id = ? AND id = ?", localSession.appName, localSession.userID, localSession.sessionID).Update("state", currentSession.State).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update session state: %w", err)
			}
		}
	}
	if err := tx.Model(&storageSession{}).Where("app_name = ? AND user_id = ? AND id = ?", localSession.appName, localSession.userID, localSession.sessionID).Update("update_time", time.Now()).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update session timestamp: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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
