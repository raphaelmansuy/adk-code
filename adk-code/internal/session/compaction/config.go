// Package compaction provides session history compaction via sliding window summarization
package compaction

const defaultPromptTemplate = `The following is a conversation history between a user and an AI agent.
Summarize the conversation concisely, focusing on:
1. Key decisions and outcomes
2. Important context and state changes
3. Unresolved questions or pending tasks
4. Tool calls and their results

Keep the summary under 500 tokens while preserving critical information.

Conversation History:
%s
`

// Config holds configuration for session history compaction
type Config struct {
	// Invocation-based triggering
	InvocationThreshold int `json:"invocation_threshold"`
	OverlapSize         int `json:"overlap_size"`

	// Token-aware triggering (adk-code enhancement)
	TokenThreshold int     `json:"token_threshold"`
	SafetyRatio    float64 `json:"safety_ratio"`

	// Prompt configuration
	PromptTemplate string `json:"prompt_template"`
}

// DefaultConfig returns the default compaction configuration
func DefaultConfig() *Config {
	return &Config{
		InvocationThreshold: 5,
		OverlapSize:         2,
		TokenThreshold:      700000,
		SafetyRatio:         0.7,
		PromptTemplate:      defaultPromptTemplate,
	}
}
