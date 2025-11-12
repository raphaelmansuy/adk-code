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

package app

import (
	"context"
	"fmt"

	"code_agent/display"
	"code_agent/persistence"
)

// SessionInitializer handles session creation and retrieval
type SessionInitializer struct {
	manager        *persistence.SessionManager
	bannerRenderer *display.BannerRenderer
}

// NewSessionInitializer creates a new session initializer
func NewSessionInitializer(manager *persistence.SessionManager, bannerRenderer *display.BannerRenderer) *SessionInitializer {
	return &SessionInitializer{
		manager:        manager,
		bannerRenderer: bannerRenderer,
	}
}

// InitializeSession gets or creates a session
func (s *SessionInitializer) InitializeSession(ctx context.Context, userID, sessionName string) error {
	sess, err := s.manager.GetSession(ctx, userID, sessionName)
	if err != nil {
		// Session doesn't exist, create it
		_, err = s.manager.CreateSession(ctx, userID, sessionName)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		fmt.Printf("✨ Created new session: %s\n\n", sessionName)
	} else {
		// Use enhanced session resume header with event count and tokens
		resumeInfo := s.bannerRenderer.RenderSessionResumeInfo(sessionName, sess.Events().Len(), 0)
		fmt.Print(resumeInfo)
	}
	return nil
}
