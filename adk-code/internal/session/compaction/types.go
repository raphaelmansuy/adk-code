// Package compaction provides session history compaction via sliding window summarization
package compaction

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/adk/session"
)

// CompactionMetadata is stored in event.CustomMetadata["_adk_compaction"]
type CompactionMetadata struct {
	StartTimestamp       time.Time `json:"start_timestamp"`
	EndTimestamp         time.Time `json:"end_timestamp"`
	StartInvocationID    string    `json:"start_invocation_id,omitempty"`
	EndInvocationID      string    `json:"end_invocation_id,omitempty"`
	CompactedContentJSON string    `json:"compacted_content_json"`
	EventCount           int       `json:"event_count"`
	OriginalTokens       int       `json:"original_tokens"`
	CompactedTokens      int       `json:"compacted_tokens"`
	CompressionRatio     float64   `json:"compression_ratio"`
}

const CompactionMetadataKey = "_adk_compaction"

// IsCompactionEvent checks if an event contains compaction metadata
func IsCompactionEvent(event *session.Event) bool {
	if event == nil || event.CustomMetadata == nil {
		return false
	}
	_, exists := event.CustomMetadata[CompactionMetadataKey]
	return exists
}

// GetCompactionMetadata extracts compaction data from event
func GetCompactionMetadata(event *session.Event) (*CompactionMetadata, error) {
	if !IsCompactionEvent(event) {
		return nil, fmt.Errorf("event is not a compaction event")
	}

	data := event.CustomMetadata[CompactionMetadataKey]

	// Marshal to JSON and unmarshal to struct
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var metadata CompactionMetadata
	if err := json.Unmarshal(jsonData, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// SetCompactionMetadata sets compaction data on an event
func SetCompactionMetadata(event *session.Event, metadata *CompactionMetadata) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	if event.CustomMetadata == nil {
		event.CustomMetadata = make(map[string]any)
	}

	// Convert to map for storage
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	var dataMap map[string]any
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return err
	}

	event.CustomMetadata[CompactionMetadataKey] = dataMap
	return nil
}
