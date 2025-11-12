// Package display provides rich terminal display functionality.
// This facade re-exports all public types and constructors for ease of use.
package display

import (
	bn "code_agent/display/banner"
	rdr "code_agent/display/renderer"
)

// ============================================================================
// Renderer Types (from display/renderer)
// ============================================================================

// Renderer is the main display renderer for formatting output
type Renderer = rdr.Renderer

// NewRenderer creates a new renderer with the specified output format
func NewRenderer(outputFormat string) (*Renderer, error) {
	return rdr.NewRenderer(outputFormat)
}

// MarkdownRenderer renders markdown content to formatted output
type MarkdownRenderer = rdr.MarkdownRenderer

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer() (*MarkdownRenderer, error) {
	return rdr.NewMarkdownRenderer()
}

// ============================================================================
// Banner Types (from display/banner)
// ============================================================================

// BannerRenderer renders banner messages
type BannerRenderer = bn.BannerRenderer

// NewBannerRenderer creates a new banner renderer
func NewBannerRenderer(renderer *Renderer) *BannerRenderer {
	return bn.NewBannerRenderer(renderer)
}
