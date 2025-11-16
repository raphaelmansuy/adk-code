// Package tracking provides token usage tracking for the code agent.
package tracking

import (
	"sync"
	"time"

	"google.golang.org/genai"
)

// TokenMetrics represents token usage information for a single API call.
type TokenMetrics struct {
	PromptTokens   int32
	CachedTokens   int32
	ResponseTokens int32
	ThoughtTokens  int32
	ToolUseTokens  int32
	TotalTokens    int32
	Timestamp      time.Time
	RequestID      string
}

// SessionTokens tracks cumulative token usage across a session.
type SessionTokens struct {
	mu                  sync.RWMutex
	TotalPromptTokens   int64
	TotalCachedTokens   int64
	TotalResponseTokens int64
	TotalThoughtTokens  int64
	TotalToolUseTokens  int64
	TotalTokens         int64
	RequestCount        int
	Metrics             []TokenMetrics
	SessionStartTime    time.Time
	// Track previous API response totals to calculate per-request deltas
	// API returns cumulative values, so we need to subtract previous to get current request's cost
	PreviousPromptTotal   int32
	PreviousCachedTotal   int32
	PreviousResponseTotal int32
	PreviousThoughtTotal  int32
	PreviousToolUseTotal  int32
}

// NewSessionTokens creates a new session token tracker.
func NewSessionTokens() *SessionTokens {
	return &SessionTokens{
		Metrics:          make([]TokenMetrics, 0),
		SessionStartTime: time.Now(),
	}
}

// RecordMetrics records token usage from a GenerateContentResponseUsageMetadata.
// For multi-turn conversations, the API returns cumulative token counts.
// We calculate the per-request delta for each component (prompt, response, cached, etc.)
// to show accurate current request usage.
func (st *SessionTokens) RecordMetrics(metadata *genai.GenerateContentResponseUsageMetadata, requestID string) {
	if metadata == nil {
		return
	}

	st.mu.Lock()
	defer st.mu.Unlock()

	// Calculate per-request deltas for each component
	// The API returns cumulative values, so we subtract the previous total to get this request's cost
	promptDelta := metadata.PromptTokenCount - st.PreviousPromptTotal
	responseDelta := metadata.CandidatesTokenCount - st.PreviousResponseTotal
	cachedDelta := metadata.CachedContentTokenCount - st.PreviousCachedTotal
	thoughtDelta := metadata.ThoughtsTokenCount - st.PreviousThoughtTotal
	toolUseDelta := metadata.ToolUsePromptTokenCount - st.PreviousToolUseTotal

	// Ensure we don't get negative values (safeguard against API quirks)
	if promptDelta < 0 {
		promptDelta = metadata.PromptTokenCount
	}
	if responseDelta < 0 {
		responseDelta = metadata.CandidatesTokenCount
	}
	if cachedDelta < 0 {
		cachedDelta = metadata.CachedContentTokenCount
	}
	if thoughtDelta < 0 {
		thoughtDelta = metadata.ThoughtsTokenCount
	}
	if toolUseDelta < 0 {
		toolUseDelta = metadata.ToolUsePromptTokenCount
	}

	// Total for this request = input (prompt) + output (response) + cached
	// This is the actual cost of this single request
	perRequestTotal := promptDelta + responseDelta + cachedDelta + thoughtDelta + toolUseDelta

	metric := TokenMetrics{
		PromptTokens:   promptDelta,
		CachedTokens:   cachedDelta,
		ResponseTokens: responseDelta,
		ThoughtTokens:  thoughtDelta,
		ToolUseTokens:  toolUseDelta,
		TotalTokens:    perRequestTotal, // Only this request's cost, not cumulative
		Timestamp:      time.Now(),
		RequestID:      requestID,
	}

	st.Metrics = append(st.Metrics, metric)

	// Accumulate the per-request deltas for session totals
	st.TotalPromptTokens += int64(promptDelta)
	st.TotalCachedTokens += int64(cachedDelta)
	st.TotalResponseTokens += int64(responseDelta)
	st.TotalThoughtTokens += int64(thoughtDelta)
	st.TotalToolUseTokens += int64(toolUseDelta)
	st.TotalTokens += int64(perRequestTotal)

	// Update previous totals for next request's delta calculation
	st.PreviousPromptTotal = metadata.PromptTokenCount
	st.PreviousResponseTotal = metadata.CandidatesTokenCount
	st.PreviousCachedTotal = metadata.CachedContentTokenCount
	st.PreviousThoughtTotal = metadata.ThoughtsTokenCount
	st.PreviousToolUseTotal = metadata.ToolUsePromptTokenCount

	st.RequestCount++
}

