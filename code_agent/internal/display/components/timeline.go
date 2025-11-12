package components

import (
	"fmt"
	"strings"
)

// EventType represents different types of events in the agent execution
type EventType string

// Event types
const (
	EventTypeThinking  EventType = "thinking"
	EventTypeExecuting EventType = "executing"
	EventTypeResult    EventType = "result"
	EventTypeSuccess   EventType = "success"
	EventTypeWarning   EventType = "warning"
	EventTypeError     EventType = "error"
	EventTypeProgress  EventType = "progress"
)

// EventTypeIcon returns the emoji icon for an event type
func EventTypeIcon(eventType EventType) string {
	switch eventType {
	case EventTypeThinking:
		return "🧠"
	case EventTypeExecuting:
		return "🔧"
	case EventTypeResult:
		return "📊"
	case EventTypeSuccess:
		return "✓"
	case EventTypeWarning:
		return "⚠️"
	case EventTypeError:
		return "❌"
	case EventTypeProgress:
		return "📍"
	default:
		return "•"
	}
}

// TimelineEvent represents a single event in the operation timeline
type TimelineEvent struct {
	ToolName string
	Status   string // "pending", "executing", "completed", "failed"
}

// EventTimeline tracks a sequence of operations
type EventTimeline struct {
	events []TimelineEvent
}

// NewEventTimeline creates a new event timeline
func NewEventTimeline() *EventTimeline {
	return &EventTimeline{
		events: make([]TimelineEvent, 0),
	}
}

// AppendEvent adds an operation to the timeline
func (et *EventTimeline) AppendEvent(toolName, status string) {
	et.events = append(et.events, TimelineEvent{
		ToolName: toolName,
		Status:   status,
	})
}

// RenderTimeline returns a formatted timeline string
func (et *EventTimeline) RenderTimeline() string {
	if len(et.events) == 0 {
		return ""
	}

	// Build timeline string like: [read_file] → [grep_search] → [write_file]
	var parts []string
	for i, event := range et.events {
		// Use short tool name (last part after underscore)
		shortName := strings.TrimPrefix(event.ToolName, "list_")
		shortName = strings.TrimPrefix(shortName, "search_")
		shortName = strings.TrimPrefix(shortName, "read_")
		shortName = strings.TrimPrefix(shortName, "write_")
		shortName = strings.TrimPrefix(shortName, "execute_")

		parts = append(parts, fmt.Sprintf("[%s]", shortName))

		// Add arrow between events (except after last)
		if i < len(et.events)-1 {
			parts = append(parts, "→")
		}
	}

	return "Timeline: " + strings.Join(parts, " ")
}

// GetEventCount returns the number of events in the timeline
func (et *EventTimeline) GetEventCount() int {
	return len(et.events)
}

// UpdateLastEventStatus updates the status of the most recent event
func (et *EventTimeline) UpdateLastEventStatus(status string) {
	if len(et.events) > 0 {
		et.events[len(et.events)-1].Status = status
	}
}

// RenderProgress returns a simple progress indicator
func (et *EventTimeline) RenderProgress() string {
	if len(et.events) == 0 {
		return ""
	}

	completed := 0
	for _, event := range et.events {
		if event.Status == "completed" {
			completed++
		}
	}

	total := len(et.events)
	percent := (completed * 100) / total

	// Simple progress bar: [████░░░░░░] 40% (2 of 5)
	barWidth := 10
	filledWidth := (completed * barWidth) / total
	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

	return fmt.Sprintf("Progress: [%s] %d%% (%d of %d operations)", bar, percent, completed, total)
}
