# ADR-009: Session History Compaction via Sliding Window Summarization

**Status:** Proposed  
**Date:** 2025-01-16  
**Authors:** ADK-Code Team  
**Deciders:** Technical Lead, Architecture Team

---

## Context

Session histories in `adk-code` grow unbounded as users interact with agents. Each event (user messages, agent responses, function calls/responses) is stored in the session, leading to:

1. **Token Budget Exhaustion**: Long conversations exceed model context windows (e.g., Gemini 2.0: 1M tokens)
2. **Database Bloat**: SQLite session storage grows linearly with O(n) events
3. **Performance Degradation**: Event retrieval and processing slows as history lengthens
4. **Cost Escalation**: API costs scale with input token count on every turn

**Key Architectural Decision:** Compaction uses the **Agent's current LLM model** for summarization, ensuring consistency with user's model choice and eliminating the need for separate API configuration.

**Current Implementation Gap:**
- `research/adk-python` implements sliding window compaction via LLM-based summarization
- `research/adk-go` has **no compaction** mechanism (confirmed via codebase grep)
- `adk-code` inherits from `adk-go` → **no compaction support**

**Token Doubling Issue:**
Investigation revealed that without compaction, token usage doubles every turn as full history is resent:
```
Turn 1: 100 tokens
Turn 2: 100 (history) + 150 (new) = 250 tokens  
Turn 3: 250 (history) + 200 (new) = 450 tokens  ← Exponential growth
```

---

## Decision

Implement **Sliding Window Compaction** in `adk-code` following the mathematical model from `research/adk-python`, adapted for Go and enhanced with token-aware triggering.

### Core Principle: Immutability with Selective Context

**CRITICAL**: Session history is **immutable and append-only** in storage, but **selective** when building LLM context. This design follows ADK Python's proven model:

✅ **Original events are NEVER deleted or modified in storage**  
✅ **Compaction creates a new Event** with metadata stored in `CustomMetadata` field  
✅ **Compaction event is appended** to the session (not replacing original events in storage)  
✅ **All events remain in storage** for full audit trail and debugging  
✅ **Context building is selective**: When a compaction event exists, **original events within its range are excluded** from LLM context and **replaced by the summary**

**Key Distinction:**

- **Storage layer**: ALL events preserved (immutable)
- **Context layer**: Only summaries sent for compacted ranges (token-efficient)

**Architectural Constraint:**

⚠️ **Cannot modify upstream ADK Go types**: `EventActions` is defined in `google.golang.org/adk/session` (v0.1.0). We cannot add fields to it.

✅ **Solution**: Use `event.CustomMetadata["_adk_compaction"]` to store compaction data. This field already exists in `model.LLMResponse` and is serialized by the persistence layer.

Example session after two compactions:

```text
Storage (ALL events preserved):
[E₁, E₂, E₃, E₄, E₅, C(1-3), E₆, E₇, E₈, C(4-7), E₉, E₁₀]

LLM Context (compacted ranges replaced by summaries):
[C(1-3), E₆, E₇, E₈, C(4-7), E₉, E₁₀]
 └─────┘ └──────────┘ └─────┘ └──────┘
 Summary    Kept      Summary   Kept
```

**How it works:**

- `C(1-3)` replaces `E₁, E₂, E₃` in LLM context (but not in storage)
- `C(4-7)` replaces `E₄, E₅, E₆, E₇` in LLM context (overlap with E₆, E₇)
- Recent uncompacted events (`E₉, E₁₀`) always included in full

### Architecture

