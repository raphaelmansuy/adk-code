package compaction

import (
	"context"
	"fmt"

	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
)

// Coordinator orchestrates the compaction process
type Coordinator struct {
	config         *Config
	selector       *Selector
	agentLLM       model.LLM
	sessionService session.Service
}

// NewCoordinator creates a new compaction coordinator
func NewCoordinator(
	config *Config,
	selector *Selector,
	agentLLM model.LLM,
	sessionService session.Service,
) *Coordinator {
	return &Coordinator{
		config:         config,
		selector:       selector,
		agentLLM:       agentLLM,
		sessionService: sessionService,
	}
}

// RunCompaction triggers compaction if thresholds are met
func (c *Coordinator) RunCompaction(
	ctx context.Context,
	sess session.Session,
) error {
	if sess == nil {
		return fmt.Errorf("session is nil")
	}

	// Get all events (unfiltered)
	events := sess.Events()
	eventList := make([]*session.Event, 0, events.Len())
	for event := range events.All() {
		eventList = append(eventList, event)
	}

	// Select events to compact
	toCompact, err := c.selector.SelectEventsToCompact(eventList)
	if err != nil {
		return fmt.Errorf("error selecting events for compaction: %w", err)
	}

	// If no events to compact, return early
	if len(toCompact) == 0 {
		return nil // No compaction needed
	}

	// Create summarizer with agent's LLM
	summarizer := NewLLMSummarizer(c.agentLLM, c.config)

	// Summarize selected events
	compactionEvent, err := summarizer.Summarize(ctx, toCompact)
	if err != nil {
		return fmt.Errorf("error summarizing events: %w", err)
	}

	// Append compaction event to session
	// Original events remain in storage
	return c.sessionService.AppendEvent(ctx, sess, compactionEvent)
}
