// Package main - CLI flag parsing and command handling
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	codingagent "code_agent/agent"
	"code_agent/display"
	"code_agent/tracking"
)

// CLIConfig holds parsed command-line flags
type CLIConfig struct {
	OutputFormat      string
	TypewriterEnabled bool
	SessionName       string
	DBPath            string
	WorkingDirectory  string
}

// ParseCLIFlags parses command-line arguments and returns config and remaining args
func ParseCLIFlags() (CLIConfig, []string) {
	outputFormat := flag.String("output-format", "rich", "Output format: rich, plain, or json")
	typewriterEnabled := flag.Bool("typewriter", false, "Enable typewriter effect for text output")
	sessionName := flag.String("session", "", "Session name (optional, defaults to 'default')")
	dbPath := flag.String("db", "", "Database path for sessions (optional, defaults to ~/.code_agent/sessions.db)")
	workingDirectory := flag.String("working-directory", "", "Working directory for the agent (optional, defaults to current directory)")
	flag.Parse()

	return CLIConfig{
		OutputFormat:      *outputFormat,
		TypewriterEnabled: *typewriterEnabled,
		SessionName:       *sessionName,
		DBPath:            *dbPath,
		WorkingDirectory:  *workingDirectory,
	}, flag.Args()
}

// HandleCLICommands processes special CLI commands (new-session, list-sessions, etc.)
// Returns true if a command was handled (and program should exit)
func HandleCLICommands(ctx context.Context, args []string, dbPath string) bool {
	if len(args) == 0 {
		return false
	}

	cmd := args[0]

	switch cmd {
	case "new-session":
		if len(args) < 2 {
			fmt.Println("Usage: code-agent new-session <session-name>")
			os.Exit(1)
		}
		handleNewSession(ctx, args[1], dbPath)
		return true

	case "list-sessions":
		handleListSessions(ctx, dbPath)
		return true

	case "delete-session":
		if len(args) < 2 {
			fmt.Println("Usage: code-agent delete-session <session-name>")
			os.Exit(1)
		}
		handleDeleteSession(ctx, args[1], dbPath)
		return true

	default:
		return false
	}
}

// handleBuiltinCommand handles built-in REPL commands like /help, /tools, etc.
// Returns true if a command was handled, false if input should be sent to agent
// Note: /exit and /quit are handled separately in main.go to break the loop
func handleBuiltinCommand(input string, renderer *display.Renderer, sessionTokens *tracking.SessionTokens) bool {
	switch input {
	case "/prompt":
		fmt.Print(renderer.Yellow("\n=== System Prompt ===\n\n"))
		fmt.Print(renderer.Dim(codingagent.EnhancedSystemPrompt))
		fmt.Print(renderer.Yellow("\n\n=== End of Prompt ===\n\n"))
		return true

	case "/help":
		printHelpMessage(renderer)
		return true

	case "/tools":
		printToolsList(renderer)
		return true

	case "/tokens":
		summary := sessionTokens.GetSummary()
		fmt.Print(tracking.FormatSessionSummary(summary))
		return true

	default:
		return false
	}
}

// printHelpMessage displays the help message
func printHelpMessage(renderer *display.Renderer) {
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
}

// printToolsList displays the available tools
func printToolsList(renderer *display.Renderer) {
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
}
