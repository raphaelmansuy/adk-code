package compaction

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// LLMSummarizer generates summaries of conversation history using an LLM
type LLMSummarizer struct {
	llm    model.LLM
	config *Config
}

// NewLLMSummarizer creates a new LLM summarizer
func NewLLMSummarizer(llm model.LLM, config *Config) *LLMSummarizer {
	return &LLMSummarizer{
		llm:    llm,
		config: config,
	}
}

// Summarize generates a summary of the provided events
func (ls *LLMSummarizer) Summarize(
	ctx context.Context,
	events []*session.Event,
) (*session.Event, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("cannot summarize empty event list")
	}

	// Format events for prompt
	conversationText := ls.formatEvents(events)
	prompt := fmt.Sprintf(ls.config.PromptTemplate, conversationText)

	// Create LLM request
	llmRequest := &model.LLMRequest{
		Model: ls.llm.Name(),
		Contents: []*genai.Content{
			{
				Role: "user",
				Parts: []*genai.Part{
					{Text: prompt},
				},
			},
		},
		Config: &genai.GenerateContentConfig{},
	}

	// Generate content using the agent's LLM
	var summaryContent *genai.Content
	var usageMetadata *genai.GenerateContentResponseUsageMetadata

	responseStream := ls.llm.GenerateContent(ctx, llmRequest, false)
	for resp := range responseStream {
		if resp == nil {
			continue
		}
		if resp.Content != nil {
			summaryContent = resp.Content
			usageMetadata = resp.UsageMetadata
			break
		}
	}

	if summaryContent == nil {
		return nil, fmt.Errorf("no summary content generated")
	}

	// Ensure role is 'model' (following ADK Python)
	summaryContent.Role = "model"

	// Calculate metrics
	originalTokens := ls.countTokens(events)
	compactedTokens := 0
	if usageMetadata != nil {
		compactedTokens = int(usageMetadata.TotalTokenCount)
	}

	// Serialize summary content to JSON
	summaryJSON, err := json.Marshal(summaryContent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal summary content: %w", err)
	}

	// Create compaction metadata
	startTime := events[0].Timestamp
	endTime := events[len(events)-1].Timestamp
	if endTime.IsZero() {
		endTime = time.Now()
	}

	metadata := &CompactionMetadata{
		StartTimestamp:       startTime,
		EndTimestamp:         endTime,
		StartInvocationID:    events[0].InvocationID,
		EndInvocationID:      events[len(events)-1].InvocationID,
		CompactedContentJSON: string(summaryJSON),
		EventCount:           len(events),
		OriginalTokens:       originalTokens,
		CompactedTokens:      compactedTokens,
	}

	// Calculate compression ratio safely
	if compactedTokens > 0 {
		metadata.CompressionRatio = float64(originalTokens) / float64(compactedTokens)
	}

	// Create compaction event (following ADK Python pattern)
	compactionEvent := &session.Event{
		ID:           uuid.NewString(),
		InvocationID: uuid.NewString(),
		Author:       "user",
		Timestamp:    time.Now(),
		LLMResponse: model.LLMResponse{
			Content: summaryContent,
		},
	}

	// Store compaction metadata in CustomMetadata
	if err := SetCompactionMetadata(compactionEvent, metadata); err != nil {
		return nil, fmt.Errorf("failed to set compaction metadata: %w", err)
	}

	return compactionEvent, nil
}

// formatEvents formats events for the summarization prompt
func (ls *LLMSummarizer) formatEvents(events []*session.Event) string {
	var sb strings.Builder

	for _, event := range events {
		if event == nil {
			continue
		}

		// Skip compaction events in the summary text
		if IsCompactionEvent(event) {
			sb.WriteString(fmt.Sprintf("[COMPACTED SUMMARY from %s]\n", event.Timestamp.Format(time.RFC3339)))
			if metadata, err := GetCompactionMetadata(event); err == nil {
				sb.WriteString(fmt.Sprintf("Events: %d, Tokens: %d->%d\n", metadata.EventCount, metadata.OriginalTokens, metadata.CompactedTokens))
			}
			continue
		}

		// Format the author and content
		if event.LLMResponse.Content != nil && len(event.LLMResponse.Content.Parts) > 0 {
			for _, part := range event.LLMResponse.Content.Parts {
				if part != nil && part.Text != "" {
					sb.WriteString(fmt.Sprintf("%s: %s\n", event.Author, part.Text))
				}
			}
		}
	}

	return sb.String()
}

// countTokens estimates token count for events
// This is a simple estimation - actual token count depends on the model
func (ls *LLMSummarizer) countTokens(events []*session.Event) int {
	totalTokens := 0

	for _, event := range events {
		if event == nil {
			continue
		}

		// Rough estimation: ~4 characters per token for English text
		if event.LLMResponse.Content != nil {
			for _, part := range event.LLMResponse.Content.Parts {
				if part != nil && part.Text != "" {
					totalTokens += len(part.Text) / 4
				}
			}
		}

		// Add tokens from usage metadata if available
		if event.LLMResponse.UsageMetadata != nil {
			if event.LLMResponse.UsageMetadata.PromptTokenCount > 0 {
				totalTokens += int(event.LLMResponse.UsageMetadata.PromptTokenCount)
			}
			if event.LLMResponse.UsageMetadata.CandidatesTokenCount > 0 {
				totalTokens += int(event.LLMResponse.UsageMetadata.CandidatesTokenCount)
			}
		}
	}

	return totalTokens
}
