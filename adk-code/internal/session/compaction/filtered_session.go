package compaction

import (
	"encoding/json"
	"iter"
	"time"

	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// FilteredSession wraps a session to provide compaction-aware event filtering
type FilteredSession struct {
	Underlying session.Session
}

// NewFilteredSession creates a new filtered session
func NewFilteredSession(underlying session.Session) *FilteredSession {
	return &FilteredSession{Underlying: underlying}
}

// ID returns the session ID (pass-through)
func (fs *FilteredSession) ID() string {
	return fs.Underlying.ID()
}

// AppName returns the application name (pass-through)
func (fs *FilteredSession) AppName() string {
	return fs.Underlying.AppName()
}

// UserID returns the user ID (pass-through)
func (fs *FilteredSession) UserID() string {
	return fs.Underlying.UserID()
}

// State returns the session state (pass-through)
func (fs *FilteredSession) State() session.State {
	return fs.Underlying.State()
}

// LastUpdateTime returns the last update time (pass-through)
func (fs *FilteredSession) LastUpdateTime() time.Time {
	return fs.Underlying.LastUpdateTime()
}

// Events returns a filtered view that excludes compacted events
func (fs *FilteredSession) Events() session.Events {
	return NewFilteredEvents(fs.Underlying.Events())
}

// FilteredEvents implements session.Events with compaction filtering
type FilteredEvents struct {
	underlying session.Events
	filtered   []*session.Event
}

// NewFilteredEvents creates a new filtered events iterator
func NewFilteredEvents(underlying session.Events) *FilteredEvents {
	filtered := filterCompactedEvents(underlying)
	return &FilteredEvents{
		underlying: underlying,
		filtered:   filtered,
	}
}

// All returns an iterator over all filtered events
func (fe *FilteredEvents) All() iter.Seq[*session.Event] {
	return func(yield func(*session.Event) bool) {
		for _, event := range fe.filtered {
			if !yield(event) {
				return
			}
		}
	}
}

// Len returns the number of filtered events
func (fe *FilteredEvents) Len() int {
	return len(fe.filtered)
}

// At returns the event at the specified index
func (fe *FilteredEvents) At(i int) *session.Event {
	if i >= 0 && i < len(fe.filtered) {
		return fe.filtered[i]
	}
	return nil
}

// filterCompactedEvents implements the filtering logic
// It excludes original events that are within compacted ranges
// but includes the compaction summaries
func filterCompactedEvents(events session.Events) []*session.Event {
	allEvents := make([]*session.Event, 0, events.Len())
	for event := range events.All() {
		allEvents = append(allEvents, event)
	}

	// Find all compaction time ranges
	type timeRange struct {
		start time.Time
		end   time.Time
	}
	compactionRanges := make([]timeRange, 0)

	for _, event := range allEvents {
		if metadata, err := GetCompactionMetadata(event); err == nil {
			compactionRanges = append(compactionRanges, timeRange{
				start: metadata.StartTimestamp,
				end:   metadata.EndTimestamp,
			})
		}
	}

	// Filter events: include compaction summaries and non-compacted events
	filtered := make([]*session.Event, 0, events.Len())

	for _, event := range allEvents {
		if IsCompactionEvent(event) {
			// Include compaction event (contains summary)
			// But restore Content from the stored summary
			metadata, err := GetCompactionMetadata(event)
			if err == nil {
				var summaryContent genai.Content
				if err := json.Unmarshal([]byte(metadata.CompactedContentJSON), &summaryContent); err == nil {
					// Create a copy of the event with the restored summary content
					filteredEvent := *event
					filteredEvent.LLMResponse.Content = &summaryContent
					filtered = append(filtered, &filteredEvent)
				} else {
					// If unmarshaling fails, include the event as-is
					filtered = append(filtered, event)
				}
			} else {
				// If getting metadata fails, include the event as-is
				filtered = append(filtered, event)
			}
		} else {
			// Check if this event is within any compacted range
			withinCompactedRange := false
			for _, cr := range compactionRanges {
				// Check if event timestamp is within this compaction range
				if !event.Timestamp.Before(cr.start) && !event.Timestamp.After(cr.end) {
					withinCompactedRange = true
					break
				}
			}

			// Include only if NOT within a compacted range
			if !withinCompactedRange {
				filtered = append(filtered, event)
			}
		}
	}

	return filtered
}
