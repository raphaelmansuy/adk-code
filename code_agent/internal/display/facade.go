// Package display provides rich terminal display functionality.
// This facade re-exports all public types and constructors for ease of use.
package display

import (
	bn "code_agent/internal/display/banner"
	"code_agent/internal/display/components"
	"code_agent/internal/display/events"
	rdr "code_agent/internal/display/renderer"
	"code_agent/internal/display/streaming"

	"code_agent/internal/tracking"

	"google.golang.org/adk/session"
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

// ============================================================================
// Component Types (from display/components)
// ============================================================================

// Paginator handles displaying long content with pagination
type Paginator = components.Paginator

// NewPaginator creates a new paginator instance
func NewPaginator(renderer *Renderer) *Paginator {
	return components.NewPaginator(renderer)
}

// Spinner provides animated progress indication
type Spinner = components.Spinner

// NewSpinner creates a new spinner
func NewSpinner(renderer *Renderer, message string) *Spinner {
	return components.NewSpinner(renderer, message)
}

// TypewriterPrinter handles typewriter-style output
type TypewriterPrinter = components.TypewriterPrinter

// TypewriterConfig holds configuration for the typewriter effect
type TypewriterConfig = components.TypewriterConfig

// NewTypewriterPrinter creates a new typewriter printer
func NewTypewriterPrinter(config *TypewriterConfig) *TypewriterPrinter {
	return components.NewTypewriterPrinter(config)
}

// DefaultTypewriterConfig returns the default typewriter configuration
func DefaultTypewriterConfig() *TypewriterConfig {
	return components.DefaultTypewriterConfig()
}

// ============================================================================
// Streaming Types (from display/streaming)
// ============================================================================

// StreamingDisplay manages streaming message display with deduplication
type StreamingDisplay = streaming.StreamingDisplay

// NewStreamingDisplay creates a new streaming display manager
func NewStreamingDisplay(renderer *Renderer, typewriter *TypewriterPrinter) *StreamingDisplay {
	return streaming.NewStreamingDisplay(renderer, typewriter)
}

// ============================================================================
// Event Handling (from display/events)
// ============================================================================

// PrintEventEnhanced processes and displays agent events
func PrintEventEnhanced(renderer *Renderer, streamDisplay *StreamingDisplay,
	event *session.Event, spinner *Spinner, activeToolName *string, toolRunning *bool,
	sessionTokens *tracking.SessionTokens, requestID string, timeline *EventTimeline) {
	events.PrintEventEnhanced(renderer, streamDisplay, event, spinner, activeToolName, toolRunning, sessionTokens, requestID, timeline)
}
