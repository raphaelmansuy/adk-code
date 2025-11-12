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

package display

import (
	"testing"
	"time"
)

func TestDefaultTypewriterConfig(t *testing.T) {
	cfg := DefaultTypewriterConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.BaseDelay == 0 {
		t.Fatal("expected non-zero base delay")
	}
	if cfg.RandomFactor < 0 || cfg.RandomFactor > 1 {
		t.Fatalf("expected RandomFactor between 0 and 1, got: %f", cfg.RandomFactor)
	}
}

func TestNewTypewriterPrinter(t *testing.T) {
	cfg := DefaultTypewriterConfig()
	tp := NewTypewriterPrinter(cfg)
	if tp == nil {
		t.Fatal("expected non-nil typewriter printer")
	}
}

func TestTypewriterPrinter_SetEnabled(t *testing.T) {
	cfg := DefaultTypewriterConfig()
	tp := NewTypewriterPrinter(cfg)

	tp.SetEnabled(true)
	if !tp.IsEnabled() {
		t.Fatal("expected typewriter to be enabled")
	}

	tp.SetEnabled(false)
	if tp.IsEnabled() {
		t.Fatal("expected typewriter to be disabled")
	}
}

func TestTypewriterPrinter_IsEnabled(t *testing.T) {
	cfg := DefaultTypewriterConfig()
	tp := NewTypewriterPrinter(cfg)

	// Default from config
	expected := cfg.Enabled
	if tp.IsEnabled() != expected {
		t.Fatalf("expected enabled=%v, got: %v", expected, tp.IsEnabled())
	}
}

func TestTypewriterPrinter_SetSpeed(t *testing.T) {
	cfg := DefaultTypewriterConfig()
	tp := NewTypewriterPrinter(cfg)

	tp.SetSpeed(0.5)
	// SetSpeed should not panic
}

func TestTypewriterPrinter_PrintInstant(t *testing.T) {
	cfg := DefaultTypewriterConfig()
	tp := NewTypewriterPrinter(cfg)

	// PrintInstant should not panic
	tp.PrintInstant("instant text")
}

func TestTypewriterPrinter_PrintfInstant(t *testing.T) {
	cfg := DefaultTypewriterConfig()
	tp := NewTypewriterPrinter(cfg)

	// PrintfInstant should not panic
	tp.PrintfInstant("formatted: %s", "text")
}

func TestTypewriterConfig_Customization(t *testing.T) {
	cfg := &TypewriterConfig{
		BaseDelay:    20 * time.Millisecond,
		FastDelay:    10 * time.Millisecond,
		SlowDelay:    30 * time.Millisecond,
		PauseDelay:   200 * time.Millisecond,
		Enabled:      true,
		RandomFactor: 0.5,
	}

	tp := NewTypewriterPrinter(cfg)
	if tp == nil {
		t.Fatal("expected non-nil typewriter with custom config")
	}

	if tp.IsEnabled() != cfg.Enabled {
		t.Fatalf("expected enabled=%v, got: %v", cfg.Enabled, tp.IsEnabled())
	}
}

func TestTypewriterPrinter_DisabledByDefault(t *testing.T) {
	cfg := &TypewriterConfig{
		BaseDelay:    10 * time.Millisecond,
		Enabled:      false,
		RandomFactor: 0.3,
	}

	tp := NewTypewriterPrinter(cfg)
	if tp.IsEnabled() {
		t.Fatal("expected typewriter to be disabled when Enabled=false")
	}
}
