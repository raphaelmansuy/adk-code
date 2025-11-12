// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package persistence

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

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
