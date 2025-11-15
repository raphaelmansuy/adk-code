package context

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"
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
	Model             model.LLM // The model from the main agent
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
// Uses an ADK agent with the model inherited from the main agent
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

	// Step 3: Generate summary using ADK agent with inherited model
	summary, err := generateSummaryWithAgent(ctx, req.Model, req.UserMessages)
	if err != nil {
		result.Error = fmt.Sprintf("failed to generate summary: %v", err)
		return result
	}
	result.Summary = summary

	// Step 4: Calculate compacted tokens
	// After compaction: only retained messages + summary
	compactedTokens := 0
	for _, msg := range selected {
		compactedTokens += estimateTokens(msg)
	}
	compactedTokens += estimateTokens(summary)

	result.CompactedTokens = compactedTokens
	if result.OriginalTokens > 0 {
		result.CompactionRatio = float64(result.OriginalTokens) / float64(compactedTokens)
	}
	result.Success = true

	return result
}

// generateSummaryWithAgent uses an ADK agent to generate a conversation summary
// The agent is created with the same model as the main agent
func generateSummaryWithAgent(ctx context.Context, llm model.LLM, messages []string) (string, error) {
	if llm == nil {
		// Fallback to simple summary if no model provided (for testing)
		return generateSummaryFromMessages(messages), nil
	}

	// Create a specialized compaction agent using the inherited model
	compactionAgent, err := llmagent.New(llmagent.Config{
		Name:        "conversation_compactor",
		Model:       llm,
		Description: "A specialized agent for compacting and summarizing conversations",
		Instruction: "You are a conversation summarization expert. When given a list of user messages, provide a brief 2-3 sentence summary that captures what the user is trying to accomplish and the key context. Be concise and focus on the main goals and important details.",
		Tools:       []tool.Tool{}, // No tools needed for summarization
	})
	if err != nil {
		return "", fmt.Errorf("failed to create compaction agent: %w", err)
	}

	// Format the messages for summarization
	messagesText := strings.Join(messages, "\n\n")
	prompt := fmt.Sprintf(CompactionPromptTemplate, messagesText)

	// Use the agent with a runner to generate the summary
	// Create a simple runner for the compaction agent (without session service for simplicity)
	agentRunner, err := runner.New(runner.Config{
		AppName: "compaction",
		Agent:   compactionAgent,
		// No SessionService needed for one-off summarization
	})
	if err != nil {
		return "", fmt.Errorf("failed to create runner: %w", err)
	}

	// Create the prompt as a genai.Content
	userMsg := &genai.Content{
		Role: genai.RoleUser,
		Parts: []*genai.Part{
			{Text: prompt},
		},
	}

	// Run the compaction agent
	var summary strings.Builder
	for evt, runErr := range agentRunner.Run(ctx, "system", "compaction", userMsg, agent.RunConfig{
		StreamingMode: agent.StreamingModeNone,
	}) {
		if runErr != nil {
			return "", fmt.Errorf("compaction agent error: %w", runErr)
		}

		// Collect text from events
		if evt != nil && evt.Content != nil {
			for _, part := range evt.Content.Parts {
				if part != nil && part.Text != "" {
					summary.WriteString(part.Text)
				}
			}
		}
	}

	result := strings.TrimSpace(summary.String())
	if result == "" {
		// Fallback if agent produced no output
		return generateSummaryFromMessages(messages), nil
	}

	return result, nil
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