```text
┌─────────────────────────────────────────────────────────────────┐
│                         Session Manager                          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │         Event History (IMMUTABLE, APPEND-ONLY)            │  │
│  │  [E₁, E₂, E₃, C(1-3), E₄, E₅, C(3-5), E₆, E₇, ...]     │  │
│  │   └─Original─┘ └Compact┘ └─Original─┘ └Compact┘          │  │
│  │                                                            │  │
│  │  C = Event with CustomMetadata["_adk_compaction"]        │  │
│  │  ALL events preserved for audit trail & debugging         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌──────────────────────────▼───────────────────────────────┐  │
│  │      CompactionSessionService (WRAPPER)                   │  │
│  │  Wraps: SQLiteSessionService                              │  │
│  │  • Intercepts Get() calls                                 │  │
│  │  • Returns FilteredSession wrapper                        │  │
│  └───────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌──────────────────────────▼───────────────────────────────┐  │
│  │      FilteredSession (WRAPPER)                            │  │
│  │  Wraps: session.Session                                   │  │
│  │  • Overrides Events() method                              │  │
│  │  • Applies compaction filtering                           │  │
│  └───────────────────────────────────────────────────────────┘  │
│                              │                                   │
│             Context Building │ (for LLM) - FILTERING LAYER      │
│  ┌──────────────────────────▼───────────────────────────────┐  │
│  │  FilteredEvents.All() iterator:                           │  │
│  │  1. Scan for events with CustomMetadata["_adk_compaction"]│  │
│  │  2. EXCLUDE original events within compacted time ranges  │  │
│  │  3. INCLUDE compaction summaries from CustomMetadata      │  │
│  │  4. INCLUDE all uncompacted recent events                 │  │
│  │                                                            │  │
│  │  Result: [C(1-3), C(3-5), E₆, E₇, ...]                   │  │
│  │  Token savings: 60-80% reduction in context size          │  │
│  └───────────────────────────────────────────────────────────┘  │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │           Compaction Coordinator (New)                    │  │
│  │  • Monitors invocation completion                         │  │
│  │  • Triggers compaction based on thresholds                │  │
│  │  • Manages sliding window overlap                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                   │
│            ┌─────────────────┴────────────────┐                 │
│            ▼                                   ▼                 │
│  ┌──────────────────┐              ┌──────────────────┐         │
│  │   Token Counter   │              │  Event Selector  │         │
│  │  • Tracks usage   │              │  • Windowing     │         │
│  │  • Threshold      │              │  • Overlap mgmt  │         │
│  └──────────────────┘              └──────────────────┘         │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │               LLM Summarizer (New)                        │  │
│  │  • Formats events for prompt                              │  │
│  │  • Calls Gemini API for summarization                     │  │
│  │  • Creates Event with Actions.Compaction                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │           SQLite Persistence (Standard)                   │  │
│  │  • Appends compaction event like any other event          │  │
│  │  • No special schema changes needed                       │  │
│  │  • Check event.Actions.Compaction != nil for type         │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Mathematical Model

### Definitions

Let:

- `E = {e₁, e₂, ..., eₙ}` = Sequence of events in session
- `I(e)` = Invocation ID of event `e`
- `T(e)` = Timestamp of event `e`  
- `τ(e)` = Token count of event `e`
- `θ` = Compaction invocation threshold (config parameter)
- `ω` = Overlap window size (config parameter)

### Sliding Window Function

Define sliding window `W` at time `t`:

```text
W(t, θ, ω) = {eᵢ ∈ E | i_start ≤ i ≤ i_end}

where:
  i_end = max{i | T(eᵢ) ≤ t}
  i_start = max{0, i_end - (θ + ω - 1)}
```

### Compaction Trigger Predicate

Compaction occurs when:

```text
∃t : |{I(e) | e ∈ E_new(t)}| ≥ θ

where:
  E_new(t) = {e ∈ E | T(e) > T(C_last)}
  C_last = most recent compaction event
```

### Overlap Preservation

To maintain context continuity:

```text
W_next = W_prev ∩ W_curr

Specifically:
  Overlap = {e ∈ W_prev | T(e) ∈ [T(e_end-ω), T(e_end)]}
```

### Token-Aware Enhancement (adk-code Extension)

Add adaptive triggering based on token budget:

```text
Compact if: Σ τ(e) > ρ · Λ
            e∈E_active

where:
  ρ = safety ratio (default: 0.7)
  Λ = model context window (e.g., 1M for Gemini 2.0)
  E_active = events since last compaction
```

---

## Implementation Components

### 1. Compaction Configuration

```go
// File: internal/session/compaction/config.go
package compaction

type Config struct {
    // Invocation-based triggering
    InvocationThreshold int     // θ: Number of invocations to trigger
    OverlapSize         int     // ω: Overlapping invocations for context
    
    // Token-aware triggering (adk-code enhancement)
    TokenThreshold      int     // ρ·Λ: Max tokens before forced compaction
    SafetyRatio         float64 // ρ: Fraction of context window (0.7 = 70%)
    
    // Prompt configuration
    PromptTemplate      string  // Custom prompt (optional)
    
    // Note: No SummarizerModel needed - uses Agent's current LLM
}

