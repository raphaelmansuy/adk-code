// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"

	codingagent "code_agent/agent"
	"code_agent/display"
	"code_agent/persistence"
	"code_agent/pkg/cli"
	"code_agent/pkg/models"
	"code_agent/tracking"
)

const AppVersion = "1.0.0"

// Application manages the entire code agent application lifecycle
type Application struct {
	config         *cli.CLIConfig
	ctx            context.Context
	signalHandler  *SignalHandler
	renderer       *display.Renderer
	bannerRenderer *display.BannerRenderer
	typewriter     *display.TypewriterPrinter
	streamDisplay  *display.StreamingDisplay
	modelRegistry  *models.Registry
	selectedModel  models.Config
	llmModel       model.LLM
	codingAgent    agent.Agent
	sessionManager *persistence.SessionManager
	agentRunner    *runner.Runner
	sessionTokens  *tracking.SessionTokens
	repl           *REPL
}

// New creates a new Application instance
func New(ctx context.Context, config *cli.CLIConfig) (*Application, error) {
	app := &Application{
		config: config,
	}

	// Setup signal handling
	app.signalHandler = NewSignalHandler(ctx)
	app.ctx = app.signalHandler.Context()

	// Initialize components
	if err := app.initializeDisplay(); err != nil {
		return nil, err
	}

	if err := app.initializeModel(); err != nil {
		return nil, err
	}

	if err := app.initializeAgent(); err != nil {
		return nil, err
	}

	if err := app.initializeSession(); err != nil {
		return nil, err
	}

	if err := app.initializeREPL(); err != nil {
		return nil, err
	}

	return app, nil
}

// initializeDisplay sets up display components
func (a *Application) initializeDisplay() error {
	var err error
	a.renderer, err = display.NewRenderer(a.config.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	a.bannerRenderer = display.NewBannerRenderer(a.renderer)
	a.typewriter = display.NewTypewriterPrinter(display.DefaultTypewriterConfig())
	a.typewriter.SetEnabled(a.config.TypewriterEnabled)
	a.streamDisplay = display.NewStreamingDisplay(a.renderer, a.typewriter)

	return nil
}

// initializeModel sets up the LLM model
func (a *Application) initializeModel() error {
	a.modelRegistry = models.NewRegistry()

	// Resolve which model to use
	var err error
	if a.config.Model == "" {
		a.selectedModel = a.modelRegistry.GetDefaultModel()
	} else {
		parsedProvider, parsedModel, parseErr := cli.ParseProviderModelSyntax(a.config.Model)
		if parseErr != nil {
			return fmt.Errorf("invalid model syntax: %w\nUse format: provider/model (e.g., gemini/2.5-flash)", parseErr)
		}

		defaultProvider := a.config.Backend
		if defaultProvider == "" {
			defaultProvider = "gemini"
		}

		a.selectedModel, err = a.modelRegistry.ResolveFromProviderSyntax(parsedProvider, parsedModel, defaultProvider)
		if err != nil {
			// Print available models and return error
			fmt.Printf("❌ Error: %v\n\nAvailable models:\n", err)
			for _, providerName := range a.modelRegistry.ListProviders() {
				models := a.modelRegistry.GetProviderModels(providerName)
				fmt.Printf("\n%s:\n", strings.ToUpper(providerName[:1])+strings.ToLower(providerName[1:]))
				for _, m := range models {
					fmt.Printf("  • %s/%s\n", providerName, m.ID)
				}
			}
			return fmt.Errorf("model resolution failed")
		}
	}

	// Get API key
	apiKey := a.config.APIKey
	if apiKey == "" && a.selectedModel.Backend == "gemini" {
		return fmt.Errorf("Gemini API backend requires GOOGLE_API_KEY environment variable or --api-key flag")
	}

	// Get working directory
	a.config.WorkingDirectory = a.resolveWorkingDirectory()

	// Print welcome banner
	displayName := a.selectedModel.DisplayName
	banner := a.bannerRenderer.RenderStartBanner(AppVersion, displayName, a.config.WorkingDirectory)
	fmt.Print(banner)

	// Create LLM model
	actualModelID := models.ExtractModelIDFromGemini(a.selectedModel.ID)

	switch a.selectedModel.Backend {
	case "vertexai":
		if a.config.VertexAIProject == "" {
			return fmt.Errorf("Vertex AI backend requires GOOGLE_CLOUD_PROJECT environment variable or --project flag")
		}
		if a.config.VertexAILocation == "" {
			return fmt.Errorf("Vertex AI backend requires GOOGLE_CLOUD_LOCATION environment variable or --location flag")
		}
		a.llmModel, err = models.CreateVertexAIModel(a.ctx, models.VertexAIConfig{
			Project:   a.config.VertexAIProject,
			Location:  a.config.VertexAILocation,
			ModelName: actualModelID,
		})

	case "openai":
		openaiKey := os.Getenv("OPENAI_API_KEY")
		if openaiKey == "" {
			return fmt.Errorf("OpenAI backend requires OPENAI_API_KEY environment variable")
		}
		a.llmModel, err = models.CreateOpenAIModel(a.ctx, models.OpenAIConfig{
			APIKey:    openaiKey,
			ModelName: actualModelID,
		})

	case "gemini":
		fallthrough
	default:
		a.llmModel, err = models.CreateGeminiModel(a.ctx, models.GeminiConfig{
			APIKey:    apiKey,
			ModelName: actualModelID,
		})
	}

	if err != nil {
		return fmt.Errorf("failed to create LLM model: %w", err)
	}

	return nil
}

// resolveWorkingDirectory resolves and validates the working directory
func (a *Application) resolveWorkingDirectory() string {
	workingDir := a.config.WorkingDirectory
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %v", err)
		}
	}

	// Expand ~ in the path
	if strings.HasPrefix(workingDir, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		workingDir = filepath.Join(homeDir, workingDir[1:])
	}

	return workingDir
}

