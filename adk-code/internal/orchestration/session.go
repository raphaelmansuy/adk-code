package orchestration

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"

	"adk-code/internal/config"
	"adk-code/internal/display"
	"adk-code/internal/session"
	"adk-code/internal/session/compaction"
	"adk-code/internal/tracking"
)

// sessionInitializer handles session and runner setup
type sessionInitializer struct {
	manager *session.SessionManager
	runner  *runner.Runner
	tokens  *tracking.SessionTokens
}

// InitializeSessionComponents sets up session management
func InitializeSessionComponents(ctx context.Context, cfg *config.Config, ag agent.Agent, bannerRenderer *display.BannerRenderer, agentLLM model.LLM) (*SessionComponents, error) {
	initializer := &sessionInitializer{}

	var err error
	initializer.manager, err = session.NewSessionManager("code_agent", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create session manager: %w", err)
	}

	// Generate unique session name if not specified
	if cfg.SessionName == "" {
		cfg.SessionName = GenerateUniqueSessionName()
	}

	// Initialize the session in the database
	sessionInit := NewSessionInitializer(initializer.manager, bannerRenderer)
	if err := sessionInit.InitializeSession(ctx, "user1", cfg.SessionName); err != nil {
		return nil, err
	}

	// Create agent runner
	sessionService := initializer.manager.GetService()

	// Set up compaction configuration and coordinator if enabled
	var compactionConfig *compaction.Config
	var coordinator *compaction.Coordinator

	// Wrap with compaction if enabled
	if cfg.CompactionEnabled {
		compactionConfig = &compaction.Config{
			InvocationThreshold: cfg.CompactionThreshold,
			OverlapSize:         cfg.CompactionOverlap,
			TokenThreshold:      cfg.CompactionTokens,
			SafetyRatio:         cfg.CompactionSafety,
			PromptTemplate:      compaction.DefaultConfig().PromptTemplate,
		}
		sessionService = compaction.NewCompactionService(sessionService, compactionConfig)

		// Create the compaction coordinator
		selector := compaction.NewSelector(compactionConfig)
		coordinator = compaction.NewCoordinator(
			compactionConfig,
			selector,
			agentLLM,
			sessionService,
		)
	}

	initializer.runner, err = runner.New(runner.Config{
		AppName:        "code_agent",
		Agent:          ag,
		SessionService: sessionService,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	// Initialize token tracking
	initializer.tokens = tracking.NewSessionTokens()

	return &SessionComponents{
		Manager:       initializer.manager,
		Runner:        initializer.runner,
		Tokens:        initializer.tokens,
		Coordinator:   coordinator,
		CompactionCfg: compactionConfig,
	}, nil
}
