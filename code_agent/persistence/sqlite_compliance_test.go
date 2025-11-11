package persistence

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/adk/session"
)

// TestStateHandlingCompliance verifies that the session persistence implementation
// correctly handles state prefixes according to the ADK reference implementation:
// - Prefixes are REMOVED when storing state deltas (app/user states stored without prefix)
// - Prefixes are RE-ADDED when merging states for client responses
// - Temporary keys (temp:*) are never stored
// - App/User state are shared across sessions/users, session state is unique
func TestStateHandlingCompliance(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := NewSQLiteSessionService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create session service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Create two sessions with overlapping state
	// This tests that app and user state are shared correctly
	createResp1, err := service.Create(ctx, &session.CreateRequest{
		AppName:   "myapp",
		UserID:    "user1",
		SessionID: "session1",
		State: map[string]any{
			"app:version":  "1.0",
			"user:theme":   "dark",
			"session_data": "session1_data",
			"temp:id":      "temp_value", // Should be ignored
		},
	})
	if err != nil {
		t.Fatalf("Failed to create session 1: %v", err)
	}

	// Create a second session for the same user
	createResp2, err := service.Create(ctx, &session.CreateRequest{
		AppName:   "myapp",
		UserID:    "user1",
		SessionID: "session2",
		State: map[string]any{
			"app:version":  "1.0",
			"user:theme":   "light", // Different user preference
			"session_data": "session2_data",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create session 2: %v", err)
	}

	// Verify session 1 initial state
	sess1State := createResp1.Session.State()
	appVersion, err := sess1State.Get("app:version")
	if err != nil || appVersion != "1.0" {
		t.Errorf("Session 1: expected app:version='1.0', got error=%v, value=%v", err, appVersion)
	}

	userTheme, err := sess1State.Get("user:theme")
	if err != nil || userTheme != "dark" {
		t.Errorf("Session 1: expected user:theme='dark', got error=%v, value=%v", err, userTheme)
	}

	sessionData, err := sess1State.Get("session_data")
	if err != nil || sessionData != "session1_data" {
		t.Errorf("Session 1: expected session_data='session1_data', got error=%v, value=%v", err, sessionData)
	}

	// Verify temporary key was not stored
	tempVal, err := sess1State.Get("temp:id")
	if err == nil {
		t.Errorf("Session 1: temporary key 'temp:id' should not exist, but got value=%v", tempVal)
	}
	if err != session.ErrStateKeyNotExist {
		t.Errorf("Session 1: expected ErrStateKeyNotExist for 'temp:id', got %v", err)
	}

	// Verify session 2 has different user state but same app state
	sess2State := createResp2.Session.State()
	appVersion2, err := sess2State.Get("app:version")
	if err != nil || appVersion2 != "1.0" {
		t.Errorf("Session 2: expected app:version='1.0', got error=%v, value=%v", err, appVersion2)
	}

	userTheme2, err := sess2State.Get("user:theme")
	if err != nil || userTheme2 != "light" {
		t.Errorf("Session 2: expected user:theme='light', got error=%v, value=%v", err, userTheme2)
	}

	sessionData2, err := sess2State.Get("session_data")
	if err != nil || sessionData2 != "session2_data" {
		t.Errorf("Session 2: expected session_data='session2_data', got error=%v, value=%v", err, sessionData2)
	}

	// Now append an event that modifies app and user state
	// This should update the shared app/user state, but only session1's session state
	event := &session.Event{
		ID:           "event1",
		InvocationID: "inv1",
		Author:       "agent",
		Timestamp:    time.Now(),
		Actions: session.EventActions{
			StateDelta: map[string]any{
				"app:version":  "1.1",                   // Update shared app state
				"user:theme":   "auto",                  // Update shared user state
				"session_data": "updated_session1_data", // Update session-specific state
				"temp:runtime": "discarded",             // Should be discarded
			},
		},
		LongRunningToolIDs: make([]string, 0),
	}

	err = service.AppendEvent(ctx, createResp1.Session, event)
	if err != nil {
		t.Fatalf("Failed to append event to session 1: %v", err)
	}

	// Retrieve session 1 and verify updates
	getResp1, err := service.Get(ctx, &session.GetRequest{
		AppName:   "myapp",
		UserID:    "user1",
		SessionID: "session1",
	})
	if err != nil {
		t.Fatalf("Failed to retrieve session 1: %v", err)
	}

	retrieved1State := getResp1.Session.State()

	// App state should be updated globally
	updatedAppVersion, err := retrieved1State.Get("app:version")
	if err != nil || updatedAppVersion != "1.1" {
		t.Errorf("Session 1 after event: expected app:version='1.1', got error=%v, value=%v", err, updatedAppVersion)
	}

	// User state should be updated globally
	updatedUserTheme, err := retrieved1State.Get("user:theme")
	if err != nil || updatedUserTheme != "auto" {
		t.Errorf("Session 1 after event: expected user:theme='auto', got error=%v, value=%v", err, updatedUserTheme)
	}

	// Session state should be updated for this session only
	updatedSessionData, err := retrieved1State.Get("session_data")
	if err != nil || updatedSessionData != "updated_session1_data" {
		t.Errorf("Session 1 after event: expected session_data='updated_session1_data', got error=%v, value=%v", err, updatedSessionData)
	}

	// Temporary key should not be stored
	tempVal, err = retrieved1State.Get("temp:runtime")
	if err == nil {
		t.Errorf("Session 1: temporary key 'temp:runtime' should not exist, but got value=%v", tempVal)
	}
	if err != session.ErrStateKeyNotExist {
		t.Errorf("Session 1: expected ErrStateKeyNotExist for 'temp:runtime', got %v", err)
	}

	// Retrieve session 2 and verify it sees the updated shared state but not the session1-specific update
	getResp2, err := service.Get(ctx, &session.GetRequest{
		AppName:   "myapp",
		UserID:    "user1",
		SessionID: "session2",
	})
	if err != nil {
		t.Fatalf("Failed to retrieve session 2: %v", err)
	}

	retrieved2State := getResp2.Session.State()

	// App state should be updated globally
	updatedAppVersion2, err := retrieved2State.Get("app:version")
	if err != nil || updatedAppVersion2 != "1.1" {
		t.Errorf("Session 2 after event: expected app:version='1.1', got error=%v, value=%v", err, updatedAppVersion2)
	}

	// User state should be updated globally
	updatedUserTheme2, err := retrieved2State.Get("user:theme")
	if err != nil || updatedUserTheme2 != "auto" {
		t.Errorf("Session 2 after event: expected user:theme='auto', got error=%v, value=%v", err, updatedUserTheme2)
	}

	// Session2's session state should NOT be affected by session1's event
	unchangedSessionData2, err := retrieved2State.Get("session_data")
	if err != nil || unchangedSessionData2 != "session2_data" {
		t.Errorf("Session 2 after session1 event: expected session_data='session2_data' (unchanged), got error=%v, value=%v", err, unchangedSessionData2)
	}
}
