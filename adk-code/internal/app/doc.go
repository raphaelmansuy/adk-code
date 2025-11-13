// Package app manages the application lifecycle, including initialization,
// configuration, signal handling, and graceful shutdown.
//
// The Application struct coordinates all components (display, model, agent, session)
// and provides the main entry point for the code agent. It handles signal
// interruption and ensures resources are properly cleaned up on exit.
//
// The application initialization process:
// 1. Loads configuration from environment and CLI flags
// 2. Handles special commands (new-session, list-sessions, etc.)
// 3. Creates and orchestrates all components
// 4. Initializes the interactive REPL
// 5. Starts the main event loop
//
// Example:
//
//	ctx := context.Background()
//	cfg, args := config.LoadFromEnv()
//	if clicommands.HandleSpecialCommands(ctx, args, &cfg) {
//		os.Exit(0)
//	}
//	application, err := app.New(ctx, &cfg)
//	if err != nil {
//		log.Fatalf("Failed to initialize application: %v", err)
//	}
//	application.Run()
package app
