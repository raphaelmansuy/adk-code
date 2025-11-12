// Package commands - Command handlers for CLI special commands
package commands

import (
	"context"
	"fmt"
	"os"

	clicommands "code_agent/internal/cli/commands"
	"code_agent/internal/config"
)

// HandleSpecialCommands processes special CLI commands (new-session, list-sessions, etc.)
// Returns true if a command was handled (and program should exit)
func HandleSpecialCommands(ctx context.Context, args []string, cfg *config.Config) bool {
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
		clicommands.HandleNewSession(ctx, args[1], cfg.DBPath)
		return true

	case "list-sessions":
		clicommands.HandleListSessions(ctx, cfg.DBPath)
		return true

	case "delete-session":
		if len(args) < 2 {
			fmt.Println("Usage: code-agent delete-session <session-name>")
			os.Exit(1)
		}
		clicommands.HandleDeleteSession(ctx, args[1], cfg.DBPath)
		return true

	default:
		return false
	}
}
