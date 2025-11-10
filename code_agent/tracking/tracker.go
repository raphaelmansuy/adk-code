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
}

// NewSessionTokens creates a new session token tracker.
func NewSessionTokens() *SessionTokens {
	return &SessionTokens{
		Metrics:          make([]TokenMetrics, 0),
		SessionStartTime: time.Now(),
	}
}

// RecordMetrics records token usage from a GenerateContentResponseUsageMetadata.
func (st *SessionTokens) RecordMetrics(metadata *genai.GenerateContentResponseUsageMetadata, requestID string) {
	if metadata == nil {
		return
	}

	st.mu.Lock()
	defer st.mu.Unlock()

	metric := TokenMetrics{
		PromptTokens:   metadata.PromptTokenCount,
		CachedTokens:   metadata.CachedContentTokenCount,
		ResponseTokens: metadata.CandidatesTokenCount,
		ThoughtTokens:  metadata.ThoughtsTokenCount,
		ToolUseTokens:  metadata.ToolUsePromptTokenCount,
		TotalTokens:    metadata.TotalTokenCount,
		Timestamp:      time.Now(),
		RequestID:      requestID,
	}

	st.Metrics = append(st.Metrics, metric)
	st.TotalPromptTokens += int64(metadata.PromptTokenCount)
	st.TotalCachedTokens += int64(metadata.CachedContentTokenCount)
	st.TotalResponseTokens += int64(metadata.CandidatesTokenCount)
	st.TotalThoughtTokens += int64(metadata.ThoughtsTokenCount)
	st.TotalToolUseTokens += int64(metadata.ToolUsePromptTokenCount)
	st.TotalTokens += int64(metadata.TotalTokenCount)
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
