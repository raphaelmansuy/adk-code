package display

import (
	"code_agent/internal/display/components"
	"code_agent/internal/display/streaming"
	pkgerrors "code_agent/pkg/errors"
)

// ComponentsConfig holds configuration for creating display components
type ComponentsConfig struct {
	OutputFormat      string
	TypewriterEnabled bool
	TypewriterConfig  *components.TypewriterConfig
}

// Components groups all display-related components
type Components struct {
	Renderer       *Renderer
	BannerRenderer *BannerRenderer
	Typewriter     *components.TypewriterPrinter
	StreamDisplay  *streaming.StreamingDisplay
}

// NewComponents creates all display components with the given configuration
func NewComponents(cfg ComponentsConfig) (*Components, error) {
	// Create renderer
	renderer, err := NewRenderer(cfg.OutputFormat)
	if err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create renderer", err)
	}

	// Create typewriter with optional custom config
	var typewriterCfg *components.TypewriterConfig
	if cfg.TypewriterConfig != nil {
		typewriterCfg = cfg.TypewriterConfig
	} else {
		typewriterCfg = components.DefaultTypewriterConfig()
	}
	typewriter := components.NewTypewriterPrinter(typewriterCfg)
	typewriter.SetEnabled(cfg.TypewriterEnabled)

	// Create other components
	bannerRenderer := NewBannerRenderer(renderer)
	streamDisplay := streaming.NewStreamingDisplay(renderer, typewriter)

	return &Components{
		Renderer:       renderer,
		BannerRenderer: bannerRenderer,
		Typewriter:     typewriter,
		StreamDisplay:  streamDisplay,
	}, nil
}
