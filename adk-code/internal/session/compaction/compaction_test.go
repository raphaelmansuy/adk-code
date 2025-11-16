package compaction

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/adk/session"
)

// TestCompactionMetadataStorage tests storing and retrieving compaction metadata
func TestCompactionMetadataStorage(t *testing.T) {
	event := &session.Event{
		ID:           uuid.NewString(),
		InvocationID: uuid.NewString(),
		Author:       "test",
		Timestamp:    time.Now(),
	}

	// Create compaction metadata
	metadata := &CompactionMetadata{
		StartTimestamp:       time.Now().Add(-time.Hour),
		EndTimestamp:         time.Now(),
		StartInvocationID:    "inv-1",
		EndInvocationID:      "inv-5",
		EventCount:           5,
		OriginalTokens:       1000,
		CompactedTokens:      500,
		CompressionRatio:     2.0,
		CompactedContentJSON: `{"role":"model","parts":[{"text":"summary"}]}`,
	}

	// Set metadata
	err := SetCompactionMetadata(event, metadata)
	if err != nil {
		t.Fatalf("SetCompactionMetadata failed: %v", err)
	}

	// Verify event is marked as compaction event
	if !IsCompactionEvent(event) {
		t.Error("Event should be marked as compaction event")
	}

	// Get metadata back
	retrievedMetadata, err := GetCompactionMetadata(event)
	if err != nil {
		t.Fatalf("GetCompactionMetadata failed: %v", err)
	}

	// Verify metadata
	if retrievedMetadata.EventCount != 5 {
		t.Errorf("Expected EventCount 5, got %d", retrievedMetadata.EventCount)
	}
	if retrievedMetadata.OriginalTokens != 1000 {
		t.Errorf("Expected OriginalTokens 1000, got %d", retrievedMetadata.OriginalTokens)
	}
	if retrievedMetadata.CompressionRatio != 2.0 {
		t.Errorf("Expected CompressionRatio 2.0, got %f", retrievedMetadata.CompressionRatio)
	}
}

// TestIsCompactionEvent tests event detection
func TestIsCompactionEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    *session.Event
		expected bool
	}{
		{
			name:     "nil event",
			event:    nil,
			expected: false,
		},
		{
			name:     "event without metadata",
			event:    &session.Event{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCompactionEvent(tt.event)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCompactionMetadataSerialization tests JSON serialization of metadata
func TestCompactionMetadataSerialization(t *testing.T) {
	metadata := &CompactionMetadata{
		StartTimestamp:       time.Now(),
		EndTimestamp:         time.Now(),
		StartInvocationID:    "inv-1",
		EndInvocationID:      "inv-5",
		EventCount:           10,
		OriginalTokens:       5000,
		CompactedTokens:      1500,
		CompressionRatio:     3.33,
		CompactedContentJSON: `{"role":"model","parts":[{"text":"summary"}]}`,
	}

	// Serialize to JSON
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	// Deserialize back
	var unmarshaled CompactionMetadata
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	// Verify values match
	if unmarshaled.EventCount != metadata.EventCount {
		t.Errorf("EventCount mismatch: %d != %d", unmarshaled.EventCount, metadata.EventCount)
	}
	if unmarshaled.CompressionRatio != metadata.CompressionRatio {
		t.Errorf("CompressionRatio mismatch: %f != %f", unmarshaled.CompressionRatio, metadata.CompressionRatio)
	}
}

// TestConfigDefaults tests that DefaultConfig returns valid configuration
func TestConfigDefaults(t *testing.T) {
	config := DefaultConfig()

	if config.InvocationThreshold == 0 {
		t.Error("InvocationThreshold should not be 0")
	}
	if config.OverlapSize == 0 {
		t.Error("OverlapSize should not be 0")
	}
	if config.TokenThreshold == 0 {
		t.Error("TokenThreshold should not be 0")
	}
	if config.SafetyRatio == 0 || config.SafetyRatio > 1.0 {
		t.Error("SafetyRatio should be between 0 and 1")
	}
	if config.PromptTemplate == "" {
		t.Error("PromptTemplate should not be empty")
	}
}
