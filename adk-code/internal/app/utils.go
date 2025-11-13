package app

import (
	"adk-code/internal/orchestration"
)

// GenerateUniqueSessionName is a facade for backward compatibility
func GenerateUniqueSessionName() string {
	return orchestration.GenerateUniqueSessionName()
}
