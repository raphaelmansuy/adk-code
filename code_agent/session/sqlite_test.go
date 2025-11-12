package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/adk/session"
)

func TestSessionCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Create a session
	resp, err := service.Create(ctx, &session.CreateRequest{AppName: "test_app", UserID: "user1", SessionID: "session1"})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if resp.Session.ID() != "session1" {
		t.Errorf("Expected session ID 'session1', got '%s'", resp.Session.ID())
	}
}

func TestSessionCRUDAndPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Create a session
	_, err = service.Create(ctx, &session.CreateRequest{AppName: "test_app", UserID: "user1", SessionID: "session1"})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Retrieve the session
	getResp, err := service.Get(ctx, &session.GetRequest{AppName: "test_app", UserID: "user1", SessionID: "session1"})
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}
	if getResp.Session.ID() != "session1" {
		t.Fatalf("Session ID mismatch: %s", getResp.Session.ID())
	}

	// List sessions
	listResp, err := service.List(ctx, &session.ListRequest{AppName: "test_app", UserID: "user1"})
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}
	if len(listResp.Sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(listResp.Sessions))
	}

	// Delete session
	if err := service.Delete(ctx, &session.DeleteRequest{AppName: "test_app", UserID: "user1", SessionID: "session1"}); err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Should not find it
	if _, err := service.Get(ctx, &session.GetRequest{AppName: "test_app", UserID: "user1", SessionID: "session1"}); err == nil {
		t.Fatalf("Expected error when fetching deleted session, got nil")
	}

	// Persistence across instances
	service1, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service1: %v", err)
	}
	ctx = context.Background()
	_, err = service1.Create(ctx, &session.CreateRequest{AppName: "test_app", UserID: "user1", SessionID: "session_persist"})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	service1.Close()

	service2, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service2: %v", err)
	}
	defer service2.Close()
	if _, err := service2.Get(ctx, &session.GetRequest{AppName: "test_app", UserID: "user1", SessionID: "session_persist"}); err != nil {
		t.Fatalf("Failed to retrieve session persisted in new service: %v", err)
	}
}

func TestAppendAndStateUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	createResp, err := service.Create(ctx, &session.CreateRequest{AppName: "test_app", UserID: "user1", SessionID: "session1", State: map[string]any{"initial_key": "initial_value"}})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sess := createResp.Session

	evt := &session.Event{ID: generateSessionID(), InvocationID: "inv1", Author: "agent", Timestamp: time.Now(), Actions: session.EventActions{StateDelta: map[string]any{"new_key": "new_value"}}, LongRunningToolIDs: make([]string, 0)}
	if err := service.AppendEvent(ctx, sess, evt); err != nil {
		t.Fatalf("Failed to append event: %v", err)
	}

	getResp, err := service.Get(ctx, &session.GetRequest{AppName: "test_app", UserID: "user1", SessionID: "session1"})
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}
	if getResp.Session.Events().Len() != 1 {
		t.Fatalf("Expected 1 event, got %d", getResp.Session.Events().Len())
	}
	val, _ := getResp.Session.State().Get("new_key")
	if val != "new_value" {
		t.Fatalf("Expected state new_key 'new_value', got '%v'", val)
	}
}

func TestDefaultDBPathInManager(t *testing.T) {
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	manager, err := NewSessionManager("test_app", "")
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()
	expected := filepath.Join(tmpHome, ".code_agent", "sessions.db")
	if manager.GetDBPath() != expected {
		t.Fatalf("Expected db path %s got %s", expected, manager.GetDBPath())
	}
}