func DefaultConfig() *Config {
    return &Config{
        InvocationThreshold: 5,      // Compact every 5 invocations
        OverlapSize:         2,      // Keep 2 invocations overlap
        TokenThreshold:      700000, // 700k tokens (70% of 1M)
        SafetyRatio:         0.7,
        PromptTemplate:      defaultPromptTemplate,
        // Uses Agent's current LLM - no separate model configuration
    }
}
```

### 2. Compaction Metadata Structure (CustomMetadata Approach)

**Why CustomMetadata?** ADK Go's `EventActions` is defined in upstream `google.golang.org/adk/session` package. We cannot modify it without forking. However, `model.LLMResponse` (embedded in `session.Event`) has a `CustomMetadata map[string]any` field that is already serialized by the persistence layer.

```go
// File: internal/session/compaction/types.go
package compaction

import (
    "time"
    "google.golang.org/genai"
)

// CompactionMetadata is stored in event.CustomMetadata["_adk_compaction"]
type CompactionMetadata struct {
    StartTimestamp    time.Time       `json:"start_timestamp"`
    EndTimestamp      time.Time       `json:"end_timestamp"`
    StartInvocationID string          `json:"start_invocation_id,omitempty"`
    EndInvocationID   string          `json:"end_invocation_id,omitempty"`
    
    // Summary stored as serialized genai.Content
    CompactedContentJSON string       `json:"compacted_content_json"`
    
    // Metrics (adk-code enhancement)
    EventCount       int              `json:"event_count"`
    OriginalTokens   int              `json:"original_tokens"`
    CompactedTokens  int              `json:"compacted_tokens"`
    CompressionRatio float64          `json:"compression_ratio"`
}

// Helper functions
const CompactionMetadataKey = "_adk_compaction"

// IsCompactionEvent checks if an event contains compaction metadata
func IsCompactionEvent(event *session.Event) bool {
    if event.CustomMetadata == nil {
        return false
    }
    _, exists := event.CustomMetadata[CompactionMetadataKey]
    return exists
}

// GetCompactionMetadata extracts compaction data from event
func GetCompactionMetadata(event *session.Event) (*CompactionMetadata, error) {
    if !IsCompactionEvent(event) {
        return nil, fmt.Errorf("event is not a compaction event")
    }
    
    data := event.CustomMetadata[CompactionMetadataKey]
    
    // Marshal to JSON and unmarshal to struct
    jsonData, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }
    
    var metadata CompactionMetadata
    if err := json.Unmarshal(jsonData, &metadata); err != nil {
        return nil, err
    }
    
    return &metadata, nil
}

// SetCompactionMetadata sets compaction data on an event
func SetCompactionMetadata(event *session.Event, metadata *CompactionMetadata) error {
    if event.CustomMetadata == nil {
        event.CustomMetadata = make(map[string]any)
    }
    
    // Convert to map for storage
    jsonData, err := json.Marshal(metadata)
    if err != nil {
        return err
    }
    
    var dataMap map[string]any
    if err := json.Unmarshal(jsonData, &dataMap); err != nil {
        return err
    }
    
    event.CustomMetadata[CompactionMetadataKey] = dataMap
    return nil
}
```

**Key Design Points:**

- ✅ Uses **existing** `CustomMetadata` field from `model.LLMResponse`
- ✅ No modifications to upstream ADK Go types required
- ✅ Already serialized/deserialized by persistence layer
- ✅ Detection: `IsCompactionEvent()` checks for `"_adk_compaction"` key
- ✅ Backward compatible - old events without this key are unaffected

### 3. Event Selector

```go
// File: internal/session/compaction/selector.go
package compaction

type Selector struct {
    config *Config
}

func (s *Selector) SelectEventsToCompact(
    events []*session.Event,
) ([]*session.Event, error) {
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
    invocationIDs := sortInvocationsByTime(invocationMap)
    
    // Calculate window: [start_idx, end_idx]
    endInvocationID := invocationIDs[len(invocationIDs)-1]
    startIdx = max(0, len(invocationIDs) - s.config.InvocationThreshold - s.config.OverlapSize)
    startInvocationID := invocationIDs[startIdx]
    
    // Collect events in window
    return filterEventsByInvocationRange(
        events, 
        startInvocationID, 
        endInvocationID,
    ), nil
}
```

### 4. LLM Summarizer

```go
// File: internal/session/compaction/summarizer.go
package compaction

