package tracking

import (
	"testing"
	"time"

	"google.golang.org/genai"
)

func TestSessionTokensRecordMetrics(t *testing.T) {
	st := NewSessionTokens()

	metadata := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:        100,
		CachedContentTokenCount: 20,
		CandidatesTokenCount:    50,
		ThoughtsTokenCount:      10,
		ToolUsePromptTokenCount: 5,
		TotalTokenCount:         185,
	}

	st.RecordMetrics(metadata, "req_1")

	if st.TotalTokens != 185 {
		t.Errorf("Expected TotalTokens=185, got %d", st.TotalTokens)
	}

	if st.TotalPromptTokens != 100 {
		t.Errorf("Expected TotalPromptTokens=100, got %d", st.TotalPromptTokens)
	}

	if st.RequestCount != 1 {
		t.Errorf("Expected RequestCount=1, got %d", st.RequestCount)
	}
}
func TestSessionTokensMultipleRecords(t *testing.T) {
	st := NewSessionTokens()

	// The API returns cumulative token counts for multi-turn conversations
	// First request: cumulative total = 150 (delta from 0 = 150)
	metadata1 := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     100,
		CandidatesTokenCount: 50,
		TotalTokenCount:      150, // Cumulative: 0 + 150 = 150
	}

	// Second request: cumulative total = 275 (delta from 150 = 125)
	// New prompt tokens: 200, New response tokens: 75
	// Prompt delta: 200 - 100 = 100, Response delta: 75 - 50 = 25, Total delta: 100 + 25 = 125
	metadata2 := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     200,
		CandidatesTokenCount: 75,
		TotalTokenCount:      275, // Cumulative: 150 + 125 = 275
	}

	st.RecordMetrics(metadata1, "req_1")
	st.RecordMetrics(metadata2, "req_2")

	// TotalTokens is the sum of per-request deltas: 150 + 125 = 275
	if st.TotalTokens != 275 {
		t.Errorf("Expected TotalTokens=275 (150 + 125), got %d", st.TotalTokens)
	}

	// TotalPromptTokens is sum of prompt deltas: 100 + 100 = 200
	if st.TotalPromptTokens != 200 {
		t.Errorf("Expected TotalPromptTokens=200 (100 + 100), got %d", st.TotalPromptTokens)
	}

	// TotalResponseTokens is sum of response deltas: 50 + 25 = 75
	if st.TotalResponseTokens != 75 {
		t.Errorf("Expected TotalResponseTokens=75 (50 + 25), got %d", st.TotalResponseTokens)
	}

	if st.RequestCount != 2 {
		t.Errorf("Expected RequestCount=2, got %d", st.RequestCount)
	}
}
func TestSessionTokensGetSummary(t *testing.T) {
	st := NewSessionTokens()

	metadata := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     100,
		CandidatesTokenCount: 50,
		TotalTokenCount:      150,
	}

	st.RecordMetrics(metadata, "req_1")

	summary := st.GetSummary()

	if summary.TotalTokens != 150 {
		t.Errorf("Expected TotalTokens=150, got %d", summary.TotalTokens)
	}

	if summary.RequestCount != 1 {
		t.Errorf("Expected RequestCount=1, got %d", summary.RequestCount)
	}

	if summary.AvgTokensPerRequest != 150.0 {
		t.Errorf("Expected AvgTokensPerRequest=150.0, got %f", summary.AvgTokensPerRequest)
	}
}
func TestGlobalTrackerGetOrCreateSession(t *testing.T) {
	gt := NewGlobalTracker()

	st1 := gt.GetOrCreateSession("session1")
	st2 := gt.GetOrCreateSession("session1")

	if st1 != st2 {
		t.Errorf("GetOrCreateSession should return same instance for same session ID")
	}

	st3 := gt.GetOrCreateSession("session2")

	if st1 == st3 {
		t.Errorf("GetOrCreateSession should return different instances for different session IDs")
	}
}
func TestGlobalTrackerGetGlobalSummary(t *testing.T) {
	gt := NewGlobalTracker()

	st1 := gt.GetOrCreateSession("session1")
	st2 := gt.GetOrCreateSession("session2")

	// Session 1: First request with 100 total tokens (50 prompt + 50 response)
	metadata1 := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     50,
		CandidatesTokenCount: 50,
		TotalTokenCount:      100,
	}

	// Session 2: First request with 200 total tokens (100 prompt + 100 response)
	metadata2 := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     100,
		CandidatesTokenCount: 100,
		TotalTokenCount:      200,
	}

	st1.RecordMetrics(metadata1, "req_1")
	st2.RecordMetrics(metadata2, "req_2")

	globalSummary := gt.GetGlobalSummary()

	if globalSummary.TotalTokens != 300 {
		t.Errorf("Expected TotalTokens=300 (100 + 200), got %d", globalSummary.TotalTokens)
	}

	if globalSummary.TotalRequests != 2 {
		t.Errorf("Expected TotalRequests=2, got %d", globalSummary.TotalRequests)
	}
}
func TestFormatTokenMetrics(t *testing.T) {
	metric := TokenMetrics{
		PromptTokens:   100,
		ResponseTokens: 50,
		TotalTokens:    150,
	}

	formatted := FormatTokenMetrics(metric)

	if formatted == "" {
		t.Errorf("FormatTokenMetrics should not return empty string")
	}

	if !contains(formatted, "prompt=100") {
		t.Errorf("FormatTokenMetrics should contain 'prompt=100', got: %s", formatted)
	}

	if !contains(formatted, "response=50") {
		t.Errorf("FormatTokenMetrics should contain 'response=50', got: %s", formatted)
	}

	if !contains(formatted, "total=150") {
		t.Errorf("FormatTokenMetrics should contain 'total=150', got: %s", formatted)
	}
}

