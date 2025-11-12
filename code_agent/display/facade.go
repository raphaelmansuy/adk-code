package display

import (
	bn "code_agent/display/banner"
	rdr "code_agent/display/renderer"
)

// Re-export Renderer type from the renderer package to keep the existing API stable.

// NewRenderer creates a new renderer via the renderer package implementation.
type Renderer = rdr.Renderer

func NewRenderer(outputFormat string) (*Renderer, error) {
	return rdr.NewRenderer(outputFormat)
}

type BannerRenderer = bn.BannerRenderer

func NewBannerRenderer(renderer *Renderer) *BannerRenderer {
	return bn.NewBannerRenderer(renderer)
}

// Alias MarkdownRenderer from renderer package for backwards compatibility with code expecting the type in the display package.
type MarkdownRenderer = rdr.MarkdownRenderer

func NewMarkdownRenderer() (*MarkdownRenderer, error) {
	return rdr.NewMarkdownRenderer()
}