// GetSummary returns a formatted summary of token usage.
func (st *SessionTokens) GetSummary() *Summary {
	st.mu.RLock()
	defer st.mu.RUnlock()

	duration := time.Since(st.SessionStartTime)

	return &Summary{
		TotalPromptTokens:   st.TotalPromptTokens,
		TotalCachedTokens:   st.TotalCachedTokens,
		TotalResponseTokens: st.TotalResponseTokens,
		TotalThoughtTokens:  st.TotalThoughtTokens,
		TotalToolUseTokens:  st.TotalToolUseTokens,
		TotalTokens:         st.TotalTokens,
		RequestCount:        st.RequestCount,
		AvgTokensPerRequest: getAverage(st.TotalTokens, int64(st.RequestCount)),
		SessionDuration:     duration,
		RequestMetrics:      st.Metrics,
	}
}

// Summary represents a summary of token usage.
type Summary struct {
	TotalPromptTokens   int64
	TotalCachedTokens   int64
	TotalResponseTokens int64
	TotalThoughtTokens  int64
	TotalToolUseTokens  int64
	TotalTokens         int64
	RequestCount        int
	AvgTokensPerRequest float64
	SessionDuration     time.Duration
	RequestMetrics      []TokenMetrics
}

// GlobalTracker tracks tokens across all sessions.
type GlobalTracker struct {
	mu       sync.RWMutex
	Sessions map[string]*SessionTokens
}

// NewGlobalTracker creates a new global token tracker.
func NewGlobalTracker() *GlobalTracker {
	return &GlobalTracker{
		Sessions: make(map[string]*SessionTokens),
	}
}

// GetOrCreateSession gets or creates a session token tracker.
func (gt *GlobalTracker) GetOrCreateSession(sessionID string) *SessionTokens {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	if st, exists := gt.Sessions[sessionID]; exists {
		return st
	}

	st := NewSessionTokens()
	gt.Sessions[sessionID] = st
	return st
}

// GetSession retrieves a session token tracker.
func (gt *GlobalTracker) GetSession(sessionID string) *SessionTokens {
	gt.mu.RLock()
	defer gt.mu.RUnlock()

	return gt.Sessions[sessionID]
}

// GetGlobalSummary returns a summary of all tokens across all sessions.
func (gt *GlobalTracker) GetGlobalSummary() *GlobalSummary {
	gt.mu.RLock()
	defer gt.mu.RUnlock()

	summary := &GlobalSummary{
		Sessions:  make(map[string]*Summary),
		StartTime: time.Now(),
	}

	for sessionID, session := range gt.Sessions {
		summary.Sessions[sessionID] = session.GetSummary()
		summary.TotalTokens += session.GetSummary().TotalTokens
		summary.TotalRequests += int64(session.GetSummary().RequestCount)
	}

	if summary.TotalRequests > 0 {
		summary.AvgTokensPerRequest = float64(summary.TotalTokens) / float64(summary.TotalRequests)
	}

	return summary
}

// GlobalSummary represents a summary of token usage across all sessions.
type GlobalSummary struct {
	Sessions            map[string]*Summary
	TotalTokens         int64
	TotalRequests       int64
	AvgTokensPerRequest float64
	StartTime           time.Time
}

// Helper function
func getAverage(total int64, count int64) float64 {
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}
