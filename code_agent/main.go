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
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/genai"

	codingagent "code_agent/agent"
	"code_agent/display"
	"code_agent/persistence"
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
	cliConfig, args := ParseCLIFlags()

	// Handle special commands (new-session, list-sessions, etc.)
	if HandleCLICommands(ctx, args, cliConfig.DBPath) {
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

	// Get API key from environment
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Print welcome banner
	modelName := "gemini-2.5-flash"
	banner := bannerRenderer.RenderStartBanner(version, modelName, workingDir)
	fmt.Print(banner)

	// Create Gemini model
	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create coding agent
	codingAgent, err := codingagent.NewCodingAgent(ctx, codingagent.Config{
		Model:            model,
		WorkingDirectory: workingDir,
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
		fmt.Printf("📖 Resumed session: %s (%d events)\n\n", cliConfig.SessionName, sess.Events().Len())
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
		if handleBuiltinCommand(input, renderer, sessionTokens) {
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
				printEventEnhanced(renderer, streamingDisplay, event, spinner, &activeToolName, &toolRunning, sessionTokens, requestID)
			}
		}

		// Stop spinner and show completion
		if !hasError {
			spinner.StopWithSuccess("Task completed")
			completion := renderer.RenderTaskComplete()
			fmt.Print(completion)
		} else {
			failure := renderer.RenderTaskFailed()
			fmt.Print(failure)
		}
	}
}
