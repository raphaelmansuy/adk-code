package cli

import (
	"testing"

	"adk-code/pkg/models"
)

func TestParseProviderModelSyntax(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedProv  string
		expectedModel string
		wantErr       bool
	}{
		// Valid cases
		{"explicit-provider", "gemini/2.5-flash", "gemini", "2.5-flash", false},
		{"explicit-vertexai", "vertexai/1.5-pro", "vertexai", "1.5-pro", false},
		{"shorthand", "flash", "", "flash", false},
		{"provider-shorthand", "gemini/flash", "gemini", "flash", false},
		{"with-latest", "gemini/latest", "gemini", "latest", false},

		// Error cases
		{"empty-string", "", "", "", true},
		{"only-slash", "/", "", "", true},
		{"trailing-slash", "gemini/", "", "", true},
		{"leading-slash", "/flash", "", "", true},
		{"too-many-slashes", "a/b/c", "", "", true},
		{"whitespace", "  gemini/flash  ", "gemini", "flash", false}, // Trimmed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, model, err := ParseProviderModelSyntax(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProviderModelSyntax(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if provider != tt.expectedProv {
					t.Errorf("ParseProviderModelSyntax(%q) provider = %q, expected %q", tt.input, provider, tt.expectedProv)
				}
				if model != tt.expectedModel {
					t.Errorf("ParseProviderModelSyntax(%q) model = %q, expected %q", tt.input, model, tt.expectedModel)
				}
			}
		})
	}
}

func TestResolveFromProviderSyntax(t *testing.T) {
	registry := models.NewRegistry()

	tests := []struct {
		name            string
		provider        string
		model           string
		defaultProvider string
		expectedModelID string
		shouldSucceed   bool
	}{
		// Valid explicit provider cases
		{"explicit-gemini-full-id", "gemini", "gemini-2.5-flash", "gemini", "gemini-2.5-flash", true},
		{"explicit-gemini-shorthand", "gemini", "flash", "gemini", "gemini-2.5-flash", true},
		{"explicit-vertexai", "vertexai", "2.5-flash", "gemini", "gemini-2.5-flash", true},

		// Default provider cases (empty provider string)
		{"default-provider-shorthand", "", "flash", "gemini", "gemini-2.5-flash", true},
		{"default-provider-explicit", "", "2.5-flash", "gemini", "gemini-2.5-flash", true},
		{"default-to-vertexai", "", "flash", "vertexai", "gemini-2.5-flash", true},

		// 1.5 models
		{"gemini-1.5-pro", "gemini", "1.5-pro", "gemini", "gemini-1.5-pro", true},
		{"gemini-pro-shorthand", "gemini", "pro", "gemini", "gemini-1.5-pro", true},

		// Error cases
		{"invalid-model", "gemini", "nonexistent", "gemini", "", false},
		{"invalid-provider", "nonexistent", "flash", "gemini", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := registry.ResolveFromProviderSyntax(tt.provider, tt.model, tt.defaultProvider)

			if (err != nil) != !tt.shouldSucceed {
				t.Errorf("ResolveFromProviderSyntax(%q, %q, %q) error = %v, shouldSucceed = %v",
					tt.provider, tt.model, tt.defaultProvider, err, tt.shouldSucceed)
				return
			}

			if !tt.shouldSucceed {
				return
			}

			if model.ID != tt.expectedModelID {
				t.Errorf("ResolveFromProviderSyntax got model ID %q, expected %q", model.ID, tt.expectedModelID)
			}
		})
	}
}

func TestGetProviderModels(t *testing.T) {
	registry := models.NewRegistry()

	geminiBakcend := registry.GetProviderModels("gemini")
	vertexAI := registry.GetProviderModels("vertexai")

	if len(geminiBakcend) == 0 {
		t.Error("Expected Gemini models, got none")
	}

	if len(vertexAI) == 0 {
		t.Error("Expected Vertex AI models, got none")
	}

	// Both providers should have the same base models
	if len(geminiBakcend) != len(vertexAI) {
		t.Errorf("Expected same number of models for both providers: gemini=%d, vertexai=%d",
			len(geminiBakcend), len(vertexAI))
	}

	// All gemini models should have gemini backend marked
	for _, model := range geminiBakcend {
		if model.Backend != "gemini" {
			t.Errorf("Model %s in gemini provider has backend %s", model.ID, model.Backend)
		}
	}

	// All vertexai models should have gemini backend marked (they're the same base models)
	for _, model := range vertexAI {
		if model.Backend != "gemini" {
			t.Errorf("Model %s in vertexai provider has backend %s", model.ID, model.Backend)
		}
	}
}