import (
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

type LLMSummarizer struct {
	llm    model.LLM  // Agent's current LLM model
	config *Config
}

const defaultPromptTemplate = `The following is a conversation history between a user and an AI agent. 
Summarize the conversation concisely, focusing on:
1. Key decisions and outcomes
2. Important context and state changes
3. Unresolved questions or pending tasks
4. Tool calls and their results

Keep the summary under 500 tokens while preserving critical information.

Conversation History:
%s
`

func (ls *LLMSummarizer) Summarize(
	ctx context.Context,
	events []*session.Event,
) (*session.Event, error) {
	// Format events for prompt
	conversationText := ls.formatEvents(events)
	prompt := fmt.Sprintf(ls.config.PromptTemplate, conversationText)
	
	// Call LLM using Agent's model
	llmRequest := &model.LLMRequest{
		Model: ls.llm.Name(),
		Contents: []*genai.Content{
			{
				Role: "user",
				Parts: []genai.Part{
					genai.Text(prompt),
				},
			},
		},
		Config: &genai.GenerateContentConfig{},
	}
	
	// Generate content using the agent's LLM
	var summaryContent *genai.Content
	var usageMetadata *genai.GenerateContentResponseUsageMetadata
	
	for resp, err := range ls.llm.GenerateContent(ctx, llmRequest, false) {
		if err != nil {
			return nil, err
		}
		if resp.Content != nil {
			summaryContent = resp.Content
			usageMetadata = resp.UsageMetadata
			break
		}
	}
	
	if summaryContent == nil {
		return nil, fmt.Errorf("no summary content generated")
	}
	
	// Ensure role is 'model' (following ADK Python)
	summaryContent.Role = "model"        // Calculate metrics (adk-code enhancement)
    originalTokens := ls.countTokens(events)
    compactedTokens := 0
    if usageMetadata != nil {
        compactedTokens = int(usageMetadata.TotalTokenCount)
    }
    
    // Serialize summary content to JSON
    summaryJSON, err := json.Marshal(summaryContent)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal summary content: %w", err)
    }
    
    // Create compaction metadata
    metadata := &CompactionMetadata{
        StartTimestamp:       events[0].Timestamp,
        EndTimestamp:         events[len(events)-1].Timestamp,
        StartInvocationID:    events[0].InvocationID,
        EndInvocationID:      events[len(events)-1].InvocationID,
        CompactedContentJSON: string(summaryJSON),
        EventCount:           len(events),
        OriginalTokens:       originalTokens,
        CompactedTokens:      compactedTokens,
        CompressionRatio:     float64(originalTokens) / float64(compactedTokens),
    }
    
    // Create compaction event (following ADK Python pattern)
    compactionEvent := session.NewEvent(uuid.NewString())
    compactionEvent.Author = "user"  // ADK Python uses "user" as author
    compactionEvent.Content = summaryContent  // For display purposes
    
    // Store compaction metadata in CustomMetadata
    if err := SetCompactionMetadata(compactionEvent, metadata); err != nil {
        return nil, fmt.Errorf("failed to set compaction metadata: %w", err)
    }
    
    return compactionEvent, nil
}

func (ls *LLMSummarizer) formatEvents(events []*session.Event) string {
    var sb strings.Builder
    for _, event := range events {
        if event.Content != nil && len(event.Content.Parts) > 0 {
            for _, part := range event.Content.Parts {
                if part.Text != nil {
                    sb.WriteString(fmt.Sprintf("%s: %s\n", 
                        event.Author, *part.Text))
                }
            }
        }
    }
    return sb.String()
}
```

### 5. Session Service Wrapper (Filtering Layer)

```go
// File: internal/session/compaction/service.go
package compaction

import (
    "context"
    "google.golang.org/adk/session"
)

// CompactionSessionService wraps the underlying session service
// to provide transparent compaction filtering when sessions are retrieved
type CompactionSessionService struct {
    underlying session.Service
    config     *Config
}

