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

	metadata1 := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     100,
		CandidatesTokenCount: 50,
		TotalTokenCount:      150,
	}

	metadata2 := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     200,
		CandidatesTokenCount: 75,
		TotalTokenCount:      275,
	}

	st.RecordMetrics(metadata1, "req_1")
	st.RecordMetrics(metadata2, "req_2")

	if st.TotalTokens != 425 {
		t.Errorf("Expected TotalTokens=425, got %d", st.TotalTokens)
	}

	if st.TotalPromptTokens != 300 {
		t.Errorf("Expected TotalPromptTokens=300, got %d", st.TotalPromptTokens)
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

	metadata1 := &genai.GenerateContentResponseUsageMetadata{
		TotalTokenCount: 100,
	}

	metadata2 := &genai.GenerateContentResponseUsageMetadata{
		TotalTokenCount: 200,
	}

	st1.RecordMetrics(metadata1, "req_1")
	st2.RecordMetrics(metadata2, "req_2")

	globalSummary := gt.GetGlobalSummary()

	if globalSummary.TotalTokens != 300 {
		t.Errorf("Expected TotalTokens=300, got %d", globalSummary.TotalTokens)
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
