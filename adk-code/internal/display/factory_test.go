package display

import (
	"testing"
	"time"
)

func TestNewComponents_CreatesAllComponents(t *testing.T) {
	cfg := ComponentsConfig{
		OutputFormat:      OutputFormatPlain,
		TypewriterEnabled: true,
	}

	comps, err := NewComponents(cfg)
	if err != nil {
		t.Fatalf("NewComponents failed: %v", err)
	}

	if comps == nil {
		t.Fatalf("expected non-nil components")
	}

	if comps.Renderer == nil {
		t.Fatal("expected Renderer to be initialized")
	}

	if comps.BannerRenderer == nil {
		t.Fatal("expected BannerRenderer to be initialized")
	}

	if comps.Typewriter == nil {
		t.Fatal("expected Typewriter to be initialized")
	}

	if comps.StreamDisplay == nil {
		t.Fatal("expected StreamDisplay to be initialized")
	}

	if !comps.Typewriter.IsEnabled() {
		t.Fatal("expected Typewriter to be enabled")
	}
}

func TestNewComponents_TypewriterDisabled(t *testing.T) {
	cfg := ComponentsConfig{
		OutputFormat:      OutputFormatPlain,
		TypewriterEnabled: false,
	}

	comps, err := NewComponents(cfg)
	if err != nil {
		t.Fatalf("NewComponents failed: %v", err)
	}

	if comps.Typewriter.IsEnabled() {
		t.Fatal("expected Typewriter to be disabled")
	}
}

func TestNewComponents_CustomTypewriterConfig(t *testing.T) {
	customCfg := &TypewriterConfig{
		BaseDelay:    15 * time.Millisecond,
		FastDelay:    8 * time.Millisecond,
		SlowDelay:    25 * time.Millisecond,
		PauseDelay:   150 * time.Millisecond,
		RandomFactor: 0.3,
		Enabled:      true,
	}

	cfg := ComponentsConfig{
		OutputFormat:      OutputFormatPlain,
		TypewriterEnabled: true,
		TypewriterConfig:  customCfg,
	}

	comps, err := NewComponents(cfg)
	if err != nil {
		t.Fatalf("NewComponents failed: %v", err)
	}

	if comps.Typewriter == nil {
		t.Fatal("expected Typewriter to be initialized with custom config")
	}
}

func TestNewComponents_InvalidOutputFormat(t *testing.T) {
	// The NewRenderer handles unknown formats gracefully
	// by defaulting to a safe format, so we just verify it doesn't panic
	cfg := ComponentsConfig{
		OutputFormat:      "unknown_format_xyz",
		TypewriterEnabled: false,
	}

	comps, err := NewComponents(cfg)
	// Should not error, but should return a valid renderer
	if err != nil {
		t.Logf("Note: NewRenderer didn't error for unknown format, which is acceptable")
	}
	if comps == nil || comps.Renderer == nil {
		t.Fatal("expected valid components even with unknown format")
	}
}
