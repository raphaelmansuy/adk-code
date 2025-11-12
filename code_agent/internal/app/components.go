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
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"

	"code_agent/display"
	"code_agent/persistence"
	"code_agent/pkg/models"
	"code_agent/tracking"
)

// DisplayComponents groups all display-related fields
type DisplayComponents struct {
	Renderer       *display.Renderer
	BannerRenderer *display.BannerRenderer
	Typewriter     *display.TypewriterPrinter
	StreamDisplay  *display.StreamingDisplay
}

// ModelComponents groups all model-related fields
type ModelComponents struct {
	Registry *models.Registry
	Selected models.Config
	LLM      model.LLM
}

// SessionComponents groups all session-related fields
type SessionComponents struct {
	Manager *persistence.SessionManager
	Runner  *runner.Runner
	Tokens  *tracking.SessionTokens
}
