package agents

import (
	"fmt"
	"regexp"
	"strings"
)

// LintSeverity represents the severity level of a linting issue
type LintSeverity string

const (
	SeverityError   LintSeverity = "error"
	SeverityWarning LintSeverity = "warning"
	SeverityInfo    LintSeverity = "info"
)

// LintIssue represents a single linting violation
type LintIssue struct {
	Rule       string       // Rule identifier (e.g., "description-too-vague")
	Severity   LintSeverity // Error, warning, or info
	Message    string       // Human-readable message
	Field      string       // Field name where issue occurs
	Suggestion string       // Optional fix suggestion
	Line       int          // Approximate line number (if known)
}

// LintResult holds all linting results for an agent
type LintResult struct {
	AgentName    string
	Passed       bool // True if no errors (warnings/info are allowed)
	Issues       []LintIssue
	ErrorCount   int
	WarningCount int
	InfoCount    int
}

// Linter performs best practices checking on agents
type Linter struct {
	rules []LintRule
}

// LintRule is an interface for custom linting rules
type LintRule interface {
	// ID returns the unique identifier for this rule
	ID() string

	// Description returns a human-readable description
	Description() string

	// Severity returns the severity level (error, warning, info)
	Severity() LintSeverity

	// Check examines an agent and returns any violations
	Check(agent *Agent) []LintIssue
}

// NewLinter creates a new linter with all built-in rules
func NewLinter() *Linter {
	return &Linter{
		rules: []LintRule{
			&DescriptionVaguenessRule{},
			&DescriptionLengthRule{},
			&NamingConventionRule{},
			&AuthorFormatRule{},
			&VersionFormatRule{},
			&EmptyTagsRule{},
			&UnusualNameCharsRule{},
			&MissingAuthorRule{},
			&MissingVersionRule{},
			&CircularDependencyRule{},
			&DependencyDoesNotExistRule{},
		},
	}
}

// Lint performs all linting checks on an agent
func (l *Linter) Lint(agent *Agent) *LintResult {
	result := &LintResult{
		AgentName: agent.Name,
		Issues:    make([]LintIssue, 0),
		Passed:    true,
	}

	for _, rule := range l.rules {
		issues := rule.Check(agent)
		for _, issue := range issues {
			result.Issues = append(result.Issues, issue)

			switch issue.Severity {
			case SeverityError:
				result.ErrorCount++
				result.Passed = false
			case SeverityWarning:
				result.WarningCount++
			case SeverityInfo:
				result.InfoCount++
			}
		}
	}

	return result
}

// LintAll performs linting on multiple agents
func (l *Linter) LintAll(agents []*Agent) map[string]*LintResult {
	results := make(map[string]*LintResult)
	for _, agent := range agents {
		results[agent.Name] = l.Lint(agent)
	}
	return results
}

// ------- Built-in Rules -------

// DescriptionVaguenessRule checks if description is too vague or weak
type DescriptionVaguenessRule struct{}

func (r *DescriptionVaguenessRule) ID() string {
	return "description-vagueness"
}

func (r *DescriptionVaguenessRule) Description() string {
	return "Description should be specific and action-oriented"
}

func (r *DescriptionVaguenessRule) Severity() LintSeverity {
	return SeverityWarning
}

func (r *DescriptionVaguenessRule) Check(agent *Agent) []LintIssue {
	var issues []LintIssue

	vaguePatterns := []struct {
		pattern string
		reason  string
	}{
		{"^[Aa] .*agent", "Avoid starting with 'A' or 'An'"},
		{"^[Aa]gent that", "Use action verbs instead of 'that'"},
		{"^[Tt]his agent", "Avoid 'this agent' - be more specific"},
		{"^[Hh]elps with", "Use stronger action verbs than 'helps with'"},
		{"^[Mm]anages", "Be more specific about what is managed"},
		{"^[Dd]oes various", "Avoid vague descriptions with 'various'"},
		{"^[Hh]andles stuff", "Be specific about what is handled"},
	}

	for _, vague := range vaguePatterns {
		if matched, _ := regexp.MatchString(vague.pattern, agent.Description); matched {
			issues = append(issues, LintIssue{
				Rule:       r.ID(),
				Severity:   r.Severity(),
				Message:    fmt.Sprintf("Description is too vague: %s", vague.reason),
				Field:      "description",
				Suggestion: "Rewrite description to be more specific and action-oriented. Example: 'Reviews code for bugs, performance issues, and security vulnerabilities'",
			})
			break
		}
	}

	// Check if description ends with period
	if len(agent.Description) > 0 && !strings.HasSuffix(agent.Description, ".") {
		issues = append(issues, LintIssue{
			Rule:       "description-period",
			Severity:   SeverityInfo,
			Message:    "Description should end with a period",
			Field:      "description",
			Suggestion: fmt.Sprintf("Change: %q", agent.Description+"."),
		})
	}

	return issues
}

