package context

import (
	"testing"
	"time"

	"adk-code/pkg/models"
)

func TestContextManager_NewContextManager(t *testing.T) {
	modelConfig := models.Config{
		Name:          "gemini-2.5-flash",
		ContextWindow: 1_000_000,
	}

	cm := NewContextManager(modelConfig)

	if cm == nil {
		t.Fatal("Expected non-nil ContextManager")
	}

	if cm.tokens.ContextWindow != 1_000_000 {
		t.Errorf("Expected context window of 1M, got %d", cm.tokens.ContextWindow)
	}

	if cm.tokens.Reserved != 100_000 {
		t.Errorf("Expected reserved of 100K, got %d", cm.tokens.Reserved)
	}
}

func TestContextManager_AddItem(t *testing.T) {
	modelConfig := models.Config{
		Name:          "gemini-2.5-flash",
		ContextWindow: 1_000_000,
	}

	cm := NewContextManager(modelConfig)

	item := ResponseItem{
		ID:        "item-1",
		Type:      ItemMessage,
		Role:      "user",
		Content:   "Hello, world!",
		Timestamp: time.Now(),
	}

	err := cm.AddItem(item)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if cm.GetItemCount() != 1 {
		t.Errorf("Expected 1 item, got %d", cm.GetItemCount())
	}
}

func TestContextManager_TruncatesOutput(t *testing.T) {
	modelConfig := models.Config{
		Name:          "gemini-2.5-flash",
		ContextWindow: 1_000_000,
	}

	cm := NewContextManager(modelConfig)

	// Create large output
	largeContent := ""
	for i := 0; i < 500; i++ {
		largeContent += "This is a very long line with lots of text\n"
	}

	item := ResponseItem{
		ID:        "item-1",
		Type:      ItemToolOutput,
		Role:      "tool",
		Content:   largeContent,
		Timestamp: time.Now(),
	}

	err := cm.AddItem(item)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Get history and check that content was truncated
	history, _ := cm.GetHistory()
	if len(history) != 1 {
		t.Fatalf("Expected 1 item in history")
	}

	if len(history[0].Content) >= len(largeContent) {
		t.Errorf("Expected content to be truncated")
	}

	// Should have elision marker
	if len(history[0].Content) > 0 && len(largeContent) > 10*1024 {
		// Content should be truncated
		t.Logf("Original size: %d, Truncated size: %d", len(largeContent), len(history[0].Content))
	}
}

func TestContextManager_DetectsCompactionNeeded(t *testing.T) {
	modelConfig := models.Config{
		Name:          "gemini-2.5-flash",
		ContextWindow: 1000, // Small context for testing
	}

	cm := NewContextManager(modelConfig)

	// Fill context to >70% (need more content since 1 token = 4 chars)
	// 1000 token window, 70% = 700 tokens, so need 2800+ chars
	largeContent := ""
	for i := 0; i < 3000; i++ {
		largeContent += "x"
	}

	item := ResponseItem{
		ID:        "item-1",
		Type:      ItemMessage,
		Role:      "user",
		Content:   largeContent,
		Timestamp: time.Now(),
	}

	err := cm.AddItem(item)
	if err != ErrCompactionNeeded {
		t.Errorf("Expected ErrCompactionNeeded, got %v", err)
	}
}

func TestContextManager_TokenInfo(t *testing.T) {
	modelConfig := models.Config{
		Name:          "gemini-2.5-flash",
		ContextWindow: 1_000_000,
	}

	cm := NewContextManager(modelConfig)

	item := ResponseItem{
		ID:        "item-1",
		Type:      ItemMessage,
		Role:      "user",
		Content:   "Test message",
		Timestamp: time.Now(),
	}

	cm.AddItem(item)

	info := cm.TokenInfo()

	if info.UsedTokens == 0 {
		t.Errorf("Expected non-zero used tokens")
	}

	if info.AvailableTokens != 900_000 { // 1M - 100K reserved
		t.Errorf("Expected 900K available tokens, got %d", info.AvailableTokens)
	}
}

func TestContextManager_Clear(t *testing.T) {
	modelConfig := models.Config{
		Name:          "gemini-2.5-flash",
		ContextWindow: 1_000_000,
	}

	cm := NewContextManager(modelConfig)

	item := ResponseItem{
		ID:        "item-1",
		Type:      ItemMessage,
		Role:      "user",
		Content:   "Test",
		Timestamp: time.Now(),
	}

	cm.AddItem(item)
	cm.Clear()

	if cm.GetItemCount() != 0 {
		t.Errorf("Expected 0 items after clear, got %d", cm.GetItemCount())
	}

	info := cm.TokenInfo()
	if info.UsedTokens != 0 {
		t.Errorf("Expected 0 used tokens after clear, got %d", info.UsedTokens)
	}
}

func TestContextManager_CustomThreshold(t *testing.T) {
	modelConfig := models.Config{
		Name:          "gemini-2.5-flash",
		ContextWindow: 1_000_000,
	}

	// Test with custom threshold of 80%
	cm := NewContextManagerWithOptions(modelConfig, nil, 0.80)

	threshold := cm.GetCompactThreshold()
	if threshold != 0.80 {
		t.Errorf("Expected threshold of 0.80, got %f", threshold)
	}

	// Test SetCompactThreshold
	cm.SetCompactThreshold(0.60)
	threshold = cm.GetCompactThreshold()
	if threshold != 0.60 {
		t.Errorf("Expected threshold of 0.60 after update, got %f", threshold)
	}
}

func TestContextManager_InvalidThreshold(t *testing.T) {
	modelConfig := models.Config{
		Name:          "gemini-2.5-flash",
		ContextWindow: 1_000_000,
	}

	// Test with invalid threshold (>1.0) - should default to 0.70
	cm := NewContextManagerWithOptions(modelConfig, nil, 1.5)

	threshold := cm.GetCompactThreshold()
	if threshold != 0.70 {
		t.Errorf("Expected default threshold of 0.70 for invalid input, got %f", threshold)
	}

	// Test SetCompactThreshold with invalid value - should not update
	cm.SetCompactThreshold(1.2)
	threshold = cm.GetCompactThreshold()
	if threshold != 0.70 {
		t.Errorf("Expected threshold to remain 0.70 after invalid update, got %f", threshold)
	}
}