// initializeAgent creates the coding agent
func (a *Application) initializeAgent() error {
	var err error
	a.codingAgent, err = codingagent.NewCodingAgent(a.ctx, codingagent.Config{
		Model:            a.llmModel,
		WorkingDirectory: a.config.WorkingDirectory,
		EnableThinking:   a.config.EnableThinking,
		ThinkingBudget:   a.config.ThinkingBudget,
	})
	if err != nil {
		return fmt.Errorf("failed to create coding agent: %w", err)
	}

	return nil
}

// initializeSession sets up session management
func (a *Application) initializeSession() error {
	var err error
	a.sessionManager, err = persistence.NewSessionManager("code_agent", a.config.DBPath)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	// Generate unique session name if not specified
	if a.config.SessionName == "" {
		a.config.SessionName = GenerateUniqueSessionName()
	}

	// Initialize the session
	sessionInit := NewSessionInitializer(a.sessionManager, a.bannerRenderer)
	if err := sessionInit.InitializeSession(a.ctx, "user1", a.config.SessionName); err != nil {
		return err
	}

	// Create agent runner
	sessionService := a.sessionManager.GetService()
	a.agentRunner, err = runner.New(runner.Config{
		AppName:        "code_agent",
		Agent:          a.codingAgent,
		SessionService: sessionService,
	})
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	// Initialize token tracking
	a.sessionTokens = tracking.NewSessionTokens()

	return nil
}

// initializeREPL sets up the REPL
func (a *Application) initializeREPL() error {
	var err error
	a.repl, err = NewREPL(REPLConfig{
		UserID:           "user1",
		SessionName:      a.config.SessionName,
		Renderer:         a.renderer,
		BannerRenderer:   a.bannerRenderer,
		StreamingDisplay: a.streamDisplay,
		TypewriterPrint:  a.typewriter,
		Runner:           a.agentRunner,
		SessionTokens:    a.sessionTokens,
		ModelRegistry:    a.modelRegistry,
		SelectedModel:    a.selectedModel,
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
	if a.sessionManager != nil {
		a.sessionManager.Close()
	}
	if a.signalHandler != nil {
		a.signalHandler.Cancel()
	}
}
