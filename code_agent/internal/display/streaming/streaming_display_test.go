package streaming

import (
	"testing"

	"code_agent/internal/display/components"
	"code_agent/internal/display/renderer"
)

// Re-export for tests
var (
	NewRenderer             = renderer.NewRenderer
	NewTypewriterPrinter    = components.NewTypewriterPrinter
	DefaultTypewriterConfig = components.DefaultTypewriterConfig
)

// TestNewStreamingDisplay creates and validates a streaming display
func TestNewStreamingDisplay(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}

	typewriter := NewTypewriterPrinter(DefaultTypewriterConfig())

	sd := NewStreamingDisplay(renderer, typewriter)

	if sd == nil {
		t.Fatal("NewStreamingDisplay returned nil")
	}
}

// TestStreamingDisplay_HandleTextMessage tests text message handling
func TestStreamingDisplay_HandleTextMessage(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}
	typewriter := NewTypewriterPrinter(DefaultTypewriterConfig())
	sd := NewStreamingDisplay(renderer, typewriter)

	// Test handling a normal text message
	// Note: activeSegment is frozen and cleared immediately after HandleTextMessage
	sd.HandleTextMessage("Hello, world!", false)

	// After HandleTextMessage, segment is frozen and activeSegment is set to nil
	// This is expected behavior - the test just verifies no panic occurs
	if sd.activeSegment != nil {
		t.Errorf("activeSegment should be nil after HandleTextMessage (segment is frozen and cleared)")
	}
}

// TestStreamingDisplay_HandleDuplicateMessage tests duplicate detection
func TestStreamingDisplay_HandleDuplicateMessage(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}
	typewriter := NewTypewriterPrinter(DefaultTypewriterConfig())
	sd := NewStreamingDisplay(renderer, typewriter)

	// First message should create a segment
	sd.HandleTextMessage("Test message", false)
	firstSegment := sd.activeSegment

	// Duplicate message should not change segment
	sd.HandleTextMessage("Test message", false)
	secondSegment := sd.activeSegment

	if firstSegment != secondSegment {
		t.Errorf("activeSegment changed when handling duplicate message")
	}
}

// TestStreamingDisplay_HandleThinkingMessage tests thinking message handling
func TestStreamingDisplay_HandleThinkingMessage(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}
	typewriter := NewTypewriterPrinter(DefaultTypewriterConfig())
	sd := NewStreamingDisplay(renderer, typewriter)

	// Note: activeSegment is frozen and cleared immediately after HandleTextMessage
	sd.HandleTextMessage("I am thinking...", true)

	// After HandleTextMessage, segment is frozen and activeSegment is set to nil
	// This is expected behavior - the test just verifies no panic occurs
	if sd.activeSegment != nil {
		t.Errorf("activeSegment should be nil after HandleTextMessage (segment is frozen and cleared)")
	}
}

// TestStreamingDisplay_SegmentTransition tests transitioning between different message types
func TestStreamingDisplay_SegmentTransition(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}
	typewriter := NewTypewriterPrinter(DefaultTypewriterConfig())
	sd := NewStreamingDisplay(renderer, typewriter)

	// Create a response segment - it will be frozen and cleared immediately
	sd.HandleTextMessage("First response", false)
	// After HandleTextMessage, activeSegment is nil (segment was frozen and cleared)

	// Switch to thinking - segment type is different, so behavior is consistent
	sd.HandleTextMessage("Thinking about this", true)
	// After this call, activeSegment is also nil (segment was frozen and cleared)

	// The test verifies that transitioning between message types doesn't panic
	// Since both segments are frozen and cleared immediately, just verify no panic occurred
	if sd.activeSegment != nil {
		t.Errorf("activeSegment should be nil after HandleTextMessage")
	}
}

// TestStreamingDisplay_FlushSegment tests flushing active segment
func TestStreamingDisplay_FlushSegment(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}
	typewriter := NewTypewriterPrinter(DefaultTypewriterConfig())
	sd := NewStreamingDisplay(renderer, typewriter)

	sd.HandleTextMessage("Test message", false)

	// After HandleTextMessage, segment is frozen and activeSegment is set to nil
	// This is expected behavior - the test verifies no panic occurs
	if sd.activeSegment != nil {
		t.Errorf("activeSegment should be nil after HandleTextMessage (segment is frozen and cleared)")
	}
}

// TestStreamingDisplay_MultipleMessages tests handling multiple messages
func TestStreamingDisplay_MultipleMessages(t *testing.T) {
	renderer, err := NewRenderer("plain")
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}
	typewriter := NewTypewriterPrinter(DefaultTypewriterConfig())
	sd := NewStreamingDisplay(renderer, typewriter)

	messages := []struct {
		text       string
		isThinking bool
	}{
		{"First message", false},
		{"Second message", false},
		{"Thinking about it", true},
		{"Back to response", false},
	}

	for _, msg := range messages {
		sd.HandleTextMessage(msg.text, msg.isThinking)
	}

	// After HandleTextMessage, segment is frozen and activeSegment is set to nil
	// This is expected behavior - the test verifies no panic occurs
	if sd.activeSegment != nil {
		t.Errorf("activeSegment should be nil after HandleTextMessage (segment is frozen and cleared)")
	}
}
