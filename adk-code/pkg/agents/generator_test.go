package agents

import (
	"os"
	"strings"
	"testing"
)

func TestNewAgentGenerator(t *testing.T) {
	gen := NewAgentGenerator()
	if gen == nil {
		t.Fatal("expected generator, got nil")
	}

	if gen.templates == nil {
		t.Error("expected templates map, got nil")
	}
}

func TestGetAvailableTemplates(t *testing.T) {
	templates := GetAvailableTemplates()
	if len(templates) == 0 {
		t.Fatal("expected available templates")
	}

	expectedCount := 3 // subagent, skill, command
	if len(templates) != expectedCount {
		t.Errorf("expected %d templates, got %d", expectedCount, len(templates))
	}
}

func TestGenerateAgentSubagent(t *testing.T) {
	gen := NewAgentGenerator()

	input := AgentGeneratorInput{
		Name:         "test-agent",
		Description:  "A test agent for unit testing.",
		TemplateType: TemplateSubagent,
		Author:       "test@example.com",
		Version:      "1.0.0",
		Tags:         []string{"test", "example"},
	}

	agent, err := gen.GenerateAgent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agent.Name != "test-agent" {
		t.Errorf("expected name 'test-agent', got %q", agent.Name)
	}

	if agent.Description != "A test agent for unit testing." {
		t.Errorf("expected description 'A test agent for unit testing.', got %q", agent.Description)
	}

	if agent.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", agent.Version)
	}

	if agent.Author != "test@example.com" {
		t.Errorf("expected author 'test@example.com', got %q", agent.Author)
	}

	if len(agent.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(agent.Tags))
	}

	if agent.RawYAML == "" {
		t.Error("expected RawYAML to be set")
	}

	if agent.Content == "" {
		t.Error("expected Content to be set")
	}
}

func TestGenerateAgentMissingName(t *testing.T) {
	gen := NewAgentGenerator()

	input := AgentGeneratorInput{
		Description: "A test agent.",
	}

	_, err := gen.GenerateAgent(input)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestGenerateAgentMissingDescription(t *testing.T) {
	gen := NewAgentGenerator()

	input := AgentGeneratorInput{
		Name: "test-agent",
	}

	_, err := gen.GenerateAgent(input)
	if err == nil {
		t.Error("expected error for missing description")
	}
}

func TestGenerateAgentShortDescription(t *testing.T) {
	gen := NewAgentGenerator()

	input := AgentGeneratorInput{
		Name:        "test-agent",
		Description: "Short",
	}

	_, err := gen.GenerateAgent(input)
	if err == nil {
		t.Error("expected error for short description")
	}
}

func TestGenerateAgentInvalidName(t *testing.T) {
	gen := NewAgentGenerator()

	input := AgentGeneratorInput{
		Name:        "TestAgent", // Not kebab-case
		Description: "A test agent for unit testing.",
	}

	_, err := gen.GenerateAgent(input)
	if err == nil {
		t.Error("expected error for invalid name format")
	}
}

func TestGenerateAgentDefaultVersion(t *testing.T) {
	gen := NewAgentGenerator()

	input := AgentGeneratorInput{
		Name:        "test-agent",
		Description: "A test agent for unit testing.",
	}

	agent, err := gen.GenerateAgent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agent.Version != "1.0.0" {
		t.Errorf("expected default version '1.0.0', got %q", agent.Version)
	}
}

func TestGenerateAgentFrontmatter(t *testing.T) {
	input := AgentGeneratorInput{
		Name:        "test-agent",
		Description: "Test description.",
		Version:     "1.0.0",
		Author:      "author@example.com",
		Tags:        []string{"tag1", "tag2"},
	}

	frontmatter := generateFrontmatter(input)

	if !strings.Contains(frontmatter, "---") {
		t.Error("frontmatter should start and end with ---")
	}

	if !strings.Contains(frontmatter, "name: test-agent") {
		t.Error("frontmatter should contain name")
	}

	if !strings.Contains(frontmatter, "description: Test description.") {
		t.Error("frontmatter should contain description")
	}

	if !strings.Contains(frontmatter, "version: 1.0.0") {
		t.Error("frontmatter should contain version")
	}

	if !strings.Contains(frontmatter, "author: author@example.com") {
		t.Error("frontmatter should contain author")
	}

	if !strings.Contains(frontmatter, "- tag1") {
		t.Error("frontmatter should contain tags")
	}
}

func TestGenerateAgentTemplates(t *testing.T) {
	gen := NewAgentGenerator()

	templates := []TemplateType{TemplateSubagent, TemplateSkill, TemplateCommand}
	for _, tmpl := range templates {
		content, err := gen.Template(tmpl)
		if err != nil {
			t.Errorf("failed to get %s template: %v", tmpl, err)
		}

		if content == "" {
			t.Errorf("template for %s is empty", tmpl)
		}
	}
}

func TestGenerateAgentUnknownTemplate(t *testing.T) {
	gen := NewAgentGenerator()

	input := AgentGeneratorInput{
		Name:         "test-agent",
		Description:  "A test agent for unit testing.",
		TemplateType: TemplateType("unknown"),
	}

	_, err := gen.GenerateAgent(input)
	if err == nil {
		t.Error("expected error for unknown template type")
	}
}

func TestWriteAgent(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "adk-agent-test-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gen := NewAgentGenerator()

	input := AgentGeneratorInput{
		Name:        "test-agent",
		Description: "A test agent for unit testing.",
	}

	agent, err := gen.GenerateAgent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filePath, err := gen.WriteAgent(agent, tmpDir)
	if err != nil {
		t.Fatalf("failed to write agent: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("agent file was not created: %v", err)
	}

	// Check file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read agent file: %v", err)
	}

	if !strings.Contains(string(content), "---") {
		t.Error("agent file should contain YAML frontmatter")
	}

	if !strings.Contains(string(content), "name: test-agent") {
		t.Error("agent file should contain agent name")
	}
}

func TestWriteAgentExistsError(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "adk-agent-test-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gen := NewAgentGenerator()

	input := AgentGeneratorInput{
		Name:        "test-agent",
		Description: "A test agent for unit testing.",
	}

	agent, err := gen.GenerateAgent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Write agent once
	_, err = gen.WriteAgent(agent, tmpDir)
	if err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// Try to write again - should fail
	_, err = gen.WriteAgent(agent, tmpDir)
	if err == nil {
		t.Error("expected error when writing agent that already exists")
	}
}

func TestCustomizeTemplate(t *testing.T) {
	gen := NewAgentGenerator()

	newContent := "# Custom Template\n\nCustom content"
	err := gen.CustomizeTemplate(TemplateSubagent, newContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := gen.Template(TemplateSubagent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved != newContent {
		t.Errorf("customized template not returned correctly")
	}
}
