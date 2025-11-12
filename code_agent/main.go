// Code Agent - A CLI coding assistant powered by Google ADK Go
package main

import (
	"context"
	"log"
	"os"

	"code_agent/internal/app"
	"code_agent/pkg/cli"
)

func main() {
	ctx := context.Background()

	// Parse command-line flags
	cliConfig, args := cli.ParseCLIFlags()

	// Handle special commands (new-session, list-sessions, etc.)
	if cli.HandleCLICommands(ctx, args, cliConfig.DBPath) {
		os.Exit(0)
	}

	// Create and run application
	application, err := app.New(ctx, &cliConfig)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	application.Run()
}
