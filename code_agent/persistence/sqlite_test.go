package persistence

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/adk/session"
)

func TestSessionCreation(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Create a session
	resp, err := service.Create(ctx, &session.CreateRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if resp.Session.ID() != "session1" {
		t.Errorf("Expected session ID 'session1', got '%s'", resp.Session.ID())
	}

	if resp.Session.AppName() != "test_app" {
		t.Errorf("Expected app name 'test_app', got '%s'", resp.Session.AppName())
	}

	if resp.Session.UserID() != "user1" {
		t.Errorf("Expected user ID 'user1', got '%s'", resp.Session.UserID())
	}
}

func TestSessionRetrieval(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Create a session
	createResp, err := service.Create(ctx, &session.CreateRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
		State:     map[string]any{"key": "value"},
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Retrieve the session
	getResp, err := service.Get(ctx, &session.GetRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}

	if getResp.Session.ID() != createResp.Session.ID() {
		t.Errorf("Expected session ID '%s', got '%s'", createResp.Session.ID(), getResp.Session.ID())
	}

	// Check state
	val, err := getResp.Session.State().Get("key")
	if err != nil {
		t.Errorf("Failed to get state value: %v", err)
	}

	if val != "value" {
		t.Errorf("Expected state value 'value', got '%v'", val)
	}
}

func TestSessionListing(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Create multiple sessions
	for i := 1; i <= 3; i++ {
		_, err := service.Create(ctx, &session.CreateRequest{
			AppName:   "test_app",
			UserID:    "user1",
			SessionID: "session" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatalf("Failed to create session %d: %v", i, err)
		}
	}

	// List sessions
	listResp, err := service.List(ctx, &session.ListRequest{
		AppName: "test_app",
		UserID:  "user1",
	})
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(listResp.Sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(listResp.Sessions))
	}
}

func TestSessionDeletion(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Create a session
	_, err = service.Create(ctx, &session.CreateRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Delete the session
	err = service.Delete(ctx, &session.DeleteRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Try to retrieve the deleted session (should fail)
	_, err = service.Get(ctx, &session.GetRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err == nil {
		t.Error("Expected error when retrieving deleted session, got nil")
	}
}

func TestSessionPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create and close first service
	service1, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}

	ctx := context.Background()

	// Create a session
	_, err = service1.Create(ctx, &session.CreateRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	service1.Close()

	// Create new service with same database
	service2, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service2.Close()

	// Try to retrieve the session (should exist)
	getResp, err := service2.Get(ctx, &session.GetRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to retrieve session from new service: %v", err)
	}

	if getResp.Session.ID() != "session1" {
		t.Errorf("Expected session ID 'session1', got '%s'", getResp.Session.ID())
	}
}

func TestAppendEvent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Create a session
	createResp, err := service.Create(ctx, &session.CreateRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	sess := createResp.Session

	// Create an event
	event := &session.Event{
		ID:           generateSessionID(),
		InvocationID: "inv1",
		Author:       "agent",
		Timestamp:    time.Now(),
		Actions: session.EventActions{
			StateDelta: make(map[string]any),
		},
		LongRunningToolIDs: make([]string, 0),
	}

	// Append event
	err = service.AppendEvent(ctx, sess, event)
	if err != nil {
		t.Fatalf("Failed to append event: %v", err)
	}

	// Retrieve session and verify event was added
	getResp, err := service.Get(ctx, &session.GetRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}

	if getResp.Session.Events().Len() != 1 {
		t.Errorf("Expected 1 event, got %d", getResp.Session.Events().Len())
	}
}

func TestSessionManager(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	manager, err := NewSessionManager("test_app", dbPath)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Create a session
	sess, err := manager.CreateSession(ctx, "user1", "my-session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if sess.ID() != "my-session" {
		t.Errorf("Expected session ID 'my-session', got '%s'", sess.ID())
	}

	// List sessions
	sessions, err := manager.ListSessions(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	// Get session
	retrievedSess, err := manager.GetSession(ctx, "user1", "my-session")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrievedSess.ID() != "my-session" {
		t.Errorf("Expected session ID 'my-session', got '%s'", retrievedSess.ID())
	}

	// Delete session
	err = manager.DeleteSession(ctx, "user1", "my-session")
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify deletion
	sessions, err = manager.ListSessions(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to list sessions after deletion: %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions after deletion, got %d", len(sessions))
	}
}

func TestDatabasePathDefault(t *testing.T) {
	// Save and restore home directory
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Create a temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)

	// Create manager without specifying db path
	manager, err := NewSessionManager("test_app", "")
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	// Check that database path is in the right location
	expectedPath := filepath.Join(tmpHome, ".code_agent", "sessions.db")
	if manager.GetDBPath() != expectedPath {
		t.Errorf("Expected database path '%s', got '%s'", expectedPath, manager.GetDBPath())
	}
}

func TestTemporaryStateKeyFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Create a session
	createResp, err := service.Create(ctx, &session.CreateRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	sess := createResp.Session

	// Create an event with both temporary and persistent state keys
	event := &session.Event{
		ID:           generateSessionID(),
		InvocationID: "inv1",
		Author:       "agent",
		Timestamp:    time.Now(),
		Actions: session.EventActions{
			StateDelta: map[string]any{
				"persistent_key":      "should_persist",
				"temp:invocation_var": "should_not_persist",
				"another_persistent":  "also_persist",
				"temp:another_temp":   "also_not_persist",
			},
		},
		LongRunningToolIDs: make([]string, 0),
	}

	// Append event
	err = service.AppendEvent(ctx, sess, event)
	if err != nil {
		t.Fatalf("Failed to append event: %v", err)
	}

	// Retrieve session and verify the event was added
	getResp, err := service.Get(ctx, &session.GetRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}

	if getResp.Session.Events().Len() != 1 {
		t.Errorf("Expected 1 event, got %d", getResp.Session.Events().Len())
	}

	// Verify the persisted event has only non-temporary keys in its state delta
	persistedEvent := getResp.Session.Events().At(0)
	if persistedEvent == nil {
		t.Fatalf("Event is nil")
	}

	// Verify temporary keys were filtered
	if _, hasTempKey := persistedEvent.Actions.StateDelta["temp:invocation_var"]; hasTempKey {
		t.Errorf("Temporary key 'temp:invocation_var' should not be persisted, but was found")
	}
	if _, hasTempKey := persistedEvent.Actions.StateDelta["temp:another_temp"]; hasTempKey {
		t.Errorf("Temporary key 'temp:another_temp' should not be persisted, but was found")
	}

	// Verify persistent keys were kept
	if val, ok := persistedEvent.Actions.StateDelta["persistent_key"]; !ok {
		t.Errorf("Persistent key 'persistent_key' should have been persisted, but was not found")
	} else if val != "should_persist" {
		t.Errorf("Expected 'persistent_key' value 'should_persist', got '%v'", val)
	}

	if val, ok := persistedEvent.Actions.StateDelta["another_persistent"]; !ok {
		t.Errorf("Persistent key 'another_persistent' should have been persisted, but was not found")
	} else if val != "also_persist" {
		t.Errorf("Expected 'another_persistent' value 'also_persist', got '%v'", val)
	}
}

func TestSessionStateUpdateOnAppendEvent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Create a session with initial state
	createResp, err := service.Create(ctx, &session.CreateRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
		State:     map[string]any{"initial_key": "initial_value"},
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	sess := createResp.Session

	// Verify initial state is set
	val, err := sess.State().Get("initial_key")
	if err != nil {
		t.Errorf("Failed to get initial state: %v", err)
	}
	if val != "initial_value" {
		t.Errorf("Expected initial_key 'initial_value', got '%v'", val)
	}

	// Create an event with state deltas
	event := &session.Event{
		ID:           generateSessionID(),
		InvocationID: "inv1",
		Author:       "agent",
		Timestamp:    time.Now(),
		Actions: session.EventActions{
			StateDelta: map[string]any{
				"new_key":     "new_value",
				"updated_key": "updated_value",
				"temp:ignore": "should_not_persist",
			},
		},
		LongRunningToolIDs: make([]string, 0),
	}

	// Append event
	err = service.AppendEvent(ctx, sess, event)
	if err != nil {
		t.Fatalf("Failed to append event: %v", err)
	}

	// Retrieve session and verify state was updated
	getResp, err := service.Get(ctx, &session.GetRequest{
		AppName:   "test_app",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}

	retrievedSession := getResp.Session

	// Verify initial state still exists
	val, err = retrievedSession.State().Get("initial_key")
	if err != nil {
		t.Errorf("Failed to get initial_key: %v", err)
	}
	if val != "initial_value" {
		t.Errorf("Expected initial_key 'initial_value', got '%v'", val)
	}

	// Verify new key was added from event
	val, err = retrievedSession.State().Get("new_key")
	if err != nil {
		t.Errorf("Failed to get new_key: %v", err)
	}
	if val != "new_value" {
		t.Errorf("Expected new_key 'new_value', got '%v'", val)
	}

	// Verify updated key was added from event
	val, err = retrievedSession.State().Get("updated_key")
	if err != nil {
		t.Errorf("Failed to get updated_key: %v", err)
	}
	if val != "updated_value" {
		t.Errorf("Expected updated_key 'updated_value', got '%v'", val)
	}

	// Verify temporary key was NOT persisted
	_, err = retrievedSession.State().Get("temp:ignore")
	if err == nil {
		t.Errorf("Temporary key 'temp:ignore' should not have been persisted, but was found")
	}
	if err != session.ErrStateKeyNotExist {
		t.Errorf("Expected ErrStateKeyNotExist for 'temp:ignore', got %v", err)
	}
}
