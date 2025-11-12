package app

import (
	"code_agent/internal/orchestration"
)

// GenerateUniqueSessionName is a facade for backward compatibility
func GenerateUniqueSessionName() string {
	return orchestration.GenerateUniqueSessionName()
}
