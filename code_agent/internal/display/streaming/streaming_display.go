package streaming

import (
	"sync"

	rdr "code_agent/internal/display/renderer"

	"google.golang.org/genai"
)

// Re-export types
type Renderer = rdr.Renderer

// StreamingDisplay manages streaming message display with deduplication
type StreamingDisplay struct {
	mu            sync.RWMutex
	renderer      *Renderer
	dedupe        *MessageDeduplicator
	activeSegment *StreamingSegment
	typewriter    *TypewriterPrinter
}

// NewStreamingDisplay creates a new streaming display manager
func NewStreamingDisplay(renderer *Renderer, typewriter *TypewriterPrinter) *StreamingDisplay {
	return &StreamingDisplay{
		renderer:   renderer,
		dedupe:     NewMessageDeduplicator(),
		typewriter: typewriter,
	}
}

// HandleTextMessage processes a text message from the agent
func (sd *StreamingDisplay) HandleTextMessage(text string, isThinking bool) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	// Check for duplicates
	if sd.dedupe.IsDuplicate(text) {
		return
	}

	// Determine message type
	msgType := MessageTypeResponse
	if isThinking {
		msgType = MessageTypeThinking
	}

	// If we have an active segment of a different type, freeze it
	if sd.activeSegment != nil && !sd.activeSegment.IsFrozen() {
		sd.activeSegment.Freeze()
		sd.activeSegment = nil
	}

	// Create new segment if needed
	if sd.activeSegment == nil {
		sd.activeSegment = NewStreamingSegment(
			msgType,
			sd.renderer.MarkdownRenderer(),
			sd.typewriter,
			sd.renderer.OutputFormat(),
		)
	}

	// Append text and freeze immediately (no streaming for now)
	sd.activeSegment.AppendText(text)
	sd.activeSegment.Freeze()
	sd.activeSegment = nil
}

// HandleToolCall processes a tool call
func (sd *StreamingDisplay) HandleToolCall(toolName string, args map[string]any) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	// Freeze any active segment
	if sd.activeSegment != nil && !sd.activeSegment.IsFrozen() {
		sd.activeSegment.Freeze()
		sd.activeSegment = nil
	}

	// Render tool call using renderer
	output := sd.renderer.RenderToolCall(toolName, args)
	if sd.typewriter != nil && sd.typewriter.IsEnabled() {
		sd.typewriter.Print(output)
	} else {
		sd.renderer.RenderText(output)
	}
}

// HandleToolResult processes a tool result
func (sd *StreamingDisplay) HandleToolResult(toolName string, result map[string]any) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	// Render tool result using renderer
	output := sd.renderer.RenderToolResult(toolName, result)
	if sd.typewriter != nil && sd.typewriter.IsEnabled() {
		sd.typewriter.Print(output)
	} else {
		sd.renderer.RenderText(output)
	}
}

// HandleEvent processes an event from the agent
func (sd *StreamingDisplay) HandleEvent(event *genai.Content) {
	if event == nil || len(event.Parts) == 0 {
		return
	}

	for _, part := range event.Parts {
		// Handle text content
		if part.Text != "" {
			// Determine if this is thinking text
			isThinking := false
			// You could add logic here to detect thinking patterns
			sd.HandleTextMessage(part.Text, isThinking)
		}

		// Handle function calls
		if part.FunctionCall != nil {
			args := make(map[string]any)
			for k, v := range part.FunctionCall.Args {
				args[k] = v
			}
			sd.HandleToolCall(part.FunctionCall.Name, args)
		}

		// Handle function responses
		if part.FunctionResponse != nil {
			result := make(map[string]any)
			if part.FunctionResponse.Response != nil {
				for k, v := range part.FunctionResponse.Response {
					result[k] = v
				}
			}
			sd.HandleToolResult(part.FunctionResponse.Name, result)
		}
	}
}

// FreezeActiveSegment freezes the currently active segment
func (sd *StreamingDisplay) FreezeActiveSegment() {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	if sd.activeSegment != nil && !sd.activeSegment.IsFrozen() {
		sd.activeSegment.Freeze()
		sd.activeSegment = nil
	}
}

// Cleanup cleans up streaming display resources
func (sd *StreamingDisplay) Cleanup() {
	sd.FreezeActiveSegment()
	if sd.dedupe != nil {
		sd.dedupe.Stop()
	}
}
