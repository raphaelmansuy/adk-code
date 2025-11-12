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

import "fmt"

// ComponentsConfig holds configuration for creating display components
type ComponentsConfig struct {
	OutputFormat      string
	TypewriterEnabled bool
	TypewriterConfig  *TypewriterConfig
}

// Components groups all display-related components
type Components struct {
	Renderer       *Renderer
	BannerRenderer *BannerRenderer
	Typewriter     *TypewriterPrinter
	StreamDisplay  *StreamingDisplay
}

// NewComponents creates all display components with the given configuration
func NewComponents(cfg ComponentsConfig) (*Components, error) {
	// Create renderer
	renderer, err := NewRenderer(cfg.OutputFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	// Create typewriter with optional custom config
	var typewriterCfg *TypewriterConfig
	if cfg.TypewriterConfig != nil {
		typewriterCfg = cfg.TypewriterConfig
	} else {
		typewriterCfg = DefaultTypewriterConfig()
	}
	typewriter := NewTypewriterPrinter(typewriterCfg)
	typewriter.SetEnabled(cfg.TypewriterEnabled)

	// Create other components
	bannerRenderer := NewBannerRenderer(renderer)
	streamDisplay := NewStreamingDisplay(renderer, typewriter)

	return &Components{
		Renderer:       renderer,
		BannerRenderer: bannerRenderer,
		Typewriter:     typewriter,
		StreamDisplay:  streamDisplay,
	}, nil
}
