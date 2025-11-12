// Package cli - CLI command dispatchers
package cli

import (
	"context"
	"fmt"
	"os"

	"code_agent/display"
	"code_agent/pkg/cli/commands"
	"code_agent/pkg/models"
	"code_agent/tracking"
)

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
		commands.HandleNewSession(ctx, args[1], dbPath)
		return true

	case "list-sessions":
		commands.HandleListSessions(ctx, dbPath)
		return true

	case "delete-session":
		if len(args) < 2 {
			fmt.Println("Usage: code-agent delete-session <session-name>")
			os.Exit(1)
		}
		commands.HandleDeleteSession(ctx, args[1], dbPath)
		return true

	default:
		return false
	}
}

// HandleBuiltinCommand handles built-in REPL commands like /help, /tools, etc.
// Returns true if a command was handled, false if input should be sent to agent
// Note: /exit and /quit are handled separately in repl.go to break the loop
func HandleBuiltinCommand(input string, renderer *display.Renderer, sessionTokens *tracking.SessionTokens, modelRegistry *models.Registry, currentModel models.Config) bool {
	return commands.HandleBuiltinCommand(input, renderer, sessionTokens, modelRegistry, currentModel)
}
