package app

import (
	"fmt"
	"time"
)

// GenerateUniqueSessionName creates a unique session name based on timestamp
// Format: session-YYYYMMDD-HHMMSS (e.g., session-20251110-221530)
func GenerateUniqueSessionName() string {
	now := time.Now()
	return fmt.Sprintf("session-%d%02d%02d-%02d%02d%02d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second())
}
