package app

import (
	"regexp"
	"testing"
)

// TestGenerateUniqueSessionNameFormat ensures the generated session name follows the expected pattern
func TestGenerateUniqueSessionNameFormat(t *testing.T) {
	name := GenerateUniqueSessionName()
	matched := regexp.MustCompile(`^session-[0-9]{8}-[0-9]{6}$`).MatchString(name)
	if !matched {
		t.Fatalf("unexpected session name format: %s", name)
	}
}
