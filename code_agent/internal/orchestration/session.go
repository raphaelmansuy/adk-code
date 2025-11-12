package orchestration

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"

	"code_agent/display"
	"code_agent/internal/config"
	"code_agent/session"
	"code_agent/tracking"
)

// sessionInitializer handles session and runner setup
type sessionInitializer struct {
	manager *session.SessionManager
	runner  *runner.Runner
	tokens  *tracking.SessionTokens
}

// InitializeSessionComponents sets up session management
func InitializeSessionComponents(ctx context.Context, cfg *config.Config, ag agent.Agent, bannerRenderer *display.BannerRenderer) (*SessionComponents, error) {
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
		Manager: initializer.manager,
		Runner:  initializer.runner,
		Tokens:  initializer.tokens,
	}, nil
}
