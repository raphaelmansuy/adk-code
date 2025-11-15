// Package context provides context window management and token tracking
package context

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrCompactionNeeded = errors.New("context compaction needed")
	ErrContextOverflow  = errors.New("context window overflow")
)

// ResponseItem represents one turn item (message, tool call, output, etc)
type ResponseItem struct {
	ID        string    // Unique identifier
	Type      ItemType  // message, tool_call, tool_output, etc
	Role      string    // user, assistant, system
	Content   string    // Item content
	Tokens    int       // Estimated tokens for this item
	Timestamp time.Time // When this item was added
}

// ItemType represents the type of response item
type ItemType string

const (
	ItemMessage       ItemType = "message"
	ItemToolCall      ItemType = "tool_call"
	ItemToolOutput    ItemType = "tool_output"
	ItemReasoning     ItemType = "reasoning"
	ItemGhostSnapshot ItemType = "ghost_snapshot"
)

// TokenBudget tracks and enforces token limits
type TokenBudget struct {
	ContextWindow    int     // Model's total context window
	Reserved         int     // Tokens reserved for output (10% typically)
	UsedTokens       int     // Tokens used so far in this turn
	PreviousTotal    int     // Total tokens from all previous turns
	MaxItemBytes     int     // Max bytes per truncated output
	CompactThreshold float64 // Compact at 70% of window
}

// ContextConfig defines model-specific settings
type ContextConfig struct {
	ModelName           string
	ContextWindow       int
	OutputTruncateBytes int     // Default: 10 KiB
	OutputTruncateLines int     // Default: 256
	TruncateHeadLines   int     // Default: 128
	TruncateTailLines   int     // Default: 128
	CompactThreshold    float64 // Default: 0.70 (70%)
}

// TokenInfo summarizes current token state
type TokenInfo struct {
	UsedTokens       int
	AvailableTokens  int
	PercentageUsed   float64
	CompactThreshold float64
	TotalTurns       int
	EstimatedOutput  int // Estimated tokens for next output
}

// TurnTokenInfo tracks tokens for a single turn
type TurnTokenInfo struct {
	TurnNumber      int
	InputTokens     int
	OutputTokens    int
	TotalTokens     int
	Timestamp       time.Time
	CompactionEvent bool // True if compaction occurred this turn
}
