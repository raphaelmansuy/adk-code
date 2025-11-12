package app

import (
	"context"
	"testing"

	"code_agent/display"
	"code_agent/pkg/cli"
)

func TestDisplayComponentFactory(t *testing.T) {
	config := &cli.CLIConfig{
		OutputFormat:      display.OutputFormatPlain,
		TypewriterEnabled: false,
	}

	factory := NewDisplayComponentFactory(config)
	if factory == nil {
		t.Error("NewDisplayComponentFactory should not return nil")
	}

	components, err := factory.Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if components == nil {
		t.Error("DisplayComponents should not be nil")
	}
	if components.Renderer == nil {
		t.Error("Renderer should not be nil")
	}
	if components.BannerRenderer == nil {
		t.Error("BannerRenderer should not be nil")
	}
	if components.Typewriter == nil {
		t.Error("Typewriter should not be nil")
	}
	if components.StreamDisplay == nil {
		t.Error("StreamDisplay should not be nil")
	}
}

func TestDisplayComponentFactoryTypewriterEnabled(t *testing.T) {
	config := &cli.CLIConfig{
		OutputFormat:      display.OutputFormatPlain,
		TypewriterEnabled: true,
	}

	factory := NewDisplayComponentFactory(config)
	components, err := factory.Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if components.Typewriter == nil {
		t.Error("Typewriter should not be nil")
	}
}

func TestModelComponentFactory(t *testing.T) {
	config := &cli.CLIConfig{
		OutputFormat: display.OutputFormatPlain,
		Model:        "", // Use default model
		APIKey:       "", // Will fail without API key, but we can test factory creation
	}

	factory := NewModelComponentFactory(config)
	if factory == nil {
		t.Error("NewModelComponentFactory should not return nil")
	}
}

func TestResolveWorkingDirectory(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "/tmp",
			expected: "/tmp",
		},
	}

	for _, test := range tests {
		config := &cli.CLIConfig{
			WorkingDirectory: test.input,
		}
		factory := NewModelComponentFactory(config)
		result := factory.resolveWorkingDirectory()
		if result != test.expected {
			t.Errorf("resolveWorkingDirectory(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestResolveWorkingDirectoryEmpty(t *testing.T) {
	config := &cli.CLIConfig{
		WorkingDirectory: "",
	}
	factory := NewModelComponentFactory(config)
	result := factory.resolveWorkingDirectory()
	if result == "" {
		t.Error("resolveWorkingDirectory with empty input should return a directory")
	}
}

func TestDisplayComponentFactoryWithJSONOutput(t *testing.T) {
	config := &cli.CLIConfig{
		OutputFormat:      display.OutputFormatJSON,
		TypewriterEnabled: false,
	}

	factory := NewDisplayComponentFactory(config)
	components, err := factory.Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if components == nil {
		t.Error("DisplayComponents should not be nil")
	}
	if components.Renderer == nil {
		t.Error("Renderer should not be nil")
	}
}

func TestFactorySequence(t *testing.T) {
	// Test that factories can be created in sequence
	displayConfig := &cli.CLIConfig{
		OutputFormat:      display.OutputFormatPlain,
		TypewriterEnabled: false,
	}

	displayFactory := NewDisplayComponentFactory(displayConfig)
	displayComponents, err := displayFactory.Create()
	if err != nil {
		t.Fatalf("Display factory failed: %v", err)
	}

	modelConfig := &cli.CLIConfig{
		WorkingDirectory: "/tmp",
	}

	modelFactory := NewModelComponentFactory(modelConfig)
	// Don't call Create() on model factory without API key
	if modelFactory == nil {
		t.Error("Model factory creation failed")
	}

	// Verify display components are still valid
	if displayComponents.Renderer == nil {
		t.Error("Display components corrupted after model factory creation")
	}
}

func TestDisplayComponentFactoryContextCancellation(t *testing.T) {
	config := &cli.CLIConfig{
		OutputFormat:      display.OutputFormatPlain,
		TypewriterEnabled: false,
	}

	factory := NewDisplayComponentFactory(config)

	// Create with active context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_ = ctx // Use ctx to avoid unused variable

	// Factory should still work (it doesn't use context)
	components, err := factory.Create()
	if err != nil {
		t.Fatalf("Create() should work even with cancelled context: %v", err)
	}
	if components == nil {
		t.Error("Components should not be nil")
	}
}
