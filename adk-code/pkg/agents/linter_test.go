package agents

import (
	"testing"
)

func TestDescriptionVaguenessRule(t *testing.T) {
	tests := []struct {
		name        string
		agent       *Agent
		shouldError bool
		ruleID      string
	}{
		{
			name: "vague description starting with A",
			agent: &Agent{
				Name:        "test-agent",
				Description: "A agent that does things",
			},
			shouldError: true,
			ruleID:      "description-vagueness",
		},
		{
			name: "good description",
			agent: &Agent{
				Name:        "test-agent",
				Description: "Reviews code for bugs and security issues.",
			},
			shouldError: false,
		},
		{
			name: "description without period",
			agent: &Agent{
				Name:        "test-agent",
				Description: "Reviews code for bugs and security issues",
			},
			shouldError: true, // Should have info about period
			ruleID:      "description-period",
		},
	}

	linter := NewLinter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := linter.Lint(tt.agent)
			if tt.shouldError && result.ErrorCount == 0 && result.WarningCount == 0 && result.InfoCount == 0 {
				t.Errorf("expected linting issues, got none")
			}
		})
	}
}

func TestDescriptionLengthRule(t *testing.T) {
	tests := []struct {
		name       string
		desc       string
		shouldFail bool
	}{
		{
			name:       "too short",
			desc:       "Short",
			shouldFail: true,
		},
		{
			name:       "minimum valid length",
			desc:       "A valid description.",
			shouldFail: false,
		},
		{
			name:       "too long",
			desc:       string(make([]byte, 1025)),
			shouldFail: true,
		},
		{
			name:       "exactly max length",
			desc:       string(make([]byte, 1024)),
			shouldFail: false,
		},
	}

	linter := NewLinter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Name:        "test-agent",
				Description: tt.desc,
			}
			result := linter.Lint(agent)
			hasLengthError := false
			for _, issue := range result.Issues {
				if issue.Rule == "description-length" {
					hasLengthError = true
					break
				}
			}
			if tt.shouldFail && !hasLengthError {
				t.Errorf("expected description-length error, got none")
			}
			if !tt.shouldFail && hasLengthError {
				t.Errorf("expected no description-length error, got one")
			}
		})
	}
}

func TestNamingConventionRule(t *testing.T) {
	tests := []struct {
		name       string
		agentName  string
		shouldFail bool
	}{
		{
			name:       "valid kebab-case",
			agentName:  "code-reviewer",
			shouldFail: false,
		},
		{
			name:       "single word",
			agentName:  "reviewer",
			shouldFail: false,
		},
		{
			name:       "uppercase invalid",
			agentName:  "CodeReviewer",
			shouldFail: true,
		},
		{
			name:       "underscores invalid",
			agentName:  "code_reviewer",
			shouldFail: true,
		},
		{
			name:       "spaces invalid",
			agentName:  "code reviewer",
			shouldFail: true,
		},
	}

	linter := NewLinter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Name:        tt.agentName,
				Description: "A good description.",
			}
			result := linter.Lint(agent)
			hasNamingError := false
			for _, issue := range result.Issues {
				if issue.Rule == "naming-convention" {
					hasNamingError = true
					break
				}
			}
			if tt.shouldFail && !hasNamingError {
				t.Errorf("expected naming-convention error for %q, got none", tt.agentName)
			}
			if !tt.shouldFail && hasNamingError {
				t.Errorf("expected no naming-convention error for %q, got one", tt.agentName)
			}
		})
	}
}

func TestUnusualNameCharsRule(t *testing.T) {
	tests := []struct {
		name       string
		agentName  string
		shouldFail bool
	}{
		{
			name:       "valid",
			agentName:  "code-reviewer-123",
			shouldFail: false,
		},
		{
			name:       "uppercase invalid",
			agentName:  "Code-Reviewer",
			shouldFail: true,
		},
		{
			name:       "special chars invalid",
			agentName:  "code@reviewer",
			shouldFail: true,
		},
		{
			name:       "underscore invalid",
			agentName:  "code_reviewer",
			shouldFail: true,
		},
	}

	linter := NewLinter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Name:        tt.agentName,
				Description: "A good description.",
			}
			result := linter.Lint(agent)
			hasCharError := false
			for _, issue := range result.Issues {
				if issue.Rule == "unusual-name-chars" {
					hasCharError = true
					break
				}
			}
			if tt.shouldFail && !hasCharError {
				t.Errorf("expected unusual-name-chars error for %q, got none", tt.agentName)
			}
			if !tt.shouldFail && hasCharError {
				t.Errorf("expected no unusual-name-chars error for %q, got one", tt.agentName)
			}
		})
	}
}

func TestVersionFormatRule(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		shouldFail bool
	}{
		{
			name:       "valid semver",
			version:    "1.0.0",
			shouldFail: false,
		},
		{
			name:       "valid with pre-release",
			version:    "1.0.0-alpha",
			shouldFail: false,
		},
		{
			name:       "valid with build",
			version:    "1.0.0+build.1",
			shouldFail: false,
		},
		{
			name:       "empty is allowed",
			version:    "",
			shouldFail: false,
		},
		{
			name:       "invalid format",
			version:    "1.0",
			shouldFail: true,
		},
		{
			name:       "invalid with letters",
			version:    "version-1",
			shouldFail: true,
		},
	}

	linter := NewLinter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Name:        "test-agent",
				Description: "A good description.",
				Version:     tt.version,
			}
			result := linter.Lint(agent)
			hasVersionError := false
			for _, issue := range result.Issues {
				if issue.Rule == "version-format" {
					hasVersionError = true
					break
				}
			}
			if tt.shouldFail && !hasVersionError {
				t.Errorf("expected version-format error for %q, got none", tt.version)
			}
			if !tt.shouldFail && hasVersionError {
				t.Errorf("expected no version-format error for %q, got one", tt.version)
			}
		})
	}
}

