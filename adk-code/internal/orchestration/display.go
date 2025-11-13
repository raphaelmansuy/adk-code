package orchestration

import (
	"fmt"

	"adk-code/internal/config"
	"adk-code/internal/display"
)

// displayInitializer handles display component setup
type displayInitializer struct {
	renderer       *display.Renderer
	bannerRenderer *display.BannerRenderer
	typewriter     *display.TypewriterPrinter
	streamDisplay  *display.StreamingDisplay
}

// InitializeDisplayComponents sets up display components
func InitializeDisplayComponents(cfg *config.Config) (*DisplayComponents, error) {
	initializer := &displayInitializer{}

	var err error
	initializer.renderer, err = display.NewRenderer(cfg.OutputFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	initializer.typewriter = display.NewTypewriterPrinter(display.DefaultTypewriterConfig())
	initializer.typewriter.SetEnabled(cfg.TypewriterEnabled)
	initializer.streamDisplay = display.NewStreamingDisplay(initializer.renderer, initializer.typewriter)
	initializer.bannerRenderer = display.NewBannerRenderer(initializer.renderer)

	return &DisplayComponents{
		Renderer:       initializer.renderer,
		BannerRenderer: initializer.bannerRenderer,
		Typewriter:     initializer.typewriter,
		StreamDisplay:  initializer.streamDisplay,
	}, nil
}
