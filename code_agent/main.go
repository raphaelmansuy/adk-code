// Code Agent - A CLI coding assistant powered by Google ADK Go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/genai"

	codingagent "code_agent/agent"
	"code_agent/display"
	"code_agent/persistence"
	"code_agent/pkg/cli"
	"code_agent/pkg/models"
	"code_agent/tracking"
)

const version = "1.0.0"

func main() {
	// Setup signal handling for graceful Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Track if we've received Ctrl+C once
	ctrlCCount := 0

	// Handle signals in a goroutine
	go func() {
		for sig := range sigChan {
			ctrlCCount++
			if sig == syscall.SIGINT {
				if ctrlCCount == 1 {
					fmt.Println("\n\n⚠️  Interrupted by user (Ctrl+C)")
					fmt.Println("Cancelling current operation...")
				} else {
					fmt.Println("\n\n⚠️  Ctrl+C pressed again - forcing exit")
					os.Exit(130) // Standard exit code for SIGINT
				}
			}
			cancel()
		}
	}()

	// Parse command-line flags
	cliConfig, args := cli.ParseCLIFlags()

	// Handle special commands (new-session, list-sessions, etc.)
	if cli.HandleCLICommands(ctx, args, cliConfig.DBPath) {
		os.Exit(0)
	}

	// Generate unique session name if not specified
	// This ensures each run without --session gets a new session
	if cliConfig.SessionName == "" {
		cliConfig.SessionName = generateUniqueSessionName()
	}

	// Create renderer
	renderer, err := display.NewRenderer(cliConfig.OutputFormat)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}

	bannerRenderer := display.NewBannerRenderer(renderer)

	// Create typewriter printer
	typewriter := display.NewTypewriterPrinter(display.DefaultTypewriterConfig())
	typewriter.SetEnabled(cliConfig.TypewriterEnabled)

	// Create streaming display
	streamingDisplay := display.NewStreamingDisplay(renderer, typewriter)

	// Create model registry
	modelRegistry := models.NewRegistry()

	// Resolve which model to use based on provider/model syntax
	var selectedModel models.Config

	if cliConfig.Model == "" {
		// No model specified, use default
		selectedModel = modelRegistry.GetDefaultModel()
	} else {
		// Parse the provider/model syntax
		parsedProvider, parsedModel, parseErr := cli.ParseProviderModelSyntax(cliConfig.Model)
		if parseErr != nil {
			log.Fatalf("Invalid model syntax: %v\nUse format: provider/model (e.g., gemini/2.5-flash)", parseErr)
		}

		// Determine default provider if not specified
		defaultProvider := cliConfig.Backend
		if defaultProvider == "" {
			defaultProvider = "gemini"
		}

		// Resolve the model using provider-aware lookup
		resolvedModel, modelErr := modelRegistry.ResolveFromProviderSyntax(
			parsedProvider,
			parsedModel,
			defaultProvider,
		)
		if modelErr != nil {
			log.Fatalf("❌ Error: %v\n\nAvailable models:\n", modelErr)
			for _, providerName := range modelRegistry.ListProviders() {
				models := modelRegistry.GetProviderModels(providerName)
				fmt.Printf("\n%s:\n", strings.ToUpper(providerName[:1])+strings.ToLower(providerName[1:]))
				for _, m := range models {
					fmt.Printf("  • %s/%s\n", providerName, m.ID)
				}
			}
			os.Exit(1)
		}
		selectedModel = resolvedModel
	}

	// Get API key from environment
	apiKey := cliConfig.APIKey
	if apiKey == "" && selectedModel.Backend == "gemini" {
		log.Fatal("Gemini API backend requires GOOGLE_API_KEY environment variable or --api-key flag")
	}

	// Get working directory
	workingDir := cliConfig.WorkingDirectory
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

	// Print welcome banner (before model creation so user knows what backend is being used)
	displayName := selectedModel.DisplayName
	banner := bannerRenderer.RenderStartBanner(version, displayName, workingDir)
	fmt.Print(banner)

	// Create model based on selected model configuration
	var llmModel model.LLM
	var modelErr error

	// Extract the actual model ID for the API (remove -vertex suffix if present)
	actualModelID := models.ExtractModelIDFromGemini(selectedModel.ID)

	switch selectedModel.Backend {
	case "vertexai":
		if cliConfig.VertexAIProject == "" {
			log.Fatal("Vertex AI backend requires GOOGLE_CLOUD_PROJECT environment variable or --project flag")
		}
		if cliConfig.VertexAILocation == "" {
			log.Fatal("Vertex AI backend requires GOOGLE_CLOUD_LOCATION environment variable or --location flag")
		}
		llmModel, modelErr = models.CreateVertexAIModel(ctx, models.VertexAIConfig{
			Project:   cliConfig.VertexAIProject,
			Location:  cliConfig.VertexAILocation,
			ModelName: actualModelID,
		})

	case "openai":
		openaiKey := os.Getenv("OPENAI_API_KEY")
		if openaiKey == "" {
			log.Fatal("OpenAI backend requires OPENAI_API_KEY environment variable")
		}
		llmModel, modelErr = models.CreateOpenAIModel(ctx, models.OpenAIConfig{
			APIKey:    openaiKey,
			ModelName: actualModelID,
		})

	case "gemini":
		fallthrough
	default:
		llmModel, modelErr = models.CreateGeminiModel(ctx, models.GeminiConfig{
			APIKey:    apiKey,
			ModelName: actualModelID,
		})
	}

	if modelErr != nil {
		log.Fatalf("Failed to create LLM model: %v", modelErr)
	}

	// Create coding agent
	codingAgent, err := codingagent.NewCodingAgent(ctx, codingagent.Config{
		Model:            llmModel,
		WorkingDirectory: workingDir,
		EnableThinking:   cliConfig.EnableThinking,
		ThinkingBudget:   cliConfig.ThinkingBudget,
	})
	if err != nil {
		log.Fatalf("Failed to create coding agent: %v", err)
	}

	// Create session manager with persistent storage
	sessionManager, err := persistence.NewSessionManager("code_agent", cliConfig.DBPath)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}
	defer sessionManager.Close()

	// Get or create the session
	userID := "user1"
	sess, err := sessionManager.GetSession(ctx, userID, cliConfig.SessionName)
	if err != nil {
		// Session doesn't exist, create it
		_, err = sessionManager.CreateSession(ctx, userID, cliConfig.SessionName)
		if err != nil {
			log.Fatalf("Failed to create session: %v", err)
		}
		fmt.Printf("✨ Created new session: %s\n\n", cliConfig.SessionName)
	} else {
		// Use enhanced session resume header with event count and tokens
		resumeInfo := bannerRenderer.RenderSessionResumeInfo(cliConfig.SessionName, sess.Events().Len(), 0)
		fmt.Print(resumeInfo)
	}

	// Create runner with persistent session service
	sessionService := sessionManager.GetService()
	agentRunner, err := runner.New(runner.Config{
		AppName:        "code_agent",
		Agent:          codingAgent,
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	// Initialize token tracking
	sessionTokens := tracking.NewSessionTokens()

	// Track last operation status for prompt indicator
	lastOperationSuccess := false

	// Show welcome message
	welcome := bannerRenderer.RenderWelcome()
	fmt.Print(welcome)

	// Setup readline with history persistence
	historyFile := filepath.Join(os.Getenv("HOME"), ".code_agent_history")

	l, err := readline.NewEx(&readline.Config{
		Prompt:          renderer.Cyan(renderer.Bold("❯") + " "),
		HistoryFile:     historyFile,
		HistoryLimit:    500,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		FuncFilterInputRune: func(r rune) (rune, bool) {
			// Allow all characters
			return r, true
		},
	})
	if err != nil {
		log.Fatalf("Failed to create readline instance: %v", err)
	}
	defer l.Close()

	// Interactive loop with readline
	for {
		// Check if context was cancelled (e.g., by Ctrl+C signal)
		select {
		case <-ctx.Done():
			fmt.Printf("\n%s\n", renderer.Cyan("Goodbye! Happy coding! 👋"))
			return
		default:
		}

		// Readline handles the prompt and history navigation automatically
		input, err := l.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				// Handle Ctrl+C from readline - break the loop to exit gracefully
				fmt.Printf("\n%s\n", renderer.Cyan("Goodbye! Happy coding! 👋"))
				return
			} else {
				break
			}
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Save to history
		l.SaveHistory(input)

		// Check for exit commands first
		if input == "/exit" || input == "/quit" {
			goodbye := renderer.Cyan("Goodbye! Happy coding! 👋")
			fmt.Printf("\n%s\n", goodbye)
			break
		}

		// Handle built-in commands
		if cli.HandleBuiltinCommand(input, renderer, sessionTokens, modelRegistry, selectedModel) {
			continue
		}

		// Create user message
		userMsg := &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{Text: input},
			},
		}

		// Run agent with enhanced spinner
		spinner := display.NewSpinner(renderer, "Agent is thinking")
		spinner.Start()

		// Create event timeline for this request
		timeline := display.NewEventTimeline()

		hasError := false
		var activeToolName string
		toolRunning := false
		requestID := fmt.Sprintf("req_%d", sessionTokens.GetSummary().RequestCount+1)

	agentLoop:
		for event, err := range agentRunner.Run(ctx, userID, cliConfig.SessionName, userMsg, agent.RunConfig{
			StreamingMode: agent.StreamingModeNone,
		}) {
			// Check if context was cancelled (Ctrl+C)
			select {
			case <-ctx.Done():
				spinner.StopWithError("Task interrupted")
				fmt.Printf("\n%s\n", renderer.Yellow("⚠️  Task cancelled by user"))
				hasError = true
				break agentLoop
			default:
			}

			if err != nil {
				spinner.StopWithError("Error occurred")
				errMsg := renderer.RenderError(err)
				fmt.Print(errMsg)
				hasError = true
				break agentLoop
			}

			if event != nil {
				display.PrintEventEnhanced(renderer, streamingDisplay, event, spinner, &activeToolName, &toolRunning, sessionTokens, requestID, timeline)
			}
		}

		// Stop spinner and show completion
		if !hasError {
			spinner.StopWithSuccess("Task completed")
			completion := renderer.RenderTaskComplete()
			fmt.Print(completion)
			lastOperationSuccess = true
		} else {
			failure := renderer.RenderTaskFailed()
			fmt.Print(failure)
			lastOperationSuccess = false
		}

		// Display event timeline if there were operations
		if timeline.GetEventCount() > 0 {
			fmt.Printf("%s\n", timeline.RenderTimeline())

			// Show progress indicator if multiple operations were performed
			if timeline.GetEventCount() > 1 {
				fmt.Printf("%s\n", timeline.RenderProgress())
			}
		}

		// Display token metrics for this request
		summary := sessionTokens.GetSummary()
		if summary.TotalTokens > 0 {
			metrics := renderer.RenderTokenMetrics(
				summary.TotalPromptTokens,
				summary.TotalCachedTokens,
				summary.TotalResponseTokens,
				summary.TotalTokens,
			)
			if metrics != "" {
				fmt.Printf("%s\n", metrics)
			}
		}

		// Update prompt based on last operation status
		if lastOperationSuccess {
			l.SetPrompt(renderer.Green("✓ ") + renderer.Cyan(renderer.Bold("❯")+" "))
		} else {
			l.SetPrompt(renderer.Cyan(renderer.Bold("❯") + " "))
		}
	}
}
