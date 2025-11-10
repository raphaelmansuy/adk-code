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
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	codingagent "code_agent/agent"
)

const (
	// ANSI color codes for pretty output
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

func main() {
	ctx := context.Background()

	// Get API key from environment
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	// Print welcome banner
	printBanner()

	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	fmt.Printf("%s%sWorking directory:%s %s\n\n", colorBold, colorCyan, colorReset, workingDir)

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

	// Interactive loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("%s%s╭─ Enter your coding task (or 'exit' to quit)%s\n", colorBold, colorGreen, colorReset)

	for {
		fmt.Printf("%s%s╰─❯%s ", colorBold, colorGreen, colorReset)

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Printf("\n%s%sGoodbye! Happy coding! 👋%s\n", colorBold, colorCyan, colorReset)
			break
		}

		// Create user message
		userMsg := &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{Text: input},
			},
		}

		// Run agent
		fmt.Printf("\n%s%s🤖 Agent:%s Thinking...\n\n", colorBold, colorBlue, colorReset)

		hasError := false
		for event, err := range agentRunner.Run(ctx, userID, sessionID, userMsg, agent.RunConfig{
			StreamingMode: agent.StreamingModeNone,
		}) {
			if err != nil {
				fmt.Printf("%s%sError:%s %v\n", colorBold, colorRed, colorReset, err)
				hasError = true
				break
			}

			if event != nil {
				printEvent(event)
			}
		}

		if !hasError {
			fmt.Printf("\n%s%s✓ Task completed%s\n\n", colorBold, colorGreen, colorReset)
		} else {
			fmt.Printf("\n%s%s✗ Task failed%s\n\n", colorBold, colorRed, colorReset)
		}

		fmt.Printf("%s%s╭─ Next task?%s\n", colorBold, colorGreen, colorReset)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   ██████╗ ██████╗ ██████╗ ███████╗     █████╗  ██████╗   ║
║  ██╔════╝██╔═══██╗██╔══██╗██╔════╝    ██╔══██╗██╔════╝   ║
║  ██║     ██║   ██║██║  ██║█████╗      ███████║██║  ███╗  ║
║  ██║     ██║   ██║██║  ██║██╔══╝      ██╔══██║██║   ██║  ║
║  ╚██████╗╚██████╔╝██████╔╝███████╗    ██║  ██║╚██████╔╝  ║
║   ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝    ╚═╝  ╚═╝ ╚═════╝   ║
║                                                           ║
║            AI-Powered Coding Assistant                    ║
║            Built with Google ADK Go                       ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Printf("%s%s%s%s\n", colorBold, colorCyan, banner, colorReset)
}

func printEvent(event *session.Event) {
	if event.Content != nil && len(event.Content.Parts) > 0 {
		for _, part := range event.Content.Parts {
			if part.Text != "" {
				// Check if it's a tool call or response
				text := part.Text
				if strings.Contains(text, "read_file") || strings.Contains(text, "write_file") ||
					strings.Contains(text, "execute_command") || strings.Contains(text, "list_directory") {
					fmt.Printf("%s%s🔧 Tool:%s %s\n", colorBold, colorYellow, colorReset, text)
				} else {
					fmt.Printf("%s", text)
				}
			}

			// Handle function calls
			if part.FunctionCall != nil {
				fmt.Printf("%s%s🔧 Calling tool:%s %s\n", colorBold, colorYellow, colorReset, part.FunctionCall.Name)
				// Print arguments if they're not too long
				if len(fmt.Sprintf("%v", part.FunctionCall.Args)) < 200 {
					fmt.Printf("%s   Args:%s %v\n", colorBold, colorReset, part.FunctionCall.Args)
				}
			}

			// Handle function responses
			if part.FunctionResponse != nil {
				fmt.Printf("%s%s✓ Tool result:%s %s\n", colorBold, colorGreen, colorReset, part.FunctionResponse.Name)
				// Print response if it's not too long
				responseStr := fmt.Sprintf("%v", part.FunctionResponse.Response)
				if len(responseStr) < 500 {
					fmt.Printf("%s   Result:%s %s\n", colorBold, colorReset, responseStr)
				} else {
					fmt.Printf("%s   Result:%s [Large output - %d bytes]\n", colorBold, colorReset, len(responseStr))
				}
			}
		}
	}
}
