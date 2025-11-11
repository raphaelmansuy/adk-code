package persistence

import (
	"testing"
)

// TestExtractStateDeltasFunction tests the extractStateDeltas function directly
func TestExtractStateDeltasFunction(t *testing.T) {
	state := map[string]any{
		"app:version":        1,
		"app:config":         "test",
		"user:name":          "Alice",
		"user:count":         10,
		"session_data":       "value",
		"temp:invocation_id": "should_be_excluded",
	}

	appDelta, userDelta, sessionDelta := extractStateDeltas(state)

	// Verify app delta has unprefixed keys
	if val, ok := appDelta["version"]; !ok {
		t.Error("appDelta should have 'version' key (without prefix)")
	} else if val != 1 {
		t.Errorf("Expected appDelta['version']=1, got %v", val)
	}

	if val, ok := appDelta["config"]; !ok {
		t.Error("appDelta should have 'config' key (without prefix)")
	} else if val != "test" {
		t.Errorf("Expected appDelta['config']='test', got %v", val)
	}

	if _, ok := appDelta["app:version"]; ok {
		t.Error("appDelta should NOT have 'app:version' key (prefix should be removed)")
	}

	// Verify user delta has unprefixed keys
	if val, ok := userDelta["name"]; !ok {
		t.Error("userDelta should have 'name' key (without prefix)")
	} else if val != "Alice" {
		t.Errorf("Expected userDelta['name']='Alice', got %v", val)
	}

	if val, ok := userDelta["count"]; !ok {
		t.Error("userDelta should have 'count' key (without prefix)")
	} else if val != 10 {
		t.Errorf("Expected userDelta['count']=10, got %v", val)
	}

	if _, ok := userDelta["user:name"]; ok {
		t.Error("userDelta should NOT have 'user:name' key (prefix should be removed)")
	}

	// Verify session delta
	if val, ok := sessionDelta["session_data"]; !ok {
		t.Error("sessionDelta should have 'session_data' key")
	} else if val != "value" {
		t.Errorf("Expected sessionDelta['session_data']='value', got %v", val)
	}

	// Verify temp keys are excluded
	if _, ok := sessionDelta["temp:invocation_id"]; ok {
		t.Error("sessionDelta should NOT have 'temp:invocation_id' key")
	}
}

// TestMergeStatesFunction tests the mergeStates function directly
func TestMergeStatesFunction(t *testing.T) {
	appState := map[string]any{
		"version": 1,
		"config":  "test",
	}

	userState := map[string]any{
		"name":  "Alice",
		"count": 10,
	}

	sessionState := map[string]any{
		"session_data": "value",
	}

	merged := mergeStates(appState, userState, sessionState)

	// Verify session state keys have no prefix
	if val, ok := merged["session_data"]; !ok {
		t.Error("merged should have 'session_data' key")
	} else if val != "value" {
		t.Errorf("Expected merged['session_data']='value', got %v", val)
	}

	// Verify app state keys have re-added prefix
	if val, ok := merged["app:version"]; !ok {
		t.Error("merged should have 'app:version' key with prefix")
	} else if val != 1 {
		t.Errorf("Expected merged['app:version']=1, got %v", val)
	}

	if val, ok := merged["app:config"]; !ok {
		t.Error("merged should have 'app:config' key with prefix")
	} else if val != "test" {
		t.Errorf("Expected merged['app:config']='test', got %v", val)
	}

	// Verify user state keys have re-added prefix
	if val, ok := merged["user:name"]; !ok {
		t.Error("merged should have 'user:name' key with prefix")
	} else if val != "Alice" {
		t.Errorf("Expected merged['user:name']='Alice', got %v", val)
	}

	if val, ok := merged["user:count"]; !ok {
		t.Error("merged should have 'user:count' key with prefix")
	} else if val != 10 {
		t.Errorf("Expected merged['user:count']=10, got %v", val)
	}

	// Verify that unprefixed keys from app/user don't exist
	if _, ok := merged["version"]; ok {
		t.Error("merged should NOT have 'version' key (should have 'app:version')")
	}

	if _, ok := merged["name"]; ok {
		t.Error("merged should NOT have 'name' key (should have 'user:name')")
	}
}
