package context

import (
	"testing"
)

func TestTokenTracker_RecordTurn(t *testing.T) {
	tt := NewTokenTracker("session-1", "gemini-2.5-flash", 1_000_000)

	tt.RecordTurn(100, 50)
	tt.RecordTurn(200, 75)

	if tt.GetTurnCount() != 2 {
		t.Errorf("Expected 2 turns, got %d", tt.GetTurnCount())
	}

	if tt.GetTotalTokens() != 425 { // 100+50+200+75
		t.Errorf("Expected 425 total tokens, got %d", tt.GetTotalTokens())
	}
}

func TestTokenTracker_AverageTurnSize(t *testing.T) {
	tt := NewTokenTracker("session-1", "gemini-2.5-flash", 1_000_000)

	tt.RecordTurn(100, 100) // 200 tokens
	tt.RecordTurn(150, 150) // 300 tokens
	tt.RecordTurn(200, 100) // 300 tokens

	avg := tt.AverageTurnSize()
	expected := (200 + 300 + 300) / 3 // 266

	if avg != expected {
		t.Errorf("Expected average of %d, got %d", expected, avg)
	}
}

func TestTokenTracker_EstimateRemainingTurns(t *testing.T) {
	tt := NewTokenTracker("session-1", "gemini-2.5-flash", 1_000_000)

	// Record 3 turns with avg 300 tokens each
	tt.RecordTurn(150, 150)
	tt.RecordTurn(150, 150)
	tt.RecordTurn(150, 150)

	remaining := tt.EstimateRemainingTurns(1_000_000, 100_000)

	// (1,000,000 - 100,000 - 900) / 300 = 2,997
	if remaining < 2900 || remaining > 3000 {
		t.Errorf("Expected around 2997 remaining turns, got %d", remaining)
	}
}

func TestTokenTracker_RecordCompaction(t *testing.T) {
	tt := NewTokenTracker("session-1", "gemini-2.5-flash", 1_000_000)

	tt.RecordTurn(100, 100)
	tt.RecordCompaction()

	// Verify compaction was recorded (would need to expose turns or add getter)
	// For now, just verify it doesn't panic
	if tt.GetTurnCount() != 1 {
		t.Errorf("Expected 1 turn after compaction")
	}
}

func TestTokenTracker_EmptyTracker(t *testing.T) {
	tt := NewTokenTracker("session-1", "gemini-2.5-flash", 1_000_000)

	if tt.GetTurnCount() != 0 {
		t.Errorf("Expected 0 turns in empty tracker")
	}

	if tt.GetTotalTokens() != 0 {
		t.Errorf("Expected 0 tokens in empty tracker")
	}

	if tt.AverageTurnSize() != 0 {
		t.Errorf("Expected 0 average for empty tracker")
	}
}