func TestFormatTokenMetrics_WithThinkingTokens(t *testing.T) {
	metric := TokenMetrics{
		PromptTokens:   100,
		ResponseTokens: 50,
		ThoughtTokens:  20,
		CachedTokens:   10,
		TotalTokens:    160,
	}

	formatted := FormatTokenMetrics(metric)

	if formatted == "" {
		t.Errorf("FormatTokenMetrics should not return empty string")
	}

	if !contains(formatted, "prompt=100") {
		t.Errorf("FormatTokenMetrics should contain 'prompt=100', got: %s", formatted)
	}

	if !contains(formatted, "response=50") {
		t.Errorf("FormatTokenMetrics should contain 'response=50', got: %s", formatted)
	}

	if !contains(formatted, "thoughts=20") {
		t.Errorf("FormatTokenMetrics should contain 'thoughts=20' when thinking tokens are present, got: %s", formatted)
	}

	if !contains(formatted, "cached=10") {
		t.Errorf("FormatTokenMetrics should contain 'cached=10' when cached tokens are present, got: %s", formatted)
	}

	if !contains(formatted, "total=160") {
		t.Errorf("FormatTokenMetrics should contain 'total=160', got: %s", formatted)
	}
}

func TestFormatSessionSummary(t *testing.T) {
	summary := &Summary{
		TotalTokens:         100,
		TotalPromptTokens:   60,
		TotalResponseTokens: 40,
		RequestCount:        1,
		SessionDuration:     time.Second * 5,
	}

	formatted := FormatSessionSummary(summary)

	if formatted == "" {
		t.Errorf("FormatSessionSummary should not return empty string")
	}

	if !contains(formatted, "Token Usage Summary") {
		t.Errorf("FormatSessionSummary should contain 'Token Usage Summary'")
	}
}

func TestFormatSessionSummary_WithThinkingTokens(t *testing.T) {
	summary := &Summary{
		TotalTokens:         200,
		TotalPromptTokens:   100,
		TotalResponseTokens: 60,
		TotalThoughtTokens:  30,
		TotalCachedTokens:   10,
		RequestCount:        2,
		SessionDuration:     time.Second * 10,
	}

	formatted := FormatSessionSummary(summary)

	if formatted == "" {
		t.Errorf("FormatSessionSummary should not return empty string")
	}

	if !contains(formatted, "Token Usage Summary") {
		t.Errorf("FormatSessionSummary should contain 'Token Usage Summary'")
	}

	if !contains(formatted, "Thoughts:") {
		t.Errorf("FormatSessionSummary should contain 'Thoughts:' when thinking tokens are present, got: %s", formatted)
	}

	if !contains(formatted, "30") {
		t.Errorf("FormatSessionSummary should show thinking token count (30), got: %s", formatted)
	}
}

func TestFormatGlobalSummary(t *testing.T) {
	summary := &GlobalSummary{
		Sessions:            make(map[string]*Summary),
		TotalTokens:         300,
		TotalRequests:       2,
		AvgTokensPerRequest: 150.0,
		StartTime:           time.Now(),
	}

	summary.Sessions["session1"] = &Summary{
		TotalTokens:  150,
		RequestCount: 1,
	}
	summary.Sessions["session2"] = &Summary{
		TotalTokens:  150,
		RequestCount: 1,
	}

	formatted := FormatGlobalSummary(summary)

	if formatted == "" {
		t.Errorf("FormatGlobalSummary should not return empty string")
	}

	if !contains(formatted, "Global Token Usage Report") {
		t.Errorf("FormatGlobalSummary should contain 'Global Token Usage Report'")
	}
}

func TestNilMetadata(t *testing.T) {
	st := NewSessionTokens()
	st.RecordMetrics(nil, "req_1")

	if st.RequestCount != 0 {
		t.Errorf("RecordMetrics should handle nil metadata gracefully")
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
