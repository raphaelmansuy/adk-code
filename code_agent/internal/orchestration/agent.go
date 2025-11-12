package orchestration

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"

	"code_agent/internal/config"
	agentprompts "code_agent/internal/prompts"
)

// InitializeAgentComponent creates the coding agent
func InitializeAgentComponent(ctx context.Context, cfg *config.Config, llm model.LLM) (agent.Agent, error) {
	ag, err := agentprompts.NewCodingAgent(ctx, agentprompts.Config{
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
