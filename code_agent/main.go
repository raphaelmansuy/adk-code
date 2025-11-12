// Code Agent - A CLI coding assistant powered by Google ADK Go
package main

import (
	"context"
	"log"
	"os"

	"code_agent/internal/app"
	clicommands "code_agent/internal/cli/commands"
	"code_agent/internal/config"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment and CLI flags
	cfg, args := config.LoadFromEnv()

	// Handle special commands (new-session, list-sessions, etc.)
	if clicommands.HandleSpecialCommands(ctx, args, &cfg) {
		os.Exit(0)
	}

	// Create and run application
	application, err := app.New(ctx, &cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	application.Run()
}
