// Package commands provides CLI command handlers organized by functionality.
package commands

import (
	"context"
	"fmt"
	"log"
	"os"

	"adk-code/internal/config"
	"adk-code/internal/session"
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
		HandleNewSession(ctx, args[1], cfg.DBPath)
		return true

	case "list-sessions":
		HandleListSessions(ctx, cfg.DBPath)
		return true

	case "delete-session":
		if len(args) < 2 {
			fmt.Println("Usage: code-agent delete-session <session-name>")
			os.Exit(1)
		}
		HandleDeleteSession(ctx, args[1], cfg.DBPath)
		return true

	default:
		return false
	}
}

// HandleNewSession creates a new session
func HandleNewSession(ctx context.Context, sessionName string, dbPath string) {
	manager, err := session.NewSessionManager("code_agent", dbPath)
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

// HandleListSessions lists all sessions
func HandleListSessions(ctx context.Context, dbPath string) {
	manager, err := session.NewSessionManager("code_agent", dbPath)
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

// HandleDeleteSession deletes a session
func HandleDeleteSession(ctx context.Context, sessionName string, dbPath string) {
	manager, err := session.NewSessionManager("code_agent", dbPath)
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
