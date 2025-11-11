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

package main

import (
	"testing"
)

func TestModelRegistry(t *testing.T) {
	registry := NewModelRegistry()

	tests := []struct {
		name          string
		modelID       string
		expectedFound bool
	}{
		{"gemini-2.5-flash", "gemini-2.5-flash", true},
		{"gemini-1.5-pro", "gemini-1.5-pro", true},
		{"gemini-2.5-flash-vertex", "gemini-2.5-flash-vertex", true},
		{"nonexistent", "nonexistent-model", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := registry.GetModel(tt.modelID)
			if tt.expectedFound {
				if err != nil {
					t.Errorf("Expected model %s to be found, got error: %v", tt.modelID, err)
				}
				if model.ID != tt.modelID {
					t.Errorf("Expected model ID %s, got %s", tt.modelID, model.ID)
				}
			} else {
				if err == nil {
					t.Errorf("Expected model %s to not be found, but it was", tt.modelID)
				}
			}
		})
	}
}

func TestModelResolve(t *testing.T) {
	registry := NewModelRegistry()

	tests := []struct {
		name            string
		modelID         string
		backend         string
		expectedBackend string
		shouldSucceed   bool
	}{
		{"explicit-model", "gemini-1.5-pro", "", "gemini", true},
		{"explicit-backend", "", "vertexai", "vertexai", true},
		{"both-specified", "gemini-1.5-pro-vertex", "vertexai", "vertexai", true},
		{"neither", "", "", "gemini", true}, // Should use default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := registry.ResolveModel(tt.modelID, tt.backend)
			if !tt.shouldSucceed {
				if err == nil {
					t.Errorf("Expected resolve to fail")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if model.Backend != tt.expectedBackend {
				t.Errorf("Expected backend %s, got %s", tt.expectedBackend, model.Backend)
			}
		})
	}
}

func TestListModelsByBackend(t *testing.T) {
	registry := NewModelRegistry()

	geminiBakcend := registry.ListModelsByBackend("gemini")
	vertexAI := registry.ListModelsByBackend("vertexai")

	if len(geminiBakcend) == 0 {
		t.Error("Expected Gemini models, got none")
	}

	if len(vertexAI) == 0 {
		t.Error("Expected Vertex AI models, got none")
	}

	// Verify that Gemini models have correct backend
	for _, model := range geminiBakcend {
		if model.Backend != "gemini" {
			t.Errorf("Model %s has backend %s, expected gemini", model.ID, model.Backend)
		}
	}

	// Verify that Vertex AI models have correct backend
	for _, model := range vertexAI {
		if model.Backend != "vertexai" {
			t.Errorf("Model %s has backend %s, expected vertexai", model.ID, model.Backend)
		}
	}
}

func TestExtractModelID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"gemini-2.5-flash", "gemini-2.5-flash"},
		{"gemini-2.5-flash-vertex", "gemini-2.5-flash"},
		{"gemini-1.5-pro", "gemini-1.5-pro"},
		{"gemini-1.5-pro-vertex", "gemini-1.5-pro"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExtractModelIDFromGemini(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDefaultModel(t *testing.T) {
	registry := NewModelRegistry()
	defaultModel := registry.GetDefaultModel()

	if !defaultModel.IsDefault {
		t.Error("Default model should have IsDefault=true")
	}

	if defaultModel.Backend == "" {
		t.Error("Default model should have a backend")
	}
}
