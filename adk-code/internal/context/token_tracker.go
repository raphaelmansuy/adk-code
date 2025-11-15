package context

import (
	"sync"
	"time"
)

// TokenTracker maintains detailed token usage across turns
type TokenTracker struct {
	mu          sync.RWMutex
	sessionID   string
	modelName   string
	turns       []TurnTokenInfo
	totalTokens int
	startTime   time.Time
}

// NewTokenTracker creates a new token tracker
func NewTokenTracker(sessionID, modelName string, contextWindow int) *TokenTracker {
	return &TokenTracker{
		sessionID: sessionID,
		modelName: modelName,
		turns:     []TurnTokenInfo{},
		startTime: time.Now(),
	}
}

// RecordTurn logs token usage for a turn
func (tt *TokenTracker) RecordTurn(inputTokens, outputTokens int) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	turn := TurnTokenInfo{
		TurnNumber:   len(tt.turns) + 1,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  inputTokens + outputTokens,
		Timestamp:    time.Now(),
	}

	tt.turns = append(tt.turns, turn)
	tt.totalTokens += turn.TotalTokens
}

// RecordCompaction marks that compaction occurred this turn
func (tt *TokenTracker) RecordCompaction() {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	if len(tt.turns) > 0 {
		tt.turns[len(tt.turns)-1].CompactionEvent = true
	}
}

// AverageTurnSize returns average tokens per turn
func (tt *TokenTracker) AverageTurnSize() int {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if len(tt.turns) == 0 {
		return 0
	}

	return tt.totalTokens / len(tt.turns)
}

// EstimateRemainingTurns estimates how many more turns fit in context
func (tt *TokenTracker) EstimateRemainingTurns(window, reserved int) int {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	available := window - reserved - tt.totalTokens
	avgTurnSize := tt.AverageTurnSize()

	if avgTurnSize == 0 {
		return 0
	}

	return available / avgTurnSize
}

// GetTotalTokens returns the total tokens used across all turns
func (tt *TokenTracker) GetTotalTokens() int {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	return tt.totalTokens
}

// GetTurnCount returns the number of turns recorded
func (tt *TokenTracker) GetTurnCount() int {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	return len(tt.turns)
}
