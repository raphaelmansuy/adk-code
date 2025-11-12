package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/adk/agent"

	"code_agent/internal/config"
)

const AppVersion = "1.0.0"

// Application manages the entire code agent application lifecycle
type Application struct {
	config        *config.Config
	ctx           context.Context
	signalHandler *SignalHandler
	display       *DisplayComponents
	model         *ModelComponents
	agent         agent.Agent
	session       *SessionComponents
	repl          *REPL
}

// New creates a new Application instance
func New(ctx context.Context, cfg *config.Config) (*Application, error) {
	app := &Application{
		config: cfg,
	}

	// Setup signal handling
	app.signalHandler = NewSignalHandler(ctx)
	app.ctx = app.signalHandler.Context()

	// Initialize display components
	var err error
	app.display, err = initializeDisplayComponents(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize model components
	app.model, err = initializeModelComponents(app.ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Resolve working directory
	cfg.WorkingDirectory = app.resolveWorkingDirectory()

	// Print welcome banner
	displayName := app.model.Selected.DisplayName
	banner := app.display.BannerRenderer.RenderStartBanner(AppVersion, displayName, cfg.WorkingDirectory)
	fmt.Print(banner)

	// Initialize agent
	app.agent, err = initializeAgentComponent(app.ctx, cfg, app.model.LLM)
	if err != nil {
		return nil, err
	}

	// Initialize session components
	app.session, err = initializeSessionComponents(app.ctx, cfg, app.agent, app.display.BannerRenderer)
	if err != nil {
		return nil, err
	}

	// Initialize REPL
	if err := app.initializeREPL(); err != nil {
		return nil, err
	}

	return app, nil
}

// resolveWorkingDirectory resolves and validates the working directory
func (a *Application) resolveWorkingDirectory() string {
	workingDir := a.config.WorkingDirectory
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("failed to get current working directory: %v", err)
		}
	}

	// Expand ~ in the path
	if strings.HasPrefix(workingDir, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("failed to get home directory: %v", err)
		}
		workingDir = filepath.Join(homeDir, workingDir[1:])
	}

	return workingDir
}

// initializeREPL sets up the REPL
func (a *Application) initializeREPL() error {
	var err error
	a.repl, err = NewREPL(REPLConfig{
		UserID:           "user1",
		SessionName:      a.config.SessionName,
		Renderer:         a.display.Renderer,
		BannerRenderer:   a.display.BannerRenderer,
		StreamingDisplay: a.display.StreamDisplay,
		TypewriterPrint:  a.display.Typewriter,
		Runner:           a.session.Runner,
		SessionTokens:    a.session.Tokens,
		ModelRegistry:    a.model.Registry,
		SelectedModel:    a.model.Selected,
	})
	if err != nil {
		return fmt.Errorf("failed to create REPL: %w", err)
	}

	return nil
}

// Run starts the application
func (a *Application) Run() {
	defer a.Close()
	a.repl.Run(a.ctx)
}

// Close cleans up application resources
func (a *Application) Close() {
	if a.repl != nil {
		a.repl.Close()
	}
	if a.session != nil && a.session.Manager != nil {
		a.session.Manager.Close()
	}
	if a.signalHandler != nil {
		a.signalHandler.Cancel()
	}
}