// DescriptionLengthRule checks description length constraints
type DescriptionLengthRule struct{}

func (r *DescriptionLengthRule) ID() string {
	return "description-length"
}

func (r *DescriptionLengthRule) Description() string {
	return "Description must be 10-1024 characters"
}

func (r *DescriptionLengthRule) Severity() LintSeverity {
	return SeverityError
}

func (r *DescriptionLengthRule) Check(agent *Agent) []LintIssue {
	var issues []LintIssue

	if len(agent.Description) < 10 {
		issues = append(issues, LintIssue{
			Rule:       r.ID(),
			Severity:   r.Severity(),
			Message:    fmt.Sprintf("Description too short (%d chars, min 10)", len(agent.Description)),
			Field:      "description",
			Suggestion: "Provide a more detailed description of the agent's purpose",
		})
	}

	if len(agent.Description) > 1024 {
		issues = append(issues, LintIssue{
			Rule:       r.ID(),
			Severity:   r.Severity(),
			Message:    fmt.Sprintf("Description too long (%d chars, max 1024)", len(agent.Description)),
			Field:      "description",
			Suggestion: "Condense the description to be more concise",
		})
	}

	return issues
}

// NamingConventionRule checks agent naming conventions
type NamingConventionRule struct{}

func (r *NamingConventionRule) ID() string {
	return "naming-convention"
}

func (r *NamingConventionRule) Description() string {
	return "Agent names should follow kebab-case convention"
}

func (r *NamingConventionRule) Severity() LintSeverity {
	return SeverityWarning
}

func (r *NamingConventionRule) Check(agent *Agent) []LintIssue {
	var issues []LintIssue

	if !isKebabCase(agent.Name) {
		issues = append(issues, LintIssue{
			Rule:       r.ID(),
			Severity:   r.Severity(),
			Message:    "Agent name should use kebab-case (lowercase with hyphens)",
			Field:      "name",
			Suggestion: fmt.Sprintf("Use: %s", toKebabCase(agent.Name)),
		})
	}

	return issues
}

// AuthorFormatRule checks author field format
type AuthorFormatRule struct{}

func (r *AuthorFormatRule) ID() string {
	return "author-format"
}

func (r *AuthorFormatRule) Description() string {
	return "Author field should be properly formatted"
}

func (r *AuthorFormatRule) Severity() LintSeverity {
	return SeverityInfo
}

func (r *AuthorFormatRule) Check(agent *Agent) []LintIssue {
	var issues []LintIssue

	if agent.Author != "" {
		// Check if looks like email
		if !isValidEmail(agent.Author) && !isValidName(agent.Author) {
			issues = append(issues, LintIssue{
				Rule:       r.ID(),
				Severity:   r.Severity(),
				Message:    "Author should be an email or display name",
				Field:      "author",
				Suggestion: "Provide author as 'name@example.com' or 'John Doe'",
			})
		}
	}

	return issues
}

// VersionFormatRule checks semantic version format
type VersionFormatRule struct{}

func (r *VersionFormatRule) ID() string {
	return "version-format"
}

func (r *VersionFormatRule) Description() string {
	return "Version should follow semantic versioning"
}

func (r *VersionFormatRule) Severity() LintSeverity {
	return SeverityWarning
}

func (r *VersionFormatRule) Check(agent *Agent) []LintIssue {
	var issues []LintIssue

	if agent.Version != "" && !isValidSemanticVersion(agent.Version) {
		issues = append(issues, LintIssue{
			Rule:       r.ID(),
			Severity:   r.Severity(),
			Message:    fmt.Sprintf("Version %q does not follow semantic versioning", agent.Version),
			Field:      "version",
			Suggestion: "Use format: major.minor.patch (e.g., 1.0.0, 1.2.3)",
		})
	}

	return issues
}

// EmptyTagsRule checks for empty tags array
type EmptyTagsRule struct{}

func (r *EmptyTagsRule) ID() string {
	return "empty-tags"
}

func (r *EmptyTagsRule) Description() string {
	return "Agent should have at least one tag for discoverability"
}

func (r *EmptyTagsRule) Severity() LintSeverity {
	return SeverityInfo
}

func (r *EmptyTagsRule) Check(agent *Agent) []LintIssue {
	var issues []LintIssue

	if len(agent.Tags) == 0 {
		issues = append(issues, LintIssue{
			Rule:       r.ID(),
			Severity:   r.Severity(),
			Message:    "Agent has no tags",
			Field:      "tags",
			Suggestion: "Add relevant tags for discoverability (e.g., ['code-review', 'quality-assurance'])",
		})
	}

	return issues
}

