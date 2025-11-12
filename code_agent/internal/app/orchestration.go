package app

import (
	"context"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"

	"code_agent/display"
	"code_agent/internal/config"
	"code_agent/internal/orchestration"
)

// initializeDisplayComponents is a facade for backward compatibility
func initializeDisplayComponents(cfg *config.Config) (*DisplayComponents, error) {
	return orchestration.InitializeDisplayComponents(cfg)
}

// initializeModelComponents is a facade for backward compatibility
func initializeModelComponents(ctx context.Context, cfg *config.Config) (*ModelComponents, error) {
	return orchestration.InitializeModelComponents(ctx, cfg)
}

// initializeSessionComponents is a facade for backward compatibility
func initializeSessionComponents(ctx context.Context, cfg *config.Config, ag agent.Agent, bannerRenderer *display.BannerRenderer) (*SessionComponents, error) {
	return orchestration.InitializeSessionComponents(ctx, cfg, ag, bannerRenderer)
}

// initializeAgentComponent is a facade for backward compatibility
func initializeAgentComponent(ctx context.Context, cfg *config.Config, llm model.LLM) (agent.Agent, error) {
	return orchestration.InitializeAgentComponent(ctx, cfg, llm)
}
