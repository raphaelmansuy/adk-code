// Code Agent - A CLI coding assistant powered by Google ADK Go
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
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
	outputFormat := flag.String("output-format", "rich", "Output format: rich, plain, or json")
	typewriterEnabled := flag.Bool("typewriter", false, "Enable typewriter effect for text output")
	sessionName := flag.String("session", "", "Session name (optional, defaults to 'default')")
	dbPath := flag.String("db", "", "Database path for sessions (optional, defaults to ~/.code_agent/sessions.db)")
	flag.Parse()

	// Handle special commands (new-session, list-sessions, etc.)
	args := flag.Args()
	if len(args) > 0 {
		cmd := args[0]
		if cmd == "new-session" {
			if len(args) < 2 {
				fmt.Println("Usage: code-agent new-session <session-name>")
				os.Exit(1)
			}
			handleNewSession(ctx, args[1], *dbPath)
			os.Exit(0)
		} else if cmd == "list-sessions" {
			handleListSessions(ctx, *dbPath)
			os.Exit(0)
		} else if cmd == "delete-session" {
			if len(args) < 2 {
				fmt.Println("Usage: code-agent delete-session <session-name>")
				os.Exit(1)
			}
			handleDeleteSession(ctx, args[1], *dbPath)
			os.Exit(0)
		}
	}

	// Generate unique session name if not specified
	// This ensures each run without --session gets a new session
	if *sessionName == "" {
		*sessionName = generateUniqueSessionName()
	}

	// Create renderer
	renderer, err := display.NewRenderer(*outputFormat)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}

	bannerRenderer := display.NewBannerRenderer(renderer)

	// Create typewriter printer
	typewriter := display.NewTypewriterPrinter(display.DefaultTypewriterConfig())
	typewriter.SetEnabled(*typewriterEnabled)

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
	sessionManager, err := persistence.NewSessionManager("code_agent", *dbPath)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}
	defer sessionManager.Close()

	// Get or create the session
	userID := "user1"
	sess, err := sessionManager.GetSession(ctx, userID, *sessionName)
	if err != nil {
		// Session doesn't exist, create it
		sess, err = sessionManager.CreateSession(ctx, userID, *sessionName)
		if err != nil {
			log.Fatalf("Failed to create session: %v", err)
		}
		fmt.Printf("✨ Created new session: %s\n\n", *sessionName)
	} else {
		fmt.Printf("📖 Resumed session: %s (%d events)\n\n", *sessionName, sess.Events().Len())
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

		if input == "/exit" || input == "/quit" {
			goodbye := renderer.Cyan("Goodbye! Happy coding! 👋")
			fmt.Printf("\n%s\n", goodbye)
			break
		}

		// Debug command to show system prompt
		if input == "/prompt" {
			fmt.Print(renderer.Yellow("\n=== System Prompt ===\n\n"))
			fmt.Print(renderer.Dim(codingagent.EnhancedSystemPrompt))
			fmt.Print(renderer.Yellow("\n\n=== End of Prompt ===\n\n"))
			continue
		}

		// Help command
		if input == "/help" {
			fmt.Print("\n" + renderer.Cyan("════════════════════════════════════════════════════════════════\n"))
			fmt.Print(renderer.Cyan("                       Code Agent Help\n"))
			fmt.Print(renderer.Cyan("════════════════════════════════════════════════════════════════\n") + "\n")

			fmt.Print(renderer.Bold("🤖 Natural Language Requests:\n"))
			fmt.Print("   Just type what you want in plain English!\n\n")

			fmt.Print(renderer.Bold("⌨️  Built-in Commands:\n"))
			fmt.Print("   • " + renderer.Bold("/help") + " - Show this help message\n")
			fmt.Print("   • " + renderer.Bold("/tools") + " - List all available tools\n")
			fmt.Print("   • " + renderer.Bold("/prompt") + " - Display the system prompt\n")
			fmt.Print("   • " + renderer.Bold("/tokens") + " - Show token usage statistics\n")
			fmt.Print("   • " + renderer.Bold("/exit") + " - Exit the agent\n")

			fmt.Print(renderer.Bold("\n📚 Session Management (CLI commands):\n"))
			fmt.Print("   • " + renderer.Bold("./code-agent new-session <name>") + " - Create a new session\n")
			fmt.Print("   • " + renderer.Bold("./code-agent list-sessions") + " - List all sessions\n")
			fmt.Print("   • " + renderer.Bold("./code-agent delete-session <name>") + " - Delete a session\n")
			fmt.Print("   • " + renderer.Bold("./code-agent --session <name>") + " - Resume a specific session\n")

			fmt.Print(renderer.Bold("\n💡 Example Requests:\n"))
			fmt.Print("   ❯ Add error handling to main.go\n")
			fmt.Print("   ❯ Create a README.md with project overview\n")
			fmt.Print("   ❯ Refactor the calculate function\n")
			fmt.Print("   ❯ Run tests and fix any failures\n")
			fmt.Print("   ❯ Add comments to all Python files\n\n")

			fmt.Print(renderer.Yellow("📖 More info: ") + "See USER_GUIDE.md for detailed documentation\n\n")
			continue
		}

		// Tools listing command
		if input == "/tools" {
			fmt.Print("\n" + renderer.Cyan("════════════════════════════════════════════════════════════════\n"))
			fmt.Print(renderer.Cyan("                    Available Tools\n"))
			fmt.Print(renderer.Cyan("════════════════════════════════════════════════════════════════\n") + "\n")

			fmt.Print(renderer.Bold("📝 Core Editing Tools:\n"))
			fmt.Print("   ✓ " + renderer.Bold("read_file") + " - Read file contents (supports line ranges)\n")
			fmt.Print("   ✓ " + renderer.Bold("write_file") + " - Create or overwrite files (atomic, safe)\n")
			fmt.Print("   ✓ " + renderer.Bold("search_replace") + " - Make targeted changes (RECOMMENDED)\n")
			fmt.Print("   ✓ " + renderer.Bold("edit_lines") + " - Edit by line number (structural changes)\n")
			fmt.Print("   ✓ " + renderer.Bold("apply_patch") + " - Apply unified diff patches (standard)\n")
			fmt.Print("   ✓ " + renderer.Bold("apply_v4a_patch") + " - Apply V4A semantic patches (NEW!)\n")

			fmt.Print(renderer.Bold("\n🔍 Discovery Tools:\n"))
			fmt.Print("   ✓ " + renderer.Bold("list_files") + " - Explore directory structure\n")
			fmt.Print("   ✓ " + renderer.Bold("search_files") + " - Find files by pattern (*.go, test_*.py)\n")
			fmt.Print("   ✓ " + renderer.Bold("grep_search") + " - Search text in files (with line numbers)\n")

			fmt.Print(renderer.Bold("\n⚡ Execution Tools:\n"))
			fmt.Print("   ✓ " + renderer.Bold("execute_command") + " - Run shell commands (pipes, redirects)\n")
			fmt.Print("   ✓ " + renderer.Bold("execute_program") + " - Run programs directly (no quoting issues)\n\n")

			fmt.Print("💡 Tip: Type " + renderer.Cyan("'/help'") + " for usage examples and patterns\n\n")
			continue
		}

		// Token usage reporting command
		if input == "/tokens" {
			summary := sessionTokens.GetSummary()
			fmt.Print(tracking.FormatSessionSummary(summary))
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

		for event, err := range agentRunner.Run(ctx, userID, *sessionName, userMsg, agent.RunConfig{
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

func printEventEnhanced(renderer *display.Renderer, streamDisplay *display.StreamingDisplay,
	event *session.Event, spinner *display.Spinner, activeToolName *string, toolRunning *bool,
	sessionTokens *tracking.SessionTokens, requestID string) {

	if event.Content == nil || len(event.Content.Parts) == 0 {
		return
	}

	// Record token metrics if available and update spinner with metrics
	if event.UsageMetadata != nil {
		sessionTokens.RecordMetrics(event.UsageMetadata, requestID)
		// Create token metrics for spinner display
		metric := &tracking.TokenMetrics{
			PromptTokens:   event.UsageMetadata.PromptTokenCount,
			CachedTokens:   event.UsageMetadata.CachedContentTokenCount,
			ResponseTokens: event.UsageMetadata.CandidatesTokenCount,
			ThoughtTokens:  event.UsageMetadata.ThoughtsTokenCount,
			ToolUseTokens:  event.UsageMetadata.ToolUsePromptTokenCount,
			TotalTokens:    event.UsageMetadata.TotalTokenCount,
		}
		// Update spinner with metrics if it's actively running
		if *toolRunning {
			spinner.UpdateWithMetrics("Processing", metric)
		} else {
			spinner.UpdateWithMetrics("Agent is thinking", metric)
		}
	}

	// Create tool renderer with enhanced features
	toolRenderer := display.NewToolRenderer(renderer)
	toolResultParser := display.NewToolResultParser(nil)

	for _, part := range event.Content.Parts {
		// Handle text content
		if part.Text != "" {
			// Only stop spinner for actual agent responses (not tool-related text)
			text := part.Text
			isToolRelated := strings.Contains(text, "read_file") ||
				strings.Contains(text, "write_file") ||
				strings.Contains(text, "execute_command") ||
				strings.Contains(text, "list_directory") ||
				strings.Contains(text, "grep_search") ||
				strings.Contains(text, "search_replace") ||
				strings.Contains(text, "edit_lines") ||
				strings.Contains(text, "apply_patch")

			if !isToolRelated {
				// This is actual agent response text, stop spinner
				spinner.Stop()

				// Detect if this is thinking/reasoning text
				isThinking := strings.Contains(strings.ToLower(text), "thinking") ||
					strings.Contains(strings.ToLower(text), "analyzing") ||
					strings.Contains(strings.ToLower(text), "considering")

				if isThinking {
					// Update spinner message instead of stopping
					spinner.Update("Analyzing your request")
				} else {
					// Render the actual text content
					output := renderer.RenderPartContent(part)
					fmt.Print(output)
				}
			}
		}

		// Handle function calls - show what tool is being executed
		if part.FunctionCall != nil {
			// First, stop the current spinner to print the tool banner
			spinner.Stop()

			*activeToolName = part.FunctionCall.Name
			*toolRunning = true

			args := make(map[string]any)
			for k, v := range part.FunctionCall.Args {
				args[k] = v
			}

			// Show what tool is being executed
			output := toolRenderer.RenderToolExecution(part.FunctionCall.Name, args)
			fmt.Print(output)

			// Start spinner with context-aware message for the tool execution
			spinnerMessage := getToolSpinnerMessage(part.FunctionCall.Name, args)
			spinner.Update(spinnerMessage)
			spinner.Start()
		}

		// Handle function responses - show the result
		if part.FunctionResponse != nil {
			// Stop spinner now that tool is complete
			spinner.Stop()
			*toolRunning = false

			result := make(map[string]any)
			if part.FunctionResponse.Response != nil {
				for k, v := range part.FunctionResponse.Response {
					result[k] = v
				}
			}

			// Use enhanced result parser for structured output
			parsedResult := toolResultParser.ParseToolResult(part.FunctionResponse.Name, result)
			if parsedResult != "" {
				// Show parsed result
				fmt.Print("\n")
				fmt.Print(parsedResult)
				fmt.Print("\n")
			}

			// Show basic result indicator (compact version)
			resultOutput := renderer.RenderToolResult(part.FunctionResponse.Name, result)
			fmt.Print(resultOutput)

			// Restart spinner for next operation (agent might still be working)
			// Update message and restart
			spinner.Update("Processing")
			spinner.Start()
		}
	}
}

// getToolSpinnerMessage returns a context-aware spinner message for tool execution
func getToolSpinnerMessage(toolName string, args map[string]any) string {
	switch toolName {
	case "read_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Reading %s", filepath.Base(path))
		}
		return "Reading file"
	case "write_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Writing %s", filepath.Base(path))
		}
		return "Writing file"
	case "search_replace", "replace_in_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Editing %s", filepath.Base(path))
		}
		return "Editing file"
	case "edit_lines":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Modifying %s", filepath.Base(path))
		}
		return "Modifying file"
	case "apply_patch", "apply_v4a_patch":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Applying patch to %s", filepath.Base(path))
		}
		return "Applying patch"
	case "list_directory", "list_files":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Listing %s", filepath.Base(path))
		}
		return "Listing directory"
	case "search_files":
		if pattern, ok := args["pattern"].(string); ok {
			return fmt.Sprintf("Searching for %s", pattern)
		}
		return "Searching files"
	case "grep_search":
		if pattern, ok := args["pattern"].(string); ok {
			return fmt.Sprintf("Searching for '%s'", pattern)
		}
		return "Searching code"
	case "execute_command":
		if command, ok := args["command"].(string); ok {
			// Truncate long commands
			if len(command) > 40 {
				command = command[:37] + "..."
			}
			return fmt.Sprintf("Running: %s", command)
		}
		return "Running command"
	case "execute_program":
		if program, ok := args["program"].(string); ok {
			return fmt.Sprintf("Executing %s", filepath.Base(program))
		}
		return "Executing program"
	default:
		return fmt.Sprintf("Running %s", toolName)
	}
}

