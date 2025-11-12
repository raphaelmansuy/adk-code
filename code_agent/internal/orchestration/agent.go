package orchestration

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"

	codingagent "code_agent/agent"
	"code_agent/internal/config"
)

// InitializeAgentComponent creates the coding agent
func InitializeAgentComponent(ctx context.Context, cfg *config.Config, llm model.LLM) (agent.Agent, error) {
	ag, err := codingagent.NewCodingAgent(ctx, codingagent.Config{
		Model:            llm,
		WorkingDirectory: cfg.WorkingDirectory,
		EnableThinking:   cfg.EnableThinking,
		ThinkingBudget:   cfg.ThinkingBudget,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create coding agent: %w", err)
	}

	return ag, nil
}
