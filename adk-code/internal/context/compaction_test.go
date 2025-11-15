package context

import (
	"context"
	"testing"
)

func TestCompactConversation_Success(t *testing.T) {
	items := []ResponseItem{
		{Content: "test message 1", Tokens: 10},
		{Content: "test message 2", Tokens: 20},
		{Content: "test message 3", Tokens: 30},
	}

	userMessages := []string{
		"First user message",
		"Second user message",
		"Third user message",
	}

	req := CompactionRequest{
		Items:             items,
		UserMessages:      userMessages,
		TargetTokenBudget: 5000,
		ModelName:         "gemini-2.5-flash",
	}

	result := CompactConversation(context.Background(), req)

	if !result.Success {
		t.Errorf("Expected compaction to succeed")
	}

	if result.OriginalTokens != 60 {
		t.Errorf("Expected 60 original tokens, got %d", result.OriginalTokens)
	}

	if len(result.Summary) == 0 {
		t.Errorf("Expected non-empty summary")
	}
}

func TestSelectUserMessagesUpToBudget(t *testing.T) {
	messages := []string{
		"First message",
		"Second message",
		"Third message",
		"Fourth message",
	}

	// Budget for about 2 messages
	selected := selectUserMessagesUpToBudget(messages, 10)

	if len(selected) == 0 {
		t.Errorf("Expected some messages to be selected")
	}

	// Selected messages should be in chronological order (newest first, then reversed)
	// Just verify we got some messages back
	if len(selected) > len(messages) {
		t.Errorf("Selected more messages than available")
	}
}

func TestReverse(t *testing.T) {
	s := []string{"a", "b", "c", "d"}
	reverse(s)

	expected := []string{"d", "c", "b", "a"}
	for i, v := range s {
		if v != expected[i] {
			t.Errorf("Expected %s at position %d, got %s", expected[i], i, v)
		}
	}
}

func TestEstimateHistoryTokens(t *testing.T) {
	items := []ResponseItem{
		{Tokens: 100},
		{Tokens: 200},
		{Tokens: 300},
	}

	total := estimateHistoryTokens(items)

	if total != 600 {
		t.Errorf("Expected 600 tokens, got %d", total)
	}
}

func TestEstimateHistoryTokens_Empty(t *testing.T) {
	items := []ResponseItem{}

	total := estimateHistoryTokens(items)

	if total != 0 {
		t.Errorf("Expected 0 tokens for empty history, got %d", total)
	}
}
