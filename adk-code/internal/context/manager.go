package context

import (
	"sync"

	"adk-code/pkg/models"
	"google.golang.org/adk/model"
)

// ContextManager maintains conversation history and enforces context limits
type ContextManager struct {
	mu     sync.RWMutex
	items  []ResponseItem // Ordered conversation history
	tokens TokenBudget    // Token tracking
	config ContextConfig  // Model-specific limits
	llm    model.LLM      // LLM instance for compaction (optional)
}

// NewContextManager creates a context manager for a specific model
func NewContextManager(modelConfig models.Config) *ContextManager {
	contextWindow := modelConfig.ContextWindow
	if contextWindow == 0 {
		// Default to 1M tokens if not specified (Gemini 2.5 Flash default)
		contextWindow = 1_000_000
	}

	return &ContextManager{
		items: []ResponseItem{},
		tokens: TokenBudget{
			ContextWindow:    contextWindow,
			Reserved:         contextWindow / 10, // 10% reserved
			CompactThreshold: 0.70,
		},
		config: contextConfigFromModel(modelConfig),
		llm:    nil, // Set via SetModel if needed for compaction
	}
}

// NewContextManagerWithModel creates a context manager with an LLM for compaction
func NewContextManagerWithModel(modelConfig models.Config, llm model.LLM) *ContextManager {
	cm := NewContextManager(modelConfig)
	cm.llm = llm
	return cm
}

// SetModel sets the LLM to use for compaction
func (cm *ContextManager) SetModel(llm model.LLM) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.llm = llm
}

// GetModel returns the LLM used for compaction
func (cm *ContextManager) GetModel() model.LLM {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.llm
}

// contextConfigFromModel creates a ContextConfig from a model Config
func contextConfigFromModel(modelConfig models.Config) ContextConfig {
	contextWindow := modelConfig.ContextWindow
	if contextWindow == 0 {
		contextWindow = 1_000_000
	}

	return ContextConfig{
		ModelName:           modelConfig.Name,
		ContextWindow:       contextWindow,
		OutputTruncateBytes: 10 * 1024, // 10 KiB
		OutputTruncateLines: 256,
		TruncateHeadLines:   128,
		TruncateTailLines:   128,
		CompactThreshold:    0.70, // 70%
	}
}

// AddItem records a new conversation item
func (cm *ContextManager) AddItem(item ResponseItem) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Truncate output if needed
	if item.Type == ItemToolOutput {
		item.Content = cm.truncateOutput(item.Content)
	}

	// Estimate tokens for this item
	item.Tokens = estimateTokens(item.Content)

	cm.items = append(cm.items, item)
	cm.tokens.UsedTokens += item.Tokens

	// Check if compaction is needed
	if cm.needsCompaction() {
		return ErrCompactionNeeded
	}

	return nil
}

// GetHistory returns conversation history prepared for model
func (cm *ContextManager) GetHistory() ([]ResponseItem, TokenInfo) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Normalize: ensure call/output pairs are consistent
	normalized := cm.normalizeHistory(cm.items)

	tokenInfo := TokenInfo{
		UsedTokens:       cm.tokens.UsedTokens,
		AvailableTokens:  cm.tokens.ContextWindow - cm.tokens.Reserved,
		PercentageUsed:   float64(cm.tokens.UsedTokens) / float64(cm.tokens.ContextWindow),
		CompactThreshold: cm.tokens.CompactThreshold,
	}

	return normalized, tokenInfo
}

// TokenInfo returns current token usage information
func (cm *ContextManager) TokenInfo() TokenInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return TokenInfo{
		UsedTokens:       cm.tokens.UsedTokens,
		AvailableTokens:  cm.tokens.ContextWindow - cm.tokens.Reserved,
		PercentageUsed:   float64(cm.tokens.UsedTokens) / float64(cm.tokens.ContextWindow),
		CompactThreshold: cm.tokens.CompactThreshold,
	}
}

// needsCompaction returns true if conversation should be compacted
func (cm *ContextManager) needsCompaction() bool {
	percentUsed := float64(cm.tokens.UsedTokens) / float64(cm.tokens.ContextWindow)
	return percentUsed > cm.tokens.CompactThreshold
}

// truncateOutput applies head+tail truncation to output
func (cm *ContextManager) truncateOutput(content string) string {
	if len(content) <= cm.config.OutputTruncateBytes {
		return content
	}

	return truncateHeadTail(
		content,
		cm.config.OutputTruncateLines,
		cm.config.TruncateHeadLines,
		cm.config.TruncateTailLines,
		cm.config.OutputTruncateBytes,
	)
}

// normalizeHistory ensures history invariants
func (cm *ContextManager) normalizeHistory(items []ResponseItem) []ResponseItem {
	// For now, just return items as-is
	// In a full implementation, this would ensure:
	// - Every tool call has corresponding output
	// - Every output has corresponding call
	// - No orphaned items
	return items
}

// estimateTokens estimates the number of tokens in text
// Using a simple heuristic: ~1 token per 4 characters
func estimateTokens(text string) int {
	// Simple heuristic: 1 token ≈ 4 characters
	return len(text) / 4
}

// Clear removes all items from the context manager
func (cm *ContextManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.items = []ResponseItem{}
	cm.tokens.UsedTokens = 0
}

// GetItemCount returns the number of items in the conversation history
func (cm *ContextManager) GetItemCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.items)
}
