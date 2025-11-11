// Package main - Utility functions
package main

import (
	"fmt"
	"time"
)

// generateUniqueSessionName creates a unique session name based on timestamp
// Format: session-YYYYMMDD-HHMMSS (e.g., session-20251110-221530)
func generateUniqueSessionName() string {
	now := time.Now()
	return fmt.Sprintf("session-%d%02d%02d-%02d%02d%02d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second())
}