// handleNewSession creates a new session
func handleNewSession(ctx context.Context, sessionName string, dbPath string) {
	manager, err := persistence.NewSessionManager("code_agent", dbPath)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	userID := "user1"
	_, err = manager.CreateSession(ctx, userID, sessionName)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	fmt.Printf("✨ Created new session: %s\n", sessionName)
}

// handleListSessions lists all sessions
func handleListSessions(ctx context.Context, dbPath string) {
	manager, err := persistence.NewSessionManager("code_agent", dbPath)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	userID := "user1"
	sessions, err := manager.ListSessions(ctx, userID)
	if err != nil {
		log.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) == 0 {
		fmt.Println("📭 No sessions found")
		return
	}

	fmt.Println("📋 Sessions:")
	for i, sess := range sessions {
		eventCount := sess.Events().Len()
		fmt.Printf("%d. %s (%d events)\n", i+1, sess.ID(), eventCount)
	}
}

// handleDeleteSession deletes a session
func handleDeleteSession(ctx context.Context, sessionName string, dbPath string) {
	manager, err := persistence.NewSessionManager("code_agent", dbPath)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	userID := "user1"
	err = manager.DeleteSession(ctx, userID, sessionName)
	if err != nil {
		log.Fatalf("Failed to delete session: %v", err)
	}

	fmt.Printf("🗑️  Deleted session: %s\n", sessionName)
}

// generateUniqueSessionName creates a unique session name based on timestamp
// Format: session-YYYYMMDD-HHMMSS (e.g., session-20251110-221530)
func generateUniqueSessionName() string {
	now := time.Now()
	return fmt.Sprintf("session-%d%02d%02d-%02d%02d%02d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second())
}
