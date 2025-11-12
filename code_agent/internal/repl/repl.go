package repl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	sessionpkg "google.golang.org/adk/session"
	"google.golang.org/genai"

	"code_agent/internal/cli"
	"code_agent/internal/display"
	"code_agent/pkg/models"
	"code_agent/tracking"
)

// Config holds configuration for the REPL
type Config struct {
	UserID           string
	SessionName      string
	Renderer         *display.Renderer
	BannerRenderer   *display.BannerRenderer
	StreamingDisplay *display.StreamingDisplay
	TypewriterPrint  *display.TypewriterPrinter
	Runner           *runner.Runner
	SessionTokens    *tracking.SessionTokens
	ModelRegistry    *models.Registry
	SelectedModel    models.Config
}

// REPL manages the read-eval-print loop
type REPL struct {
	config       Config
	readline     *readline.Instance
	historyFile  string
	lastOpStatus bool
}

// New creates a new REPL instance
func New(config Config) (*REPL, error) {
	historyFile := filepath.Join(os.Getenv("HOME"), ".code_agent_history")

	l, err := readline.NewEx(&readline.Config{
		Prompt:          config.Renderer.Cyan(config.Renderer.Bold("❯") + " "),
		HistoryFile:     historyFile,
		HistoryLimit:    500,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		FuncFilterInputRune: func(r rune) (rune, bool) {
			return r, true
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create readline instance: %w", err)
	}

	return &REPL{
		config:      config,
		readline:    l,
		historyFile: historyFile,
	}, nil
}

// Close closes the REPL resources
func (r *REPL) Close() error {
	return r.readline.Close()
}

// Run starts the REPL loop
func (r *REPL) Run(ctx context.Context) {
	// Show welcome message
	welcome := r.config.BannerRenderer.RenderWelcome()
	fmt.Print(welcome)

	for {
		// Check if context was cancelled
		select {
		case <-ctx.Done():
			fmt.Printf("\n%s\n", r.config.Renderer.Cyan("Goodbye! Happy coding! 👋"))
			return
		default:
		}

		// Read input
		input, err := r.readline.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				fmt.Printf("\n%s\n", r.config.Renderer.Cyan("Goodbye! Happy coding! 👋"))
				return
			}
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Save to history
		r.readline.SaveHistory(input)

		// Check for exit commands
		if input == "/exit" || input == "/quit" {
			goodbye := r.config.Renderer.Cyan("Goodbye! Happy coding! 👋")
			fmt.Printf("\n%s\n", goodbye)
			break
		}

		// Handle built-in commands
		if cli.HandleBuiltinCommand(input, r.config.Renderer, r.config.SessionTokens, r.config.ModelRegistry, r.config.SelectedModel) {
			continue
		}

		// Process user message
		r.processUserMessage(ctx, input)
	}
}

// processUserMessage handles a user input message
func (r *REPL) processUserMessage(ctx context.Context, input string) {
	// Create user message
	userMsg := &genai.Content{
		Role: genai.RoleUser,
		Parts: []*genai.Part{
			{Text: input},
		},
	}

	// Run agent with enhanced spinner
	spinner := display.NewSpinner(r.config.Renderer, "Agent is thinking")
	spinner.Start()

	// Create event timeline for this request
	timeline := display.NewEventTimeline()

	hasError := false
	var activeToolName string
	toolRunning := false
	requestID := fmt.Sprintf("req_%d", r.config.SessionTokens.GetSummary().RequestCount+1)

	// Run the agent in a goroutine and receive results through a channel
	// This allows us to respond to context cancellation while the agent is thinking
	type eventResult struct {
		event *sessionpkg.Event
		err   error
	}

	eventChan := make(chan eventResult, 1)
	go func() {
		for evt, err := range r.config.Runner.Run(ctx, r.config.UserID, r.config.SessionName, userMsg, agent.RunConfig{
			StreamingMode: agent.StreamingModeNone,
		}) {
			// Send result through channel (non-blocking due to buffer)
			eventChan <- eventResult{evt, err}

			// Check if context was cancelled - if so, stop processing more events
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
		close(eventChan)
	}()

agentLoop:
	for {
		// Check for context cancellation and event arrival at the same level
		// This ensures we respond immediately to Ctrl+C during reasoning
		select {
		case <-ctx.Done():
			spinner.StopWithError("Task interrupted")
			fmt.Printf("\n%s\n", r.config.Renderer.Yellow("⚠️  Task cancelled by user"))
			hasError = true
			break agentLoop
		case result, ok := <-eventChan:
			// Channel closed - agent finished
			if !ok {
				break agentLoop
			}

			// Handle the result
			if result.err != nil {
				spinner.StopWithError("Error occurred")
				errMsg := r.config.Renderer.RenderError(result.err)
				fmt.Print(errMsg)
				hasError = true
				break agentLoop
			}

			if result.event != nil {
				display.PrintEventEnhanced(r.config.Renderer, r.config.StreamingDisplay, result.event, spinner, &activeToolName, &toolRunning, r.config.SessionTokens, requestID, timeline)
			}
		}
	}

	// Stop spinner and show completion
	if !hasError {
		spinner.StopWithSuccess("Task completed")
		completion := r.config.Renderer.RenderTaskComplete()
		fmt.Print(completion)
		r.lastOpStatus = true
	} else {
		failure := r.config.Renderer.RenderTaskFailed()
		fmt.Print(failure)
		r.lastOpStatus = false
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
	summary := r.config.SessionTokens.GetSummary()
	if summary.TotalTokens > 0 {
		metrics := r.config.Renderer.RenderTokenMetrics(
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
	if r.lastOpStatus {
		r.readline.SetPrompt(r.config.Renderer.Green("✓ ") + r.config.Renderer.Cyan(r.config.Renderer.Bold("❯")+" "))
	} else {
		r.readline.SetPrompt(r.config.Renderer.Cyan(r.config.Renderer.Bold("❯") + " "))
	}
}
