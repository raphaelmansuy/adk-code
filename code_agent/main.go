// Code Agent - A CLI coding assistant powered by Google ADK Go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
	ctx := context.Background()

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

	// Interactive loop
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Show prompt
		promptText := renderer.Bold("❯") + " "
		fmt.Print(renderer.Cyan(promptText))

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
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

		for event, err := range agentRunner.Run(ctx, userID, cliConfig.SessionName, userMsg, agent.RunConfig{
			StreamingMode: agent.StreamingModeNone,
		}) {
			if err != nil {
				spinner.StopWithError("Error occurred")
				errMsg := renderer.RenderError(err)
				fmt.Print(errMsg)
				hasError = true
				break
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

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
}
