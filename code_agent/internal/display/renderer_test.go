package display

import (
	"errors"
	"strings"
	"testing"
)

func TestNewRenderer_Plain(t *testing.T) {
	r, err := NewRenderer(OutputFormatPlain)
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
}

func TestNewRenderer_Rich(t *testing.T) {
	r, err := NewRenderer(OutputFormatRich)
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
}

func TestRenderer_Bold_Plain(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.Bold("test")
	if !strings.Contains(result, "test") {
		t.Fatalf("expected 'test' in result, got: %s", result)
	}
}

func TestRenderer_Dim(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.Dim("dimmed text")
	if !strings.Contains(result, "dimmed text") {
		t.Fatalf("expected 'dimmed text' in result, got: %s", result)
	}
}

func TestRenderer_Red_PlainFormat(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.Red("error")
	if result != "error" {
		t.Fatalf("expected 'error' in plain format, got: %s", result)
	}
}

func TestRenderer_Green_PlainFormat(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.Green("success")
	if result != "success" {
		t.Fatalf("expected 'success' in plain format, got: %s", result)
	}
}

func TestRenderer_Yellow_PlainFormat(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.Yellow("warning")
	if result != "warning" {
		t.Fatalf("expected 'warning' in plain format, got: %s", result)
	}
}

func TestRenderer_Blue_PlainFormat(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.Blue("info")
	if result != "info" {
		t.Fatalf("expected 'info' in plain format, got: %s", result)
	}
}

func TestRenderer_Cyan_PlainFormat(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.Cyan("info")
	if result != "info" {
		t.Fatalf("expected 'info' in plain format, got: %s", result)
	}
}

func TestRenderer_SuccessCheckmark(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.SuccessCheckmark("done")
	if !strings.Contains(result, "done") {
		t.Fatalf("expected 'done' in result, got: %s", result)
	}
}

func TestRenderer_ErrorX(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.ErrorX("failed")
	if !strings.Contains(result, "failed") {
		t.Fatalf("expected 'failed' in result, got: %s", result)
	}
}

func TestRenderer_RenderError(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	err := errors.New("test error")
	result := r.RenderError(err)
	if !strings.Contains(result, "test error") {
		t.Fatalf("expected error text in result, got: %s", result)
	}
}

func TestRenderer_RenderWarning(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.RenderWarning("test warning")
	if !strings.Contains(result, "test warning") && !strings.Contains(result, "warning") {
		t.Fatalf("expected warning in result, got: %s", result)
	}
}

func TestRenderer_RenderInfo(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.RenderInfo("test info")
	if result == "" {
		t.Fatal("expected non-empty info message")
	}
}

func TestRenderer_RenderBanner(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	banner := r.RenderBanner("1.0.0", "test-model", "/test/dir")
	if banner == "" {
		t.Fatal("expected non-empty banner")
	}
	if !strings.Contains(banner, "1.0.0") {
		t.Fatalf("expected version in banner, got: %s", banner)
	}
}

func TestRenderer_RenderMarkdown(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.RenderMarkdown("# Hello\n\nWorld")
	if !strings.Contains(result, "Hello") {
		t.Fatalf("expected markdown content in result, got: %s", result)
	}
}

func TestRenderer_RenderText(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.RenderText("plain text")
	if result != "plain text" {
		t.Fatalf("expected 'plain text', got: %s", result)
	}
}

func TestRenderer_RenderToolCall(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.RenderToolCall("read_file", map[string]any{"path": "/test.txt"})
	// The result contains the human-readable action, not necessarily the tool name
	if result == "" {
		t.Fatalf("expected non-empty result, got empty string")
	}
	// Check that it contains some indication of the action
	if !strings.Contains(result, "Reading") && !strings.Contains(result, "read_file") {
		t.Fatalf("expected readable action in result, got: %s", result)
	}
}

func TestRenderer_RenderToolResult(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.RenderToolResult("read_file", map[string]any{"content": "test"})
	if result == "" {
		t.Fatal("expected non-empty tool result")
	}
}

func TestRenderer_RenderAgentThinking(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.RenderAgentThinking()
	if result == "" {
		t.Fatal("expected non-empty agent thinking message")
	}
}

func TestRenderer_RenderAgentWorking(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.RenderAgentWorking("analyzing code")
	if !strings.Contains(result, "analyzing code") {
		t.Fatalf("expected action in result, got: %s", result)
	}
}

func TestRenderer_RenderAgentResponse(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	result := r.RenderAgentResponse("Here's the solution...")
	if !strings.Contains(result, "solution") {
		t.Fatalf("expected response text in result, got: %s", result)
	}
}
