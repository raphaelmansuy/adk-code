package compaction

import (
	"sort"
	"time"

	"google.golang.org/adk/session"
)

// Selector selects events to be compacted based on configured thresholds
type Selector struct {
	config *Config
}

// NewSelector creates a new event selector
func NewSelector(config *Config) *Selector {
	return &Selector{config: config}
}

// SelectEventsToCompact selects events that should be compacted based on the configured thresholds
func (s *Selector) SelectEventsToCompact(events []*session.Event) ([]*session.Event, error) {
	if len(events) == 0 {
		return nil, nil
	}

	// Find last compaction event using CustomMetadata
	lastCompactionIdx := -1
	for i := len(events) - 1; i >= 0; i-- {
		if IsCompactionEvent(events[i]) {
			lastCompactionIdx = i
			break
		}
	}

	// Count unique invocations since last compaction
	invocationMap := make(map[string]time.Time)
	startIdx := lastCompactionIdx + 1

	for i := startIdx; i < len(events); i++ {
		if events[i].InvocationID != "" {
			invocationMap[events[i].InvocationID] = events[i].Timestamp
		}
	}

	// Check invocation threshold
	if len(invocationMap) < s.config.InvocationThreshold {
		return nil, nil // Not enough invocations
	}

	// Sort invocation IDs by timestamp
	invocationIDs := s.sortInvocationsByTime(invocationMap)

	// Calculate window: need to select based on threshold and overlap
	// We want to compact events from earlier invocations, keeping recent ones
	if len(invocationIDs) < s.config.InvocationThreshold {
		return nil, nil
	}

	// Select window: from start of threshold window to end of threshold window
	windowSize := s.config.InvocationThreshold + s.config.OverlapSize
	var startInvocationID, endInvocationID string

	if len(invocationIDs) > windowSize {
		// Slide window: compact oldest invocations, keep recent ones
		startIdx := len(invocationIDs) - windowSize
		endIdx := startIdx + s.config.InvocationThreshold - 1

		startInvocationID = invocationIDs[startIdx]
		endInvocationID = invocationIDs[endIdx]
	} else {
		// Window fits all invocations since last compaction
		startInvocationID = invocationIDs[0]
		endInvocationID = invocationIDs[len(invocationIDs)-1]
	}

	// Collect events in window
	return s.filterEventsByInvocationRange(events, startInvocationID, endInvocationID), nil
}

// sortInvocationsByTime returns sorted invocation IDs by timestamp
func (s *Selector) sortInvocationsByTime(invocationMap map[string]time.Time) []string {
	type invocation struct {
		id        string
		timestamp time.Time
	}

	invocations := make([]invocation, 0, len(invocationMap))
	for id, ts := range invocationMap {
		invocations = append(invocations, invocation{id, ts})
	}

	sort.Slice(invocations, func(i, j int) bool {
		return invocations[i].timestamp.Before(invocations[j].timestamp)
	})

	ids := make([]string, len(invocations))
	for i, inv := range invocations {
		ids[i] = inv.id
	}

	return ids
}

// filterEventsByInvocationRange returns events within the specified invocation range (inclusive)
func (s *Selector) filterEventsByInvocationRange(
	events []*session.Event,
	startInvocationID, endInvocationID string,
) []*session.Event {
	result := make([]*session.Event, 0, len(events))
	inRange := false

	for _, event := range events {
		if event.InvocationID == startInvocationID {
			inRange = true
		}

		if inRange {
			result = append(result, event)
		}

		if event.InvocationID == endInvocationID {
			inRange = false
		}
	}

	return result
}
