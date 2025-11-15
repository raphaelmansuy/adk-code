package context

import (
	"context"
	"fmt"
)

const (
	// Compaction prompt template
	CompactionPromptTemplate = `Summarize this conversation concisely:

User messages:
%s

Please provide a brief 2-3 sentence summary of what the user is trying to accomplish and key context.`

	// Token budget for compaction
	CompactUserMessageMaxTokens = 20000
)

// CompactionRequest describes what needs compacting
type CompactionRequest struct {
	Items             []ResponseItem
	UserMessages      []string
	TargetTokenBudget int
	ModelName         string
}

// CompactionResult is the output of compaction
type CompactionResult struct {
	OriginalTokens   int
	CompactedTokens  int
	Summary          string
	RetainedMessages []string
	CompactionRatio  float64
	Success          bool
	Error            string
}

// CompactConversation reduces conversation size while preserving intent
func CompactConversation(
	ctx context.Context,
	req CompactionRequest,
) CompactionResult {
	result := CompactionResult{
		Success:          false,
		RetainedMessages: []string{},
	}

	// Step 1: Estimate original tokens
	result.OriginalTokens = estimateHistoryTokens(req.Items)

	// Step 2: Select user messages to retain (newest first, up to budget)
	selected := selectUserMessagesUpToBudget(
		req.UserMessages,
		CompactUserMessageMaxTokens,
	)
	result.RetainedMessages = selected

	// Step 3: Generate summary (would call LLM in real implementation)
	// For now, placeholder - actual implementation calls the model
	summary := generateSummaryFromMessages(req.UserMessages)
	result.Summary = summary

	// Step 4: Build compacted history
	// [Initial context] + [selected user messages] + [summary]
	compactedTokens := estimateHistoryTokens(req.Items) // Would recount after compaction
	result.CompactedTokens = compactedTokens
	if compactedTokens > 0 {
		result.CompactionRatio = float64(result.OriginalTokens) / float64(compactedTokens)
	}
	result.Success = true

	return result
}

// selectUserMessagesUpToBudget selects messages respecting byte budget
func selectUserMessagesUpToBudget(messages []string, maxTokens int) []string {
	maxBytes := maxTokens * 4 // Rough estimate: 1 token ≈ 4 bytes

	var selected []string
	remaining := maxBytes

	// Iterate newest → oldest
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if len(msg) <= remaining {
			selected = append(selected, msg)
			remaining -= len(msg)
		} else if remaining > 0 {
			// Truncate this message
			truncated := msg[:remaining]
			selected = append(selected, truncated)
			break
		} else {
			break
		}
	}

	// Reverse back to chronological order
	reverse(selected)
	return selected
}

func reverse(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func generateSummaryFromMessages(messages []string) string {
	// Placeholder - actual implementation would call LLM
	if len(messages) == 0 {
		return ""
	}
	return fmt.Sprintf(
		"Conversation spanning %d user messages with focus on code tasks",
		len(messages),
	)
}

func estimateHistoryTokens(items []ResponseItem) int {
	total := 0
	for _, item := range items {
		total += item.Tokens
	}
	return total
}