func TestListProviders(t *testing.T) {
	registry := models.NewRegistry()
	providers := registry.ListProviders()

	if len(providers) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(providers))
	}

	// Check that expected providers are present
	providerMap := make(map[string]bool)
	for _, p := range providers {
		providerMap[p] = true
	}

	expectedProviders := []string{"gemini", "vertexai", "openai"}
	for _, expected := range expectedProviders {
		if !providerMap[expected] {
			t.Errorf("Expected provider %q not found in list", expected)
		}
	}

	// Verify sorted order
	if len(providers) >= 2 && providers[0] > providers[1] {
		t.Errorf("Providers not in sorted order: %v", providers)
	}
}

func TestProviderParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected models.Provider
		isValid  bool
	}{
		{"gemini", "gemini", models.ProviderGemini, true},
		{"vertexai", "vertexai", models.ProviderVertexAI, true},
		{"case-insensitive", "GEMINI", models.ProviderGemini, true},
		{"invalid", "invalid", models.Provider(""), false},
		{"empty", "", models.Provider(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.ParseProvider(tt.input)
			if tt.isValid {
				if result != tt.expected {
					t.Errorf("ParseProvider(%q) = %q, expected %q", tt.input, result, tt.expected)
				}
			} else {
				if result != "" {
					t.Errorf("ParseProvider(%q) should return empty, got %q", tt.input, result)
				}
			}
		})
	}
}

func TestProviderMetadata(t *testing.T) {
	tests := []struct {
		provider models.Provider
		checkFn  func(models.ProviderMetadata) bool
	}{
		{
			models.ProviderGemini,
			func(m models.ProviderMetadata) bool {
				return m.DisplayName == "Gemini API" && m.Name == "gemini" && m.Icon == "🔷"
			},
		},
		{
			models.ProviderVertexAI,
			func(m models.ProviderMetadata) bool {
				return m.DisplayName == "Vertex AI" && m.Name == "vertexai" && m.Icon == "🔶"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.provider.String(), func(t *testing.T) {
			meta := models.GetProviderMetadata(tt.provider)
			if !tt.checkFn(meta) {
				t.Errorf("Provider metadata check failed for %s: %+v", tt.provider, meta)
			}
			if len(meta.Requirements) == 0 {
				t.Errorf("Provider %s should have requirements", tt.provider)
			}
		})
	}
}

// Integration test: Test the full flow from parsing to model resolution
func TestParseAndResolveFlow(t *testing.T) {
	registry := models.NewRegistry()

	tests := []struct {
		name            string
		input           string
		defaultProvider string
		expectedSuccess bool
		expectedModelID string
	}{
		{"simple-shorthand", "flash", "gemini", true, "gemini-2.5-flash"},
		{"explicit-provider", "gemini/flash", "vertexai", true, "gemini-2.5-flash"},
		{"full-model-id", "gemini/gemini-2.5-flash", "gemini", true, "gemini-2.5-flash"},
		{"invalid-syntax", "/flash", "gemini", false, ""},
		{"invalid-model", "gemini/nonexistent", "gemini", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, model, parseErr := ParseProviderModelSyntax(tt.input)
			if parseErr != nil {
				if tt.expectedSuccess {
					t.Errorf("Unexpected parse error: %v", parseErr)
				}
				return
			}

			resolved, resolveErr := registry.ResolveFromProviderSyntax(provider, model, tt.defaultProvider)
			if (resolveErr != nil) != !tt.expectedSuccess {
				t.Errorf("Unexpected resolve error: %v", resolveErr)
				return
			}

			if tt.expectedSuccess && resolved.ID != tt.expectedModelID {
				t.Errorf("Expected model ID %q, got %q", tt.expectedModelID, resolved.ID)
			}
		})
	}
}