// UnusualNameCharsRule checks for unusual characters in names
type UnusualNameCharsRule struct{}

func (r *UnusualNameCharsRule) ID() string {
	return "unusual-name-chars"
}

func (r *UnusualNameCharsRule) Description() string {
	return "Agent names should only contain lowercase letters, numbers, and hyphens"
}

func (r *UnusualNameCharsRule) Severity() LintSeverity {
	return SeverityError
}

func (r *UnusualNameCharsRule) Check(agent *Agent) []LintIssue {
	var issues []LintIssue

	validChars := regexp.MustCompile(`^[a-z0-9\-]+$`)
	if !validChars.MatchString(agent.Name) {
		issues = append(issues, LintIssue{
			Rule:       r.ID(),
			Severity:   r.Severity(),
			Message:    "Agent name contains invalid characters",
			Field:      "name",
			Suggestion: "Use only lowercase letters, numbers, and hyphens",
		})
	}

	return issues
}

// MissingAuthorRule checks for missing author
type MissingAuthorRule struct{}

func (r *MissingAuthorRule) ID() string {
	return "missing-author"
}

func (r *MissingAuthorRule) Description() string {
	return "Agent should specify an author"
}

func (r *MissingAuthorRule) Severity() LintSeverity {
	return SeverityInfo
}

func (r *MissingAuthorRule) Check(agent *Agent) []LintIssue {
	var issues []LintIssue

	if agent.Author == "" {
		issues = append(issues, LintIssue{
			Rule:       r.ID(),
			Severity:   r.Severity(),
			Message:    "Author field is not specified",
			Field:      "author",
			Suggestion: "Add author field to credit the creator",
		})
	}

	return issues
}

// MissingVersionRule checks for missing version
type MissingVersionRule struct{}

func (r *MissingVersionRule) ID() string {
	return "missing-version"
}

func (r *MissingVersionRule) Description() string {
	return "Agent should specify a version"
}

func (r *MissingVersionRule) Severity() LintSeverity {
	return SeverityInfo
}

func (r *MissingVersionRule) Check(agent *Agent) []LintIssue {
	var issues []LintIssue

	if agent.Version == "" {
		issues = append(issues, LintIssue{
			Rule:       r.ID(),
			Severity:   r.Severity(),
			Message:    "Version field is not specified",
			Field:      "version",
			Suggestion: "Add version field starting with 1.0.0",
		})
	}

	return issues
}

// CircularDependencyRule checks for circular dependencies (requires DependencyGraph)
type CircularDependencyRule struct{}

func (r *CircularDependencyRule) ID() string {
	return "circular-dependency"
}

func (r *CircularDependencyRule) Description() string {
	return "Agent should not have circular dependencies"
}

func (r *CircularDependencyRule) Severity() LintSeverity {
	return SeverityError
}

func (r *CircularDependencyRule) Check(agent *Agent) []LintIssue {
	// This rule requires external context (dependency graph)
	// For now, return no issues - should be integrated with DependencyGraph
	return []LintIssue{}
}

// DependencyDoesNotExistRule checks if dependencies are valid
type DependencyDoesNotExistRule struct{}

func (r *DependencyDoesNotExistRule) ID() string {
	return "invalid-dependency"
}

func (r *DependencyDoesNotExistRule) Description() string {
	return "All dependencies should reference existing agents"
}

func (r *DependencyDoesNotExistRule) Severity() LintSeverity {
	return SeverityError
}

func (r *DependencyDoesNotExistRule) Check(agent *Agent) []LintIssue {
	// This rule requires external context (agent registry)
	// For now, return no issues - should be integrated with discovery
	return []LintIssue{}
}

// ------- Helper Functions -------

// isKebabCase checks if a string follows kebab-case convention
func isKebabCase(s string) bool {
	// Must be lowercase letters, numbers, and hyphens
	// Cannot start or end with hyphen
	if len(s) == 0 {
		return false
	}
	if s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	return regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(s)
}

// toKebabCase converts a string to kebab-case
func toKebabCase(s string) string {
	// Insert hyphens before uppercase letters (for camelCase)
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' && s[i-1] != '-' && s[i-1] != ' ' && s[i-1] != '_' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	s = result.String()

	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces, underscores with hyphens
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")

	// Remove multiple consecutive hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")

	return s
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}

// isValidName checks if a string is a valid display name
func isValidName(name string) bool {
	if len(name) < 2 || len(name) > 100 {
		return false
	}
	// Allow letters, spaces, hyphens, apostrophes
	return regexp.MustCompile(`^[a-zA-Z\s\-']+$`).MatchString(name)
}

// isValidSemanticVersion checks if string follows semantic versioning
func isValidSemanticVersion(version string) bool {
	return regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`).MatchString(version)
}
