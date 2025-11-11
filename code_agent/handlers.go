// Package main - Session management handlers
package main

import (
	"context"
	"fmt"
	"log"

	"code_agent/persistence"
)

// handleNewSession creates a new session
func handleNewSession(ctx context.Context, sessionName string, dbPath string) {
	manager, err := persistence.NewSessionManager("code_agent", dbPath)
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

// handleListSessions lists all sessions
func handleListSessions(ctx context.Context, dbPath string) {
	manager, err := persistence.NewSessionManager("code_agent", dbPath)
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

// handleDeleteSession deletes a session
func handleDeleteSession(ctx context.Context, sessionName string, dbPath string) {
	manager, err := persistence.NewSessionManager("code_agent", dbPath)
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