// NewCompactionService creates a wrapper around the session service
func NewCompactionService(underlying session.Service, config *Config) *CompactionSessionService {
    return &CompactionSessionService{
        underlying: underlying,
        config:     config,
    }
}

// Get wraps the underlying Get to return a filtered session
func (c *CompactionSessionService) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
    resp, err := c.underlying.Get(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // Wrap the session with filtering layer
    filteredSession := NewFilteredSession(resp.Session)
    
    return &session.GetResponse{
        Session: filteredSession,
    }, nil
}

// Pass-through methods (delegate to underlying service)
func (c *CompactionSessionService) Create(ctx context.Context, req *session.CreateRequest) (*session.CreateResponse, error) {
    return c.underlying.Create(ctx, req)
}

func (c *CompactionSessionService) List(ctx context.Context, req *session.ListRequest) (*session.ListResponse, error) {
    return c.underlying.List(ctx, req)
}

func (c *CompactionSessionService) Delete(ctx context.Context, req *session.DeleteRequest) error {
    return c.underlying.Delete(ctx, req)
}

func (c *CompactionSessionService) AppendEvent(ctx context.Context, sess session.Session, event *session.Event) error {
    return c.underlying.AppendEvent(ctx, sess, event)
}
```

```go
// File: internal/session/compaction/filtered_session.go
package compaction

import (
    "iter"
    "time"
    "google.golang.org/adk/session"
)

// FilteredSession wraps a session to provide compaction-aware event filtering
type FilteredSession struct {
    underlying session.Session
}

func NewFilteredSession(underlying session.Session) *FilteredSession {
    return &FilteredSession{underlying: underlying}
}

// Pass-through methods
func (fs *FilteredSession) ID() string                    { return fs.underlying.ID() }
func (fs *FilteredSession) AppName() string               { return fs.underlying.AppName() }
func (fs *FilteredSession) UserID() string                { return fs.underlying.UserID() }
func (fs *FilteredSession) State() session.State          { return fs.underlying.State() }
func (fs *FilteredSession) LastUpdateTime() time.Time     { return fs.underlying.LastUpdateTime() }

// Events returns a filtered view that excludes compacted events
func (fs *FilteredSession) Events() session.Events {
    return NewFilteredEvents(fs.underlying.Events())
}
```

```go
// File: internal/session/compaction/filtered_events.go
package compaction

import (
    "encoding/json"
    "iter"
    "google.golang.org/adk/session"
    "google.golang.org/genai"
)

// FilteredEvents implements session.Events with compaction filtering
type FilteredEvents struct {
    underlying session.Events
    filtered   []*session.Event
}

func NewFilteredEvents(underlying session.Events) *FilteredEvents {
    filtered := filterCompactedEvents(underlying)
    return &FilteredEvents{
        underlying: underlying,
        filtered:   filtered,
    }
}

func (fe *FilteredEvents) All() iter.Seq[*session.Event] {
    return func(yield func(*session.Event) bool) {
        for _, event := range fe.filtered {
            if !yield(event) {
                return
            }
        }
    }
}

func (fe *FilteredEvents) Len() int {
    return len(fe.filtered)
}

func (fe *FilteredEvents) At(i int) *session.Event {
    if i >= 0 && i < len(fe.filtered) {
        return fe.filtered[i]
    }
    return nil
}

