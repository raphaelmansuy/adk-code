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
	"code_agent/internal/orchestration"
	"code_agent/internal/repl"
	"code_agent/internal/runtime"
)

const AppVersion = "1.0.0"

// Application manages the entire code agent application lifecycle
type Application struct {
	config        *config.Config
	ctx           context.Context
	signalHandler *runtime.SignalHandler
	display       *DisplayComponents
	model         *ModelComponents
	agent         agent.Agent
	mcp           *MCPComponents
	session       *SessionComponents
	repl          *repl.REPL
}

// New creates a new Application instance using the builder pattern
func New(ctx context.Context, cfg *config.Config) (*Application, error) {
	app := &Application{
		config: cfg,
	}

	// Setup signal handling
	app.signalHandler = runtime.NewSignalHandler(ctx)
	app.ctx = app.signalHandler.Context()

	// Resolve working directory early (needed for banner)
	cfg.WorkingDirectory = app.resolveWorkingDirectory()

	// Use builder pattern to orchestrate all components
	components, err := orchestration.NewOrchestrator(app.ctx, cfg).
		WithDisplay().
		WithModel().
		WithAgent().
		WithSession().
		Build()

	if err != nil {
		return nil, err
	}

	// Assign built components to application
	app.display = components.Display
	app.model = components.Model
	app.agent = components.Agent
	app.mcp = components.MCP
	app.session = components.Session

	// Print welcome banner
	displayName := app.model.Selected.DisplayName
	banner := app.display.BannerRenderer.RenderStartBanner(AppVersion, displayName, cfg.WorkingDirectory)
	fmt.Print(banner)

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
	a.repl, err = repl.New(repl.Config{
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
		MCPComponents:    a.mcp,
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
