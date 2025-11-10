// Code Agent - A CLI coding assistant powered by Google ADK Go
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	codingagent "code_agent/agent"
	"code_agent/display"
)

const version = "1.0.0"

func main() {
	ctx := context.Background()

	// Parse command-line flags
	outputFormat := flag.String("output-format", "rich", "Output format: rich, plain, or json")
	typewriterEnabled := flag.Bool("typewriter", false, "Enable typewriter effect for text output")
	flag.Parse()

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

	// Create session service (in-memory for simplicity)
	sessionService := session.InMemoryService()

	// Create runner
	agentRunner, err := runner.New(runner.Config{
		AppName:        "code_agent",
		Agent:          codingAgent,
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	// Start interactive session
	userID := "user1"
	sessionID := "session1"

	// Create the session
	_, err = sessionService.Create(ctx, &session.CreateRequest{
		AppName:   "code_agent",
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

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

		if input == "exit" || input == "quit" {
			goodbye := renderer.Cyan("Goodbye! Happy coding! 👋")
			fmt.Printf("\n%s\n", goodbye)
			break
		}

		// Debug command to show system prompt
		if input == "debug prompt" || input == "show prompt" || input == ".prompt" {
			fmt.Print(renderer.Yellow("\n=== System Prompt ===\n\n"))
			fmt.Print(renderer.Dim(codingagent.EnhancedSystemPrompt))
			fmt.Print(renderer.Yellow("\n\n=== End of Prompt ===\n\n"))
			continue
		}

		// Help command
		if input == "help" || input == ".help" {
			fmt.Print(renderer.Cyan("\n════════════════════════════════════════════════════════════════\n"))
			fmt.Print(renderer.Cyan("  ") + renderer.Bold("Code Agent Help") + "\n")
			fmt.Print(renderer.Cyan("════════════════════════════════════════════════════════════════\n\n"))

			fmt.Print(renderer.Bold("🤖 Natural Language Requests:\n"))
			fmt.Print(renderer.Dim("   Just type what you want in plain English!\n\n"))

			fmt.Print(renderer.Bold("⌨️  Built-in Commands:\n"))
			fmt.Print(renderer.Green("  • ") + renderer.Bold("help") + " / " + renderer.Bold(".help") + " - Show this help message\n")
			fmt.Print(renderer.Green("  • ") + renderer.Bold(".tools") + " - List all available tools\n")
			fmt.Print(renderer.Green("  • ") + renderer.Bold(".prompt") + " - Display the system prompt\n")
			fmt.Print(renderer.Green("  • ") + renderer.Bold("exit") + " / " + renderer.Bold("quit") + " - Exit the agent\n")

			fmt.Print(renderer.Bold("\n💡 Example Requests:\n"))
			fmt.Print(renderer.Green("  ❯ ") + renderer.Dim("Add error handling to main.go\n"))
			fmt.Print(renderer.Green("  ❯ ") + renderer.Dim("Create a README.md with project overview\n"))
			fmt.Print(renderer.Green("  ❯ ") + renderer.Dim("Refactor the calculate function\n"))
			fmt.Print(renderer.Green("  ❯ ") + renderer.Dim("Run tests and fix any failures\n"))
			fmt.Print(renderer.Green("  ❯ ") + renderer.Dim("Add comments to all Python files\n\n"))

			fmt.Print(renderer.Yellow("📖 More info: ") + renderer.Dim("See USER_GUIDE.md for detailed documentation\n\n"))
			continue
		}

		// Tools listing command
		if input == ".tools" {
			fmt.Print(renderer.Cyan("\n════════════════════════════════════════════════════════════════\n"))
			fmt.Print(renderer.Cyan("  ") + renderer.Bold("Available Tools") + "\n")
			fmt.Print(renderer.Cyan("════════════════════════════════════════════════════════════════\n\n"))

			fmt.Print(renderer.Bold("📝 Core Editing Tools:\n"))
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("read_file") + " - Read file contents (supports line ranges)\n")
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("write_file") + " - Create or overwrite files (atomic, safe)\n")
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("search_replace") + " - Make targeted changes (RECOMMENDED)\n")
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("edit_lines") + " - Edit by line number (structural changes)\n")
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("apply_patch") + " - Apply unified diff patches (standard)\n")
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("apply_v4a_patch") + " - Apply V4A semantic patches (NEW!)\n")

			fmt.Print(renderer.Bold("\n🔍 Discovery Tools:\n"))
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("list_files") + " - Explore directory structure\n")
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("search_files") + " - Find files by pattern (*.go, test_*.py)\n")
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("grep_search") + " - Search text in files (with line numbers)\n")

			fmt.Print(renderer.Bold("\n⚡ Execution Tools:\n"))
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("execute_command") + " - Run shell commands (pipes, redirects)\n")
			fmt.Print(renderer.Green("  ✓ ") + renderer.Bold("execute_program") + " - Run programs directly (no quoting issues)\n\n")

			fmt.Print(renderer.Dim("💡 Tip: Type ") + renderer.Cyan("'help'") + renderer.Dim(" for usage examples and patterns\n\n"))
			continue
		}

		// Create user message
		userMsg := &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{Text: input},
			},
		}

		// Run agent with spinner
		spinner := display.NewSpinner(renderer, "Agent is thinking")
		spinner.Start()

		hasError := false
		var activeToolName string

		for event, err := range agentRunner.Run(ctx, userID, sessionID, userMsg, agent.RunConfig{
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
				printEventEnhanced(renderer, streamingDisplay, event, spinner, &activeToolName)
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
	event *session.Event, spinner *display.Spinner, activeToolName *string) {

	if event.Content == nil || len(event.Content.Parts) == 0 {
		return
	}

	// Create tool renderer with enhanced features
	toolRenderer := display.NewToolRenderer(renderer)
	toolResultParser := display.NewToolResultParser(nil)

	for _, part := range event.Content.Parts {
		// Handle text content
		if part.Text != "" {
			// Stop spinner once for text output
			spinner.Stop()

			// Detect if this is thinking/reasoning text
			isThinking := strings.Contains(strings.ToLower(part.Text), "thinking") ||
				strings.Contains(strings.ToLower(part.Text), "analyzing") ||
				strings.Contains(strings.ToLower(part.Text), "considering")

			if isThinking {
				// Render as thinking
				output := renderer.RenderAgentWorking("Thinking")
				fmt.Print(output)
			}

			// Render the actual text content
			output := renderer.RenderPartContent(part)
			fmt.Print(output)
		}

		// Handle function calls - show what tool is being executed
		if part.FunctionCall != nil {
			// Stop spinner
			spinner.Stop()

			args := make(map[string]any)
			for k, v := range part.FunctionCall.Args {
				args[k] = v
			}

			*activeToolName = part.FunctionCall.Name

			// Use enhanced tool renderer with "is doing" verb tense
			output := toolRenderer.RenderToolExecution(part.FunctionCall.Name, args)
			fmt.Print(output)
		}

		// Handle function responses - show the result
		if part.FunctionResponse != nil {
			// Stop spinner
			spinner.Stop()

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
			output := renderer.RenderToolResult(part.FunctionResponse.Name, result)
			fmt.Print(output)
		}
	}
}