// filterCompactedEvents implements the filtering logic
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
            // But replace Content with the stored summary
            metadata, _ := GetCompactionMetadata(event)
            var summaryContent genai.Content
            json.Unmarshal([]byte(metadata.CompactedContentJSON), &summaryContent)
            
            // Create a new event with the summary content
            filteredEvent := *event
            filteredEvent.Content = &summaryContent
            filtered = append(filtered, &filteredEvent)
        } else {
            // Check if this event is within any compacted range
            withinCompactedRange := false
            for _, cr := range compactionRanges {
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
```

### 6. Compaction Coordinator

```go
// File: internal/session/compaction/coordinator.go
package compaction

import (
    "context"
    "google.golang.org/adk/session"
)

type Coordinator struct {
    config         *Config
    selector       *Selector
    agentLLM       model.LLM  // Agent's LLM model for summarization
    sessionService session.Service
}

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

func (c *Coordinator) RunCompaction(
    ctx context.Context,
    sess session.Session,
) error {
    // Get all events (unfiltered)
    events := sess.Events()
    eventList := make([]*session.Event, 0, events.Len())
    for event := range events.All() {
        eventList = append(eventList, event)
    }
    
    // Select events to compact
    toCompact, err := c.selector.SelectEventsToCompact(eventList)
    if err != nil || len(toCompact) == 0 {
        return err // No compaction needed
    }
    
    // Create summarizer with agent's LLM
    summarizer := &LLMSummarizer{
        llm:    c.agentLLM,
        config: c.config,
    }
    
    // Summarize selected events
    compactionEvent, err := summarizer.Summarize(ctx, toCompact)
    if err != nil {
        return err
    }
    
    // Append compaction event to session
    // Original events remain in storage
    return c.sessionService.AppendEvent(ctx, sess, compactionEvent)
}
```

### 7. Integration into Session Manager

```go
// File: internal/session/manager.go (modification)
package session

import (
    "adk-code/internal/session/compaction"
    "adk-code/internal/session/persistence"
)

func NewSessionManager(appName, dbPath string) (*SessionManager, error) {
    // ... existing dbPath handling ...
    
    // Create base persistence service
    baseSvc, err := persistence.NewSQLiteSessionService(dbPath)
    if err != nil {
        return nil, pkgerrors.Wrap(pkgerrors.CodeInternal, "failed to create session service", err)
    }
    
    // Wrap with compaction layer
    compactionConfig := compaction.DefaultConfig()
    compactionSvc := compaction.NewCompactionService(baseSvc, compactionConfig)
    
    return &SessionManager{
        sessionService: compactionSvc,  // Use wrapped service
        dbPath:         dbPath,
        appName:        appName,
    }, nil
}
```

```go
// File: internal/repl/repl.go or orchestration layer
// Add compaction coordinator hook after invocation completes

func (r *REPL) runWithCompaction(
    ctx context.Context,
    userMsg *genai.Content,
    requestID string,
) {
    // ... existing event processing loop ...
    
agentLoop:
    for {
        select {
        case <-ctx.Done():
            break agentLoop
        case result, ok := <-eventChan:
            if !ok {
                break agentLoop
            }
            // Process event...
        }
    }
    
    // After all events processed, trigger compaction asynchronously
    if !hasError && r.config.CompactionEnabled {
        go func() {
            // Get current session
            sess, err := r.config.SessionManager.GetSession(
                context.Background(),
                r.config.UserID,
                r.config.SessionName,
            )
            if err != nil {
                log.Printf("Failed to get session for compaction: %v", err)
                return
            }
            
            // Create coordinator with agent's LLM
            coordinator := compaction.NewCoordinator(
                r.config.CompactionConfig,
                compaction.NewSelector(r.config.CompactionConfig),
                r.config.Agent.LLM(),  // Use agent's current LLM
                r.config.SessionManager.GetService(),
            )
            
            if err := coordinator.RunCompaction(context.Background(), sess); err != nil {
                log.Printf("Compaction failed: %v", err)
            }
        }()
    }
}
```---

## Consequences

### Positive

✅ **Token Efficiency**: Reduces context size by 60-80% by replacing verbose events with summaries  
✅ **Cost Reduction**: Lower API costs due to significantly reduced input tokens  
✅ **Scalability**: Supports arbitrarily long conversations within model limits  
✅ **Immutable Audit Trail**: All original events preserved in storage for debugging and compliance  
✅ **Simple Storage**: No schema changes needed; compaction event is just another event  
✅ **Selective Context**: Filtering layer excludes compacted events from LLM context while keeping them in storage  
✅ **Token Tracking**: Enhanced with compression metrics (improvement over adk-python)  
✅ **Proven Design**: Follows battle-tested ADK Python implementation pattern exactly

### Negative

⚠️ **LLM Dependency**: Summarization requires additional API call (cost: ~500 tokens/summary)  
⚠️ **Latency**: Compaction adds 1-2s delay post-invocation (mitigated by async execution)  
⚠️ **Lossy Compression**: Fine-grained details may be lost in summaries  
⚠️ **Complexity**: Additional config parameters require tuning  
⚠️ **Testing Burden**: Requires integration tests with real LLM calls

### Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Summarization errors lose critical context | Keep overlap window (ω ≥ 2), store raw events permanently |
| Compaction loop consumes too many tokens | Add max compaction limit (e.g., 5 summaries/session) |
| Database schema changes break existing sessions | Migration script, backward-compatible event structure |
| Async compaction fails silently | Structured logging, Prometheus metrics (future ADR) |

---

## Compatibility

- **ADK Python Compatibility**: ✅ **Fully aligned** with Python implementation's immutable model
- **adk-go Compatibility**: Not applicable (upstream lacks compaction)
- **Backward Compatibility**: ✅ Existing sessions work without modification; compaction is opt-in via config
- **Storage Schema**: ✅ **No schema changes required** - compaction event uses existing Event structure with Actions.Compaction field
- **Event Detection**: Check `event.Actions.Compaction != nil` to identify compaction events
- **Migration**: None needed - new field is nullable and backward compatible

---

## Alternatives Considered

### 1. Fixed-Size Ring Buffer (Rejected)

**Approach**: Keep last N events, discard oldest  
**Rejected Because**: Loses all historical context; no summarization

### 2. Hierarchical Summarization (Deferred)

**Approach**: Multi-level summaries (hour → day → week)  
**Deferred Because**: Over-engineering for MVP; can extend later

### 3. Manual User-Triggered Compaction (Rejected)

**Approach**: `/compact` REPL command  
**Rejected Because**: Poor UX; automation is better

### 4. Token-Only Triggering (Rejected)

**Approach**: Only compact when token threshold exceeded  
**Rejected Because**: Unpredictable timing; harder to debug

---

## Testing Strategy

```go
// File: internal/session/compaction/coordinator_test.go

func TestCompactionE2E(t *testing.T) {
    tests := []struct {
        name           string
        invocations    int
        threshold      int
        overlap        int
        expectedEvents int
    }{
        {
            name:           "no_compaction_below_threshold",
            invocations:    3,
            threshold:      5,
            overlap:        2,
            expectedEvents: 3, // No compaction event created
        },
        {
            name:           "single_compaction_at_threshold",
            invocations:    5,
            threshold:      5,
            overlap:        2,
            expectedEvents: 6, // 5 original + 1 compaction
        },
        {
            name:           "multiple_compactions",
            invocations:    12,
            threshold:      5,
            overlap:        2,
            expectedEvents: 14, // 12 original + 2 compactions
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock session with tt.invocations events
            // Run compaction coordinator
            // Assert expected number of events including compaction events
        })
    }
}
```

---

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1-2)

- [ ] Create `internal/session/compaction` package
- [ ] Implement `Config` and `Selector`
- [ ] Define `CompactionMetadata` struct (stored in `CustomMetadata`)
- [ ] Implement helper functions: `IsCompactionEvent()`, `GetCompactionMetadata()`, `SetCompactionMetadata()`
- [ ] Create wrapper types: `CompactionSessionService`, `FilteredSession`, `FilteredEvents`
- [ ] ~~Database migration~~ (NOT NEEDED - `CustomMetadata` already stored)

### Phase 2: Summarization (Week 3)

- [ ] Implement `LLMSummarizer` using Agent's LLM model
- [ ] Add prompt template configuration
- [ ] Token counting utilities
- [ ] Ensure compatibility with all model backends (Gemini, Vertex AI, OpenAI)

### Phase 3: Coordination (Week 4)

- [ ] Implement `Coordinator` with async execution
- [ ] Integrate with existing `Runner`
- [ ] Add compaction metrics to `internal/tracking`

### Phase 4: Configuration & CLI (Week 5)

- [ ] Add compaction flags to CLI
- [ ] Environment variable support
- [ ] `/compaction-status` REPL command

### Phase 5: Testing & Documentation (Week 6)

- [ ] Unit tests (≥80% coverage)
- [ ] Integration tests with real Gemini API
- [ ] Update `docs/ARCHITECTURE.md`
- [ ] Update `docs/QUICK_REFERENCE.md`

---

## Configuration Example

```toml
# ~/.code_agent/config.toml
[compaction]
enabled = true
invocation_threshold = 5
overlap_size = 2
token_threshold = 700000
safety_ratio = 0.7

# Custom prompt (optional)
# prompt_template = "file://~/.code_agent/compaction_prompt.txt"

# Note: Compaction automatically uses the Agent's current model
# No separate model configuration needed
```

---

## Metrics to Track

```go
// Prometheus metrics (future ADR)
var (
    compactionTriggersTotal = prometheus.NewCounter(...)
    compactionDurationSeconds = prometheus.NewHistogram(...)
    compactionCompressionRatio = prometheus.NewGauge(...)
    compactionErrorsTotal = prometheus.NewCounter(...)
)
```

---

## Immutability Design Comparison

### ADK Python Implementation (Reference)

```python
# From research/adk-python/src/google/adk/sessions/base_session_service.py
async def append_event(self, session: Session, event: Event) -> Event:
    """Appends an event to a session object."""
    # ... validation ...
    session.events.append(event)  # APPEND ONLY - never deletes
    return event

# From research/adk-python/src/google/adk/apps/llm_event_summarizer.py
async def maybe_summarize_events(self, *, events: list[Event]) -> Optional[Event]:
    # ... summarization logic ...
    compaction = EventCompaction(
        start_timestamp=events[0].timestamp,
        end_timestamp=events[-1].timestamp,
        compacted_content=summary_content,
    )
    actions = EventActions(compaction=compaction)
    return Event(author='user', actions=actions, invocation_id=Event.new_id())
```

### adk-code Implementation (This ADR)

```go
// Aligned with Python: append-only storage, selective context
func (c *Coordinator) RunCompaction(ctx context.Context, sess session.Session) error {
    // ... selection logic ...
    compactionEvent, err := c.summarizer.Summarize(ctx, toCompact)
    
    // Store compaction metadata in CustomMetadata (not Actions)
    // This is the key difference from ADK Python due to Go type constraints
    SetCompactionMetadata(compactionEvent, metadata)
    
    // Append to session (like Python's session.events.append())
    return c.sessionService.AppendEvent(ctx, sess, compactionEvent)
}

// Context filtering via WRAPPER PATTERN (unique to adk-code)
type FilteredSession struct {
    underlying session.Session
}

func (fs *FilteredSession) Events() session.Events {
    return NewFilteredEvents(fs.underlying.Events())
}

// FilteredEvents.All() implements filtering (like Python's _process_compaction_events)
func (fe *FilteredEvents) All() iter.Seq[*session.Event] {
    return func(yield func(*session.Event) bool) {
        for _, event := range fe.filtered {
            // Filtered list already excludes events within compacted ranges
            if !yield(event) {
                return
            }
        }
    }
}

func filterCompactedEvents(events session.Events) []*session.Event {
    for _, event := range allEvents {
        if IsCompactionEvent(event) {
            // Include compaction summary (from CustomMetadata)
            metadata, _ := GetCompactionMetadata(event)
            // Deserialize and use summary content
        } else if !isWithinCompactedRange(event, compactionRanges) {
            // Include original event ONLY if NOT within compacted range
        }
        // else: Event is within compacted range → EXCLUDE from context
    }
}
```

**Key Principles**:

1. **Storage**: Complete event history maintained (immutable)
2. **Metadata**: Uses `CustomMetadata["_adk_compaction"]` instead of `Actions.Compaction`
3. **Filtering**: Wrapper pattern intercepts `Session.Events()` calls
4. **Context**: Compacted events REPLACE originals in LLM requests (selective)
5. **Result**: Token efficiency without losing audit trail or modifying ADK types

---

## References

1. **Source Analysis**:
   - `research/adk-python/src/google/adk/apps/compaction.py`
   - `research/adk-python/src/google/adk/apps/llm_event_summarizer.py`
   - `research/adk-python/src/google/adk/runners.py` (lines 388-400)

2. **Related ADRs**:
   - ADR-003: Session Persistence Design
   - ADR-006: Token Tracking Implementation (proposed)

3. **External Resources**:
   - [Gemini API Context Caching](https://ai.google.dev/gemini-api/docs/caching)
   - [Token Optimization Strategies](https://cloud.google.com/vertex-ai/generative-ai/docs/context-cache)

---

## Decision Outcome

**Approved**: Implement sliding window compaction with token-aware triggering as specified above.

**Next Steps**:

1. Create GitHub issue with implementation phases
2. Set up branch: `feature/session-compaction`
3. Begin Phase 1 implementation

**Reviewers**: @tech-lead @architecture-team