func TestEmptyTagsRule(t *testing.T) {
	tests := []struct {
		name           string
		tags           []string
		shouldHaveInfo bool
	}{
		{
			name:           "no tags",
			tags:           []string{},
			shouldHaveInfo: true,
		},
		{
			name:           "with tags",
			tags:           []string{"code-review", "quality"},
			shouldHaveInfo: false,
		},
	}

	linter := NewLinter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Name:        "test-agent",
				Description: "A good description.",
				Tags:        tt.tags,
			}
			result := linter.Lint(agent)
			hasTagsInfo := false
			for _, issue := range result.Issues {
				if issue.Rule == "empty-tags" {
					hasTagsInfo = true
					break
				}
			}
			if tt.shouldHaveInfo && !hasTagsInfo {
				t.Errorf("expected empty-tags info, got none")
			}
			if !tt.shouldHaveInfo && hasTagsInfo {
				t.Errorf("expected no empty-tags info, got one")
			}
		})
	}
}

func TestMissingAuthorRule(t *testing.T) {
	tests := []struct {
		name           string
		author         string
		shouldHaveInfo bool
	}{
		{
			name:           "no author",
			author:         "",
			shouldHaveInfo: true,
		},
		{
			name:           "with author",
			author:         "john@example.com",
			shouldHaveInfo: false,
		},
	}

	linter := NewLinter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Name:        "test-agent",
				Description: "A good description.",
				Author:      tt.author,
			}
			result := linter.Lint(agent)
			hasAuthorInfo := false
			for _, issue := range result.Issues {
				if issue.Rule == "missing-author" {
					hasAuthorInfo = true
					break
				}
			}
			if tt.shouldHaveInfo && !hasAuthorInfo {
				t.Errorf("expected missing-author info, got none")
			}
			if !tt.shouldHaveInfo && hasAuthorInfo {
				t.Errorf("expected no missing-author info, got one")
			}
		})
	}
}

func TestMissingVersionRule(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		shouldHaveInfo bool
	}{
		{
			name:           "no version",
			version:        "",
			shouldHaveInfo: true,
		},
		{
			name:           "with version",
			version:        "1.0.0",
			shouldHaveInfo: false,
		},
	}

	linter := NewLinter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Name:        "test-agent",
				Description: "A good description.",
				Version:     tt.version,
			}
			result := linter.Lint(agent)
			hasVersionInfo := false
			for _, issue := range result.Issues {
				if issue.Rule == "missing-version" {
					hasVersionInfo = true
					break
				}
			}
			if tt.shouldHaveInfo && !hasVersionInfo {
				t.Errorf("expected missing-version info, got none")
			}
			if !tt.shouldHaveInfo && hasVersionInfo {
				t.Errorf("expected no missing-version info, got one")
			}
		})
	}
}

func TestLintResult(t *testing.T) {
	agent := &Agent{
		Name:        "bad-agent",
		Description: "short",
		Author:      "",
		Version:     "invalid",
	}

	linter := NewLinter()
	result := linter.Lint(agent)

	if result.AgentName != "bad-agent" {
		t.Errorf("expected agent name 'bad-agent', got %q", result.AgentName)
	}

	if result.Passed {
		t.Errorf("expected Passed=false for agent with errors")
	}

	if result.ErrorCount == 0 {
		t.Errorf("expected error count > 0, got %d", result.ErrorCount)
	}
}

func TestLintAll(t *testing.T) {
	agents := []*Agent{
		{
			Name:        "good-agent",
			Description: "A good agent description.",
			Version:     "1.0.0",
		},
		{
			Name:        "bad-agent",
			Description: "short",
		},
	}

	linter := NewLinter()
	results := linter.LintAll(agents)

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	goodResult, exists := results["good-agent"]
	if !exists {
		t.Fatal("expected result for good-agent")
	}
	if !goodResult.Passed {
		t.Errorf("expected good-agent to pass linting")
	}

	badResult, exists := results["bad-agent"]
	if !exists {
		t.Fatal("expected result for bad-agent")
	}
	if badResult.Passed {
		t.Errorf("expected bad-agent to fail linting")
	}
}

// Helper function tests
func TestIsKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"code-reviewer", true},
		{"agent", true},
		{"my-awesome-agent", true},
		{"123-agent", true},
		{"CodeReviewer", false},
		{"code_reviewer", false},
		{"code reviewer", false},
		{"-agent", false},
		{"agent-", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isKebabCase(tt.input)
		if result != tt.expected {
			t.Errorf("isKebabCase(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CodeReviewer", "code-reviewer"},
		{"code_reviewer", "code-reviewer"},
		{"code reviewer", "code-reviewer"},
		{"Code Reviewer", "code-reviewer"},
		{"codeReviewer", "code-reviewer"},
	}

	for _, tt := range tests {
		result := toKebabCase(tt.input)
		if result != tt.expected {
			t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@example.co.uk", true},
		{"invalid", false},
		{"@example.com", false},
		{"test@", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isValidEmail(tt.email)
		if result != tt.expected {
			t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, result, tt.expected)
		}
	}
}

func TestIsValidSemanticVersion(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		{"1.0.0", true},
		{"1.2.3", true},
		{"0.0.1", true},
		{"1.0.0-alpha", true},
		{"1.0.0-beta.1", true},
		{"1.0.0+build.1", true},
		{"1.0", false},
		{"1", false},
		{"version-1", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isValidSemanticVersion(tt.version)
		if result != tt.expected {
			t.Errorf("isValidSemanticVersion(%q) = %v, want %v", tt.version, result, tt.expected)
		}
	}
}
